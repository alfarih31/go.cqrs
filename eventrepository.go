package ycq

import (
	"context"
	"time"
)

type EventRepository interface {
	Append(ctx context.Context, streamId string, events []EventMessage, expectedVersion *int) error
	Link(ctx context.Context, streamId string, eventIds []string, expectedVersion *int) error
	DeleteStream(ctx context.Context, streamId string) error
	Read(ctx context.Context) EventRepositoryReader
	HasEvent(ctx context.Context, id string) (bool, error)
	GetStreamIdOf(ctx context.Context, eventId string) (string, error)
	GetVersionInStream(ctx context.Context, streamId, eventId string) (*int, error)
	IsEventInStream(ctx context.Context, streamId, eventId string) (bool, error)
}

type EventRepositoryReader interface {
	Stream(streamId string) EventRepositoryReader
	FromTime(date time.Time) EventRepositoryReader
	FromId(id int) EventRepositoryReader
	ToTime(date time.Time) EventRepositoryReader
	ToId(id int) EventRepositoryReader
	Forward() EventRepositoryReader
	Backward() EventRepositoryReader
	Limit(count int) EventRepositoryReader
	Event(id string) (EventMessage, error)
	Events(ids []string) ([]EventMessage, error)
	Count() (int, error)
	ToList() ([]EventMessage, error)
	Last(streamId string) (EventMessage, error)
}
