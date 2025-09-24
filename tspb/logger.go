package tspb

import (
	"io"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/time/rate"
)

// RateLimitedLogger is the replacement of progressbar.ProgressBar if program is not running in terminal.
type RateLimitedLogger struct {
	// Rate indicates how often logging is performed.
	//
	// By default, only log every 10 seconds.
	Rate *rate.Sometimes
	// Logger is the log.Logger instance that is used for logging.
	//
	// Defaults to log.Default.
	Logger *log.Logger

	description string
	size, n     int64
}

func newRateLimitedLogger(description string, size int64) *RateLimitedLogger {
	return &RateLimitedLogger{
		Rate:        &rate.Sometimes{Interval: 10 * time.Second},
		Logger:      log.Default(),
		description: description,
		size:        size,
	}
}

func (l *RateLimitedLogger) add(d int64) {
	l.n += d
	l.Rate.Do(func() {
		if l.size <= 0 {
			l.Logger.Printf("%s: %s", l.description, humanize.IBytes(uint64(l.n)))
		} else {
			l.Logger.Printf("%s: %s/%s (%.2f%%)", l.description, humanize.IBytes(uint64(l.n)), humanize.IBytes(uint64(l.size)), 100.0*float64(l.n)/float64(l.size))
		}
	})
}

// implements io.ReadCloser.
var _ io.ReadCloser = &RateLimitedLogger{}

func (l *RateLimitedLogger) Read(p []byte) (n int, err error) {
	n = len(p)
	l.add(int64(n))
	return
}

func (l *RateLimitedLogger) Close() error {
	return nil
}

// implements io.Writer.
var _ io.Writer = &RateLimitedLogger{}

func (l *RateLimitedLogger) Write(p []byte) (n int, err error) {
	n = len(p)
	l.add(int64(n))
	return
}
