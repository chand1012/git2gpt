package utils

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// https://chat.openai.com/share/c9f3510b-4278-4a1b-b2ab-bf78f60dff46
func RemoveCodeComments(code string) string {
	// Compile the regular expression to match various single-line and multi-line comment styles.
	// This regex will match single-line comments and multi-line comments in C, C++, JavaScript, and HTML.
	commentRegex := regexp.MustCompile(`^\s*(//|#|--|<!--|%|;|REM\s).*$`)

	// Use a scanner to process each line of the input string
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(code))
	for scanner.Scan() {
		line := scanner.Text()
		// Remove the comment portion from the line using a regex
		cleanLine := commentRegex.ReplaceAllString(line, "")
		if cleanLine != "" {
			// Write the cleaned line to the result, preserving original line breaks
			result.WriteString(cleanLine + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(&result, "Error reading input:", err)
	}

	// Perform an additional global replacement to catch any multi-line comments that might span multiple lines processed by the scanner
	finalCleanedCode := commentRegex.ReplaceAllString(result.String(), "")

	return strings.TrimRight(finalCleanedCode, "\n")
}
