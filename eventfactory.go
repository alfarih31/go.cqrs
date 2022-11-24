// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"
)

// EventFactory is the interface that an event factory should implement.
//
// An event factory returns instances of an event given the event type
// as a string.
// An event factory is required during deserialisation of events by the
// eventstore or repository depending on your implementation.
//
// The eventstore will return a string describing the event type. To unmarshal
// the contents of the persisted event which will typically be in some serialised
// format such as JSON an instance of the event type will need to be created.
type EventFactory interface {
	GetEvent(string) Event
	RegisterDelegate(eventName string, delegate func() Event) error
}

// DelegateEventFactory uses delegate functions to instantiate event instances
// given the name of the event type as a string.
type DelegateEventFactory struct {
	eventFactories map[string]func() Event
}

// NewDelegateEventFactory constructs a new DelegateEventFactory
func NewDelegateEventFactory() *DelegateEventFactory {
	return &DelegateEventFactory{
		eventFactories: make(map[string]func() Event),
	}
}

// RegisterDelegate registers a delegate that will return an event instance given
// an event type name as a string.
//
// If an attempt is made to register multiple delegates for an event type, an error
// is returned.
func (t *DelegateEventFactory) RegisterDelegate(eventName string, delegate func() Event) error {
	if _, ok := t.eventFactories[eventName]; ok {
		return fmt.Errorf("Factory delegate already registered for type: \"%s\"", eventName)
	}
	t.eventFactories[eventName] = delegate
	return nil
}

// GetEvent returns an event instance given an event type as a string.
//
// An appropriate delegate must be registered for the event type.
// If an appropriate delegate is not registered, the method will return nil.
func (t *DelegateEventFactory) GetEvent(typeName string) Event {
	if f, ok := t.eventFactories[typeName]; ok {
		return f()
	}
	return nil
}
