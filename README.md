# Goengine

Goengine is a simple implementation of a concurrent shared-state design pattern
in Go. It is inspired by a [blog
post](https://blog.mozilla.org/services/2014/03/12/sane-concurrency-with-go/)
by Rob Miller from Mozilla, as well as the [Reactor
pattern](https://en.wikipedia.org/wiki/Reactor_pattern).

To achieve this, Goengine uses a main loop which listens for instructions
("Actions") and dispatches handlers ("Reducers") accordingly. State is
protected by a [`RWMutex`](https://golang.org/pkg/sync/#RWMutex). The result
from a handler function flows from the event loop via a channel. All these
details are abstracted away from the user.

Goengine is not ready for production. In particular, it needs to be
stress-tested to verify that it does not suffer from race conditions.

## Installation

```bash
go get github.com/weijiekoh/goengine
```

## Example code

Look at [`goengine_test.go`](./goengine_test.go) for example code.

## Usage

First, import `goengine`:

```golang
import (
    goengine "github.com/weijiekoh/goengine" 
)
```

Next, define a handler. It must follow the method signature shown below,
and if it needs to return data and/or an `error`, it should do so via a
`Response` value. In the example below, the handler uses `Get()`, `Set()`,
and the given `StateKey` to manipulate the engine state, and returns a
`Response` with `nil` values. (See below for an example that returns a Response
with a non-`nil` value.

```golang
func storeNum(
    eng *goengine.Engine, sk goengine.StateKey, num interface{},
) goengine.Response {
	nums := eng.Get(sk).([]int)
	nums = append(nums, num.(int))
	eng.Set(sk, nums)
	return goengine.Response{Data: nil, Err: nil}
}
```

Next, create and run the engine, and register your handler:

```golang
engine := goengine.BuildEngine()
engine.Run()
numsSk, numsRk := engine.Register(make([]int, 0), storeNum)
```

The handler may use the state key, `numsSk`, to access the state via `Get`
and `Set`. **Do not** run `engine.Set` outside of a handler as this may lead to
a race condition.

You need the reducer key (`numsRk`) to trigger your handler:

```golang
engine.Act(numsRk, val)
```

The following example shows how to access the response from a handler. Note
that you need to handle data type conversions as the State is agnostic about
them.

```golang
package main

import (
	goengine "github.com/weijiekoh/goengine"
	"log"
)

func incrementNum(
	eng *goengine.Engine, sk goengine.StateKey, num interface{},
) goengine.Response {
	return goengine.Response{Data: num.(int) + 1, Err: nil}
}

func main() {
	vals := []int{0, 1, 2, 3, 4}

	// Create and run the engine
	engine := goengine.BuildEngine()
	engine.Run()

	// Register the incrementNum handler
	_, rk := engine.Register(nil, incrementNum)

	result := make([]int, 0)

	for _, val := range vals {
		// Dispatch incrementNum using the given ReducerKey rk
		response, _ := engine.Act(rk, val)
		result = append(result, response.Data.(int))
	}

	log.Println(result)
}
```
