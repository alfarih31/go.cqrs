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

func (e *SqlDomainRepo) Load(ctx context.Context, streamId string, aggregateRoot AggregateRoot) error {
	if err := e.ValidateDependencies(); err != nil {
		return err
	}

	msgs, err := e.repo.Read(ctx).Stream(streamId).Forward().ToList()
	if err != nil {
		return err
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
		return NewEventMessage(em.EventID(), event, em.Version()), nil
	})
	if err != nil {
		return err
	}

	aggregateRoot.RebuildFromEvents(evs)

	return nil
}

func (e *SqlDomainRepo) Save(ctx context.Context, streamId string, aggregate AggregateRoot, expectedVersion *int) error {
	changes := aggregate.GetChanges()

	err := e.repo.Append(ctx, streamId, changes, expectedVersion)
	if err != nil {
		return err
	}

	// Set version to current version on saved
	aggregate.setVersion(aggregate.CurrentVersion())
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
