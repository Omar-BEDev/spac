/*
Copyright 2026 Omar-BEDev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"fmt"
	"strings"
)

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
