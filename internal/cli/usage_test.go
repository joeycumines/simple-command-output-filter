package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI_init(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errorIs   error
		checkFunc func(*testing.T, *CLI)
	}{
		{
			name:      "no arguments",
			args:      []string{},
			wantError: true,
			errorIs:   errNoCommand,
		},
		{
			name:      "help flag",
			args:      []string{"-h"},
			wantError: true, // help flag causes error
		},
		{
			name:      "simple command",
			args:      []string{"echo", "hello"},
			wantError: false,
			checkFunc: func(t *testing.T, c *CLI) {
				if c.command != "echo" {
					t.Errorf("Expected command to be 'echo', got %q", c.command)
				}
				if len(c.args) != 1 || c.args[0] != "hello" {
					t.Errorf("Expected args to be ['hello'], got %v", c.args)
				}
			},
		},
		{
			name:      "with pattern",
			args:      []string{"-p", "hello*", "echo", "hello world"},
			wantError: false,
			checkFunc: func(t *testing.T, c *CLI) {
				if c.command != "echo" {
					t.Errorf("Expected command to be 'echo', got %q", c.command)
				}
				if len(c.compiledPatterns) != 1 {
					t.Errorf("Expected 1 compiled pattern, got %d", len(c.compiledPatterns))
				}
			},
		},
		{
			name:      "with pattern and invert match",
			args:      []string{"-p", "hello*", "-v", "echo", "hello world"},
			wantError: false,
			checkFunc: func(t *testing.T, c *CLI) {
				if !c.invertMatch {
					t.Errorf("Expected invertMatch to be true")
				}
			},
		},
		{
			name:      "with valid pattern some might find confusing",
			args:      []string{"-p", "hello[", "echo", "hello world"},
			wantError: false,
		},
		{
			name:      "with double dash separator",
			args:      []string{"-p", "hello*", "--", "echo", "-v", "hello world"},
			wantError: false,
			checkFunc: func(t *testing.T, c *CLI) {
				if c.command != "echo" {
					t.Errorf("Expected command to be 'echo', got %q", c.command)
				}
				if len(c.args) != 2 || c.args[0] != "-v" || c.args[1] != "hello world" {
					t.Errorf("Expected args to be ['-v', 'hello world'], got %v", c.args)
				}
				if c.invertMatch {
					t.Errorf("Expected invertMatch to be false")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			cli := &CLI{
				Input:  strings.NewReader(""),
				Output: &output,
				ErrOut: &output,
			}

			err := cli.init(tc.args)

			// check error state
			if (err != nil) != tc.wantError {
				t.Errorf("init() error = %v, wantError %v", err, tc.wantError)
				return
			}

			// check specific error if expected
			if tc.errorIs != nil && err != nil && !strings.Contains(err.Error(), tc.errorIs.Error()) {
				t.Errorf("init() error = %v, expected to contain %v", err, tc.errorIs)
			}

			// run additional checks if provided
			if err == nil && tc.checkFunc != nil {
				tc.checkFunc(t, cli)
			}
		})
	}
}
