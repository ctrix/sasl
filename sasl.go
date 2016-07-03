// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package sasl

import (
	"errors"
)

var (
	ErrInvalidState = errors.New("Invalid state")
	ErrAuthn        = errors.New("Authentication error")
	ErrAuthz        = errors.New("Authorization error")
	ErrTooManySteps = errors.New("Step called too many times")
)

// State represents the current state of a Mechanism's underlying state machine.
type State int8

const (
	Initial State = iota
	AuthTextSent
	ResponseSent
	ValidServerResponse
)

// Mechanism represents a SASL mechanism.
//
// A Mechanism is stateful and therefore should not be shared between
// goroutines.
type Mechanism struct {
	Name  string
	Start func(state State) (more bool, resp []byte, err error)
	Next  func(state State, challenge []byte) (more bool, resp []byte, err error)

	state State
	resp  []byte
	err   error
}

// Step attempts to transition the SASL mechanism to its next state. If Step is
// called after a previous invocation generates an error (and the Mechanism has
// not been reset to its initial state), Step panics.
func (m *Mechanism) Step(challenge []byte) (more bool, err error) {
	if m.Err() != nil {
		panic(m.Err())
	}

	switch m.state {
	case Initial:
		more, m.resp, m.err = m.Start(m.state)
		m.state = AuthTextSent
		return more, m.err
	case AuthTextSent:
		more, m.resp, m.err = m.Next(m.state, challenge)
		m.state = ResponseSent
		return more, m.err
	case ResponseSent:
		more, m.resp, m.err = m.Next(m.state, challenge)
		m.state = ValidServerResponse
		return more, m.err
	case ValidServerResponse:
		more, m.resp, m.err = m.Next(m.state, challenge)
		return more, m.err
	}

	return false, ErrInvalidState
}

// Err returns any errors generated by the SASL mechanism.
func (m *Mechanism) Err() error {
	return m.err
}

// State returns the internal state of the SASL mechanism.
func (m *Mechanism) State() State {
	return m.state
}