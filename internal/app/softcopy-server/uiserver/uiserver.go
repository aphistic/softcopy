package uiserver

import (
	"net/http"

	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"
	"github.com/gorilla/mux"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver/template"
	"github.com/aphistic/softcopy/internal/pkg/api"
)

func NewProcess() nacelle.Process {
	return basehttp.NewServer(&uiServer{})
}

type uiServer struct {
	Logger    nacelle.Logger           `service:"logger"`
	API       *api.Client              `service:"api"`
	Importers *importers.ImportRunners `service:"importers"`

	tpl *template.Template
}

func (us *uiServer) Init(config nacelle.Config, server *http.Server) error {
	tpl, err := template.LoadTemplates()
	if err != nil {
		return err
	}
	us.tpl = tpl

	r := mux.NewRouter()
	r.HandleFunc("/importers", us.GetImporters)

	for _, importer := range us.Importers.Runners() {
		webImporter, ok := importer.(importers.ImporterWebHandler)
		if !ok {
			continue
		}

		us.Logger.Info("Setting up web interface for %s", importer.Name())
		subRouter := r.PathPrefix("/importers/" + importer.Name()).Subrouter().StrictSlash(true)
		err = webImporter.SetupWebHandlers(subRouter)
		if err != nil {
			return err
		}
	}

	server.Handler = r

	return nil
}

func (us *uiServer) GetImporters(w http.ResponseWriter, r *http.Request) {
	t, err := us.tpl.Template("importers.html.tpl")
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

func (us *uiServer) GetImporter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name, ok := vars["name"]
	if !ok {
		us.Logger.Error("could not get importer name")
		return
	}
	path := vars["path"]

	us.Logger.Debug("got name: %s, path: %s", name, path)
}
