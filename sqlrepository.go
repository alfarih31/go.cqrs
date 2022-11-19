package ycq

import (
	"context"
	"fmt"
	"github.com/jetbasrawi/go.cqrs/internal/transformer"
)

// SqlDomainRepo is an implementation of the DomainRepository
// that uses Get for persistence
type SqlDomainRepo struct {
	*DomainRepositoryBase
	repo EventRepository
}

func (e *SqlDomainRepo) Load(ctx context.Context, aggregateTypeName, aggregateId string) (AggregateRoot, error) {
	if err := e.ValidateDependencies(); err != nil {
		return nil, err
	}

	// Get aggregate
	aggregate := e.aggregateFactory.GetAggregate(aggregateTypeName, aggregateId)
	if aggregate == nil {
		return nil, &ErrAggregateNotFound{
			AggregateID:   aggregateId,
			AggregateType: aggregateTypeName,
		}
	}

	streamName, err := e.streamNameDelegate.GetStreamName(aggregateTypeName, aggregateId)
	if err != nil {
		return nil, err
	}

	msgs, err := e.repo.Read(ctx).Stream(streamName).Forward().ToList()
	if err != nil {
		return nil, err
	}

	evs, err := transformer.ArrayTransformer[EventMessage, EventMessage](msgs, func(em EventMessage) (EventMessage, error) {
		event := e.eventFactory.GetEvent(em.Event().Name())
		if event == nil {
			return nil, &ErrEventNotFound{
				EventName: em.Event().Name(),
			}
		}

		if err = event.Unmarshal(em.Event().Data().(string)); err != nil {
			return nil, &ErrUnexpected{Err: err}
		}
		return NewEventMessage(em.AggregateID(), event, em.Version()), nil
	})
	if err != nil {
		return nil, err
	}

	if len(evs) == 0 {
		return nil, &ErrRepositoryExecution{
			Err: &ErrAggregateNotFound{
				AggregateID:   aggregateId,
				AggregateType: aggregateTypeName,
			},
		}
	}

	aggregate.RebuildFromEvents(evs)

	return aggregate, nil
}

func (e *SqlDomainRepo) Save(ctx context.Context, aggregate AggregateRoot, expectedVersion *int) error {
	if e.streamNameDelegate == nil {
		return fmt.Errorf("the domain repository has no stream name delegate")
	}

	changes := aggregate.GetChanges()

	streamName, err := e.streamNameDelegate.GetStreamName(TypeOf(aggregate), aggregate.AggregateID())
	if err != nil {
		return err
	}

	err = e.repo.Append(ctx, streamName, changes, expectedVersion)
	if err != nil {
		return err
	}

	aggregate.ClearChanges()

	for _, v := range changes {
		e.eventBus.PublishEvent(v)
	}

	return nil
}

// NewSqlDomainRepository constructs a new CommonDomainRepository
func NewSqlDomainRepository(repo EventRepository, eventBus EventBus) (DomainRepository, error) {
	if repo == nil {
		return nil, fmt.Errorf("nil Eventstore injected into repository")
	}

	base, err := NewDomainRepositoryBase(eventBus)
	if err != nil {
		return nil, err
	}

	return &SqlDomainRepo{
		DomainRepositoryBase: base,
		repo:                 repo,
	}, nil
}
