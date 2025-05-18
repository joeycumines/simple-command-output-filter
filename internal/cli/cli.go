package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"regexp"
)

type CLI struct {
	Input            io.Reader
	Output           io.Writer
	ErrOut           io.Writer
	flagSet          *flag.FlagSet
	command          string
	errorMode        errorMode
	rawPatterns      stringSliceFlag
	patternFiles     stringSliceFlag
	compiledPatterns []*regexp.Regexp
	args             []string
	invertMatch      bool // like grep -v
}

var errNoCommand = errors.New("no command specified")

func (x *CLI) Main(args []string) int {
	if err := x.init(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		_, _ = fmt.Fprintf(x.ErrOut, "Error initializing: %s\n", err)

		if errors.Is(err, errNoCommand) {
			_, _ = fmt.Fprintln(x.ErrOut)
			x.usage()
		}

		return 2
	}

	if err := x.run(); err != nil {
		if errors.Is(err, errDueToMode) {
			// everything was ok, but we either had, or didn't have content
			// (whichever was the opposite of what we wanted)
			return 1
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code := exitErr.ExitCode()
			if code == -1 {
				// N.B. cbf but might want to return 128 + signal number under posix
				return 1 // signal exit code
			}
			if code <= 0 {
				return 1 // unexpected case
			}
			return code // general case
		}
		_, _ = fmt.Fprintf(x.ErrOut, "Error running command: %s\n", err)
		return 1 // general case
	}

	return 0
}
