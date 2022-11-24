// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"context"
	"fmt"
	parser "github.com/alfarih31/nb-go-parser"
	"net/url"

	"github.com/jetbasrawi/go.geteventstore"
)

// DomainRepository is the interface that all domain repositories should implement.
type DomainRepository interface {
	//Load an aggregate of the given type and ID
	Load(ctx context.Context, streamId string, aggregateRoot AggregateRoot) error

	//Save the aggregate.
	Save(ctx context.Context, streamId string, aggregate AggregateRoot, expectedVersion *int) error

	SetEventFactory(factory EventFactory)
}

type DomainRepositoryBase struct {
	eventBus     EventBus
	eventFactory EventFactory
}

func (d *DomainRepositoryBase) SetEventFactory(factory EventFactory) {
	d.eventFactory = factory
}

func (d *DomainRepositoryBase) ValidateDependencies() error {
	if d.eventFactory == nil {
		return fmt.Errorf("the domain has no Event Factory")
	}

	return nil
}

func NewDomainRepositoryBase(eventBus EventBus) (*DomainRepositoryBase, error) {
	if eventBus == nil {
		return nil, fmt.Errorf("nil EventBus injected into repository")
	}

	return &DomainRepositoryBase{
		eventBus: eventBus,
	}, nil
}

// EventStoreDomainRepo is an implementation of the DomainRepository
// that uses GetEventStore for persistence
type EventStoreDomainRepo struct {
	*DomainRepositoryBase
	eventStore *goes.Client
}

// NewEventStoreDomainRepository constructs a new DomainRepository
func NewEventStoreDomainRepository(eventStore *goes.Client, eventBus EventBus) (*EventStoreDomainRepo, error) {
	if eventStore == nil {
		return nil, fmt.Errorf("nil Eventstore injected into repository")
	}

	base, err := NewDomainRepositoryBase(eventBus)
	if err != nil {
		return nil, err
	}

	d := &EventStoreDomainRepo{
		DomainRepositoryBase: base,
		eventStore:           eventStore,
	}
	return d, nil
}

// Load will load all events from a stream and apply those events to an aggregate
// of the type specified.
//
// The aggregate type and id will be passed to the configured StreamNamer to
// get the stream name.
func (r *EventStoreDomainRepo) Load(ctx context.Context, streamId string, aggregateRoot AggregateRoot) error {
	if err := r.ValidateDependencies(); err != nil {
		return err
	}

	stream := r.eventStore.NewStreamReader(streamId)
	for stream.Next() {
		switch err := stream.Err().(type) {
		case nil:
			break
		case *url.Error, *goes.ErrTemporarilyUnavailable:
			return &ErrRepositoryUnavailable{}
		case *goes.ErrNoMoreEvents:
			return nil
		case *goes.ErrUnauthorized:
			return &ErrUnauthorized{}
		case *goes.ErrNotFound:
			return &ErrAggregateNotFound{AggregateID: aggregateRoot.AggregateID()}
		default:
			return &ErrUnexpected{Err: err}
		}

		event := r.eventFactory.GetEvent(stream.EventResponse().Event.EventType)

		//TODO: No test for meta
		meta := make(map[string]string)
		stream.Scan(event, &meta)
		if stream.Err() != nil {
			return stream.Err()
		}
		em := NewEventMessage(parser.String(aggregateRoot.AggregateID()).ToStringPtr(), event, Int(stream.EventResponse().Event.EventNumber))
		for k, v := range meta {
			em.SetHeader(k, v)
		}
		aggregateRoot.Apply(em)
		aggregateRoot.IncrementVersion()
	}

	return nil

}

// Save persists an aggregate
func (r *EventStoreDomainRepo) Save(ctx context.Context, streamId string, aggregate AggregateRoot, expectedVersion *int) error {
	if err := r.ValidateDependencies(); err != nil {
		return err
	}

	resultEvents := aggregate.GetChanges()

	if len(resultEvents) > 0 {

		evs := make([]*goes.Event, len(resultEvents))

		for k, v := range resultEvents {
			//TODO: There is no test for this code
			v.SetHeader("AggregateID", aggregate.AggregateID())
			evs[k] = goes.NewEvent("", v.Event().Name(), v.Event(), v.GetHeaders())
		}

		streamWriter := r.eventStore.NewStreamWriter(streamId)
		err := streamWriter.Append(expectedVersion, evs...)
		switch e := err.(type) {
		case nil:
			break
		case *goes.ErrConcurrencyViolation:
			return &ErrConcurrencyViolation{Aggregate: aggregate, ExpectedVersion: expectedVersion, StreamName: streamId}
		case *goes.ErrUnauthorized:
			return &ErrUnauthorized{}
		case *goes.ErrTemporarilyUnavailable:
			return &ErrRepositoryUnavailable{}
		default:
			return &ErrUnexpected{Err: e}
		}
	}

	aggregate.ClearChanges()

	for k, v := range resultEvents {
		if expectedVersion == nil {
			r.eventBus.PublishEvent(v)
		} else {
			em := NewEventMessage(v.EventID(), v.Event(), Int(*expectedVersion+k+1))
			r.eventBus.PublishEvent(em)
		}
	}

	return nil
}
