package uiserver

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver/frontend"
	"github.com/aphistic/softcopy/internal/pkg/logging"
	"github.com/go-chi/chi"
)

type staticController struct {
	logger logging.Logger

	loader *frontend.Loader
}

func newStaticController(loader *frontend.Loader, logger logging.Logger) *staticController {
	return &staticController{
		logger: logger,
		loader: loader,
	}
}

func (sc *staticController) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/*", sc.GetAsset)
	return r
}

func (sc *staticController) GetAsset(w http.ResponseWriter, req *http.Request) {
	routeCtx := chi.RouteContext(req.Context())
	assetPath := strings.TrimPrefix(routeCtx.RoutePath, "/")
	data, err := sc.loader.ReadStaticFile(assetPath)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sc.logger.Error("could not load static asset %s: %s", assetPath, err)
		return
	}

	io.WriteString(w, string(data))
}
