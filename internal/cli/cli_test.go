package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCLI_Main_integration(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedCode   int
	}{
		{
			name:           "no arguments",
			args:           []string{},
			expectedOutput: "",
			expectedCode:   2,
		},
		{
			name:           "help flag",
			args:           []string{"-h"},
			expectedOutput: "",
			expectedCode:   0,
		},
		{
			name:           "simple echo command",
			args:           []string{"echo", "hello world"},
			expectedOutput: "",
			expectedCode:   0,
		},
		{
			name:           "with matching pattern",
			args:           []string{"-p", "hello*", "echo", "hello world"},
			expectedOutput: "hello world\n",
			expectedCode:   0,
		},
		{
			name:           "with non-matching pattern",
			args:           []string{"-p", "foo*", "echo", "hello world"},
			expectedOutput: "",
			expectedCode:   0,
		},
		{
			name:           "with matching pattern and invert-match",
			args:           []string{"-p", "hello*", "-v", "echo", "hello world"},
			expectedOutput: "",
			expectedCode:   0,
		},
		{
			name:           "with non-matching pattern and invert-match",
			args:           []string{"-p", "foo*", "-v", "echo", "hello world"},
			expectedOutput: "hello world\n",
			expectedCode:   0,
		},
		{
			name:           "with invalid command",
			args:           []string{"command_that_does_not_exist_12345"},
			expectedOutput: "",
			expectedCode:   1,
		},
		{
			name:           "error mode no-content, no output",
			args:           []string{"-e", "no-content", "echo", "-n", ""}, // echo -n produces no output
			expectedOutput: "",
			expectedCode:   1, // Should exit 1 due to no-content mode and no output
		},
		{
			name:           "error mode no-content, with output",
			args:           []string{"-e", "no-content", "-p", "hello*", "echo", "hello world"},
			expectedOutput: "hello world\n",
			expectedCode:   0, // Should exit 0
		},
		{
			name:           "error mode on-content, with output",
			args:           []string{"-e", "on-content", "-p", "hello*", "echo", "hello world"},
			expectedOutput: "hello world\n",
			expectedCode:   1, // Should exit 1 due to on-content mode and output
		},
		{
			name:           "error mode on-content, no output",
			args:           []string{"-e", "on-content", "-p", "foo", "echo", "hello world"},
			expectedOutput: "",
			expectedCode:   0, // Should exit 0
		},
		{
			name:           "error mode on-content, no output (no patterns, inverted)",
			args:           []string{"-e", "on-content", "-v", "echo", "hello world"}, // -v with no patterns means all output
			expectedOutput: "hello world\n",
			expectedCode:   1, // Should exit 1 due to on-content and output
		},
		{
			name:           "error mode no-content, with output (no patterns, inverted)",
			args:           []string{"-e", "no-content", "-v", "echo", "hello world"}, // -v with no patterns means all output
			expectedOutput: "hello world\n",
			expectedCode:   0, // Should exit 0
		},
		{
			name:           "invalid error mode value",
			args:           []string{"-e", "bogus", "echo", "hello"},
			expectedOutput: "", // Error message will be on stderr
			expectedCode:   2,  // Should exit 2 due to init error
		},
	}

	origStdin, origStdout, origStderr := os.Stdin, os.Stdout, os.Stderr
	defer func() {
		os.Stdin, os.Stdout, os.Stderr = origStdin, origStdout, origStderr
	}()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			cli := &CLI{
				Input:  strings.NewReader(""),
				Output: &stdout,
				ErrOut: &stderr,
			}

			exitCode := cli.Main(tc.args)

			if exitCode != tc.expectedCode {
				t.Errorf("Expected exit code %d, got %d", tc.expectedCode, exitCode)
			}

			if got := stdout.String(); got != tc.expectedOutput {
				t.Errorf("Expected stdout output %q, got %q", tc.expectedOutput, got)
			}

			if exitCode != tc.expectedCode || stdout.String() != tc.expectedOutput {
				t.Logf("Stderr output: %s", stderr.String())
			}
		})
	}
}
