package orm

import (
	"fmt"
	parser "github.com/alfarih31/nb-go-parser"
	_logger "github.com/jetbasrawi/go.cqrs/internal/orm/logger"
	"github.com/jetbasrawi/go.cqrs/internal/orm/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

	"gorm.io/gorm"
	_gormLogger "gorm.io/gorm/logger"
)

type OrmDriver uint

const (
	OrmDriverPostgres OrmDriver = iota + 1
	OrmDriverMysql
)

type DB interface {
	GetQuery() *models.Query
	GetDB() *gorm.DB
}

type conn struct {
	db *gorm.DB
	q  *models.Query
}

func (c *conn) GetDB() *gorm.DB {
	return c.db
}

func (c *conn) GetQuery() *models.Query {
	return c.q
}

func New(driver OrmDriver, dsn string, debug ...bool) (DB, error) {
	isDebug := parser.GetOptBoolArg(debug, false)

	// Go to warn if not debug
	logLevel := _gormLogger.Info
	if !isDebug {
		logLevel = _gormLogger.Warn
	}

	// Create logger
	logger := _logger.New()

	var d gorm.Dialector
	switch driver {
	case OrmDriverPostgres:
		d = postgres.New(postgres.Config{
			DSN: dsn,
		})
	case OrmDriverMysql:
		d = mysql.New(mysql.Config{
			DSN: dsn,
		})
	default:
		return nil, fmt.Errorf("unknown driver")
	}

	db, err := gorm.Open(d, &gorm.Config{
		Logger:      logger.LogMode(logLevel),
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	c := &conn{
		db: db,
		q:  models.Use(db),
	}
	models.SetDefault(db)

	return c, nil
}
