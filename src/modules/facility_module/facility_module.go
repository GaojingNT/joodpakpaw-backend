package facility_module

import (
	"gorm.io/gorm"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/tracing"
)

type (
	module struct {
		db     *gorm.DB
		cacher *cache.Cache
		tracer *tracing.MyTracer
	}
	Controller struct {
		m *module
	}
	Service struct {
		tracer *tracing.MyTracer
		logger *logger.Logger
	}
	Repository struct {
		db     *gorm.DB
		tracer *tracing.MyTracer
		logger *logger.Logger
	}
)

func NewModule() module {
	return module{
		db:     database.CurrentDatabase(),
		cacher: cache.CurrentCacher(),
		tracer: tracing.CurrentTracer(),
	}
}

func (m *module) Controller() Controller {
	return Controller{
		m: m,
	}
}

func NewService() Service {
	return Service{
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
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
func (c Controller) facilityService() Service { return NewService() }

// Register service repositories
func (s Service) facilityRepository() Repository { return NewRepository() }
