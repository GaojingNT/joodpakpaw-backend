package routes

import (
	"joodpakpaw/app/grpc_server"
	"joodpakpaw/src/modules/user_module"
	pbUser "joodpakpaw/src/modules/user_module/proto/user"
)

func GRPCRegisterServices(s *grpc_server.GrpcServer) {
	// Initialize module instances
	userModule := user_module.NewModule()

	// GRPC services ------------------------------------------------------------------
	pbUser.RegisterUserServiceServer(s.Grpc, userModule.GrpcController())
}
