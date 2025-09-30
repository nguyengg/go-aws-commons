package executor

// NewCallerRunsOnFullExecutor returns a new executor that will execute the command on same goroutine as caller if a
// new pool of n goroutines is full.
//
// Passing n == 0 effectively returns an executor that executes the function on the same goroutine as caller.
// Panics if n < 0.
//
// If you want to be able to queue more task then there are available goroutines, use New with WithCapacity.
func NewCallerRunsOnFullExecutor(n int) Executor {
	return New(CallerRunsOnFullPolicy, n)
}

type callerRunsOnFullExecutor struct {
	*baseExecutor
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
