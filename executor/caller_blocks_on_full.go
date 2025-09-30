package executor

// NewCallerBlocksOnFullExecutor returns a new executor that will block if there is no immediately available goroutine
// to pick up the work.
//
// Passing n == 0 effectively returns an executor that executes the function on the same goroutine as caller.
// Panics if n < 0.
//
// If you want to be able to queue more task then there are available goroutines, use New with WithCapacity.
func NewCallerBlocksOnFullExecutor(n int) Executor {
	return New(CallerBlocksOnFullPolicy, n)
}

type callerBlocksOnFullExecutor struct {
	*baseExecutor
}

func (ex *callerBlocksOnFullExecutor) Execute(f func()) (err error) {
	select {
	case <-ex.done:
		return ErrClosed
	case ex.inputs <- f:
		return nil
	}
}
