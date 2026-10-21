package config

import "os"

type appConfig struct {
	Env                     string
	HttpPort                string
	GrpcPort                string
	AppName                 string
	ServiceName             string
	AppFrontendUrl          string
	AppFrontendLocalhostUrl string
	MLServiceURL            string
}

func NewAppConfig() *appConfig {
	return &appConfig{
		Env:                     os.Getenv("ENV"),
		HttpPort:                os.Getenv("HTTP_PORT"),
		GrpcPort:                os.Getenv("GRPC_PORT"),
		AppName:                 os.Getenv("APP_NAME"),
		ServiceName:             os.Getenv("SERVICE_NAME"),
		AppFrontendUrl:          os.Getenv("APP_FRONTEND_URL"),
		AppFrontendLocalhostUrl: os.Getenv("APP_FRONTEND_LOCALHOST"),
		MLServiceURL:            os.Getenv("APP_ML_URL"),
	}
}
