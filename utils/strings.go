package utils

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// RemoveCodeComments removes single-line and multiline comments from the provided code string.
func RemoveCodeComments(code string) string {
	// Regex for single-line comments.
	singleLineCommentRegex := regexp.MustCompile(`^\s*(//|#|--|<!--|%|;|REM\s).*$`)

	// Regex for multiline comments in C, JavaScript, Go, and HTML.
	multiLineCommentRegex := regexp.MustCompile(`(?s)/\*.*?\*/|<!--.*?-->`)
 
	// Use a scanner to process each line of the input string.
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(code))
	for scanner.Scan() {
		line := scanner.Text()
		// First remove multiline comments as they may span across multiple lines.
		line = multiLineCommentRegex.ReplaceAllString(line, "")
		// Then remove any single-line comment parts that remain.
		cleanLine := singleLineCommentRegex.ReplaceAllString(line, "")
		if cleanLine != "" {
			// Write the cleaned line to the result, preserving original line breaks.
			result.WriteString(cleanLine + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(&result, "Error reading input:", err)
	}

	// Additional cleanup in case of multiline comments spanning across multiple scanned lines.
	finalCleanedCode := multiLineCommentRegex.ReplaceAllString(result.String(), "")

	return strings.TrimRight(finalCleanedCode, "\n")
}
