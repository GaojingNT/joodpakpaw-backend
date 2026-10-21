package booking_module

import (
	"gorm.io/gorm"
	"joodpakpaw/config"
	"joodpakpaw/internal/helpers"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
	"joodpakpaw/src/modules/property_module"
	"joodpakpaw/src/modules/user_module"
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
func (c Controller) bookingService() Service { return NewService() }

// Register service repositories
func (s Service) bookingRepository() Repository               { return NewRepository() }
func (s Service) propertyRepository() property_module.Repository { return property_module.NewRepository() }
func (s Service) userRepository() user_module.Repository       { return user_module.NewRepository() }
