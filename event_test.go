// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"encoding/json"
	"math/rand"

	. "gopkg.in/check.v1"
)

var _ = Suite(&EventSuite{})

type EventSuite struct {
}

type SomeEvent struct {
	Item  string `json:"item"`
	Count int    `json:"count"`
}

func (s SomeEvent) Data() interface{} {
	return s
}

func (s SomeEvent) Name() string {
	return "SomeEvent"
}

func (s SomeEvent) Unmarshal(raw string) error {
	return json.Unmarshal([]byte(raw), &s)
}

func (s SomeEvent) Marshal() (string, error) {
	jb, err := json.Marshal(s.Data())
	if err != nil {
		return "", err
	}

	return string(jb), nil
}

type SomeOtherEvent struct {
	OrderID string `json:"order_id"`
}

func (s SomeOtherEvent) Data() interface{} {
	return s
}

func (s SomeOtherEvent) Name() string {
	return "SomeOtherEvent"
}

func (s SomeOtherEvent) Unmarshal(raw string) error {
	return json.Unmarshal([]byte(raw), &s)
}

func (s SomeOtherEvent) Marshal() (string, error) {
	jb, err := json.Marshal(s.Data())
	if err != nil {
		return "", err
	}

	return string(jb), nil
}

func NewTestEventMessage(id string) *EventDescriptor {
	ev := &SomeEvent{Item: NewUUID(), Count: rand.Intn(100)}
	return NewEventMessage(&id, ev, nil)
}

func (s *EventSuite) TestNewEventMessage(c *C) {
	id := NewUUID()
	ev := &SomeEvent{Item: "Some String", Count: 43}

	em := NewEventMessage(&id, ev, nil)

	c.Assert(*em.id, Equals, id)
	c.Assert(em.event, Equals, ev)
	c.Assert(em.headers, NotNil)
}

func (s *EventSuite) TestShouldGetTypeOfEvent(c *C) {
	se := &SomeEvent{"Some String", 42}
	em := &EventDescriptor{event: se}
	c.Assert(em.Event().Name(), Equals, "SomeEvent")
}

//TODO: Do i need this still?
//func (s *EventSuite) TestShouldGetTypeOfAggregate(c *C) {
//em := &EventMessage{aggregate: &SomeAggregate{}}
//c.Assert(em.AggregateType(), Equals, "SomeAggregate")
//}

func (s *EventSuite) TestShouldGetEventVersion(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, Int(5))

	c.Assert(*em.Version(), Equals, 5)
}

func (s *EventSuite) TestShouldGetHeaders(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, nil)
	em.headers["a"] = "b"

	h := em.GetHeaders()

	c.Assert(h, DeepEquals, em.headers)
}

func (s *EventSuite) TestShouldGetEvent(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, nil)
	got := em.Event()
	c.Assert(got, DeepEquals, em.event)
}

func (s *EventSuite) TestAddHeaderInt(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, nil)

	em.SetHeader("a", 3)

	c.Assert(em.headers["a"], Equals, 3)
}

func (s *EventSuite) TestAddHeaderString(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, nil)

	em.SetHeader("a", "abc")

	c.Assert(em.headers["a"], Equals, "abc")
}

func (s *EventSuite) TestAddHeaderStruct(c *C) {
	ev := &SomeEvent{"Some data", 456}
	em := NewEventMessage(nil, ev, nil)

	em.SetHeader("a", ev)

	c.Assert(em.headers["a"], DeepEquals, ev)
}
