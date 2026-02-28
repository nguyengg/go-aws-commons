// Package tspb provides progress logger that become a progress bar in interactive mode (terminal present), or logs
// at interval otherwise.
package tspb

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/time/rate"
)

// DefaultBytes is a convenient method for building io.WriteCloser loggers without explicitly using NewBuilder.
func DefaultBytes(size int64, desc string, optFns ...func(*Builder)) io.WriteCloser {
	b := &Builder{
		Size: size,
		Rate: &rate.Sometimes{Interval: 5 * time.Second},

		prefix:  strings.TrimSuffix(desc, " ") + " ",
		options: append(defaultBytesOptions, progressbar.OptionSetDescription(desc)),
	}
	for _, fn := range optFns {
		fn(b)
	}

	return b.Build()
}

// DefaultCounter is a variant of DefaultBytes that sets up the progressbar and logger for a counter instead.
func DefaultCounter(desc string, optFns ...func(*Builder)) io.WriteCloser {
	b := &Builder{
		Size: -1,
		Rate: &rate.Sometimes{Interval: 5 * time.Second},

		prefix:  strings.TrimSuffix(desc, " ") + " ",
		options: append(defaultCounterOptions, progressbar.OptionSetDescription(desc)),
	}
	for _, fn := range optFns {
		fn(b)
	}

	return b.Build()
}

var defaultBytesOptions = []progressbar.Option{
	// matching progressbar.DefaultBytes.
	progressbar.OptionSetWriter(os.Stderr),
	progressbar.OptionShowBytes(true),
	progressbar.OptionShowTotalBytes(true),
	progressbar.OptionSetWidth(10),
	progressbar.OptionThrottle(1 * time.Second), // 65ms is too short imo.
	progressbar.OptionShowCount(),
	progressbar.OptionOnCompletion(func() {
		_, _ = fmt.Fprint(os.Stderr, "\n")
	}),
	progressbar.OptionSpinnerType(14),
	progressbar.OptionFullWidth(),
	progressbar.OptionSetRenderBlankState(true),
	// my own additions.
	progressbar.OptionUseIECUnits(true),
	progressbar.OptionSetElapsedTime(true),
	progressbar.OptionSetPredictTime(true),
	progressbar.OptionShowElapsedTimeOnFinish(),
}

var defaultCounterOptions = []progressbar.Option{
	progressbar.OptionSetWriter(os.Stderr),
	progressbar.OptionShowBytes(false),
	progressbar.OptionShowTotalBytes(false),
	progressbar.OptionSetWidth(10),
	progressbar.OptionThrottle(1 * time.Second),
	progressbar.OptionShowCount(),
	progressbar.OptionOnCompletion(func() {
		_, _ = fmt.Fprint(os.Stderr, "\n")
	}),
}
