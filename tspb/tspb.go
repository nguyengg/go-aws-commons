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

// DefaultBytes creates and customises a ProgressLogger for reading/writing bytes progress.
//
// The size and desc arguments are passed to the underlying progressbar.ProgressBar as-is if progressbar.ProgressBar is
// used. Otherwise, desc will be used as a prefix for log messages such as "desc x MiB / y GiB (z %) [elapsedTime]".
//
// If you need to customise the progressbar further, pass WithProgressBarOptions.
func DefaultBytes(size int64, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
	logger := &ProgressLogger{
		Size:   size,
		Rate:   &rate.Sometimes{Interval: 5 * time.Second},
		Logger: log.Default(),

		options: defaultBytesOptions,
	}
	for _, fn := range optFns {
		fn(logger)
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		logger.bar = progressbar.NewOptions64(size, append([]progressbar.Option{progressbar.OptionSetDescription(desc)}, logger.options...)...)
	} else if logger.Log == nil {
		if desc != "" {
			logger.Log = CreateSimpleLogFunction(logger.Logger, desc+" ", true)
		} else {
			logger.Log = logger.defaultLogBytes
		}
	}

	return logger
}

// DefaultBytesWriter is a variant of DefaultBytes that is given an io.Writer instead.
//
// If w is an os.File, its name and size will be used to generate a sensible progressbar's description or log message
// prefix. The desc argument can include `{name}` or `{basename}` which will be replaced with the actual file's name
// ([os.File.Name]) or basename accordingly.
func DefaultBytesWriter(w io.Writer, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
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

	return DefaultBytes(size, desc, optFns...)
}

// DefaultBytesReader is a variant of DefaultBytesWriter that is given an io.Reader instead.
//
// If r is an os.File, its name and size will be used to generate a sensible progressbar's description or log message
// prefix. The desc argument can include `{name}` or `{basename}` which will be replaced with the actual file's name
// ([os.File.Name]) or basename accordingly.
func DefaultBytesReader(r io.Reader, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
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

	return DefaultBytes(size, desc, optFns...)
}

// DefaultCounter is a variant of DefaultBytes that sets up the progressbar and logger for a counter instead.
func DefaultCounter(n int, desc string, optFns ...func(*ProgressLogger)) *ProgressLogger {
	logger := &ProgressLogger{
		Size:   int64(n),
		Rate:   &rate.Sometimes{Interval: 5 * time.Second},
		Logger: log.Default(),

		options: defaultCounterOptions,
	}
	for _, fn := range optFns {
		fn(logger)
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		logger.bar = progressbar.NewOptions(n, append([]progressbar.Option{progressbar.OptionSetDescription(desc)}, logger.options...)...)
	} else if logger.Log == nil {
		if desc != "" {
			logger.Log = CreateSimpleLogFunction(logger.Logger, desc+" ", false)
		} else {
			logger.Log = CreateSimpleLogFunction(logger.Logger, "", false)
		}
	}

	return logger
}

// WithProgressBarOptions provides a way to customise the progressbar options manually.
//
// This will replace existing options from DefaultBytes and DefaultCounter.
func WithProgressBarOptions(options ...progressbar.Option) func(*ProgressLogger) {
	return func(logger *ProgressLogger) {
		logger.options = options
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
