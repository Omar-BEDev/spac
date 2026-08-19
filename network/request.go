// Package network implements the HTTP calls made by the spac command.
//
// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). Decisions are commented inline to make the
// reasoning behind them explicit.
package network

import (
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

// Send performs a single HTTP request with the given method and target URL.
// It drains and discards the response body so the connection can be reused,
// then returns the response status line (for example "200 OK").
func Send(method, targetURL string) (string, error) {
	req, err := http.NewRequest(strings.ToUpper(method), targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	// Decision: drain the body instead of closing it unread to allow HTTP
	// keep-alive connection reuse on the client side.
	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.Status, nil
}
