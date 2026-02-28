// Package executor is inspired by Java Executor and ThreadPoolExecutor, especially its RejectedExecutionHandler.
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

// FullBufferPolicy controls the behaviour of the Executor when its task buffer is full.
//
// This is the equivalent of [RejectedExecutionHandler] for Java's [ThreadPoolExecutor].
//
// [RejectedExecutionHandler]: https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/concurrent/RejectedExecutionHandler.html
// [ThreadPoolExecutor]: https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/concurrent/ThreadPoolExecutor.html
type FullBufferPolicy int

const (
	// CallerBlocksOnFullPolicy will cause Executor.Execute to block if the task buffer is full.
	//
	// This is the default behaviour.
	CallerBlocksOnFullPolicy FullBufferPolicy = iota
	// CallerRunsOnFullPolicy will execute the task in the same goroutine that calls Executor.Execute.
	//
	// This is the equivalent of Java's [CallerRunsPolicy].
	//
	// [CallerRunsPolicy]: https://docs.oracle.com/en/java/javase/21/docs/api/java.base/java/util/concurrent/ThreadPoolExecutor.CallerRunsPolicy.html
	CallerRunsOnFullPolicy
	// TODO support DropOnFullPolicy to drop task overflow.
)

// ErrClosed is returned by Executor.Execute if Executor.Close has already been called.
var ErrClosed = errors.New("executor already closed")

// Options allows customisations of the executor returned by New.
type Options struct {
	// capacity is the capacity of inputs channel.
	capacity int
}

// WithCapacity changes the capacity of the executor's task buffer.
//
// With a capacity that is higher than the number of goroutines, Executor.Execute would take longer to start kicking in
// its full-buffer policy (either CallerRunsOnFullPolicy or CallerBlocksOnFullPolicy). Useful if you have a list of
// tasks that is larger than the number of CPUs.
//
// Panics if capacity < 0.
func WithCapacity(capacity int) func(*Options) {
	return func(opts *Options) {
		opts.capacity = capacity
	}
}

// New creates a new Executor with the given full-buffer policy and pool size.
//
// Passing poolSize == 0 effectively returns an executor that executes the function on the same goroutine as caller.
// Panics if poolSize < 0.
//
// By default, the input buffer capacity is the same as the pool size. Use WithCapacity to change this.
func New(policy FullBufferPolicy, poolSize int, optFns ...func(*Options)) (ex Executor) {
	if poolSize == 0 {
		return &callerAlwaysRunsExecutor{}
	}

	opts := &Options{capacity: poolSize}
	for _, fn := range optFns {
		fn(opts)
	}

	base := &baseExecutor{
		inputs: make(chan func(), opts.capacity),
		done:   make(chan struct{}),
	}

	switch policy {
	case CallerRunsOnFullPolicy:
		ex = &callerRunsOnFullExecutor{base}
	default:
		ex = &callerBlocksOnFullExecutor{base}
	}

	for range poolSize {
		base.wg.Go(base.poll)
	}

	return ex
}

type baseExecutor struct {
	inputs chan func()
	done   chan struct{}
	wg     sync.WaitGroup
}

func (ex *baseExecutor) poll() {
	ex.wg.Go(func() {
		defer func() {
			_ = recover()
		}()

		for fn := range ex.inputs {
			fn()
		}
	})
}

func (ex *baseExecutor) Wait() {
	ex.wg.Wait()
}

func (ex *baseExecutor) Close() error {
	select {
	case <-ex.done:
		return ErrClosed
	default:
		close(ex.done)
		close(ex.inputs)
		return nil
	}
}
