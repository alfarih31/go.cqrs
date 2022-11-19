package main

import (
	_env "github.com/alfarih31/nb-go-env"
	"github.com/jetbasrawi/go.cqrs/internal/orm/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func main() {
	checkErr := func(err error) {
		if err != nil {
			log.Fatalln(err)
		}
	}

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

	EventStoreModel := g.GenerateModel("event_store", gen.FieldRelateModel(field.HasMany, "EventStreams", &model.EventStream{}, &field.RelateConfig{
		RelateSlicePointer: true,
		GORMTag:            "foreignKey:event_id",
	}))

	EventStreamModel := g.GenerateModel("event_stream", gen.FieldRelateModel(field.BelongsTo, "Event", &model.EventStore{}, &field.RelateConfig{
		RelatePointer: true,
		GORMTag:       "foreignKey:event_id;references:event_id",
	}))

	g.ApplyBasic(EventStoreModel, EventStreamModel)

	g.Execute()
}
