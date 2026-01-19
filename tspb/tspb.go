package tspb

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

		options: append(defaultBytesOptions, progressbar.OptionSetDescription(desc)),
	}

	b.WithMessagePrefix(strings.TrimSuffix(desc, " ") + " ")

	for _, fn := range optFns {
		fn(b)
	}

	return b.Build()
}

// DefaultBytesWriter is a variant of DefaultBytes that is given an io.Writer instead.
//
// If w is an os.File, its name and size will be used to generate a sensible progressbar's description or log message
// prefix. The desc argument can include `{name}` or `{basename}` which will be replaced with the actual file's name
// ([os.File.Name]) or basename accordingly.
func DefaultBytesWriter(w io.Writer, desc string, optFns ...func(builder *Builder)) io.WriteCloser {
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
func DefaultBytesReader(r io.Reader, desc string, optFns ...func(*Builder)) io.WriteCloser {
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
func DefaultCounter(desc string, optFns ...func(*Builder)) io.WriteCloser {
	b := &Builder{
		Size: -1,
		Rate: &rate.Sometimes{Interval: 5 * time.Second},

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
