package cli

import (
	"fmt"
	"strings"
)

// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode) to add the "run -tests" command parser.
//
// Grammar:
//
//	spac>> run -tests "<path to tests.json>"

// IsRunTestsCommand reports whether input is a "run -tests" command,
// ignoring letter case.
func IsRunTestsCommand(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return lower == "run -tests" || strings.HasPrefix(lower, "run -tests ")
}

// ParseRunTests extracts the tests file path from a "run -tests" command
// line. The path may be double-quoted so it can contain spaces.
func ParseRunTests(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if !IsRunTestsCommand(trimmed) {
		return "", fmt.Errorf("not a run -tests command")
	}

	args := splitArgs(trimmed)
	if len(args) < 3 {
		return "", fmt.Errorf("missing tests file path")
	}

	path := unquote(args[2])
	if path == "" {
		return "", fmt.Errorf("missing tests file path")
	}
	return path, nil
}
