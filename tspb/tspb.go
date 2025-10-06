package tspb

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"golang.org/x/time/rate"
)

// New creates and customises a ProgressLogger.
//
// The size and desc arguments are passed to the underlying progressbar.ProgressBar as-is if progressbar.ProgressBar is
// used. Otherwise, desc will be used as a prefix for log messages such as "desc x MiB / y GiB (z %) [elapsedTime]".
//
// If you need to customise the progressbar further, pass WithProgressBarOptions.
func New(size int64, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
	logger := &ProgressLogger{
		Size:   size,
		Rate:   &rate.Sometimes{Interval: 5 * time.Second},
		Log:    createLogFunction(log.Default(), desc),
		Logger: log.Default(),

		options: defaultBytesOptions,
	}
	for _, fn := range optFns {
		fn(logger)
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		logger.bar = progressbar.NewOptions64(size, append([]progressbar.Option{progressbar.OptionSetDescription(desc)}, logger.options...)...)
	}

	return logger
}

// NewWithWriter is a variant of New that is given an io.Writer instead.
//
// If w is an os.File, its name and size will be used to generate a sensible progressbar's description or log message
// prefix. The desc argument can include `{name}` or `{basename}` which will be replaced with the actual file's name
// ([os.File.Name]) or basename accordingly.
func NewWithWriter(w io.Writer, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
	var size int64 = -1

	if f, ok := w.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		if desc != "" {
			desc = strings.ReplaceAll(desc, "{name}", name)
			desc = strings.ReplaceAll(desc, "{basename}", filepath.Base(name))
		} else {
			desc = fmt.Sprintf(`writing to "%s"`, filepath.Base(name))
		}
	} else if desc == "" {
		desc = "writing"
	}

	return New(size, desc, optFns...)
}

// NewWithReader is a variant of NewWithWriter that is given an io.Reader instead.
//
// If r is an os.File, its name and size will be used to generate a sensible progressbar's description or log message
// prefix. The desc argument can include `{name}` or `{basename}` which will be replaced with the actual file's name
// ([os.File.Name]) or basename accordingly.
func NewWithReader(r io.Reader, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
	var size int64 = -1

	if f, ok := r.(*os.File); ok {
		name := f.Name()
		if fi, err := f.Stat(); err == nil {
			size = fi.Size()
		}

		if desc != "" {
			desc = strings.ReplaceAll(desc, "{name}", name)
			desc = strings.ReplaceAll(desc, "{basename}", filepath.Base(name))
		} else {
			desc = fmt.Sprintf(`reading from "%s"`, filepath.Base(name))
		}
	} else if desc == "" {
		desc = "reading"
	}

	return New(size, desc, optFns...)
}

// WithProgressBarOptions provides a way to customise the progressbar options manually.
func WithProgressBarOptions(options ...progressbar.Option) func(*ProgressLogger) {
	return func(logger *ProgressLogger) {
		logger.options = options
	}
}

// WithDefaultBytesProgressBar sets sensible defaults for the progressbar and logger to display bytes.
func WithDefaultBytesProgressBar() func(*ProgressLogger) {
	return func(logger *ProgressLogger) {
		logger.options = defaultBytesOptions
	}
}

// WithDefaultCounterProgressBar sets sensible defaults for the progressbar and logger to display counter.
func WithDefaultCounterProgressBar() func(*ProgressLogger) {
	return func(logger *ProgressLogger) {
		logger.options = defaultCounterOptions
		logger.Log = func(size, written int64, elapsed time.Duration, done bool) {
			if done {
				if size > 0 && size != written {
					logger.Logger.Printf("%d / %d (%.2f%%) [%s]", written, size, 100.0*float64(written)/float64(size), elapsed)
				} else {
					logger.Logger.Printf("%d [%s]", written, elapsed)
				}
				return
			}

			if size > 0 {
				logger.Logger.Printf("%d / %d (%.2f%%) [%s]", written, size, 100.0*float64(written)/float64(size), elapsed)
			} else {
				logger.Logger.Printf("%d [%s]", written, elapsed)
			}
		}
	}
}

var defaultBytesOptions = []progressbar.Option{
	progressbar.OptionShowBytes(true),
	progressbar.OptionShowTotalBytes(true),
	progressbar.OptionShowCount(),
	progressbar.OptionShowElapsedTimeOnFinish(),
	progressbar.OptionThrottle(1 * time.Second),
	progressbar.OptionOnCompletion(func() {
		_, _ = os.Stderr.WriteString("\n")
	}),
}

var defaultCounterOptions = []progressbar.Option{
	progressbar.OptionShowBytes(false),
	progressbar.OptionShowTotalBytes(false),
	progressbar.OptionShowCount(),
	progressbar.OptionShowElapsedTimeOnFinish(),
	progressbar.OptionThrottle(1 * time.Second),
	progressbar.OptionOnCompletion(func() {
		_, _ = os.Stderr.WriteString("\n")
	}),
}
