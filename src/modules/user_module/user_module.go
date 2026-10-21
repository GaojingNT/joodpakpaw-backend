package user_module

import (
	"gorm.io/gorm"
	"joodpakpaw/config"
	"joodpakpaw/internal/helpers"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
	pbUser "joodpakpaw/src/modules/user_module/proto/user"
)

type (
	module struct {
		db         *gorm.DB
		cacher     *cache.Cache
		tracer     *tracing.MyTracer
		fileSystem *storage.FileSystem
	}
	Controller struct {
		m *module
	}
	GrpcController struct {
		m *module
		pbUser.UnimplementedUserServiceServer
	}
	Service struct {
		tracer      *tracing.MyTracer
		logger      *logger.Logger
		fileSystem  *storage.FileSystem
		mailService *helpers.MailService
	}
	Repository struct {
		db         *gorm.DB
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
	}
)

func NewModule() module {
	return module{
		db:         database.CurrentDatabase(),
		cacher:     cache.CurrentCacher(),
		tracer:     tracing.CurrentTracer(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

func (m *module) Controller() Controller {
	return Controller{
		m: m,
	}
}

func (m *module) GrpcController() GrpcController {
	return GrpcController{
		m: m,
	}
}

func NewService() Service {
	mailConfig := config.NewMailConfig()
	return Service{
		tracer:      tracing.CurrentTracer(),
		logger:      logger.CurrentLogger(),
		fileSystem:  storage.CurrentFileStorage(),
		mailService: helpers.NewMailService(mailConfig.SmtpHost, mailConfig.SmtpPort, mailConfig.SmtpUsername, mailConfig.SmtpPassword),
	}
}

func NewRepository() Repository {
	return Repository{
		db:         database.CurrentDatabase(),
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
	}
}

// Register controller services
func (c Controller) userService() Service     { return NewService() }
func (c GrpcController) userService() Service { return NewService() }

// Register service repositories
func (s Service) userRepository() Repository { return NewRepository() }
