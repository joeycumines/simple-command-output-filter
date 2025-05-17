package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_compileSinglePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   []string
		noMatch []string
	}{
		{
			name:    "exact match",
			pattern: "hello",
			match:   []string{"hello"},
			noMatch: []string{"hello world", "world hello", " hello", "hello "},
		},
		{
			name:    "wildcard at beginning",
			pattern: "*world",
			match:   []string{"world", "hello world", "  world"},
			noMatch: []string{"world hello", "world "},
		},
		{
			name:    "wildcard at end",
			pattern: "hello*",
			match:   []string{"hello", "hello world", "hello  "},
			noMatch: []string{"world hello", " hello"},
		},
		{
			name:    "wildcard in middle",
			pattern: "hello*world",
			match:   []string{"helloworld", "hello world", "hello  world", "hello-123-world"},
			noMatch: []string{"hello", "world", "helloxworldy"},
		},
		{
			name:    "multiple wildcards",
			pattern: "*hello*world*",
			match:   []string{"hello world", "xhello world", "hello worldx", "xhello worldx", "hello in the world", "helloworld"},
			noMatch: []string{"hellworld"},
		},
		{
			name:    "escaped wildcard",
			pattern: "hello**world",
			match:   []string{"hello*world"},
			noMatch: []string{"hello world", "helloworld", "hello**world"},
		},
		{
			name:    "special regex characters",
			pattern: "hello.",
			match:   []string{"hello."},
			noMatch: []string{"hello", "hello1"},
		},
		{
			name:    "complex pattern",
			pattern: "**hello**world**",
			match:   []string{"*hello*world*"},
			noMatch: []string{"**hello**world**", "helloworld"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re := compileSinglePattern(tc.pattern)

			for _, s := range tc.match {
				if !re.MatchString(s) {
					t.Errorf("pattern %q should match %q but didn't", tc.pattern, s)
				}
			}

			for _, s := range tc.noMatch {
				if re.MatchString(s) {
					t.Errorf("pattern %q shouldn't match %q but did", tc.pattern, s)
				}
			}
		})
	}
}

func TestCLI_loadAndCompilePatterns(t *testing.T) {
	t.Run("empty patterns", func(t *testing.T) {
		cli := &CLI{}
		err := cli.loadAndCompilePatterns()
		if err != nil {
			t.Fatalf("loadAndCompilePatterns() error = %v", err)
		}
		if len(cli.compiledPatterns) != 0 {
			t.Errorf("expected 0 compiled patterns, got %d", len(cli.compiledPatterns))
		}
	})

	t.Run("raw patterns only", func(t *testing.T) {
		cli := &CLI{
			rawPatterns: []string{"hello", "world*"},
		}
		err := cli.loadAndCompilePatterns()
		if err != nil {
			t.Fatalf("loadAndCompilePatterns() error = %v", err)
		}
		if len(cli.compiledPatterns) != 2 {
			t.Errorf("expected 2 compiled patterns, got %d", len(cli.compiledPatterns))
		}

		// verify the compiled patterns work as expected
		if !cli.compiledPatterns[0].MatchString("hello") {
			t.Errorf("compiled pattern 'hello' should match 'hello'")
		}
		if cli.compiledPatterns[0].MatchString("hello world") {
			t.Errorf("compiled pattern 'hello' should not match 'hello world'")
		}

		if !cli.compiledPatterns[1].MatchString("world") {
			t.Errorf("compiled pattern 'world*' should match 'world'")
		}
		if !cli.compiledPatterns[1].MatchString("world123") {
			t.Errorf("compiled pattern 'world*' should match 'world123'")
		}
	})
}

func TestCLI_loadAndCompilePatterns_withBothSources(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	patternFile := filepath.Join(tmpDir, "patterns.txt")
	filePatterns := "file_pattern1\nfile_pattern2 # comment\n# comment line\nfile_pattern3 ## not a comment\n"
	if err := os.WriteFile(patternFile, []byte(filePatterns), 0644); err != nil {
		t.Fatalf("Failed to write pattern file: %v", err)
	}

	cli := &CLI{
		rawPatterns:  []string{"raw_pattern1", "raw_pattern2*"},
		patternFiles: []string{patternFile},
	}

	err = cli.loadAndCompilePatterns()
	if err != nil {
		t.Fatalf("loadAndCompilePatterns() error = %v", err)
	}

	// expected patterns are raw plus file patterns with comments stripped
	expectedPatterns := []string{
		"raw_pattern1",
		"raw_pattern2*",
		"file_pattern1",
		"file_pattern2",
		"file_pattern3 # not a comment",
	}

	if len(cli.compiledPatterns) != len(expectedPatterns) {
		t.Errorf("expected %d compiled patterns, got %d", len(expectedPatterns), len(cli.compiledPatterns))
	}

	testStrings := map[string][]string{
		"raw_pattern1":                  {"raw_pattern1"},
		"raw_pattern2*":                 {"raw_pattern2", "raw_pattern2abc"},
		"file_pattern1":                 {"file_pattern1"},
		"file_pattern2":                 {"file_pattern2"},
		"file_pattern3 # not a comment": {"file_pattern3 # not a comment"},
	}

	for i, pattern := range expectedPatterns {
		if i < len(cli.compiledPatterns) {
			regex := cli.compiledPatterns[i]
			for _, match := range testStrings[pattern] {
				if !regex.MatchString(match) {
					t.Errorf("pattern %q should match %q but didn't", pattern, match)
				}
			}
		}
	}
}

func TestCLI_loadAndCompilePatterns_fileError(t *testing.T) {
	cli := &CLI{
		patternFiles: []string{"/non/existent/file/path"},
	}

	// load and compile patterns should return an error
	err := cli.loadAndCompilePatterns()
	if err == nil {
		t.Fatalf("loadAndCompilePatterns() should have failed with non-existent file")
	}
}

func Test_compileSinglePattern_edgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		match   []string
		noMatch []string
	}{
		{
			name:    "empty pattern",
			pattern: "",
			match:   []string{""},
			noMatch: []string{"a", " "},
		},
		{
			name:    "only wildcard",
			pattern: "*",
			match:   []string{"", "a", "abc", "  "},
			noMatch: []string{},
		},
		{
			name:    "multiple consecutive wildcards",
			pattern: "a***b",
			match:   []string{"a*b"},
			noMatch: []string{"ab"},
		},
		{
			name:    "special characters",
			pattern: "hello+.^$",
			match:   []string{"hello+.^$"},
			noMatch: []string{"hello", "hello+", "hello+.^"},
		},
		{
			name:    "pattern with invalid brackets is fine",
			pattern: "hello[world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := compileSinglePattern(tt.pattern)

			for _, s := range tt.match {
				if !re.MatchString(s) {
					t.Errorf("pattern %q should match %q but didn't", tt.pattern, s)
				}
			}

			for _, s := range tt.noMatch {
				if re.MatchString(s) {
					t.Errorf("pattern %q shouldn't match %q but did", tt.pattern, s)
				}
			}
		})
	}
}
