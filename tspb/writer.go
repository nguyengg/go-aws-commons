package tspb

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"golang.org/x/time/rate"
)

// Builder is used to create an io.WriteCloser-compatible progressbar.ProgressBar or ProgressLogger depending on whether
// os.Stderr is a terminal.
type Builder struct {
	// Size is the expected size.
	//
	// If 0 or negative, the expected size is unknown and as a result, the logger will not print the expected size.
	Size int64

	// Rate indicates how often logging is performed.
	//
	// By default, only log every 5 seconds. Specify a zero-value rate.Sometimes to apply no throttling.
	Rate *rate.Sometimes

	// LogFn is used to log progress only if progressbar.ProgressBar is not used.
	LogFn LogFunction

	prefix, donePrefix string
	options            []progressbar.Option
}

// LogFunction controls how progress is logged.
type LogFunction func(size, written int64, elapsed time.Duration, done bool)

// NewBuilder provides a fluent interface to building the io.WriteCloser-compatible progress logger.
func NewBuilder() *Builder {
	return &Builder{
		prefix:     "writing ",
		donePrefix: "wrote ",
	}
}

// WithProgressBarOptions provides a way to customise the progressbar options manually.
//
// This will replace existing options from previous invocations.
func (b *Builder) WithProgressBarOptions(options ...progressbar.Option) *Builder {
	b.options = options
	return b
}

// WithMessagePrefix customises the Log function using the given prefixes.
//
// The prefix argument will create log messages such as "{prefix}x MiB / y GiB (z %) [elapsed]". If donePrefix is
// given, only its first argument will be used to log done messages such as "{donePrefix}x MiB / y GiB (z %) [elapsed]".
func (b *Builder) WithMessagePrefix(prefix string, donePrefix ...string) *Builder {
	b.prefix = prefix

	if len(donePrefix) != 0 {
		b.donePrefix = donePrefix[0]
	} else {
		b.donePrefix = ""
	}

	return b
}

// Build will return either a progressbar.ProgressBar or ProgressLogger depending on whether os.Stderr is a terminal.
func (b *Builder) Build() io.WriteCloser {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		return b.BuildProgressBar()
	}

	return b.BuildProgressLogger()
}

// BuildProgressLogger explicitly creates a ProgressLogger.
func (b *Builder) BuildProgressLogger() *ProgressLogger {
	logFn := b.LogFn
	if logFn == nil {
		logFn = CreateSimpleLogFunction(log.Default(), b.prefix, b.donePrefix, true)
	}

	return &ProgressLogger{
		Size:  b.Size,
		Rate:  b.Rate,
		LogFn: b.LogFn,
	}
}

// BuildProgressBar explicitly creates a progressbar.ProgressBar.
func (b *Builder) BuildProgressBar() *progressbar.ProgressBar {
	if len(b.options) != 0 {
		return progressbar.NewOptions64(b.Size, b.options...)
	}

	return progressbar.NewOptions64(b.Size, defaultBytesOptions...)
}
