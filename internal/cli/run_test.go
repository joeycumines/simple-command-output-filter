package cli

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCLI_run(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		args           []string
		patterns       []string
		invertMatch    bool
		expectedOutput string
	}{
		{
			name:           "no patterns",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       nil,
			invertMatch:    false,
			expectedOutput: "", // no patterns means no output when not inverted
		},
		{
			name:           "no patterns inverted",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       nil,
			invertMatch:    true,
			expectedOutput: "hello world\n", // no patterns inverted means all output
		},
		{
			name:           "exact match pattern",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       []string{"hello world"},
			invertMatch:    false,
			expectedOutput: "hello world\n",
		},
		{
			name:           "non-matching pattern",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       []string{"foo bar"},
			invertMatch:    false,
			expectedOutput: "",
		},
		{
			name:           "non-matching pattern inverted",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       []string{"foo bar"},
			invertMatch:    true,
			expectedOutput: "hello world\n",
		},
		{
			name:           "wildcard pattern",
			command:        "echo",
			args:           []string{"hello", "world"},
			patterns:       []string{"hello*"},
			invertMatch:    false,
			expectedOutput: "hello world\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer

			cli := &CLI{
				Input:       strings.NewReader(""),
				Output:      &output,
				ErrOut:      &bytes.Buffer{},
				command:     tc.command,
				args:        tc.args,
				invertMatch: tc.invertMatch,
			}

			if len(tc.patterns) > 0 {
				cli.compiledPatterns = make([]*regexp.Regexp, 0, len(tc.patterns))
				for _, p := range tc.patterns {
					re := compileSinglePattern(p)
					cli.compiledPatterns = append(cli.compiledPatterns, re)
				}
			}

			// run the command with a timeout context to prevent test hanging
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			done := make(chan struct{})
			var runErr error

			go func() {
				runErr = cli.run()
				close(done)
			}()

			select {
			case <-ctx.Done():
				t.Fatalf("Test timed out")
			case <-done:
			}

			if runErr != nil {
				t.Fatalf("run() error = %v", runErr)
			}

			if got := output.String(); got != tc.expectedOutput {
				t.Errorf("Expected output %q, got %q", tc.expectedOutput, got)
			}
		})
	}
}

func TestCLI_run_stdinPassing(t *testing.T) {
	// test that stdin is correctly passed through to the command
	tests := []struct {
		name    string
		input   string
		command string
		args    []string
	}{
		{
			name:    "simple input",
			input:   "test input\n",
			command: "cat",
			args:    []string{},
		},
		{
			name:    "multi-line input",
			input:   "line1\nline2\nline3\n",
			command: "cat",
			args:    []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			// use cat to echo input to output
			cli := &CLI{
				Input:   strings.NewReader(tc.input),
				Output:  &stdout,
				ErrOut:  &stderr,
				command: tc.command,
				args:    tc.args,
				// This is needed for the test to pass. The default without patterns is to output nothing.
				invertMatch: true,
			}

			err := cli.run()
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}

			if got := stdout.String(); got != tc.input {
				t.Errorf("Expected output %q, got %q", tc.input, got)
			}
		})
	}
}

func TestCLI_stderrPassing(t *testing.T) {
	// test that stderr output is passed through unmodified
	var stdout, stderr bytes.Buffer

	cli := &CLI{
		Input:   strings.NewReader(""),
		Output:  &stdout,
		ErrOut:  &stderr,
		command: "bash",
		args:    []string{"-c", "echo 'standard output'; echo 'standard error' >&2"},
	}

	// only match "nothing" so we don't output anything to stdout
	re, _ := regexp.Compile("nothing-matches-this")
	cli.compiledPatterns = []*regexp.Regexp{re}

	err := cli.run()
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// stdout should be empty since our pattern doesn't match
	if got := stdout.String(); got != "" {
		t.Errorf("Expected empty stdout, got %q", got)
	}

	// stderr should contain the error message
	expectedStderr := "standard error\n"
	if got := stderr.String(); got != expectedStderr {
		t.Errorf("Expected stderr %q, got %q", expectedStderr, got)
	}
}

func TestCLI_multipleLines(t *testing.T) {
	// test with a command that outputs multiple lines
	var stdout bytes.Buffer

	cli := &CLI{
		Input:   strings.NewReader(""),
		Output:  &stdout,
		ErrOut:  &bytes.Buffer{},
		command: "bash",
		args:    []string{"-c", "echo -e 'line1\nline2\nline3\nline4'"},
	}

	// only match lines with odd numbers
	re, _ := regexp.Compile(".*[13]$")
	cli.compiledPatterns = []*regexp.Regexp{re}

	err := cli.run()
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	expected := "line1\nline3\n"
	if got := stdout.String(); got != expected {
		t.Errorf("Expected output %q, got %q", expected, got)
	}
}

func TestCLI_run_errorModes(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		args           []string
		patterns       []string
		invertMatch    bool
		errorMode      errorMode
		expectedOutput string
		expectedError  error
	}{
		// errorModeDefault
		{
			name:           "default mode, with output",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"hello"},
			errorMode:      errorModeDefault,
			expectedOutput: "hello\n",
			expectedError:  nil,
		},
		{
			name:           "default mode, no output (pattern mismatch)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"world"},
			errorMode:      errorModeDefault,
			expectedOutput: "",
			expectedError:  nil,
		},
		{
			name:           "default mode, no output (no patterns)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			errorMode:      errorModeDefault,
			expectedOutput: "", // No patterns, no output
			expectedError:  nil,
		},
		{
			name:           "default mode, with output (no patterns, inverted)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			invertMatch:    true,
			errorMode:      errorModeDefault,
			expectedOutput: "hello\n", // No patterns, inverted, all output
			expectedError:  nil,
		},

		// errorModeNoContent
		{
			name:           "no-content mode, with output",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"hello"},
			errorMode:      errorModeNoContent,
			expectedOutput: "hello\n",
			expectedError:  nil,
		},
		{
			name:           "no-content mode, no output (pattern mismatch)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"world"},
			errorMode:      errorModeNoContent,
			expectedOutput: "",
			expectedError:  errDueToMode,
		},
		{
			name:           "no-content mode, no output (no patterns)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			errorMode:      errorModeNoContent,
			expectedOutput: "",
			expectedError:  errDueToMode,
		},
		{
			name:           "no-content mode, with output (no patterns, inverted)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			invertMatch:    true,
			errorMode:      errorModeNoContent,
			expectedOutput: "hello\n",
			expectedError:  nil,
		},

		// errorModeOnContent
		{
			name:           "on-content mode, with output",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"hello"},
			errorMode:      errorModeOnContent,
			expectedOutput: "hello\n",
			expectedError:  errDueToMode,
		},
		{
			name:           "on-content mode, no output (pattern mismatch)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       []string{"world"},
			errorMode:      errorModeOnContent,
			expectedOutput: "",
			expectedError:  nil,
		},
		{
			name:           "on-content mode, no output (no patterns)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			errorMode:      errorModeOnContent,
			expectedOutput: "",
			expectedError:  nil,
		},
		{
			name:           "on-content mode, with output (no patterns, inverted)",
			command:        "echo",
			args:           []string{"hello"},
			patterns:       nil,
			invertMatch:    true,
			errorMode:      errorModeOnContent,
			expectedOutput: "hello\n",
			expectedError:  errDueToMode,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cli := &CLI{
				Input:       strings.NewReader(""),
				Output:      &stdout,
				ErrOut:      &stderr,
				command:     tc.command,
				args:        tc.args,
				invertMatch: tc.invertMatch,
				errorMode:   tc.errorMode,
			}
			if tc.patterns != nil {
				cli.compiledPatterns = make([]*regexp.Regexp, len(tc.patterns))
				for i, p := range tc.patterns {
					cli.compiledPatterns[i] = compileSinglePattern(p)
				}
			}

			err := cli.run()

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("run() error = %v, want %v", err, tc.expectedError)
			}

			if got := stdout.String(); got != tc.expectedOutput {
				t.Errorf("stdout = %q, want %q", got, tc.expectedOutput)
			}
			if tc.expectedError != nil && stderr.String() != "" {
				t.Logf("stderr with expected error: %s", stderr.String())
			}
		})
	}
}
