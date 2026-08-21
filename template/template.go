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

// Package template loads the JSON request-body template used by the spac
// console. The body is data-driven: it lives in a local JSON file (by
// default templates/body.json) and is never hardcoded in Go code.
//
// A valid template file must carry a top-level "struct" header tag that maps
// a structure name to its JSON body:
//
//	{
//	  "struct": {
//	    "product": { "name": "...", "price": 0 },
//	    "user": { "name": "...", "email": "..." }
//	  }
//	}
package template

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// DefaultPath is the default location of the request body template file,
// relative to the current working directory.
//
// Decision: it is a variable rather than a constant so tests can point it at
// a temporary file without a real repository check-out being required.
var DefaultPath = "templates/body.json"

// StructFile is a validated body template file. ActiveName is the structure
// the console serializes as the request body; Body holds its raw JSON.
type StructFile struct {
	ActiveName string
	Body       []byte
}

// Load reads and validates a body template file. The top-level "struct" tag
// is mandatory: without it the file is rejected as "not a body template".
// When name is empty the first structure in sorted name order is selected;
// when name is given it must exist in the file or an error is returned.
func Load(path, name string) (*StructFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read body template %q: %w", path, err)
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse body template %q: %w", path, err)
	}

	structRaw, ok := doc["struct"]
	if !ok {
		return nil, fmt.Errorf("file is not a body template: missing top-level struct tag")
	}

	var structures map[string]json.RawMessage
	if err := json.Unmarshal(structRaw, &structures); err != nil {
		return nil, fmt.Errorf("file is not a body template: struct tag must be a JSON object")
	}
	if len(structures) == 0 {
		return nil, fmt.Errorf("body template %q: struct tag is empty", path)
	}

	// Decision: without an explicit -struct(name) selector the first
	// structure in sorted name order is the "active" one, so the pick stays
	// deterministic instead of relying on Go map iteration order. With a
	// selector the requested structure wins and a missing name fails fast.
	var active string
	if name == "" {
		names := make([]string, 0, len(structures))
		for candidate := range structures {
			names = append(names, candidate)
		}
		sort.Strings(names)
		active = names[0]
	} else {
		raw, exists := structures[name]
		if !exists {
			return nil, fmt.Errorf("body template %q: struct %q does not exist", path, name)
		}
		active = name
		structures = map[string]json.RawMessage{name: raw}
	}
	raw := structures[active]

	// Normalize the body: re-marshal with 2-space indentation so the body is
	// clean and stable for display and sending, independent of how the
	// template file was written. This also validates the structure JSON.
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("invalid JSON in struct %q: %w", active, err)
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialize struct %q: %w", active, err)
	}

	return &StructFile{ActiveName: active, Body: pretty}, nil
}

// LoadForMethod returns the active body-template structure bytes only for
// methods that carry a body (POST, PUT, PATCH). GET, DELETE, HEAD and
// OPTIONS carry no body (nil, nil). name selects a specific structure from
// the template file; an empty name uses the default structure. If the
// template file cannot be loaded for a body-carrying request the error is
// returned so the caller can surface it instead of sending a silently wrong
// request.
func LoadForMethod(method, path, name string) ([]byte, error) {
	switch strings.ToLower(method) {
	case "post", "put", "patch":
		sf, err := Load(path, name)
		if err != nil {
			return nil, err
		}
		return sf.Body, nil
	default:
		return nil, nil
	}
}
