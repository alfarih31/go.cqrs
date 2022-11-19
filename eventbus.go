// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import "context"

// EventBus is the inteface that an event bus must implement.
type EventBus interface {
	PublishEvent(EventMessage)
	AddHandler(EventHandler, ...Event)
}

// InternalEventBus provides a lightweight in process event bus
type InternalEventBus struct {
	eventHandlers map[string]map[EventHandler]struct{}
}

// NewInternalEventBus constructs a new InternalEventBus
func NewInternalEventBus() *InternalEventBus {
	b := &InternalEventBus{
		eventHandlers: make(map[string]map[EventHandler]struct{}),
	}
	return b
}

// PublishEvent publishes events to all registered event handlers
func (b *InternalEventBus) PublishEvent(event EventMessage) {
	// Use context for concurrency safe
	ctx := context.TODO()
	if handlers, ok := b.eventHandlers[event.Event().Name()]; ok {
		for handler := range handlers {
			handler.Handle(ctx, event)
		}
	}
}

// AddHandler registers an event handler for all of the events specified in the
// variadic events parameter.
func (b *InternalEventBus) AddHandler(handler EventHandler, events ...Event) {

	for _, event := range events {
		eventName := event.Name()

		// There can be multiple handlers for any event.
		// Here we check that a map is initialized to hold these handlers
		// for a given type. If not we create one.
		if _, ok := b.eventHandlers[eventName]; !ok {
			b.eventHandlers[eventName] = make(map[EventHandler]struct{})
		}

		// Add this handler to the collection of handlers for the type.
		b.eventHandlers[eventName][handler] = struct{}{}
	}
}
