package middlewares

import (
	"gorm.io/gorm"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database"
	"joodpakpaw/pkg/logger"
	"joodpakpaw/pkg/tracing"
)

type Middleware struct {
	db     *gorm.DB
	cacher *cache.Cache
	tracer *tracing.MyTracer
	logger *logger.Logger
}

func NewMiddleware() *Middleware {
	return &Middleware{
		db:     database.CurrentDatabase(),
		cacher: cache.CurrentCacher(),
		tracer: tracing.CurrentTracer(),
		logger: logger.CurrentLogger(),
	}
}
