package executor

import (
	"errors"
	"sync"
)

// Executor is inspired by Java Executor that abstracts submitting a task and executing it.
type Executor interface {
	// Execute executes the given command.
	//
	// The return error comes from the executor's attempt to run the function, not from the function itself. If the
	// function panics, the poller can recover and continue polling unless the executor has been closed.
	Execute(func()) error

	// Close signals the goroutines in this executor to stop.
	Close() error

	// Wait waits for all goroutines to terminate.
	//
	// A common usage pattern is to Close then Wait to make sure all goroutines have a chance to finish its work
	// and terminate gracefully.
	Wait()
}

// ErrClosed is returned by Executor.Execute if Executor.Close has already been called.
var ErrClosed = errors.New("executor already closed")

// NewCallerRunsOnFullExecutor returns a new executor that will execute the command on same goroutine as caller if a
// new pool of n goroutines is full.
//
// Passing n == 0 effectively returns an always-caller-run executor. Panics if n < 0.
func NewCallerRunsOnFullExecutor(n int) Executor {
	if n == 0 {
		return &callerRunsExecutor{}
	}

	ex := &callerRunsOnFullExecutor{
		inputs: make(chan func(), n),
		done:   make(chan struct{}),
	}

	for range n {
		ex.wg.Go(ex.poll)
	}

	return ex
}

type callerRunsOnFullExecutor struct {
	inputs chan func()
	done   chan struct{}
	wg     sync.WaitGroup
}

func (ex *callerRunsOnFullExecutor) poll() {
	ex.wg.Go(func() {
		defer func() {
			_ = recover()
		}()

		for fn := range ex.inputs {
			fn()
		}
	})
}

func (ex *callerRunsOnFullExecutor) Execute(f func()) (err error) {
	select {
	case <-ex.done:
		return ErrClosed
	case ex.inputs <- f:
		return nil
	default:
		f()
		return nil
	}
}

func (ex *callerRunsOnFullExecutor) Wait() {
	ex.wg.Wait()
}

func (ex *callerRunsOnFullExecutor) Close() error {
	select {
	case <-ex.done:
		return ErrClosed
	default:
		close(ex.done)
		close(ex.inputs)
		return nil
	}
}

type callerRunsExecutor struct {
	closed bool
}

func (ex *callerRunsExecutor) Execute(f func()) error {
	if ex.closed {
		return ErrClosed
	}

	f()
	return nil
}

func (ex *callerRunsExecutor) Wait() {
}

func (ex *callerRunsExecutor) Close() error {
	if ex.closed {
		return ErrClosed
	}

	ex.closed = true
	return nil
}
