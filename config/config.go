package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Global       *Config
	IsProduction bool
)

type Config struct {
	App           *appConfig
	Auth          *authConfig
	Fiber         *fiberConfig
	Database      *databaseConfig
	Redis         *redisConfig
	Sentry        *sentryConfig
	OpenTelemetry *openTelemetryConfig
	Logger        *loggerConfig
	FileSystem    *fileSystemConfig
	Mail          *MailConfig
}

func NewConfig() *Config {
	// Load .env file (always use .env file)
	// IsProduction flag is controlled by ENV variable
	// - ENV=production → IsProduction = true (hide error messages)
	// - ENV=development → IsProduction = false (show error messages)
	godotenv.Load(".env")

	// Set IsProduction flag based on ENV variable
	if os.Getenv("ENV") == "production" {
		IsProduction = true
	}

	return &Config{
		App:           NewAppConfig(),
		Auth:          NewAuthConfig(),
		Fiber:         NewFiberConfig(),
		Database:      NewDatabaseConfig(),
		Redis:         NewRedisConfig(),
		Sentry:        NewSentryConfig(),
		OpenTelemetry: NewOpenTelemetryConfig(),
		Logger:        NewLoggerConfig(),
		FileSystem:    NewFileSystemConfig(),
		Mail:          NewMailConfig(),
	}
}
