// This package makes it easier to write apps that require concurrent
// operations with shared state. It implements a Heka- and Flow-esque pattern
// where only one goroutine modifies the state. The basic principle is that an
// Engine has a main loop which listens to an ActionChan channel and dispatches
// Reducers accordingly. Engine state is a map which uses StateKeys to look up
// data, and each Reducer is assigned its own StateKey.

package goengine

import (
	"errors"
	"sync"
)

type StateKey int
type ReducerKey int
type ReducerFunc func(*Engine, StateKey, interface{}) Response
type State map[StateKey]interface{}

type Reducer struct {
	StateKey    StateKey
	ReducerFunc ReducerFunc
}

type Action struct {
	ReducerKey   ReducerKey
	Data         interface{}
	ResponseChan chan Response
}

type Response struct {
	Data interface{}
	Err  error
}

type Engine struct {
	ActionChan     chan Action
	NextReducerKey ReducerKey
	NextStateKey   StateKey
	State          State
	Mux            sync.RWMutex
	Reducers       map[ReducerKey]Reducer
}

// BuildEngine() initialises an Engine
func BuildEngine() Engine {
	return Engine{
		ActionChan:     make(chan Action),
		NextReducerKey: 0,
		NextStateKey:   0,
		State:          make(State),
		Reducers:       make(map[ReducerKey]Reducer),
	}
}

// Register() stores a ReducerFunc in the Engine and provides a StateKey and
// ReducerKey - with which one can access the respective portion of the State
// and said Reducer function.
func (e *Engine) Register(initialState interface{}, rf ReducerFunc) (StateKey, ReducerKey) {
	// Set initialState and the StateKey
	sk := e.NextStateKey
	e.State[sk] = initialState
	e.NextStateKey += 1

	// Set the ReducerFunc and the ReducerKey
	r := Reducer{StateKey: sk, ReducerFunc: rf}
	rk := e.NextReducerKey
	e.NextReducerKey += 1
	e.Reducers[rk] = r
	return sk, rk
}

// Dispatch() looks up a ReducerFunc to run given a ReducerKey
func (e *Engine) dispatch(rk ReducerKey, data interface{}) Response {
	return e.Reducers[rk].ReducerFunc(e, e.Reducers[rk].StateKey, data)
}

// Run() launches goroutine to listen for Actions in ActionChan, Dispatch()
// them, and feed the response into ResponseChan.
func (e *Engine) Run() {
	go func() {
		for action := range (*e).ActionChan {
			response := e.dispatch(action.ReducerKey, action.Data)
			action.ResponseChan <- response
		}
	}()
}

func (e *Engine) feedAction(rk ReducerKey, data interface{}) chan Response {
	rc := make(chan Response)
	a := Action{ReducerKey: rk, Data: data, ResponseChan: rc}
	e.ActionChan <- a
	return a.ResponseChan
}

func (e *Engine) Act(rk ReducerKey, data interface{}) (Response, error) {
	responseChan := e.feedAction(rk, data)
	for response := range responseChan {
		defer close(responseChan)
		return response, nil
	}
	return Response{nil, nil},
		errors.New("Unable to return a response from responseChan")
}

// Get() returns a copy of the section of the state denoted by @sk.
func (e *Engine) Get(sk StateKey) interface{} {
	e.Mux.RLock()
	defer e.Mux.RUnlock()
	return e.State[sk]
}

// set() replaces the state denoted by @sk with @data.
func (e *Engine) Set(sk StateKey, data interface{}) {
	e.Mux.RLock()
	e.State[sk] = data
	e.Mux.RUnlock()
}
