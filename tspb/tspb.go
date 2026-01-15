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

// DefaultBytes is a convenient method for building io.WriteCloser loggers without explicitly using NewBuilder.
func DefaultBytes(size int64, desc string, optFns ...func(*Builder)) io.WriteCloser {
	b := &Builder{
		Size: size,
		Rate: &rate.Sometimes{Interval: 5 * time.Second},

		options: defaultBytesOptions,
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
func DefaultCounter(n int, desc string, optFns ...func(*Builder)) io.WriteCloser {
	b := &Builder{
		Size: int64(n),
		Rate: &rate.Sometimes{Interval: 5 * time.Second},

		options: defaultCounterOptions,
	}
	for _, fn := range optFns {
		fn(b)
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		b.Size = int64(n)
		b.options = append([]progressbar.Option{progressbar.OptionSetDescription(desc)}, b.options...)
	} else if b.LogFn == nil {
		if desc != "" {
			b.LogFn = CreateSimpleLogFunction(log.Default(), desc+" ", "", false)
		} else {
			b.LogFn = CreateSimpleLogFunction(log.Default(), "", "", false)
		}
	}

	return b.Build()
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
