// Copyright 2016 Jet Basrawi. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package ycq

import "context"

func NewMockEventHandler() *MockEventHandler {
	return &MockEventHandler{
		make([]EventMessage, 0),
	}
}

type MockEventHandler struct {
	events []EventMessage
}

func (m *MockEventHandler) Handle(ctx context.Context, event EventMessage) {
	m.events = append(m.events, event)
}
