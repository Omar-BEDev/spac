// Package template loads the JSON request-body template used by the spac
// console. The body is data-driven: it lives in a local JSON file (by
// default templates/body.json) and is never hardcoded in Go code.
//
// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). Decisions are commented inline to make the
// reasoning behind them explicit.
package template

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DefaultPath is the default location of the request body template file,
// relative to the current working directory.
//
// Decision: it is a variable rather than a constant so tests can point it at
// a temporary file without a real repository check-out being required.
var DefaultPath = "templates/body.json"

// Load reads and validates the JSON body template at path. The template is
// returned as raw bytes so it can be both displayed to the user and sent
// verbatim as the request body.
func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read body template %q: %w", path, err)
	}
	if !json.Valid(data) {
		return nil, fmt.Errorf("body template %q is not valid JSON", path)
	}
	return data, nil
}

// LoadForMethod returns the body template bytes only for POST and PUT
// requests. GET and DELETE carry no body (nil, nil). If the template file
// cannot be loaded for a POST/PUT request the error is returned so the
// caller can surface it instead of sending a silently wrong request.
func LoadForMethod(method, path string) ([]byte, error) {
	switch strings.ToLower(method) {
	case "post", "put":
		return Load(path)
	default:
		return nil, nil
	}
}
