package app

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	socketio "github.com/zishang520/socket.io/v2/socket"
	"joodpakpaw/app/grpc_server"
	"joodpakpaw/app/http_server"
	"joodpakpaw/app/scheduler"
	"joodpakpaw/config"
	"joodpakpaw/database/seeders"
	"joodpakpaw/internal/jwt"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	sentryException "joodpakpaw/pkg/sentry"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
	"joodpakpaw/pkg/wshub"
	"joodpakpaw/src/routes"
)

type ApplicationServer struct {
	HttpServer *http_server.HttpServer
	GrpcServer *grpc_server.GrpcServer
}

type IApplicationServer interface {
	StartServer() error
	StopServer() error

	initialize()
	registerServerRouter()
	handle() error
	cleanup()
}

func NewApplicationServer() *ApplicationServer {
	return &ApplicationServer{}
}

func (app *ApplicationServer) StartServer() error {
	// Initialize the application components
	app.initialize()

	// Create microservice instances
	app.HttpServer = http_server.NewHttpServer(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)
	// app.GrpcServer = grpc_server.NewGrpcServer(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)

	// Initialize global scheduler
	scheduler.GlobalScheduler = scheduler.NewScheduler(config.Global, database.DbClient, cache.AppCacher, tracing.AppTracer, storage.FileStorage)

	// Register routes
	app.registerServerRouter()

	return app.handle()
}

func (app *ApplicationServer) StopServer() error {
	// Ignore fiber child process
	if fiber.IsChild() {
		return nil
	}

	var timeMs = time.Now().UnixMilli()

	// Stop servers
	app.HttpServer.StopServer()
	// app.GrpcServer.StopServer()

	// Shutdown global scheduler
	// scheduler.GlobalScheduler.Shutdown()

	// Cleanup
	app.cleanup()

	// Print time taken to shutdown
	lastText := fmt.Sprintf("Time taken to shutdown successfully: %s ms", strconv.FormatInt(time.Now().UnixMilli()-timeMs, 10))
	log.Println(lastText)
	logger.CurrentLogger().Info(lastText)

	return nil
}

func (app *ApplicationServer) initialize() {
	// Load environment variables
	config.Global = config.NewConfig()

	// Initialize JWT
	jwt.JwtToken = jwt.NewJWT()

	// Initialize Postgresql connection
	database.DbClient = database.Initialize()

	// Initialize storage file system instance (needed by seeder too)
	storage.FileStorage = storage.NewFileSystem(
		config.Global.FileSystem.Disk,
		storage.WithS3Endpoint(config.Global.FileSystem.S3Config.Endpoint),
		storage.WithS3Bucket(config.Global.FileSystem.S3Config.Bucket),
		storage.WithS3Region(config.Global.FileSystem.S3Config.Region),
		storage.WithS3AccessKeyID(config.Global.FileSystem.S3Config.AccessKeyID),
		storage.WithS3SecretKey(config.Global.FileSystem.S3Config.SecretKey),
		storage.WithS3UseSSL(config.Global.FileSystem.S3Config.UseSSL),
		storage.WithS3Prefix(config.Global.FileSystem.S3Config.Prefix),
		storage.WithS3PublicURL(config.Global.FileSystem.S3Config.PublicURL),
	)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--db-seed":
			fmt.Println("Running database seeder...")
			seeders.SeedDatabase(database.DbClient)
			fmt.Println("Seeding completed.")
			os.Exit(0)

		case "--db-rollback":
			fmt.Println("Rolling back last migration...")
			database.Rollback()
			fmt.Println("Rollback completed.")
			os.Exit(0)
		}
	}

	// Initialize logger
	logger.AppLogger = logger.Initialize()

	// Initialize connection to cache
	cache.AppCacher = cache.NewCacher(
		cache.Initialize(),
		cache.WithPrefix(config.Global.Redis.RedisCachePrefix),
		cache.WithExpired(time.Minute*time.Duration(config.Global.Redis.RedisCacheDuration)),
	)

	// Initialize OpenTelemetry tracing
	tracing.AppTracer = tracing.InitTracer()

	// Initialize Sentry client for error logging and tracing
	sentryException.Initialize()

	// Initialize WebSocket Hub
	wshub.AppHub = wshub.NewHub()

	// Initialize Socket.IO Server
	sioServer := socketio.NewServer(nil, nil)
	sioServer.On("connection", func(clients ...any) {
		socket := clients[0].(*socketio.Socket)

		tokenSlice := socket.Handshake().Query["token"]
		if len(tokenSlice) == 0 {
			socket.Disconnect(true)
			return
		}
		token := tokenSlice[0]

		payload, err := jwt.JwtToken.Payload(token)
		if err != nil {
			socket.Disconnect(true)
			return
		}

		userMap, ok := payload["user"].(map[string]interface{})
		if !ok {
			socket.Disconnect(true)
			return
		}

		userID := int(userMap["id"].(float64))
		room := socketio.Room(fmt.Sprintf("user_%d", userID))
		socket.Join(room)
		socket.Emit("connected", map[string]interface{}{"userId": userID})
	})
	wshub.AppHub.SetSocketIOServer(sioServer)
}

func (app *ApplicationServer) registerServerRouter() {
	// Setup routes
	routes.HTTPRoutes(app.HttpServer)
	// routes.GRPCRegisterServices(app.GrpcServer)
}

func (app *ApplicationServer) cleanup() {
	// Close all database connection
	if database.CurrentDatabase() != nil {
		sqlDB, _ := database.CurrentDatabase().DB()
		sqlDB.Close()
	}

	// Close all redis connection
	if cache.CurrentCacher() != nil {
		cache.CurrentCacher().Close()
	}

	// Sentry: Flush buffered events before the program terminates.
	if config.Global.Sentry.SentryDSN != "" {
		sentry.Flush(2 * time.Second)
		sentry.Recover()
	}

	// Close all tracer connection
	if tracing.CurrentTracer().TraceProvider != nil {
		tracing.CurrentTracer().Cleanup()
	}

	// Close logger
	if logger.CurrentLogger() != nil {
		logger.CurrentLogger().Sync()

		// Check if enable write log file, closing file I/O
		if config.Global.Logger.LogMode == "s3" || config.Global.Logger.LogMode == "file" {
			logger.Close()
		}
	}

	// Stop all schedule jobs
	if scheduler.CurrentScheduler() != nil {
		scheduler.CurrentScheduler().Shutdown()
	}
}
