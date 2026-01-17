package tspb

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/time/rate"
)

// ProgressLogger always logs using the provided log.Logger instance.
//
// The zero-value ProgressLogger is ready for use.
//
// To customise the message, use NewBuilder.
type ProgressLogger struct {
	// Size is the expected size.
	//
	// If 0 or negative, the expected size is unknown and as a result, the logger will not print the expected size.
	Size int64

	// Rate indicates how often logging is performed.
	//
	// By default, only log every 5 seconds. Specify a zero-value rate.Sometimes to apply no throttling.
	Rate *rate.Sometimes

	// LogFn is used to log progress.
	LogFn LogFunction

	once, finished sync.Once
	written        int64
	start          time.Time
}

var _ io.WriteCloser = &ProgressLogger{}

func (l *ProgressLogger) init() {
	l.once.Do(func() {
		l.start = time.Now()

		if l.Rate == nil {
			l.Rate = &rate.Sometimes{Interval: 5 * time.Second}
		}

		if l.LogFn == nil {
			l.LogFn = CreateSimpleLogFunction(log.Default(), "writing ", "wrote ", true)
		}
	})
}

func (l *ProgressLogger) Write(p []byte) (n int, err error) {
	l.init()

	n = len(p)
	l.written += int64(n)

	l.Rate.Do(func() {
		elapsed := time.Since(l.start).Truncate(time.Second)
		l.LogFn(l.Size, l.written, elapsed, false)
	})
	return
}

// Close should always be called as a deferred function.
//
// If not in terminal, print a final log message showing current progress. If progressbar is used,
// [progressbar.ProgressBar.Exit] will be called to force completion hook to run without completing the progressbar
// unlike Finish.
func (l *ProgressLogger) Close() (err error) {
	l.init()

	l.finished.Do(func() {
		elapsed := time.Since(l.start).Truncate(time.Second)
		l.LogFn(l.Size, l.written, elapsed, true)
	})
	return
}

// CreateSimpleLogFunction creates a sensible no-nonsense logging function that is good enough most of the time.
func CreateSimpleLogFunction(logger *log.Logger, prefix, donePrefix string, showBytes bool) func(size, written int64, elapsed time.Duration, done bool) {
	if showBytes {
		if donePrefix != "" {
			return func(size, written int64, elapsed time.Duration, done bool) {
				if done {
					if size > 0 && size != written {
						logger.Printf("%s%s / %s (%.2f%%) [%s]", donePrefix, humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
					} else {
						logger.Printf("%s%s [%s]", donePrefix, humanize.IBytes(uint64(written)), elapsed)
					}
				} else if size > 0 {
					logger.Printf("%s%s / %s (%.2f%%) [%s]", prefix, humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
				} else {
					logger.Printf("%s%s [%s]", prefix, humanize.IBytes(uint64(written)), elapsed)
				}
			}
		}

		return func(size, written int64, elapsed time.Duration, done bool) {
			if size > 0 {
				logger.Printf("%s%s / %s (%.2f%%) [%s]", prefix, humanize.IBytes(uint64(written)), humanize.IBytes(uint64(size)), 100.0*float64(written)/float64(size), elapsed)
			} else {
				logger.Printf("%s%s [%s]", prefix, humanize.IBytes(uint64(written)), elapsed)
			}
		}
	}

	if donePrefix != "" {
		return func(size, written int64, elapsed time.Duration, done bool) {
			if done {
				if size > 0 {
					logger.Printf("%s%d / %d (%.2f%%) [%s]", donePrefix, written, size, 100.0*float64(written)/float64(size), elapsed)
				} else {
					logger.Printf("%s%d [%s]", donePrefix, written, elapsed)
				}
			} else if size > 0 {
				logger.Printf("%s%d / %d (%.2f%%) [%s]", prefix, written, size, 100.0*float64(written)/float64(size), elapsed)
			} else {
				logger.Printf("%s%d [%s]", prefix, written, elapsed)
			}
		}
	}

	return func(size, written int64, elapsed time.Duration, done bool) {
		if size > 0 {
			logger.Printf("%s%d / %d (%.2f%%) [%s]", prefix, written, size, 100.0*float64(written)/float64(size), elapsed)
		} else {
			logger.Printf("%s%d [%s]", prefix, written, elapsed)
		}
	}
}
