package apiserver

import (
	"github.com/efritz/nacelle"
	basegrpc "github.com/efritz/nacelle/base/grpc"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/internal/pkg/api"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type apiProcess struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`
}

func NewProcess() nacelle.Process {
	return basegrpc.NewServer(&apiProcess{})
}

func (ap *apiProcess) Init(config nacelle.Config, server *grpc.Server) error {
	scproto.RegisterSoftcopyServer(server, &apiServer{
		logger: ap.Logger,
		api:    ap.API,
	})
	scproto.RegisterSoftcopyAdminServer(server, &adminServer{
		logger: ap.Logger,
		api:    ap.API,
	})

	return nil
}
