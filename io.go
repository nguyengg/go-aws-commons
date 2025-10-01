package commons

import (
	"context"
	"io"
)

// CopyBufferWithContext is a custom implementation of io.CopyBuffer that is cancellable via context.
//
// Similar to io.CopyBuffer, if buf is nil, a new buffer of size 32*1024 is created.
// Unlike io.CopyBuffer, it does not matter if src implements [io.WriterTo] or dst implements [io.ReaderFrom] because
// those interfaces do not support context.
//
// The context is checked for done status after every write. As a result, having too small a buffer may introduce too
// much overhead, while having a very large buffer may cause context cancellation to have a delayed effect.
func CopyBufferWithContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

	var nr, nw int
	var read int64
	for {
		nr, err = src.Read(buf)

		if nr > 0 {
			switch nw, err = dst.Write(buf[0:nr]); {
			case err != nil:
				return written, err
			case nr < nw:
				return written, io.ErrShortWrite
			}

			select {
			case <-ctx.Done():
				return written, ctx.Err()
			default:
				read += int64(nr)
				written += int64(nw)
			}
		}

		if err == io.EOF {
			return written, nil
		}
		if err != nil {
			return written, err
		}
	}
}

// NewContextReader wraps the given io.Reader so that if the context is cancelled, [io.Reader.Read] always returns the
// error from the context.
func NewContextReader(ctx context.Context, r io.Reader) io.Reader {
	return &ctxReader{ctx, r}
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		return c.r.Read(p)
	}
}

// NewContextWriter wraps the given io.Writer so that if the context is cancelled, [io.Writer.Write] always returns the
// // error from the context.
func NewContextWriter(ctx context.Context, w io.Writer) io.Writer {
	return &ctxWriter{ctx, w}
}

type ctxWriter struct {
	ctx context.Context
	w   io.Writer
}

func (c *ctxWriter) Write(p []byte) (int, error) {
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		return c.w.Write(p)
	}
}

// Sizer implements io.Writer that tallies that number of bytes written.
type Sizer struct {
	// Size is the total number of bytes that have been written to this io.Writer.
	Size int64
}

func (s *Sizer) Write(p []byte) (n int, err error) {
	n = len(p)
	s.Size += int64(n)
	return
}
