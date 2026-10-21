package scheduler

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Log will log a message
func (ctx *SchedulerContext) Log(tag string, message string) {
	ctx.sc.Logger.Info(message, zap.String("tag", tag))
}

// DB returns the database client
func (ctx *SchedulerContext) DB() *gorm.DB {
	return ctx.sc.DbClient
}
