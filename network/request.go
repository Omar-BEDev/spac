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

// Send performs a single HTTP request with no body and returns the response
// status line. It ignores the numeric status code, which callers that only
// need the printable string can do.
func Send(method, targetURL string) (string, error) {
	status, _, err := SendWithBody(method, targetURL, nil)
	return status, err
}

// SendWithBody performs a single HTTP request with an optional JSON body.
// When body is non-empty it is sent as-is and the Content-Type header is set
// to application/json. The response body is drained so the connection can be
// reused, then the response status line (for example "200 OK") and its
// numeric code (for example 200) are returned. The numeric code is exposed
// so callers can assert on it instead of parsing the printable line.
func SendWithBody(method, targetURL string, body []byte) (string, int, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(strings.ToUpper(method), targetURL, reader)
	if err != nil {
		return "", 0, fmt.Errorf("build request: %w", err)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := DefaultClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	// Decision: drain the body instead of closing it unread to allow HTTP
	// keep-alive connection reuse on the client side.
	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.Status, resp.StatusCode, nil
}
