package ycq

import (
	"context"
	"encoding/json"
	"fmt"
	parser "github.com/alfarih31/nb-go-parser"
	"github.com/jetbasrawi/go.cqrs/internal/orm"
	"github.com/jetbasrawi/go.cqrs/internal/orm/model"
	"github.com/jetbasrawi/go.cqrs/internal/orm/models"
	"github.com/jetbasrawi/go.cqrs/internal/transformer"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"time"
)

type readDirection uint

const (
	readDirectionForward readDirection = iota + 1
	readDirectionBackward
)

type sqlEventRepository struct {
	db orm.DB
}

type sqlEventRepositoryReaderSpec struct {
	fromTime  *time.Time
	toTime    *time.Time
	fromId    *int
	toId      *int
	direction readDirection
	limit     *int
}

type sqlEventRepositoryReader struct {
	streamQuery models.IEventStreamDo
	spec        *sqlEventRepositoryReaderSpec
}

func (s *sqlEventRepositoryReader) buildEvent(m *model.EventStream) (EventMessage, error) {
	return NewEventMessage(&m.Event.EventID, &RawEvent{
		name: m.Event.EventName,
		data: m.Event.EventData,
	}, parser.Int(m.StreamVersion).ToIntPtr()), nil
}

func (s *sqlEventRepositoryReaderSpec) BuildQuery(query models.IEventStreamDo) (models.IEventStreamDo, error) {
	conds := []gen.Condition{}
	orders := []field.Expr{}
	if s.toTime != nil {
		// Can't do query if toTime exist but from fromTime not exist
		if s.fromTime == nil {
			return nil, &ErrRepositoryExecution{
				Err: fmt.Errorf("fromTime must exist if toTime exist"),
			}
		}

		conds = append(conds, models.EventStream.CreatedAt.Between(*s.fromTime, *s.toTime))
	} else if s.fromTime != nil {
		if s.toTime != nil {
			conds = append(conds, models.EventStream.CreatedAt.Between(*s.fromTime, *s.toTime))
		} else {
			conds = append(conds, models.EventStream.CreatedAt.Gte(*s.fromTime))
		}
	}

	if s.toId != nil {
		// Can't do query if toId exist but from fromId not exist
		if s.fromId == nil {
			return nil, &ErrRepositoryExecution{
				Err: fmt.Errorf("fromId must exist if toId exist"),
			}
		}

		conds = append(conds, models.EventStream.ID.Between(int64(*s.fromId), int64(*s.toId)))
	} else if s.fromTime != nil {
		conds = append(conds, models.EventStream.ID.Gte(int64(*s.fromId)))
	}

	switch s.direction {
	case readDirectionForward:
		orders = append(orders, models.EventStream.ID)
	case readDirectionBackward:
		orders = append(orders, models.EventStream.ID.Desc())
	}

	query = query.Where(conds...).Order(orders...)
	if s.limit != nil {
		query = query.Limit(*s.limit)
	}

	return query, nil
}

func (s *sqlEventRepositoryReader) Stream(streamId string) EventRepositoryReader {
	s.streamQuery = s.streamQuery.Where(models.EventStream.StreamID.Eq(streamId))
	return s
}

func (s *sqlEventRepositoryReader) FromTime(date time.Time) EventRepositoryReader {
	s.spec.fromTime = &date
	return s
}

func (s *sqlEventRepositoryReader) FromId(id int) EventRepositoryReader {
	s.spec.fromId = &id
	return s
}

func (s *sqlEventRepositoryReader) ToTime(date time.Time) EventRepositoryReader {
	s.spec.toTime = &date
	return s
}

func (s *sqlEventRepositoryReader) ToId(id int) EventRepositoryReader {
	s.spec.toId = &id

	return s
}

func (s *sqlEventRepositoryReader) Forward() EventRepositoryReader {
	s.spec.direction = readDirectionForward

	return s
}

func (s *sqlEventRepositoryReader) Backward() EventRepositoryReader {
	s.spec.direction = readDirectionBackward

	return s
}

func (s *sqlEventRepositoryReader) Limit(count int) EventRepositoryReader {
	s.spec.limit = &count

	return s
}

func (s *sqlEventRepositoryReader) Event(id string) (EventMessage, error) {
	q, err := s.spec.BuildQuery(s.streamQuery)
	if err != nil {
		return nil, err
	}

	evs, err := q.Where(models.EventStream.EventID.Eq(id)).Joins(models.EventStream.Event).Limit(1).First()

	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return s.buildEvent(evs)
}

func (s *sqlEventRepositoryReader) Events(ids []string) ([]EventMessage, error) {
	q, err := s.spec.BuildQuery(s.streamQuery)
	if err != nil {
		return nil, err
	}

	evs, err := q.Where(models.EventStream.EventID.In(ids...)).Joins(models.EventStream.Event).Find()

	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	ems, err := transformer.ArrayTransformer[*model.EventStream, EventMessage](evs, func(m *model.EventStream) (EventMessage, error) {
		return s.buildEvent(m)
	})

	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return ems, nil
}

func (s *sqlEventRepositoryReader) Count() (int, error) {
	q, err := s.spec.BuildQuery(s.streamQuery)
	if err != nil {
		return 0, err
	}

	c, err := q.Select(models.EventStream.ID).Count()
	if err != nil {
		return 0, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return int(c), nil
}

func (s *sqlEventRepositoryReader) ToList() ([]EventMessage, error) {
	q, err := s.spec.BuildQuery(s.streamQuery)
	if err != nil {
		return nil, err
	}

	evs, err := q.Joins(models.EventStream.Event).Find()

	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	ems, err := transformer.ArrayTransformer[*model.EventStream, EventMessage](evs, func(m *model.EventStream) (EventMessage, error) {
		return s.buildEvent(m)
	})

	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return ems, nil
}

func (s *sqlEventRepository) HasEvent(ctx context.Context, id string) (bool, error) {
	i, err := s.db.GetQuery().EventStore.Select(models.EventStore.ID).Where(models.EventStore.EventID.Eq(id)).Count()
	if err != nil {
		return false, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return i > 0, nil
}

func (s *sqlEventRepositoryReader) Last(streamId string) (EventMessage, error) {
	evs, err := s.streamQuery.Where(models.EventStream.StreamID.Eq(streamId)).Joins(models.EventStream.Event).Order(models.EventStream.StreamVersion.Desc()).Limit(1).First()
	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return s.buildEvent(evs)
}

func (s *sqlEventRepository) GetStreamIdOf(ctx context.Context, eventId string) (string, error) {
	// Get last stream
	evs, err := s.db.GetQuery().EventStream.WithContext(ctx).Select(models.EventStream.StreamID).Where(models.EventStream.EventID.Eq(eventId)).Limit(1).First()
	if err != nil {
		return "", &ErrRepositoryExecution{
			Err: err,
		}
	}

	return evs.EventID, nil
}

func (s *sqlEventRepository) GetVersionInStream(ctx context.Context, streamId, eventId string) (*int, error) {
	evs, err := s.db.GetQuery().EventStream.WithContext(ctx).Select(models.EventStream.StreamVersion).Where(models.EventStream.StreamID.Eq(streamId), models.EventStream.EventID.Eq(eventId)).Order(models.EventStream.StreamVersion.Desc()).Limit(1).First()
	if err != nil {
		return nil, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return parser.Int(evs.StreamVersion).ToIntPtr(), nil
}

func (s *sqlEventRepository) IsEventInStream(ctx context.Context, streamId, eventId string) (bool, error) {
	i, err := s.db.GetQuery().EventStream.Select(models.EventStream.ID).Where(models.EventStream.EventID.Eq(eventId), models.EventStream.StreamID.Eq(streamId)).Count()
	if err != nil {
		return false, &ErrRepositoryExecution{
			Err: err,
		}
	}

	return i > 0, nil
}

type eventMetadata struct {
	Timestamp     time.Time `json:"timestamp"`
	CorrelationId string    `json:"correlation_id"`
}

func (s *sqlEventRepository) appendToStream(ctx context.Context, streamId string, events []EventMessage, expectedVersion *int) error {
	if streamId == "" {
		return &ErrRepositoryExecution{
			Err: fmt.Errorf("streamId can't be empty"),
		}
	}

	evModels := make([]*model.EventStore, len(events))
	for i, ev := range events {
		ds, err := ev.Event().Marshal()
		if err != nil {
			return err
		}

		md, err := json.Marshal(eventMetadata{
			Timestamp:     time.Now(),
			CorrelationId: NewUUID(),
		})

		eventID := NewUUID()
		evModels[i] = &model.EventStore{
			EventID:   eventID,
			EventName: ev.Event().Name(),
			EventData: ds,
			Metadata:  parser.String(md).ToStringPtr(),
		}
		events[i].setID(&eventID)
	}

	err := s.db.GetQuery().Transaction(func(tx *models.Query) error {
		q := tx.WithContext(ctx)
		if err := q.EventStore.Omit(field.AssociationFields).Create(evModels...); err != nil {
			return err
		}

		evIds, err := transformer.ArrayTransformer[*model.EventStore, string](evModels, func(m *model.EventStore) (string, error) {
			return m.EventID, nil
		})

		if err != nil {
			return err
		}

		for _, evId := range evIds {
			// Get last stream
			evs, err := q.EventStream.Select(models.EventStream.StreamVersion).Where(models.EventStream.StreamID.Eq(streamId)).Order(models.EventStream.StreamVersion.Desc()).Limit(1).First()
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
			}

			// Get last version
			lastVersion := 0
			if evs != nil {
				lastVersion = int(evs.StreamVersion)
			}

			newVersion := lastVersion + 1
			if expectedVersion != nil {
				if *expectedVersion != lastVersion {
					return fmt.Errorf("Wrong expected stream version")
				}
			}

			if err := q.EventStream.Omit(field.AssociationFields).Create(&model.EventStream{
				StreamID:      streamId,
				StreamVersion: int32(newVersion),
				EventID:       evId,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return &ErrRepositoryExecution{
			Err: err,
		}
	}

	return nil
}

func (s *sqlEventRepository) Append(ctx context.Context, streamId string, events []EventMessage, expectedVersion *int) error {
	return s.appendToStream(ctx, streamId, events, expectedVersion)
}

func (s *sqlEventRepository) Link(ctx context.Context, streamId string, eventIds []string, expectedVersion *int) error {
	q := s.db.GetQuery().WithContext(ctx)
	// Make sure all eventIds in event
	es, err := q.EventStore.Select(models.EventStore.ID).Where(models.EventStore.EventID.In(eventIds...)).Find()
	if err != nil {
		return &ErrRepositoryExecution{
			Err: err,
		}
	}

	if len(es) != len(eventIds) {
		return &ErrRepositoryExecution{
			Err: fmt.Errorf("An event not exist"),
		}
	}

	for _, evId := range eventIds {
		// Get last stream
		evs, err := q.EventStream.Select(models.EventStream.StreamVersion).Where(models.EventStream.StreamID.Eq(streamId)).Order(models.EventStream.StreamVersion.Desc()).Limit(1).First()
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return &ErrRepositoryExecution{
					Err: err,
				}
			}
		}

		// Get last version
		lastVersion := 0
		if evs != nil {
			lastVersion = int(evs.StreamVersion)
		}

		newVersion := lastVersion + 1
		if expectedVersion != nil {
			if *expectedVersion != lastVersion {
				return fmt.Errorf("Wrong expected stream version")
			}

		}

		if err := q.EventStream.Omit(field.AssociationFields).Create(&model.EventStream{
			StreamID:      streamId,
			StreamVersion: int32(newVersion),
			EventID:       evId,
		}); err != nil {
			return &ErrRepositoryExecution{
				Err: err,
			}
		}
	}

	return nil
}

func (s *sqlEventRepository) DeleteStream(ctx context.Context, streamId string) error {
	if streamId == "" {
		return &ErrRepositoryExecution{
			Err: fmt.Errorf("streamId can't be empty"),
		}
	}
	_, err := s.db.GetQuery().EventStream.WithContext(ctx).Where(models.EventStream.StreamID.Eq(streamId)).Delete()
	if err != nil {
		return &ErrRepositoryExecution{
			Err: err,
		}
	}

	return nil
}

func (s *sqlEventRepository) Read(ctx context.Context) EventRepositoryReader {
	return &sqlEventRepositoryReader{
		streamQuery: s.db.GetQuery().WithContext(ctx).EventStream.ReadDB(),
		spec:        new(sqlEventRepositoryReaderSpec),
	}
}

func NewSqlEventRepository(driver, dsn string, eventBus EventBus, debug ...bool) (EventRepository, error) {
	var ormDriver orm.OrmDriver
	switch driver {
	case "postgres":
		ormDriver = orm.OrmDriverPostgres
	case "mysql":
		ormDriver = orm.OrmDriverMysql
	}

	db, err := orm.New(ormDriver, dsn, debug...)
	if err != nil {
		return nil, err
	}

	return &sqlEventRepository{
		db: db,
	}, nil
}
