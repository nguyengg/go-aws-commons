package args

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Scanner produces lines of text from command-line positional arguments and additional sources.
type Scanner struct {
	// Filter filters and transforms which lines of text are kept and returned by Scan.
	//
	// By default, the nil value means all lines are returned as-is without any transformation.
	Filter func(string) (string, bool)

	// NoStdin if true will cause Scan to never read from os.Stdin.
	NoStdin bool
}

// Scan starts scanning for text lines from a string slice, files, or from os.Stdin.
//
// The use case for this method comes from passing command-line positional arguments; if "--" exists as one of the
// arguments, the scanner will start reading from os.Stdin after exhausting args and files.
//
// Similarly, you can also pass a list of files whose content will be parsed as newline-delimited text. Error opening
// a file will not automatically stop the iterator; it is up to consumer of the iterator to stop or not, but the file
// will be skipped in case of error.
//
// If both args and files are empty, Scan will automatically scan from os.Stdin unless Scanner.NoStdin is true.
func (s *Scanner) Scan(args []string, files []string) func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {
		readFromStdin := len(args) == 0 && len(files) == 0

		filter := s.Filter
		if filter == nil {
			filter = defaultFilter
		}

		// args.
		for _, arg := range args {
			switch {
			case arg == "--":
				readFromStdin = true
			case !yield(arg, nil):
				return
			}
		}

		// files.
		for _, name := range files {
			f, err := os.Open(name)
			if err != nil {
				if !yield("", fmt.Errorf(`open file "%s" error: %w`, name, err)) {
					return
				}

				continue
			}

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				switch v, ok := s.Filter(scanner.Text()); {
				case !ok:
				case !yield(v, nil):
					return
				}
			}

			if err, _ = scanner.Err(), f.Close(); err != nil {
				if !yield("", fmt.Errorf(`read file "%s" error: %w`, name, err)) {
					return
				}
			}
		}

		// os.Stdin.
		if !s.NoStdin && readFromStdin {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				switch v, ok := s.Filter(scanner.Text()); {
				case !ok:
				case !yield(v, nil):
					return
				}
			}

			if err := scanner.Err(); err != nil {
				if !yield("", fmt.Errorf("read from stdin error: %w", err)) {
					return
				}
			}
		}

	}
}

// Scan starts scanning for text lines from a string slice, files, or from os.Stdin using a custom Scanner.
//
// This method will supply a sensible Scanner.Filter that trims the returned lines while also skipping empty lines as
// well as lines starting with "#".
//
// See Scanner.Scan for more information.
func Scan(args []string, files []string, optFns ...func(*Scanner)) func(yield func(string, error) bool) {
	s := &Scanner{Filter: defaultFilter}
	for _, fn := range optFns {
		fn(s)
	}

	return s.Scan(args, files)
}

// defaultFilter is the default Scanner.Filter.
func defaultFilter(text string) (string, bool) {
	text = strings.TrimSpace(text)
	return text, text != "" && !strings.HasPrefix(text, "#")
}
