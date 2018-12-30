package apiserver

import (
	"github.com/efritz/nacelle"
	"github.com/efritz/nacelle/process"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/api"
	scproto "github.com/aphistic/softcopy/proto"
)

type apiProcess struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`
}

func NewProcess() nacelle.Process {
	return process.NewGRPCServer(
		&apiProcess{},
	)
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
