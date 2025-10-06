package tspb

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"golang.org/x/time/rate"
)

// ProgressLogger either uses progressbar.ProgressBar if os.Stderr is a terminal, or fallbacks to logging otherwise.
//
// The zero-value ProgressLogger is ready for use and will detect whether to use progressbar.ProgressBar or log.Default
// on the first Write. In order to move this detection logic earlier, or to customise the ProgressLogger and/or the
// underlying progressbar, use DefaultBytes instead.
//
// ProgressLogger implements io.Writer to provide progress logging similar to progressbar.ProgressBar. The intended
// usage of this struct is to log progress such as writing "writing x MiB / y GiB (z %) [elapsedTime]" for progress
// reporting with throttling (once every 5 seconds), and when completed successfully log
// "wrote x MiB / y GiB (z %) GiB [elapsedTime]" once.
//
// Close should always be called to apply the completion logic of the progressbar ([progressbar.ProgressBar.Exit]).
// Finish should be called only if process has completed successfully; this will call [progressbar.ProgressBar.Finish]
// to fill the progressbar and trigger finish logic. The fallback logger will only log once if both Finish and Close are
// called.
type ProgressLogger struct {
	// Size is the expected size.
	//
	// If 0 or negative, the expected size is unknown and as a result, the logger will not print the expected size.
	Size int64

	// Rate indicates how often logging is performed.
	//
	// By default, only log every 5 seconds. Specify a zero-value rate.Sometimes to apply no throttling.
	Rate *rate.Sometimes

	// Log is used to log progress only if progressbar.ProgressBar is not used.
	//
	// By default, this uses log.Default to log with messages such as "writing x MiB / y GiB (z %) [elapsedTime]"
	// for every Write, and "wrote y GiB [elapsedTime]" on Close.
	Log func(size, written int64, elapsed time.Duration, done bool)

	// Logger is the instance that the default implementation of Log uses.
	//
	// Default to log.Default.
	Logger *log.Logger

	once, finished sync.Once
	options        []progressbar.Option
	bar            *progressbar.ProgressBar
	written        int64
	start          time.Time
}

var _ io.WriteCloser = &ProgressLogger{}

func (l *ProgressLogger) init() {
	l.once.Do(func() {
		l.start = time.Now()

		if l.bar != nil {
			return
		}

		if l.Size == 0 {
			l.Size = -1
		}

		if term.IsTerminal(int(os.Stderr.Fd())) {
			if len(l.options) != 0 {
				l.bar = progressbar.NewOptions64(l.Size, l.options...)
			} else {
				l.bar = progressbar.NewOptions64(l.Size, defaultBytesOptions...)
			}
			return
		}

		if l.Logger == nil {
			l.Logger = log.Default()
		}

		if l.Log == nil {
			l.Log = l.defaultLogBytes
		}
	})
}

func (l *ProgressLogger) Write(p []byte) (n int, err error) {
	l.init()

	if l.bar != nil {
		return l.bar.Write(p)
	}

	n = len(p)
	l.written += int64(n)

	l.Rate.Do(func() {
		elapsed := time.Now().Sub(l.start).Truncate(time.Second)
		l.Log(l.Size, l.written, elapsed, false)
	})
	return
}

func (l *ProgressLogger) Close() error {
	l.init()

	if l.bar != nil {
		return l.bar.Exit()
	}

	l.finished.Do(func() {
		elapsed := time.Now().Sub(l.start).Truncate(time.Second)
		l.Log(l.Size, l.written, elapsed, true)
	})
	return nil
}

func (l *ProgressLogger) Finish() error {
	l.init()

	if l.bar != nil {
		return l.bar.Finish()
	}

	l.finished.Do(func() {
		elapsed := time.Now().Sub(l.start).Truncate(time.Second)
		l.Log(l.Size, l.written, elapsed, true)
	})
	return nil
}

func (l *ProgressLogger) defaultLogBytes(size, written int64, elapsed time.Duration, done bool) {
	if done {
		if size > 0 && size != written {
			l.Logger.Printf("wrote %s / %s (%.2f%%) [%s]", humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
		} else {
			l.Logger.Printf("wrote %s [%s]", humanize.IBytes(uint64(written)), elapsed)
		}

		return
	}

	if size > 0 {
		l.Logger.Printf("writing %s / %s (%.2f%%) [%s]", humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
	} else {
		l.Logger.Printf("writing %s [%s]", humanize.IBytes(uint64(written)), elapsed)
	}
}

// CreateSimpleLogFunction creates a sensible no-nonsense logging function that is good enough most of the time.
func CreateSimpleLogFunction(logger *log.Logger, prefix string, showBytes bool) func(size, written int64, elapsed time.Duration, done bool) {
	if showBytes {
		return func(size, written int64, elapsed time.Duration, done bool) {
			if size > 0 || (done && size != written) {
				logger.Printf("%s%s / %s (%.2f%%) [%s]", prefix, humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
			} else {
				logger.Printf("%s%s [%s]", prefix, humanize.IBytes(uint64(written)), elapsed)
			}
		}
	}

	return func(size, written int64, elapsed time.Duration, done bool) {
		if size > 0 || (done && size != written) {
			logger.Printf("%s%d / %d (%.2f%%) [%s]", prefix, written, size, 100.0*float64(written)/float64(size), elapsed)
		} else {
			logger.Printf("%s%d [%s]", prefix, written, elapsed)
		}
	}
}
