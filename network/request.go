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

// Package network implements the HTTP calls made by the spac command.
package network

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultClient is the HTTP client used by Send. It is a variable on purpose
// so unit tests can swap in a client pointed at a local test server without
// performing real network calls.
var DefaultClient = &http.Client{Timeout: 10 * time.Second}

// Response is everything the console needs from one HTTP round trip: the
// printable status line, its numeric code, the response headers and the raw
// response body.
type Response struct {
	Status string
	Code   int
	Header http.Header
	Body   []byte
}

// Send performs a single HTTP request with no body and returns the response
// status line. It ignores the rest of the response, which callers that only
// need the printable string can do.
func Send(method, targetURL string) (string, error) {
	resp, err := SendWithBody(method, targetURL, nil, nil)
	if err != nil {
		return "", err
	}
	return resp.Status, nil
}

// SendWithBody performs a single HTTP request with optional custom headers
// and an optional JSON body. When body is non-empty it is sent as-is and the
// Content-Type header is set to application/json unless the caller supplied
// its own Content-Type in headers. The full response (status line, numeric
// code, headers and body) is returned so callers can display or assert on
// any part of it instead of parsing the printable line.
//
// Decision: headers is a plain map keyed by canonical header name; a nil map
// means no extra headers, keeping the common call sites short.
func SendWithBody(method, targetURL string, headers map[string]string, body []byte) (*Response, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(strings.ToUpper(method), targetURL, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	// Decision: read the whole body rather than draining it into io.Discard;
	// reading to EOF also drains the connection, so HTTP keep-alive reuse is
	// preserved while the body stays available for display and assertions.
	respBody, readErr := io.ReadAll(resp.Body)

	out := &Response{
		Status: resp.Status,
		Code:   resp.StatusCode,
		Header: resp.Header,
		Body:   respBody,
	}
	if readErr != nil {
		return out, fmt.Errorf("read response body: %w", readErr)
	}
	return out, nil
}
