package sftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"

	"github.com/efritz/nacelle"
	sftpclient "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/config"
	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
)

const (
	dupeFolderName = "duplicate"
)

type SFTPImporter struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`

	stopChan chan struct{}

	name string

	host     string
	port     int
	username string
	password string
	importPath     string
}

func NewSFTPImporter(name string, loader *config.OptionLoader) (*SFTPImporter, error) {
	host, err := loader.GetString("host")
	if err != nil {
		return nil, fmt.Errorf("could not get host: %s", err)
	}
	port, err := loader.GetIntOrDefault("port", 22)
	if err != nil {
		return nil, fmt.Errorf("could not get port: %s", err)
	}
	username, err := loader.GetStringOrDefault("username", "")
	if err != nil {
		return nil, fmt.Errorf("could not get username: %s", err)
	}
	password, err := loader.GetStringOrDefault("password", "")
	if err != nil {
		return nil, fmt.Errorf("could not get password: %s", err)
	}
	importPath, err := loader.GetStringOrDefault("import_path", "")
	if err != nil {
		return nil, fmt.Errorf("could not get import path: %s", err)
	}

	return &SFTPImporter{
		stopChan: make(chan struct{}),

		name: name,

		host:     host,
		port:     port,
		username: username,
		password: password,
		importPath:     importPath,
	}, nil
}

func (si *SFTPImporter) Name() string {
	return fmt.Sprintf("sftp:%s", si.name)
}

func (si *SFTPImporter) Start(ctx context.Context) error {
	si.Logger = si.Logger.WithFields(nacelle.LogFields{
		"importer": si.Name(),
	})

	si.Logger.Debug("starting sftp ssh connection to %s:%d", si.host, si.port)
	conn, err := ssh.Dial(
		"tcp", fmt.Sprintf("%s:%d", si.host, si.port),
		&ssh.ClientConfig{
			Timeout: 30 * time.Second,
			User:    si.username,
			Auth: []ssh.AuthMethod{
				ssh.Password(si.password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	)
	if err != nil {
		si.Logger.Error("error on dial: %s", err)
		return err
	}

	client, err := sftpclient.NewClient(conn)
	if err != nil {

		si.Logger.Error("error on sftp client: %s", err)
		return err
	}

	readBuf := make([]byte, 4096)
	for {
		select {
		case <-si.stopChan:
			return nil
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
			info, err := client.ReadDir(si.importPath)
			if err != nil {
				si.Logger.Error("error on readdir: %s", err)
				continue
			}

			for _, file := range info {
				if file.IsDir() {
					continue
				}

				si.Logger.Debug("found file %s", file.Name())
				filePath := path.Join(si.importPath, file.Name())
				docDate := time.Now().UTC()

				// Check if the file exists before we try downloading it
				_, err = si.API.GetFileWithDate(file.Name(), docDate)
				if err == nil {
					si.Logger.Warning(
						"file %s already exists for today, moving to duplicates",
						file.Name(),
					)
					dupePath := path.Join(si.importPath, dupeFolderName)
					_, err = client.Stat(dupePath)
					if os.IsNotExist(err) {
						// If the path doesn't exist, create it
						mkErr := client.Mkdir(dupePath)
						if mkErr != nil {
							si.Logger.Error(
								"could not create duplicates directory: %s",
								mkErr,
							)
							continue
						}
					} else if err != nil {
						si.Logger.Error(
							"could not check for duplicates directory: %s",
							err,
						)
						continue
					}

					dupeFilePath := path.Join(dupePath, file.Name())
					err = client.Rename(filePath, dupeFilePath)
					if err != nil {
						si.Logger.Error(
							"could not move file to duplicates directory: %s",
							err,
						)
						continue
					}

					continue
				} else if err != scerrors.ErrNotFound {
					si.Logger.Error("could not check if file exists: %s", err)
					continue
				}

				sftpFile, err := client.Open(filePath)
				if err != nil {
					si.Logger.Error("error opening file: %s", err)
					continue
				}

				// File doesn't exist yet, create it
				id, err := si.API.CreateFile(file.Name(), docDate)
				if err != nil {
					si.Logger.Error("could not create file: %s", err)
					sftpFile.Close()
					continue
				}

				of, err := si.API.OpenFile(id, records.FILE_MODE_WRITE)
				if err != nil {
					si.Logger.Error("could not open file for writing: %s", err)
					sftpFile.Close()
					continue
				}

				curRead := 0
				for {
					n, readErr := sftpFile.Read(readBuf)
					if readErr != nil && readErr != io.EOF {
						si.Logger.Error("could not read remote file: %s", readErr)
						sftpFile.Close()
						of.Close()
						continue
					}

					curWrite := 0
					for {
						writeN, err := of.Write(readBuf[curWrite : n-curWrite])
						if err != nil {
							si.Logger.Error("could not write file: %s", err)
							of.Close()
							sftpFile.Close()
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

				of.Close()
				sftpFile.Close()

				err = client.Remove(filePath)
				if err != nil {
					si.Logger.Error("could not remove added file: %s", err)
					continue
				}
			}
		}
	}
}

func (si *SFTPImporter) Stop() error {
	close(si.stopChan)
	return nil
}
