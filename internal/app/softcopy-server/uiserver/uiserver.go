package uiserver

import (
	"net/http"

	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"
	"github.com/efritz/nacelle/logging"
	"github.com/go-chi/chi"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver/backend"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver/frontend"
	"github.com/aphistic/softcopy/internal/pkg/api"
)

func NewProcess() nacelle.Process {
	return basehttp.NewServer(&uiServer{})
}

type uiServer struct {
	Logger    nacelle.Logger           `service:"logger"`
	API       *api.Client              `service:"api"`
	Importers *importers.ImportRunners `service:"importers"`

	tpl *backend.Template
}

func (us *uiServer) Init(config nacelle.Config, server *http.Server) error {
	us.Logger = us.Logger.WithFields(logging.Fields{
		"service": "uiserver",
	})

	backendLoader, err := backend.NewLoader(
		backend.LoaderLogger(us.Logger),
	)
	if err != nil {
		return err
	}

	frontendLoader, err := frontend.NewLoader(
		frontend.LoaderLogger(us.Logger),
	)
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	// Home page controller
	hc := newHomeController(backendLoader, us.Logger)
	r.Mount("/", hc.Router())

	// Static assets controller
	sc := newStaticController(frontendLoader, us.Logger)
	r.Mount("/static", sc.Router())

	r.HandleFunc("/importers", us.GetImporters)
	for _, importer := range us.Importers.Runners() {
		webImporter, ok := importer.(importers.ImporterWebHandler)
		if !ok {
			continue
		}

		us.Logger.Info("Setting up web interface for %s", importer.Name())
		r.Route("/importers/"+importer.Name(), webImporter.SetupWebHandlers)
	}

	server.Handler = r

	return nil
}

func (us *uiServer) GetImporters(w http.ResponseWriter, r *http.Request) {
	t, err := us.tpl.Template("templates/importers.html.tpl")
	if err != nil {
		us.Logger.Error("could not find template: %s", err)
		return
	}

	err = t.Execute(w, us.Importers.Runners())
	if err != nil {
		us.Logger.Error("could not execute template: %s", err)
		return
	}
}

// func (us *uiServer) GetImporter(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)

// 	name, ok := vars["name"]
// 	if !ok {
// 		us.Logger.Error("could not get importer name")
// 		return
// 	}
// 	path := vars["path"]

// 	us.Logger.Debug("got name: %s, path: %s", name, path)
// }
