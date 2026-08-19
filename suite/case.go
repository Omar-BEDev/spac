// Package suite loads and validates "run -tests" files. A tests file is a
// JSON list of cases, each describing one HTTP request to execute:
//
//	[
//	  { "method": "post", "url": "https://api.example.com/users",
//	    "body": { "name": "..." } },
//	  { "method": "get", "url": "https://api.example.com/health" }
//	]
//
// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). Decisions are commented inline.
package suite

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"spac/cli"
)

// Case is a single request to run from a tests file. Body is optional and,
// when present, is kept as raw JSON so it is sent exactly as written.
type Case struct {
	Method string          `json:"method"`
	URL    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// Parse validates raw tests-file JSON and returns its cases.
//
// Validation fails fast with a clear error on: non-array content, malformed
// JSON, zero cases, an unsupported method, or a missing URL. Validation of
// the URL scheme is intentionally left to the network layer, so an
// unreachable or bad-scheme case surfaces as a FAIL at run time.
func Parse(data []byte) ([]Case, error) {
	var cases []Case
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, fmt.Errorf("parse tests file: %w", err)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("tests file contains no cases")
	}
	for i, tc := range cases {
		if !cli.SupportedMethod(strings.ToLower(tc.Method)) {
			return nil, fmt.Errorf("case %d: unsupported method %q", i+1, tc.Method)
		}
		if strings.TrimSpace(tc.URL) == "" {
			return nil, fmt.Errorf("case %d: missing url", i+1)
		}
	}
	return cases, nil
}

// Load reads and parses a tests file from disk.
func Load(path string) ([]Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tests file %q: %w", path, err)
	}
	return Parse(data)
}
