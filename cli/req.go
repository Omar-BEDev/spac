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
	"net/url"
	"strings"
	"unicode"
)

// Request holds the parsed result of a "new req" command line.
type Request struct {
	URL     string
	Methods []string
}

// allowedMethods lists the HTTP methods accepted by the -method(-) flag.
// Decision: keep the set small and aligned with the existing command map.
var allowedMethods = map[string]bool{
	"post":   true,
	"get":    true,
	"put":    true,
	"delete": true,
}

// IsReqCommand reports whether input is a "new req" command, ignoring
// letter case so both "New req" and "new Req" are accepted from the console.
// Decision: the existing command map uses exact matching, this looser check
// is reserved for the interactive command line parser.
func IsReqCommand(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return lower == "new req" || strings.HasPrefix(lower, "new req ")
}

// ParseReq parses a "new req" command line. The grammar is:
//
//	spac>> new req "<api link>" [-method(post,get,put,delete)]
//
// The API link is required and may be double-quoted so the shell-style
// tokenizer keeps it together. Links containing whitespace are rejected
// because a raw space in an HTTP URL is not something a real server will
// accept (the server replies 400), and only http/https schemes are allowed.
// The -method flag is optional and defaults to "post" when omitted; each
// listed method triggers one request and duplicates are executed only once.
func ParseReq(input string) (*Request, error) {
	trimmed := strings.TrimSpace(input)
	if !IsReqCommand(trimmed) {
		return nil, fmt.Errorf("not a new req command")
	}

	args := splitArgs(trimmed)
	if len(args) < 3 {
		return nil, fmt.Errorf("missing api link")
	}

	apiLink := unquote(args[2])
	if apiLink == "" {
		return nil, fmt.Errorf("api link cannot be empty")
	}
	// Decision: rather than silently percent-encoding the URL (which would
	// hide user mistakes), fail fast with an explicit error.
	if containsWhitespace(apiLink) {
		return nil, fmt.Errorf("api link contains whitespace")
	}

	// Decision: only http and https make sense for the network sender, so
	// schemes like ftp:// or a missing scheme are rejected here instead of
	// surfacing as a confusing request error later.
	parsedURL, err := url.Parse(apiLink)
	if err != nil {
		return nil, fmt.Errorf("invalid api link %q: %w", apiLink, err)
	}
	switch parsedURL.Scheme {
	case "http", "https":
	default:
		if parsedURL.Scheme == "" {
			return nil, fmt.Errorf("api link is missing a scheme; prefix it with http:// or https://")
		}
		return nil, fmt.Errorf("unsupported api link scheme %q; use http or https", parsedURL.Scheme)
	}

	req := &Request{URL: apiLink}

	var methods []string
	haveMethods := false
	for _, arg := range args[3:] {
		parsedMethods, ok, err := parseMethodArg(arg)
		if err != nil {
			return nil, err
		}
		if ok {
			haveMethods = true
			methods = append(methods, parsedMethods...)
		}
	}

	if haveMethods {
		req.Methods = dedupe(methods)
	} else {
		// Default decision: POST is the most common create/check verb.
		req.Methods = []string{"post"}
	}

	return req, nil
}

// splitArgs splits a command line by whitespace while keeping double-quoted
// sections together so the API link can contain spaces.
func splitArgs(input string) []string {
	var (
		args    []string
		current strings.Builder
		inQuote bool
	)
	for _, r := range input {
		switch {
		case r == '"':
			inQuote = !inQuote
			current.WriteRune(r)
		case unicode.IsSpace(r) && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// unquote removes surrounding double quotes from a token when present.
func unquote(token string) string {
	if len(token) >= 2 && token[0] == '"' && token[len(token)-1] == '"' {
		return token[1 : len(token)-1]
	}
	return token
}

// containsWhitespace reports whether s contains any Unicode space character.
func containsWhitespace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// parseMethodArg recognises a "-method(...)" or "-method=..." flag and
// returns the validated method list it carries. ok is false when the token
// is not a method flag at all.
func parseMethodArg(arg string) ([]string, bool, error) {
	lower := strings.ToLower(arg)
	var inner string
	switch {
	case strings.HasPrefix(lower, "-method(") && strings.HasSuffix(arg, ")"):
		inner = arg[len("-method(") : len(arg)-1]
	case strings.HasPrefix(lower, "-method="):
		inner = arg[len("-method="):]
	default:
		return nil, false, nil
	}
	methods, err := parseMethods(strings.Split(inner, ","))
	return methods, true, err
}

// parseMethods validates and normalises a list of raw method names. It
// lowercases them, rejects anything outside the supported set, and removes
// duplicates while preserving input order.
func parseMethods(raw []string) ([]string, error) {
	var methods []string
	seen := make(map[string]bool)
	for _, m := range raw {
		m = strings.ToLower(strings.TrimSpace(m))
		if m == "" {
			continue
		}
		if !allowedMethods[m] {
			return nil, fmt.Errorf("unsupported method %q", m)
		}
		if !seen[m] {
			seen[m] = true
			methods = append(methods, m)
		}
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("no methods provided")
	}
	return methods, nil
}

// dedupe removes duplicate strings preserving the first occurrence order.
func dedupe(items []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}
