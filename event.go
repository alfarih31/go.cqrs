// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

type Event interface {
	Data() interface{}
	Name() string
	Unmarshal(rawString string) error
	Marshal() (string, error)
}

// EventMessage is the interface that a command must implement.
type EventMessage interface {
	setID(ID *string)

	// EventID returns the ID of the event
	EventID() *string

	// GetHeaders returns the key value collection of headers for the event.
	//
	// Headers are metadata about the event that do not form part of the
	// actual event but are still required to be persisted alongside the event.
	GetHeaders() map[string]interface{}

	// SetHeader sets the value of the header specified by the key
	SetHeader(string, interface{})

	// Returns the actual event which is the payload of the event message.
	Event() Event

	// Version returns the version of the event
	Version() *int
}

var _ Event = new(RawEvent)

type RawEvent struct {
	name string
	data interface{}
}

func (r *RawEvent) Name() string {
	return r.name
}

func (r *RawEvent) Unmarshal(rawString string) error {
	panic("implement me")
}

func (r *RawEvent) Marshal() (string, error) {
	panic("implement me")
}

func (r *RawEvent) Data() interface{} {
	return r.data
}

// EventDescriptor is an implementation of the event message interface.
type EventDescriptor struct {
	id      *string
	event   Event
	headers map[string]interface{}
	version *int
}

// NewEventMessage returns a new event descriptor
func NewEventMessage(eventID *string, event Event, version *int) *EventDescriptor {
	return &EventDescriptor{
		id:      eventID,
		event:   event,
		headers: make(map[string]interface{}),
		version: version,
	}
}

// setID returns the ID of the Aggregate that the event relates to.
func (c *EventDescriptor) setID(id *string) {
	c.id = id
}

// EventID returns the ID of the Aggregate that the event relates to.
func (c *EventDescriptor) EventID() *string {
	return c.id
}

// GetHeaders returns the headers for the event.
func (c *EventDescriptor) GetHeaders() map[string]interface{} {
	return c.headers
}

// SetHeader sets the value of the header specified by the key
func (c *EventDescriptor) SetHeader(key string, value interface{}) {
	c.headers[key] = value
}

// Event the event payload of the event message
func (c *EventDescriptor) Event() Event {
	return c.event
}

// Version returns the version of the event
func (c *EventDescriptor) Version() *int {
	return c.version
}
