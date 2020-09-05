package uiserver

import (
	"fmt"
	"net/http"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver/backend"
	"github.com/aphistic/softcopy/internal/pkg/logging"
	"github.com/go-chi/chi"
)

type homeController struct {
	logger logging.Logger
	loader *backend.Loader
}

func newHomeController(loader *backend.Loader, logger logging.Logger) *homeController {
	return &homeController{
		logger: logger,
		loader: loader,
	}
}

func (h *homeController) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetHome)
	return r
}

func (h *homeController) GetHome(w http.ResponseWriter, r *http.Request) {
	data, err := h.loader.ReadFile("home.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("Could not read home.html: %s", err)
		return
	}

	fmt.Fprintf(w, string(data))
}
