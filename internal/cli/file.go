package cli

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

func readPatternsFromFile(allRawPatterns []string, filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pattern file %q: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if line := stripCommentFromLine(scanner.Text()); line != `` {
			allRawPatterns = append(allRawPatterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read pattern file %q: %w", filePath, err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close pattern file %q: %w", filePath, err)
	}

	return allRawPatterns, nil
}

func stripCommentFromLine(line string) string {
	runes := []rune(line)
	result := make([]rune, 0, len(runes))
	// iterate, deduping '#' and stripping any comment (per help docs)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '#' {
			if i+1 >= len(runes) || runes[i+1] != '#' {
				// a comment was found
				for i = len(result) - 1; i >= 0 && unicode.IsSpace(result[i]); i-- {
					// trim trailing whitespace
					result = result[:i]
				}
				break
			}
			// doubled '#' found, treat as literal
			i++ // consume the second #
		}
		result = append(result, runes[i])
	}
	return string(result)
}
