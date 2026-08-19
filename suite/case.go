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

// Package suite loads and validates "run -tests" files. A tests file is a
// JSON document with a top-level "tests" tag holding the list of cases, each
// describing one HTTP request to execute:
//
//	{
//	  "tests": [
//	    { "method": "post", "url": "https://api.example.com/users",
//	      "body": { "name": "..." } },
//	    { "method": "get", "url": "https://api.example.com/health" }
//	  ]
//	}
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
// Validation fails fast with a clear error on: malformed JSON, a missing
// top-level "tests" tag (so any random object or an empty {} is rejected as
// "not a tests file"), a wrong-typed "tests" value, zero cases, an
// unsupported method, or a missing URL. Validation of the URL scheme is
// intentionally left to the network layer, so an unreachable or bad-scheme
// case surfaces as a FAIL at run time.
func Parse(data []byte) ([]Case, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse tests file: %w", err)
	}

	rawCases, ok := doc["tests"]
	if !ok {
		return nil, fmt.Errorf("file is not a tests file: missing top-level tests tag")
	}

	var entries []json.RawMessage
	if err := json.Unmarshal(rawCases, &entries); err != nil {
		return nil, fmt.Errorf("file is not a tests file: tests tag must be a list of cases")
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("tests file contains no cases")
	}

	cases := make([]Case, 0, len(entries))
	for i, raw := range entries {
		var tc Case
		if err := json.Unmarshal(raw, &tc); err != nil {
			return nil, fmt.Errorf("case %d: %w", i+1, err)
		}
		if !cli.SupportedMethod(strings.ToLower(tc.Method)) {
			return nil, fmt.Errorf("case %d: unsupported method %q", i+1, tc.Method)
		}
		if strings.TrimSpace(tc.URL) == "" {
			return nil, fmt.Errorf("case %d: missing url", i+1)
		}
		cases = append(cases, tc)
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
