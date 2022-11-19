package pkg

import (
	"database/sql"
	parser "github.com/alfarih31/nb-go-parser"
	"github.com/jetbasrawi/go.cqrs/migration/internal"
	"os"
)

type Migrator interface {
	Up() error
	Down() error
	Reset() error
}

type migrator struct {
	*internal.Migrator
}

func Create(driver, dsn string, migrationDir ...string) (Migrator, error) {
	mDir := parser.GetOptStringArg(migrationDir, "migration/sql/pg")

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	mg := internal.NewMigrator(db, mDir, os.Stdout)

	return &migrator{
		Migrator: mg,
	}, nil
}
