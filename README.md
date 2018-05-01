# Goengine

Goengine is a simple implementation of a concurrent shared-state design pattern
for Go. It is inspired by Mozilla's [Rob
Miller](https://blog.mozilla.org/services/2014/03/12/sane-concurrency-with-go/)
and the [Redux pattern](https://redux.js.org/).

The idea behind Goengine is that in order to prevent race conditions, only one
goroutine should write to global state. Goengine uses a main loop which listens
for instructions ("Actions") and dispatches handlers ("Reducers") accordingly. 

## System architecture

<img width="500" src="./engine.png" />
