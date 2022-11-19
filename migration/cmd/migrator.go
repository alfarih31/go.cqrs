package main

import (
	"fmt"
	_env "github.com/alfarih31/nb-go-env"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jetbasrawi/go.cqrs/migration/pkg"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	env, err := _env.LoadEnv(".env", true)
	if err != nil {
		log.Println(err)
	}

	dsn, err := env.GetString("DB_DSN")
	checkErr(err)

	driver, err := env.GetString("DB_DRIVER")
	checkErr(err)

	mg, err := pkg.Create(driver, dsn, env.MustGetString("MIGRATION_DIR", "migration/sql/pg"))
	checkErr(err)

	args := os.Args[1:]

	switch args[0] {
	case "up":
		err = mg.Up()
	case "down":
		err = mg.Down()
	case "reset":
		err = mg.Reset()
	case "help":
		fallthrough
	default:
		fmt.Print(usage)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

const usage = `
Migration Tool Help

migrate <command> [<args>]

Command:
  help                  show this help.
  up                    apply migration.
  down                  undo migration.
  reset                 reset migration history
`
