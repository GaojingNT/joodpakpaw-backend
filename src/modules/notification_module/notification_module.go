package notification_module

import (
	"gorm.io/gorm"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/tracing"
	"joodpakpaw/pkg/wshub"
)

type (
	module struct {
		db     *gorm.DB
		tracer *tracing.MyTracer
		hub    *wshub.Hub
	}
	Controller struct {
		m *module
	}
	Service struct {
		tracer *tracing.MyTracer
		logger *logger.Logger
		hub    *wshub.Hub
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
		tracer: tracing.CurrentTracer(),
		hub:    wshub.CurrentHub(),
	}
}

func (m *module) Controller() Controller {
	return Controller{m: m}
}

func NewService() Service {
	return Service{
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
		hub:    wshub.CurrentHub(),
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
func (c Controller) notificationService() Service { return NewService() }

// Register service repositories
func (s Service) notificationRepository() Repository { return NewRepository() }
