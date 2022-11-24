// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/jetbasrawi/go.geteventstore"
	"github.com/jetbasrawi/go.geteventstore.testfeed"
	. "gopkg.in/check.v1"
)

var (
	_ = Suite(&ComDomRepoSuite{})
)

type ComDomRepoSuite struct {
	eventBus           EventBus
	repo               *EventStoreDomainRepo
	someEvent          *SomeEvent
	someMeta           map[string]string
	someMockEvent      *mock.Event
	someOtherEvent     *SomeOtherEvent
	someOtherMeta      map[string]string
	someOtherMockEvent *mock.Event
	mux                *http.ServeMux
	server             *httptest.Server
	client             *goes.Client
	streamName         string
	stubFeed           *mock.AtomFeedSimulator
}

func (s *ComDomRepoSuite) SetUpTest(c *C) {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
	s.client, _ = goes.NewClient(nil, s.server.URL)

	s.SetupDefaultRepo(s.client)
}

func (s *ComDomRepoSuite) TearDownTest(c *C) {
	s.server.Close()
}

func (s *ComDomRepoSuite) SetupDefaultSimulator() {
	s.streamName = "astream"

	// Set up an event of type SomeEvent
	s.someEvent = &SomeEvent{Item: "Some Item", Count: 42}
	s.someMeta = map[string]string{"AggregateID": NewUUID()}
	s.someMockEvent = mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, s.someEvent, s.someMeta)

	// Set up an event of type SomeOtherEvent
	s.someOtherEvent = &SomeOtherEvent{NewUUID()}
	s.someOtherMeta = map[string]string{"AggregateID": NewUUID()}
	s.someOtherMockEvent = mock.CreateTestEventFromData(s.streamName, s.server.URL, 1, s.someOtherEvent, s.someOtherMeta)

	// Create a slice with the two events for the Atom feed simulator
	es := []*mock.Event{s.someMockEvent, s.someOtherMockEvent}

	s.SetupSimulator(es, nil)
}

func (s *ComDomRepoSuite) SetupSimulator(es []*mock.Event, m *mock.Event) {
	// Set up an AtomFeedSimulator
	u, _ := url.Parse(s.server.URL)
	sim, err := mock.NewAtomFeedSimulator(es, u, nil, -1)
	if err != nil {
		log.Fatal(err)
	}
	s.stubFeed = sim

	// Set the http handler
	s.mux.Handle("/", s.stubFeed)
}

func (s *ComDomRepoSuite) SetupDefaultRepo(client *goes.Client) {
	s.eventBus = NewInternalEventBus()

	s.repo, _ = NewEventStoreDomainRepository(client, s.eventBus)

	eventFactory := NewDelegateEventFactory()
	eventFactory.RegisterDelegate("SomeEvent",
		func() Event { return &SomeEvent{} })
	eventFactory.RegisterDelegate("SomeOtherEvent",
		func() Event { return &SomeOtherEvent{} })
	s.repo.SetEventFactory(eventFactory)
}

func (s *ComDomRepoSuite) TestCanConstructNewRepository(c *C) {
	store, _ := goes.NewClient(nil, "")
	eventBus := NewInternalEventBus()

	repo, err := NewEventStoreDomainRepository(store, eventBus)

	c.Assert(repo, NotNil)
	c.Assert(err, IsNil)
	c.Assert(repo.eventBus, NotNil)
}

func (s *ComDomRepoSuite) TestCreatingNewRepositoryWithNilEventStoreReturnsAnError(c *C) {
	eventBus := NewInternalEventBus()
	repo, err := NewEventStoreDomainRepository(nil, eventBus)

	c.Assert(repo, IsNil)
	c.Assert(err, DeepEquals, fmt.Errorf("nil Eventstore injected into repository"))
}

func (s *ComDomRepoSuite) TestCreatingNewRepositoryWithNilEventBusReturnsAnError(c *C) {
	store, _ := goes.NewClient(nil, "")
	repo, err := NewEventStoreDomainRepository(store, nil)

	c.Assert(repo, IsNil)
	c.Assert(err, DeepEquals, fmt.Errorf("nil EventBus injected into repository"))
}

func (s *ComDomRepoSuite) TestRepositoryCanLoadAggregateWithEvents(c *C) {

	s.SetupDefaultSimulator()

	id := NewUUID()
	got := NewStubAggregate(id)
	err := s.repo.Load(context.Background(), "StubAggregate", got)

	c.Assert(err, IsNil)
	c.Assert(got.AggregateID(), Equals, id)

	events := got.events

	c.Assert(events[0].Event(), DeepEquals, s.someEvent)
	c.Assert(events[1].Event(), DeepEquals, s.someOtherEvent)
}

func (s *ComDomRepoSuite) TestRepositoryIncrementsAggregateVersionForEachEvent(c *C) {
	ev1 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev2 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	ev3 := mock.CreateTestEventFromData(s.streamName, s.server.URL, 0, &SomeEvent{Item: "Some Item", Count: 42}, nil)
	es := []*mock.Event{ev1, ev2, ev3}

	s.SetupSimulator(es, nil)

	id := NewUUID()
	got := NewSomeAggregate(id)
	err := s.repo.Load(context.Background(), "SomeAggregate", got)
	c.Assert(err, IsNil)

	// Version is a zero based index. The first item is zero
	c.Assert(got.OriginalVersion(), Equals, 3)
	c.Assert(got.CurrentVersion(), Equals, 3)
}

func (s *ComDomRepoSuite) TestSaveAggregateWithUncommittedChanges(c *C) {

	someEvent := &SomeEvent{Item: "Some string", Count: 4353}
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)

		es := []*goes.Event{}
		var d json.RawMessage
		var m json.RawMessage
		ev := &goes.Event{Data: &d, MetaData: &m}
		es = append(es, ev)

		err := json.NewDecoder(r.Body).Decode(&es)
		c.Assert(err, IsNil)

		data, ok := es[0].Data.(*json.RawMessage)
		c.Assert(ok, Equals, true)
		e := &SomeEvent{}
		err = json.Unmarshal(*data, e)
		c.Assert(e, DeepEquals, someEvent)

	})

	id := NewUUID()
	agg := NewSomeAggregate(id)

	em := NewEventMessage(&id, someEvent, nil)
	agg.TrackChange(em)

	err := s.repo.Save(context.Background(), "SomeAggregate", agg, nil)

	c.Assert(err, IsNil)

}

func (s *ComDomRepoSuite) TestLoadReturnErrUnauthorized(c *C) {
	id := NewUUID()
	agg := NewSomeAggregate(id)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "")
	})

	err := s.repo.Load(context.Background(), "SomeAggregate", agg)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnauthorized{})

}

func (s *ComDomRepoSuite) TestLoadReturnErrUnavailable(c *C) {
	id := NewUUID()
	agg := NewSomeAggregate(id)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "")
	})

	err := s.repo.Load(context.Background(), "SomeAggregate", agg)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrRepositoryUnavailable{})

}

func (s *ComDomRepoSuite) TestSaveReturnErrUnauthorized(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	eventId := NewUUID()
	agg.TrackChange(NewEventMessage(&eventId, &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(context.Background(), "SomeAggregate", agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnauthorized{})

}

func (s *ComDomRepoSuite) TestSaveReturnErrUnavailable(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	eventId := NewUUID()
	agg.TrackChange(NewEventMessage(&eventId, &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(context.Background(), "SomeAggregate", agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrRepositoryUnavailable{})

}

func (s *ComDomRepoSuite) TestReturnsErrorOnLoadIfEventFactoryNotRegistered(c *C) {
	s.repo.eventFactory = nil

	agg := NewSomeAggregate(NewUUID())
	err := s.repo.Load(context.Background(), "SomeAggregate", agg)

	c.Assert(err, DeepEquals, fmt.Errorf("the domain has no Event Factory"))
}

func (s *ComDomRepoSuite) TestCanSetEventFactory(c *C) {
	eventFactory := NewDelegateEventFactory()

	s.repo.SetEventFactory(eventFactory)

	c.Assert(s.repo.eventFactory, Equals, eventFactory)
}

func (s *ComDomRepoSuite) TestAggregateNotFoundError(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	err := s.repo.Load(context.Background(), "SomeAggregate", agg)

	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrAggregateNotFound{AggregateID: id, AggregateType: TypeOf(&SomeAggregate{})})
}

func (s *ComDomRepoSuite) TestSaveUnhandledErrors(c *C) {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "")
	})

	id := NewUUID()
	agg := NewSomeAggregate(id)
	eventId := NewUUID()
	agg.TrackChange(NewEventMessage(&eventId, &SomeEvent{"Some data", 4}, nil))

	err := s.repo.Save(context.Background(), "SomeAggregate", agg, nil)
	c.Assert(err, NotNil)
	c.Assert(err, FitsTypeOf, &ErrUnexpected{})

}

func (s *ComDomRepoSuite) TestNewEventsArePublishedOnSave(c *C) {
	fakeHandler := &FakeEventHandler{}
	s.eventBus.AddHandler(fakeHandler, "SomeEvent", "SomeOtherEvent")

	eventId1 := NewUUID()
	eventId2 := NewUUID()
	em1 := NewEventMessage(&eventId1, &SomeEvent{"--------PUBLISH", 456}, nil)
	em2 := NewEventMessage(&eventId2, &SomeOtherEvent{"--------PUBLISH"}, nil)

	agg := NewSomeAggregate(NewUUID())
	agg.TrackChange(em1)
	agg.TrackChange(em2)

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c.Assert(r.Method, Equals, http.MethodPost)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, "")
	})

	err := s.repo.Save(context.Background(), "SomeAggregate", agg, Int(agg.OriginalVersion()))
	c.Assert(err, IsNil)

	//spew.Dump(s.eventBus)

	c.Assert(fakeHandler.Events, HasLen, 2)
	got1 := fakeHandler.Events[0].Version()
	got2 := fakeHandler.Events[1].Version()
	c.Assert(*got1, Equals, 1)
	c.Assert(*got2, Equals, 2)

}

//////////////////////////////////////////////////////////////////////////////
// Fakes

type FakeEventHandler struct {
	Events []EventMessage
}

func (h *FakeEventHandler) Handle(ctx context.Context, message EventMessage) {
	h.Events = append(h.Events, message)
}

func NewStubAggregate(id string) *StubAggregate {
	return &StubAggregate{
		AggregateBase: NewAggregateBase(id),
	}
}

type StubAggregate struct {
	*AggregateBase
	events []EventMessage
}

func (t *StubAggregate) RebuildFromEvents(events []EventMessage) {
	//TODO implement me
	panic("implement me")
}

func (t *StubAggregate) AggregateType() string {
	return "StubAggregate"
}

func (t *StubAggregate) Handle(command CommandMessage) error {
	return nil
}

func (t *StubAggregate) Apply(event EventMessage) {
	t.events = append(t.events, event)
}
