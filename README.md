# scope

scope is a simple tool to manage function scope in go.

## Installation

```bash
go get github.com/snowmerak/scope
```

## Usage

### Make State

```go
package riddle

import (
	"sync/atomic"
)

type State struct {
	number atomic.Int64
}
```

The state is a struct to use in the scope.  
The state can be any struct, but it must be passed as a pointer to the scope.

### Make Work

```go
var _ scope.Work[State] = AddOne

func AddOne(ctx context.Context, state *State) error {
	state.number.Add(1)
	return nil
}

var _ scope.Work[State] = AddTwo

func AddTwo(ctx context.Context, state *State) error {
	state.number.Add(2)
	return nil
}

var _ scope.Work[State] = SubOne

func SubOne(ctx context.Context, state *State) error {
	state.number.Add(-1)
	return nil
}

var _ scope.Work[State] = SetZero

func SetZero(ctx context.Context, state *State) error {
	state.number.Store(0)
	return nil
}
```

The work is a function to use in the scope.  
The work has two parameters, `context.Context` and `*State`.  
The work must return an error. This error will be checked by the `Checker`.

### Make Checker

```go
package main

import (
    "log"
)

func main() {
    printChecker := func(err error) {
        log.Println(err)
    }
}
```

The checker is a function to check the error returned by the work.

### Make CleanUp

```go
package riddle

import (
    "context"
    "fmt"
    "log"
    "github.com/snowmerak/scope"
)

var _ scope.CleanUp[State] = ResetAndPrint

func ResetAndPrint(ctx context.Context, state *State) {
	log.Printf("State: %d\n", state.number.Load())
	state.number.Store(0)
}

var _ scope.CleanUp[State] = JustPrint

func JustPrint(ctx context.Context, state *State) {
	fmt.Printf("State: %d\n", state.number.Load())
}
```

The cleanup is a function to clean up the state after the scope.  
The cleanup has two parameters, `context.Context` and `*State`.

### Use Scope

```go
package main

import (
	"context"
	"github.com/snowmerak/scope"
	"github.com/snowmerak/scope/prac/riddle"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := riddle.NewState()

	scope.Sequence(ctx, r, func(err error) {
		log.Printf("sequence error: %v", err)
	}, riddle.ResetAndPrint, riddle.AddOne, riddle.AddTwo, riddle.SubOne)
}
```

The scope is a function to manage the state and the work.  
The scope has four parameters, `context.Context`, `*State`, `scope.Checker`, and `scope.Work...`.

When the scope is ended, print the state and reset the state.

```shell
2024/07/28 15:54:31 State: 2
```

If we want to persist the state, we can use other `CleanUp` functions.

```go
package main

import (
	"context"
	"github.com/snowmerak/scope"
	"github.com/snowmerak/scope/prac/riddle"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := riddle.NewState()

	scope.Sequence(ctx, r, func(err error) {
		log.Printf("sequence error: %v", err)
	}, riddle.JustPrint, riddle.AddOne, riddle.AddTwo, riddle.SubOne)

	scope.Sequence(ctx, r, func(err error) {
		log.Printf("sequence error: %v", err)
	}, riddle.JustPrint, riddle.AddOne, riddle.AddTwo, riddle.SubOne)

	scope.Sequence(ctx, r, func(err error) {
		log.Printf("sequence error: %v", err)
	}, riddle.JustPrint, riddle.AddOne, riddle.AddTwo, riddle.SubOne)
}
```

We replace the `ResetAndPrint` function with the `JustPrint` function.

```shell
State: 2
State: 4
State: 6
```

The state is persisted.
