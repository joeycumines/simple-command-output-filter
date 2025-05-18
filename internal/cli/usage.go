package cli

import (
	"flag"
	"fmt"
)

const helpText = `simple-command-output-filter - Filter stdout of a command based on patterns.

USAGE:
  simple-command-output-filter [options] [--] command [args...]

DESCRIPTION:
  Executes the specified command and filters its standard output. Lines are
  matched against a set of patterns. By default, only lines matching any
  pattern are printed. The filter is ENTIRELY transparent otherwise: stdin and
  stderr are passed through, signals are forwarded, and the command's exit
  status is preserved.

PATTERNS:
  - Patterns match the entire line from start to end.
  - '*' (asterisk) is a wildcard, matching zero or more characters.
  - '**' (double asterisk) matches a literal asterisk character.
  - All other characters are matched literally.
  - Patterns can be specified via -p/--pattern flags or -f/--pattern-file flags.
  - If multiple patterns are provided, a line is considered a match if it
    matches ANY of the patterns.

PATTERN FILES:
  - Each line in a pattern file is treated as a separate pattern.
  - Empty lines in pattern files are ignored.
  - Pattern file lines are only read until the first '#' character, unless
    doubled ('##'), which is treated as a literal '#'.
  - If a pattern file line contains a comment, any whitespace immediately
    preceding the comment is ignored.

BEHAVIOR WITHOUT PATTERNS:
  If no patterns are provided (e.g., no -p, --pattern, -f, or --pattern-file flags are used):
    - Without -v/--invert-match: no lines will be output from the command's stdout
      (as no lines can match an empty set of patterns).
    - With    -v/--invert-match: all lines will be output from the command's stdout
      (as all lines are considered "non-matching" against an empty set of patterns).

EXIT STATUS AND ERROR MODES (-e, --error-mode):
  Alters exit status based on WRITTEN content, ONLY if the command succeeds.
  If the command fails, its original exit status is used.
  Modes:
    - 'default': (Default) Exit status mirrors the command's status.
    - 'no-content': Exit 1 if no content (command succeeded), else 0.
    - 'on-content': Exit 1 if any content (command succeeded), else 0.

OPTIONS:
`

func (x *CLI) usage() {
	_, _ = fmt.Fprint(x.ErrOut, helpText)
	x.flagSet.PrintDefaults()
}

func (x *CLI) init(args []string) error {
	x.errorMode = errorModeDefault

	x.flagSet = flag.NewFlagSet("simple-command-output-filter", flag.ContinueOnError)

	x.flagSet.Usage = x.usage

	// avoid polluting stdout
	x.flagSet.SetOutput(x.ErrOut)

	x.flagSet.Var(&x.rawPatterns, "p", "Pattern to filter by (can be specified multiple times).")
	x.flagSet.Var(&x.rawPatterns, "pattern", "Alias for -p.")
	x.flagSet.Var(&x.patternFiles, "f", "File containing patterns, one per line (can be specified multiple times).")
	x.flagSet.Var(&x.patternFiles, "pattern-file", "Alias for -f.")
	x.flagSet.BoolVar(&x.invertMatch, "v", false, "Invert match (selects non-matching lines).")
	x.flagSet.BoolVar(&x.invertMatch, "invert-match", false, "Alias for -v.")
	x.flagSet.Var(&x.errorMode, "e", "Error mode: 'default', 'no-content', or 'on-content'.")
	x.flagSet.Var(&x.errorMode, "error-mode", "Alias for -e.")

	if err := x.flagSet.Parse(args); err != nil {
		return err // inclusive of flag.ErrHelp
	}

	cmdArgs := x.flagSet.Args()
	if len(cmdArgs) == 0 {
		return errNoCommand
	}

	x.command = cmdArgs[0]
	x.args = cmdArgs[1:]

	return x.loadAndCompilePatterns()
}
