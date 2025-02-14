package s3writer

import (
	"fmt"
)

// executor is inspired by Java Executor that abstracts submitting a task and executing it.
type executor interface {
	// Execute executes the given command.
	//
	// The return error comes from the executor's attempt to run the function, not from the function itself.
	Execute(func()) error

	// Close signals the goroutines in this executor to stop.
	Close() error
}

// newCallerRunsOnFullExecutor returns a new executor that will execute the command on same goroutine as caller if a
// new pool of n goroutines is full.
//
// Passing n == 0 effectively returns an always-caller-run executor. Panics if n < 0.
func newCallerRunsOnFullExecutor(n int) executor {
	if n == 0 {
		return &callerRunsExecutor{}
	}

	inputs := make(chan func(), n)

	for range n {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// if the channel is closed then exit here.
					if _, ok := <-inputs; !ok {
						return
					}
				}
			}()

			for f := range inputs {
				f()
			}
		}()
	}

	return &callerRunsOnFullExecutor{inputs: inputs}
}

type callerRunsOnFullExecutor struct {
	inputs chan func()
}

func (ex *callerRunsOnFullExecutor) Execute(f func()) (err error) {
	// TODO if there are concurrent calls to Close and Execute, sending to inputs may panic.
	// see if this can be fixed.
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic recovered: %v", r)
			}
		}
	}()

	select {
	case ex.inputs <- f:
	default:
		f()
	}

	return nil
}

func (ex *callerRunsOnFullExecutor) Close() error {
	close(ex.inputs)
	return nil
}

type callerRunsExecutor struct {
}

func (ex callerRunsExecutor) Execute(f func()) error {
	f()
	return nil
}

func (ex callerRunsExecutor) Close() error {
	return nil
}
