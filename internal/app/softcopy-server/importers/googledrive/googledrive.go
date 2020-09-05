// https://developers.google.com/drive/api/v3/about-auth
// https://github.com/gimite/google-drive-ruby/blob/master/doc/authorization.md
// https://www.syncwithtech.org/authorizing-google-apis/

package googledrive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/drive/v3"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/config"
	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

const (
	mimeFolder     = "application/vnd.google-apps.folder"
	dupeFolderName = "duplicate"
)

type GoogleDriveImporter struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`

	stopChan chan struct{}

	name string

	clientId     string
	clientSecret string
	refreshToken string
	importPath   string
}

func NewGoogleDriveImporter(name string, loader *config.OptionLoader) (*GoogleDriveImporter, error) {
	clientId, err := loader.GetString("client_id")
	if err != nil {
		return nil, err
	}
	clientSecret, err := loader.GetString("client_secret")
	if err != nil {
		return nil, err
	}
	refreshToken, err := loader.GetStringOrDefault("refresh_token", "")
	if err != nil {
		return nil, err
	}
	importPath, _ := loader.GetStringOrDefault("import_path", "")

	return &GoogleDriveImporter{
		name: name,

		clientId:     clientId,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		importPath:   importPath,
	}, nil
}

func (gdi *GoogleDriveImporter) Name() string {
	return fmt.Sprintf("google-drive:%s", gdi.name)
}

func (gdi *GoogleDriveImporter) Start(ctx context.Context) error {
	gdi.Logger = gdi.Logger.WithFields(nacelle.LogFields{
		"importer": gdi.Name(),
	})

	gdi.Logger.Debug("starting google drive importer")

	if gdi.refreshToken == "" {
		gdi.Logger.Warning(
			"A refresh token hasn't been configured. Please see the google drive " +
				"importer documentation for more information",
		)
		select {
		case <-gdi.stopChan:
		case <-ctx.Done():
		}

		return nil
	}

	authCtx := context.Background()
	authConfig := gdi.makeAuthConfig()
	authClient := authConfig.Client(authCtx, &oauth2.Token{
		RefreshToken: gdi.refreshToken,
	})
	client, err := drive.New(authClient)
	if err != nil {
		gdi.Logger.Error("could not create drive client: %s", err)
		return err
	}

	readBuf := make([]byte, 4096)
	for {
		select {
		case <-gdi.stopChan:
			return nil
		case <-ctx.Done():
			return nil
		case <-time.After(10 * time.Second):
			// TODO This should use a watch on the import path instead of polling
			importFolder, err := gdi.findFolder(gdi.importPath, client)
			if err != nil {
				gdi.Logger.Error("error finding folder: %s", err)
				continue
			}

			files, err := gdi.findFolderFiles(importFolder.Id, client)
			if err != nil {
				gdi.Logger.Error("could not find files in folder: %s", err)
			}

			docDate := time.Now().UTC()
			for _, file := range files {
				gdi.Logger.Debug("file: %s", file.Name)

				// Check if the file exists before we try downloading it
				_, err := gdi.API.GetFileWithDate(file.Name, docDate)
				if err == nil {
					gdi.Logger.Warning(
						"file %s already exists for today, moving to duplicates",
						file.Name,
					)

					dupePath := path.Join(gdi.importPath, dupeFolderName)
					dupeFolder, err := gdi.findFolder(dupePath, client)
					if err == scerrors.ErrNotFound {
						gdi.Logger.Debug("could not find duplicates folder, creating one")
						newDupeFolder, err := client.Files.Create(&drive.File{
							Name:     dupeFolderName,
							MimeType: mimeFolder,
							Parents:  []string{importFolder.Id},
						}).
							Fields("id, parents").
							Do()
						if err != nil {
							gdi.Logger.Error("could not create dupe folder: %s", err)
							continue
						}
						dupeFolder = newDupeFolder
					}

					_, err = client.Files.Update(file.Id, nil).
						AddParents(dupeFolder.Id).
						RemoveParents(importFolder.Id).
						Fields("id", "parents").
						Do()
					if err != nil {
						gdi.Logger.Error("could not move file to dupe folder: %s", err)
					}

					continue
				} else if err != nil {
					gdi.Logger.Error("could not check for dupe path: %s", err)
					continue
				}

				fileRes, err := client.Files.Get(file.Id).Download()
				if err != nil {
					gdi.Logger.Error("could not download file %s: %s", file.Name, err)
					continue
				}

				id, err := gdi.API.CreateFile(file.Name, docDate)
				if err != nil {
					gdi.Logger.Error("could not create file: %s", err)
					fileRes.Body.Close()
					continue
				}

				of, err := gdi.API.OpenFile(id, records.FILE_MODE_WRITE)
				if err != nil {
					gdi.Logger.Error("could not open file for writing: %s", err)
					fileRes.Body.Close()
					continue
				}

				curRead := 0
				for {
					n, readErr := fileRes.Body.Read(readBuf)
					if readErr != nil && readErr != io.EOF {
						gdi.Logger.Error("could not read remote file: %s", readErr)
						fileRes.Body.Close()
						of.Close()
						continue
					}

					curWrite := 0
					for {
						writeN, err := of.Write(readBuf[curWrite : n-curWrite])
						if err != nil {
							gdi.Logger.Error("could not write file: %s", err)
							fileRes.Body.Close()
							of.Close()
							continue
						}

						curWrite += writeN

						if curWrite >= n {
							break
						}
					}

					curRead += n

					if readErr == io.EOF {
						break
					}
				}

				fileRes.Body.Close()
				of.Close()

				err = client.Files.Delete(file.Id).Do()
				if err != nil {
					gdi.Logger.Error("could not remove added file: %s", err)
					continue
				}
			}
		}
	}
}

func (gdi *GoogleDriveImporter) findFolder(folderPath string, client *drive.Service) (*drive.File, error) {
	pathParts := strings.Split(folderPath, "/")

	prevParent := "root"
	var prevFile *drive.File
	for _, part := range pathParts {
		list, err := client.Files.List().
			Q(fmt.Sprintf(`parents in "%s"`, prevParent)).
			Fields("files(id, name, mimeType, trashed)").
			Do()
		if err != nil {
			return nil, err
		}

		foundFolder := false
		for _, file := range list.Files {
			if file.MimeType == mimeFolder && file.Name == part && !file.Trashed {
				prevFile = file
				prevParent = file.Id
				foundFolder = true
				break
			}
		}
		if !foundFolder {
			return nil, scerrors.ErrNotFound
		}
	}

	return prevFile, nil
}

func (gdi *GoogleDriveImporter) findFolderFiles(folderID string, client *drive.Service) ([]*drive.File, error) {
	list, err := client.Files.List().Q(fmt.Sprintf(`parents in "%s"`, folderID)).Do()
	if err != nil {
		return nil, err
	}

	var results []*drive.File
	for _, file := range list.Files {
		if file.MimeType != mimeFolder {
			results = append(results, file)
		}
	}

	return results, nil
}

func (gdi *GoogleDriveImporter) Stop() error {
	close(gdi.stopChan)
	return nil
}

func (gdi *GoogleDriveImporter) SetupWebHandlers(router chi.Router) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authConfig := gdi.makeAuthConfig()
		authURL := authConfig.AuthCodeURL(
			"todomakethisrandom",
			oauth2.ApprovalForce,
			oauth2.AccessTypeOffline,
		)

		fmt.Fprintf(w, "<a href=\"%s\" target=\"_blank\">%s</a>", authURL, authURL)

		postURL := *r.URL
		postURL.Path = path.Join(postURL.Path, "/token")

		fmt.Fprintf(w, "<form action=\"%s\" method=\"POST\">", postURL.String())
		fmt.Fprintf(w, "Token: <input type=\"text\" name=\"token\" value=\"\" />")
		fmt.Fprintf(w, "<input type=\"submit\" value=\"Auth\">")
		fmt.Fprintf(w, "</form>")
	})
	router.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			gdi.Logger.Error("could not parse form data: %s", err)
			return
		}

		tokenData := r.Form.Get("token")
		gdi.Logger.Debug("token: %s", tokenData)

		authConfig := gdi.makeAuthConfig()

		ctx := context.Background()
		token, err := authConfig.Exchange(ctx, tokenData)
		if err != nil {
			gdi.Logger.Error("could not exchange token data: %s", err)
			return
		}

		fmt.Fprintf(w, "Refresh Token: %s", token.RefreshToken)
	})
}

func (gdi *GoogleDriveImporter) makeAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     gdi.clientId,
		ClientSecret: gdi.clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			drive.DriveScope,
		},
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
	}
}
