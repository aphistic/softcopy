package apiserver

import (
	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/pkg/api"
)

type apiServer struct {
	logger nacelle.Logger
	api    *api.Client
}
