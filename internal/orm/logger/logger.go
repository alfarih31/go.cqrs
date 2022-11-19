package logger

import (
	"context"
	"fmt"
	_logger "github.com/alfarih31/nb-go-logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	_gormLogger "gorm.io/gorm/logger"
	"time"
)

const SlowThreshold = time.Duration(200) * time.Millisecond

type LogLevel = _gormLogger.LogLevel

const (
	Debug LogLevel = iota + 5
)

type logger struct {
	_logger.Logger
	level        LogLevel
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

type DBLogger interface {
	LogMode(LogLevel) _gormLogger.Interface
	Info(context.Context, string, ...interface{})
	Warn(context.Context, string, ...interface{})
	Error(context.Context, string, ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
	Hook(hook logrus.Hook) DBLogger
	Level() _logger.LogLevel
}

func New() DBLogger {
	return &logger{
		Logger:  _logger.New("DB"),
		BeginAt: time.Now(),
	}
}

func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info(msg, data...)
}

func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warn(msg, data...)
}

func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Error(msg, data...)
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= _gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.level >= _gormLogger.Error && err != gorm.ErrRecordNotFound:
		sql, rows := fc()
		if rows == -1 {
			l.Error(ctx, err.Error(), map[string]interface{}{
				"elapsed": elapsed.String(),
				"sql":     sql,
			})
		} else {
			l.Error(ctx, err.Error(), map[string]interface{}{
				"elapsed": elapsed.String(),
				"rows":    rows,
				"sql":     sql,
			})
		}
	case elapsed > SlowThreshold && SlowThreshold != 0 && l.level >= _gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", SlowThreshold)
		if rows == -1 {
			l.Warn(ctx, slowLog, map[string]interface{}{
				"elapsed": elapsed.String(),
				"sql":     sql,
			})
		} else {
			l.Warn(ctx, slowLog, map[string]interface{}{
				"elapsed": elapsed.String(),
				"rows":    rows,
				"sql":     sql,
			})
		}
	case l.level == _gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Info(ctx, "", map[string]interface{}{
				"elapsed": elapsed.String(),
				"sql":     sql,
			})
		} else {
			l.Info(ctx, "", map[string]interface{}{
				"elapsed": elapsed.String(),
				"rows":    rows,
				"sql":     sql,
			})
		}
	}
}

func (l *logger) Hook(hook logrus.Hook) DBLogger {
	l.AddHook(hook)
	return l
}

func (l *logger) Level() _logger.LogLevel {
	return l.GetLevel()
}

func (l *logger) LogMode(level LogLevel) _gormLogger.Interface {
	newLogger := *l
	switch level {
	case _gormLogger.Silent:
		newLogger.level = _gormLogger.Silent
	case _gormLogger.Error:
		newLogger.level = _gormLogger.Error
		newLogger.SetLevel("error")
	case _gormLogger.Warn:
		newLogger.level = _gormLogger.Warn
		newLogger.SetLevel("warn")
	case _gormLogger.Info:
		newLogger.level = _gormLogger.Info
		newLogger.SetLevel("info")
	}

	return &newLogger
}
