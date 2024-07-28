package riddle

import (
	"context"
	"fmt"
	"github.com/snowmerak/scope"
	"log"
	"sync/atomic"
)

type State struct {
	number atomic.Int64
}

func NewState() *State {
	return &State{}
}

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

var _ scope.CleanUp[State] = ResetAndPrint

func ResetAndPrint(ctx context.Context, state *State) {
	log.Printf("State: %d\n", state.number.Load())
	state.number.Store(0)
}

var _ scope.CleanUp[State] = JustPrint

func JustPrint(ctx context.Context, state *State) {
	fmt.Printf("State: %d\n", state.number.Load())
}

var _ scope.Checker[State] = SimpleChecker

func SimpleChecker(err error, state *State) {
	log.Printf("SimpleChecker: %v\n", err)
	state.number.Store(0)
}
