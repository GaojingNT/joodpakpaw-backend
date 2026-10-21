package auth_module

import (
	"gorm.io/gorm"
	"joodpakpaw/config"
	"joodpakpaw/internal/helpers"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
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

func NewService(
	tracer *tracing.MyTracer,
	logger *logger.Logger,
	fileSystem *storage.FileSystem,
	mailService *helpers.MailService,
) Service {
	return Service{
		tracer:      tracer,
		logger:      logger,
		fileSystem:  fileSystem,
		mailService: mailService,
	}
}

func NewRepository(
	db *gorm.DB,
	tracer *tracing.MyTracer,
	logger *logger.Logger,
	fileSystem *storage.FileSystem,
) Repository {
	return Repository{
		db:         db,
		tracer:     tracer,
		logger:     logger,
		fileSystem: fileSystem,
	}
}

// Register controller services

func (c Controller) authService() Service {
	mailService := helpers.NewMailService(
		config.Global.Mail.SmtpHost,
		config.Global.Mail.SmtpPort,
		config.Global.Mail.SmtpUsername,
		config.Global.Mail.SmtpPassword,
	)
	return NewService(tracing.CurrentTracer(), logger.CurrentLogger(), storage.CurrentFileStorage(), mailService)
}

// Register controller repositories

func (s Service) authRepository() Repository {
	return NewRepository(database.CurrentDatabase(), tracing.CurrentTracer(), logger.CurrentLogger(), storage.CurrentFileStorage())
}
