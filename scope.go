package scope

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

type (
	Work[T any]         func(ctx context.Context, state *T) error
	Checker             func(error)
	CleanUp[T any]      func(ctx context.Context, state *T)
	Scope[T any]        func(ctx context.Context, state *T, checker Checker, cleanup CleanUp[T], works ...Work[T])
	WithCancel[T any]   func(ctx context.Context, state *T, checker Checker, cleanup CleanUp[T], works ...Work[T])
	WithTimeout[T any]  func(ctx context.Context, state *T, duration time.Duration, checker Checker, cleanup CleanUp[T], works ...Work[T])
	WithDeadline[T any] func(ctx context.Context, state *T, deadline time.Time, checker Checker, cleanup CleanUp[T], works ...Work[T])
)

// sequence runs the given functions in sequence.
// If any of the functions returns an error, the error is returned.
// If any of the functions panics, the panic is caught and returned as an error.
// If all functions succeed, nil is returned.
func sequence[T any](ctx context.Context, state *T, f ...Work[T]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("scope recovered: %v", r)
		}
	}()

	for _, w := range f {
		if err = w(ctx, state); err != nil {
			return err
		}
	}

	return nil
}

// parallel runs the given functions in parallel.
// If all functions return an error, the error is returned.
// If any of the functions panics, the panic is caught and returned as an error.
// If all functions succeed, nil is returned.
func parallel[T any](ctx context.Context, state *T, f ...Work[T]) (success int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("scope recovered: %v", r)
		}
	}()

	errs := make([]error, len(f))
	for i, w := range f {
		go func(i int, w Work[T]) {
			defer func() {
				if r := recover(); r != nil {
					errs[i] = fmt.Errorf("scope recovered: %v", r)
				}
			}()

			if err := w(ctx, state); err != nil {
				errs[i] = err
			}
		}(i, w)
	}

	nilIdx := make([]int, 4)
	for i, err := range errs {
		if err == nil {
			nilIdx = append(nilIdx, i)
		}
	}

	slices.Reverse(nilIdx)
	success = len(nilIdx)

	for _, i := range nilIdx {
		errs = append(errs[:i], errs[i+1:]...)
	}

	return success, errors.Join(errs...)
}

// Sequence runs the given functions in sequence.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func Sequence[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	if err := sequence[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ Scope[int] = Sequence[int]

// SequenceWithCancel runs the given functions in sequence with a cancel function.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func SequenceWithCancel[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := sequence[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithCancel[int] = SequenceWithCancel[int]

// SequenceWithTimeout runs the given functions in sequence with a timeout.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func SequenceWithTimeout[T any](ctx context.Context, state *T, timeout time.Duration, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := sequence[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithTimeout[int] = SequenceWithTimeout[int]

// SequenceWithDeadline runs the given functions in sequence with a deadline.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func SequenceWithDeadline[T any](ctx context.Context, state *T, deadline time.Time, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	if err := sequence[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithDeadline[int] = SequenceWithDeadline[int]

// All runs the given functions in parallel.
// If all functions return an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func All[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	if _, err := parallel[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ Scope[int] = All[int]

// AllWithCancel runs the given functions in parallel with a cancel function.
// If all functions return an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AllWithCancel[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if _, err := parallel[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithCancel[int] = AllWithCancel[int]

// AllWithTimeout runs the given functions in parallel with a timeout.
// If all functions return an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AllWithTimeout[T any](ctx context.Context, state *T, timeout time.Duration, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, err := parallel[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithTimeout[int] = AllWithTimeout[int]

// AllWithDeadline runs the given functions in parallel with a deadline.
// If all functions return an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AllWithDeadline[T any](ctx context.Context, state *T, deadline time.Time, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	if _, err := parallel[T](ctx, state, f...); err != nil {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithDeadline[int] = AllWithDeadline[int]

// Any runs the given functions in parallel.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func Any[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	success, err := parallel[T](ctx, state, f...)
	if success == 0 {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ Scope[int] = Any[int]

// AnyWithCancel runs the given functions in parallel with a cancel function.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AnyWithCancel[T any](ctx context.Context, state *T, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	success, err := parallel[T](ctx, state, f...)
	if success == 0 {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithCancel[int] = AnyWithCancel[int]

// AnyWithTimeout runs the given functions in parallel with a timeout.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AnyWithTimeout[T any](ctx context.Context, state *T, timeout time.Duration, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	success, err := parallel[T](ctx, state, f...)
	if success == 0 {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithTimeout[int] = AnyWithTimeout[int]

// AnyWithDeadline runs the given functions in parallel with a deadline.
// If any of the functions returns an error, the error is passed to the errChecker.
// If cleanUp is not nil, it is called after all functions have been executed.
func AnyWithDeadline[T any](ctx context.Context, state *T, deadline time.Time, errChecker Checker, cleanUp CleanUp[T], f ...Work[T]) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	success, err := parallel[T](ctx, state, f...)
	if success == 0 {
		errChecker(err)
	}

	if cleanUp != nil {
		cleanUp(ctx, state)
	}
}

var _ WithDeadline[int] = AnyWithDeadline[int]
