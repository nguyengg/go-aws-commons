package executor

type callerAlwaysRunsExecutor struct {
	closed bool
}

func (ex *callerAlwaysRunsExecutor) Execute(f func()) error {
	if ex.closed {
		return ErrClosed
	}

	f()
	return nil
}

func (ex *callerAlwaysRunsExecutor) Wait() {
}

func (ex *callerAlwaysRunsExecutor) Close() error {
	if ex.closed {
		return ErrClosed
	}

	ex.closed = true
	return nil
}
