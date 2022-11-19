package main

import (
	_env "github.com/alfarih31/nb-go-env"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
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

	var dialector gorm.Dialector
	switch driver {
	case "postgres":
		dialector = postgres.New(postgres.Config{
			DSN: dsn,
		})
	case "mysql":
		dialector = mysql.New(mysql.Config{
			DSN: dsn,
		})
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:      logger.Default,
		PrepareStmt: true,
	})
	checkErr(err)

	g := gen.NewGenerator(gen.Config{
		OutPath:           "./internal/orm/models",
		OutFile:           "./internal/orm/models/query.go",
		FieldNullable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithDefaultQuery + gen.WithoutContext + gen.WithQueryInterface,
	})

	g.UseDB(db)

	EventStoreModel := g.GenerateModel("event_store")

	EventStreamModel := g.GenerateModel("event_stream")

	g.ApplyBasic(EventStoreModel, EventStreamModel)

	g.Execute()
}
