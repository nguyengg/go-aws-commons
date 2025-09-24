package tspb

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// defaultByteOptions uses the same options as progressbar.DefaultBytes.
func defaultByteOptions(desc string) []progressbar.Option {
	return []progressbar.Option{
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(1 * time.Second),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	}
}

// DefaultBytes is the terminal-safe equivalent of progressbar.DefaultBytes.
func DefaultBytes(size int64, desc string, options ...progressbar.Option) io.WriteCloser {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return progressbar.NewOptions64(size, append(defaultByteOptions(desc), options...)...)
	}

	return newRateLimitedLogger(desc, size)
}
