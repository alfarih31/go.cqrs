package internal

import (
	"database/sql"
	"github.com/pressly/goose"
	"io"
	"log"
	"sync"
)

// Migrator is a database migrator
type Migrator struct {
	driver *sql.DB
	dir    string
	sync.Mutex
}

// NewMigrator creates a database migrator
func NewMigrator(driver *sql.DB, dir string, out io.Writer) *Migrator {
	goose.SetLogger(log.New(out, "migrator ", log.LstdFlags))
	return &Migrator{
		driver: driver,
		dir:    dir,
	}
}

// Up runs migration up
func (m *Migrator) Up() error {
	defer m.driver.Close()

	return goose.Up(m.driver, m.dir)
}

// Down runs migration down
func (m *Migrator) Down() error {
	defer m.driver.Close()

	return goose.Down(m.driver, m.dir)
}

// Reset resets migration into initial state
func (m *Migrator) Reset() error {
	defer m.driver.Close()

	return goose.Reset(m.driver, m.dir)
}

// Status prints the migration status
func (m *Migrator) Status() error {
	defer m.driver.Close()

	return goose.Status(m.driver, m.dir)
}

// Create creates a new migration file.
func (m *Migrator) Create(name string) error {
	defer m.driver.Close()

	return goose.Create(m.driver, m.dir, name, "sql")
}
