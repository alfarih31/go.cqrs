// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"

	. "gopkg.in/check.v1"
)

var _ = Suite(&DelegateEventFactorySuite{})

type DelegateEventFactorySuite struct {
	factory *DelegateEventFactory
}

func (s *DelegateEventFactorySuite) SetUpTest(c *C) {
	s.factory = NewDelegateEventFactory()
}

func (s *DelegateEventFactorySuite) TestNewEventFactory(c *C) {
	factory := NewDelegateEventFactory()
	c.Assert(factory.eventFactories, NotNil)
}

func (s *DelegateEventFactorySuite) TestCanRegisterEventFactoryDelegate(c *C) {
	err := s.factory.RegisterDelegate("SomeEvent",
		func() Event { return &SomeEvent{} })

	c.Assert(err, IsNil)

	c.Assert(s.factory.eventFactories[TypeOf(&SomeEvent{})](),
		DeepEquals,
		&SomeEvent{})
}

func (s *DelegateEventFactorySuite) TestDuplicateEventFactoryRegistrationReturnsAnError(c *C) {
	err := s.factory.RegisterDelegate("SomeEvent",
		func() Event { return &SomeEvent{} })

	c.Assert(err, IsNil)

	err = s.factory.RegisterDelegate("SomeEvent",
		func() Event { return &SomeEvent{} })

	c.Assert(err, NotNil)
	c.Assert(err,
		DeepEquals,
		fmt.Errorf("Factory delegate already registered for type: \"%s\"",
			TypeOf(&SomeEvent{})))
}

func (s *DelegateEventFactorySuite) TestCanGetEventInstanceFromString(c *C) {
	_ = s.factory.RegisterDelegate("SomeEvent",
		func() Event { return &SomeEvent{} })

	ev := s.factory.GetEvent(TypeOf(&SomeEvent{}))
	c.Assert(ev, DeepEquals, &SomeEvent{})
}
