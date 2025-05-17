package cli

import (
	"regexp"
	"strings"
)

// loadAndCompilePatterns handles init for the patterns and pattern files.
func (x *CLI) loadAndCompilePatterns() error {
	var allRawPatterns []string

	allRawPatterns = append(allRawPatterns, x.rawPatterns...)

	var err error
	for _, filePath := range x.patternFiles {
		allRawPatterns, err = readPatternsFromFile(allRawPatterns, filePath)
		if err != nil {
			return err
		}
	}

	// if no patterns, len(x.compiledPatterns) == 0, handled later
	if len(allRawPatterns) == 0 {
		return nil
	}

	x.compiledPatterns = make([]*regexp.Regexp, 0, len(allRawPatterns))

	for _, pStr := range allRawPatterns {
		x.compiledPatterns = append(x.compiledPatterns, compileSinglePattern(pStr))
	}

	return nil
}

// compileSinglePattern complies a regex from a single pattern string.
func compileSinglePattern(pattern string) *regexp.Regexp {
	var (
		i        int
		char     rune
		runes    = []rune(pattern)
		regexStr strings.Builder
	)

	regexStr.WriteString("^")

	for ; i < len(runes); i++ {
		char = runes[i]
		if char == '*' {
			// check for double asterisk (escaped)
			if i+1 < len(runes) && runes[i+1] == '*' {
				// match literal asterisk
				regexStr.WriteString(`\*`)
				// consume second asterisk
				i++
			} else {
				// wildcard match
				regexStr.WriteString(".*")
			}
		} else {
			// match literal character
			// N.B. ignores unicode grapheme clusters...
			regexStr.WriteString(regexp.QuoteMeta(string(char)))
		}
	}

	regexStr.WriteString("$")

	return regexp.MustCompile(regexStr.String())
}
