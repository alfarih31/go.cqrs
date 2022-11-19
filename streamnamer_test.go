// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"fmt"

	. "gopkg.in/check.v1"
)

var _ = Suite(&DelegateStreamNamerSuite{})

type DelegateStreamNamerSuite struct {
	namer *DelegateStreamNamer
}

func (s *DelegateStreamNamerSuite) SetUpTest(c *C) {
	s.namer = NewDelegateStreamNamer()
}

func (s *DelegateStreamNamerSuite) TestNewDelegateStreamNamer(c *C) {
	namer := NewDelegateStreamNamer()
	c.Assert(namer.delegates, NotNil)
}

func (s *DelegateStreamNamerSuite) TestCanRegisterStreamNameDelegate(c *C) {

	err := s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)
	agg := NewSomeAggregate(NewUUID())
	f, _ := s.namer.delegates[TypeOf(agg)]
	stream := f(TypeOf(agg), agg.AggregateID())
	c.Assert(stream, Equals, agg.AggregateID()+TypeOf(agg))
}

func (s *DelegateStreamNamerSuite) TestCanRegisterStreamNameDelegateWithMultipleAggregateRoots(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
		&SomeOtherAggregate{},
	)
	c.Assert(err, IsNil)

	ar1 := NewSomeAggregate(NewUUID())
	f, _ := s.namer.delegates[TypeOf(ar1)]
	stream := f(TypeOf(ar1), ar1.AggregateID())
	c.Assert(stream, Equals, ar1.AggregateID()+TypeOf(ar1))

	ar2 := NewSomeOtherAggregate(NewUUID())
	f2, _ := s.namer.delegates[TypeOf(ar2)]
	stream2 := f2(TypeOf(ar2), ar2.AggregateID())
	c.Assert(stream2, Equals, ar2.AggregateID()+TypeOf(ar2))
}

func (s *DelegateStreamNamerSuite) TestCanGetStreamName(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)
	agg := NewSomeAggregate(NewUUID())
	stream, err := s.namer.GetStreamName(TypeOf(agg), agg.AggregateID())
	c.Assert(err, IsNil)
	c.Assert(stream, Equals, agg.AggregateID()+TypeOf(agg))
}

func (s *DelegateStreamNamerSuite) TestGetStreamNameReturnsAnErrorIfNoDelegateRegisteredForAggregate(c *C) {
	err := s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
	)
	agg := NewSomeOtherAggregate(NewUUID())
	stream, err := s.namer.GetStreamName(TypeOf(agg), agg.AggregateID())
	c.Assert(err, NotNil)
	c.Assert(stream, Equals, "")
	c.Assert(err, DeepEquals, fmt.Errorf("there is no stream name delegate for aggregate of type \"%s\"",
		TypeOf(agg)))

}

func (s *DelegateStreamNamerSuite) TestRegisteringAStreamNameDelegateMoreThanOnceReturnsAndError(c *C) {

	err := s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
	)
	c.Assert(err, IsNil)

	err = s.namer.RegisterDelegate(func(a string, id string) string { return id + a },
		&SomeAggregate{},
	)
	c.Assert(err, DeepEquals,
		fmt.Errorf("the stream name delegate for \"%s\" is already registered with the stream namer",
			TypeOf(NewSomeAggregate(NewUUID()))))
}
