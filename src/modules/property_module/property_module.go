package property_module

import (
	"gorm.io/gorm"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/storage"
	"joodpakpaw/pkg/tracing"
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
		tracer     *tracing.MyTracer
		logger     *logger.Logger
		fileSystem *storage.FileSystem
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
	return Service{
		tracer:     tracing.CurrentTracer(),
		logger:     logger.CurrentLogger(),
		fileSystem: storage.CurrentFileStorage(),
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
func (c Controller) propertyService() Service { return NewService() }

// Register service repositories
func (s Service) propertyRepository() Repository { return NewRepository() }
func (s Service) userRepository() user_module.Repository { return user_module.NewRepository() }
