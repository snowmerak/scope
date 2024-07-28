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
