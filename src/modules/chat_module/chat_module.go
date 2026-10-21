package chat_module

import (
	"gorm.io/gorm"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
	"joodpakpaw/pkg/wshub"
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
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
		hub        *wshub.Hub
	}
	Repository struct {
		db     *gorm.DB
		tracer *tracing.MyTracer
		logger *logger.Logger
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
	return Controller{m: m}
}

func NewService() Service {
	return Service{
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
		hub:        wshub.CurrentHub(),
	}
}

func NewRepository() Repository {
	return Repository{
		db:     database.CurrentDatabase(),
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
	}
}

// Register controller services
func (c Controller) chatService() Service { return NewService() }

// Register service repositories
func (s Service) chatRepository() Repository { return NewRepository() }
