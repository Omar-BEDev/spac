// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It contains end-to-end tests for handleLine: the
// requested action must log one history entry per executed method, unknown
// commands must be reported, and failed requests must not pollute history.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"spac/history"
	"spac/network"
	"spac/ui"
)

// TestMain forces terminal decorations off for every main-package test. This
// guarantees banner/prompt colors and the spinner never leak ANSI sequences
// into captured output, so assertions stay about the messages themselves.
func TestMain(m *testing.M) {
	ui.SetEnabled(false)
	os.Exit(m.Run())
}

// captureStdout runs fn while its prints go to a pipe instead of the real
// stdout, then returns everything it printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	defer func() { os.Stdout = original }()

	fn()

	w.Close()
	out, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatalf("io.ReadAll() error: %v", err)
	}
	return string(out)
}

// TestHandleLineLogsEachMethod verifies that a successful request logs one
// history entry per listed method (in this case post and get).
func TestHandleLineLogsEachMethod(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handleLine(fmt.Sprintf(`new req "%s" -method(post,get)`, server.URL))

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read history log: %v", err)
	}
	for _, want := range []string{"new req POST ", "new req GET "} {
		if !strings.Contains(string(content), want) {
			t.Errorf("history log missing %q ; got %q", want, content)
		}
	}
}

// TestHandleLineUnknownCommand verifies that an unrecognized line is reported
// to the user as an unknown command.
func TestHandleLineUnknownCommand(t *testing.T) {
	out := captureStdout(t, func() {
		handleLine("foo")
	})
	if !strings.Contains(out, "unknown command") {
		t.Errorf("expected %q in output, got %q", "unknown command", out)
	}
}

// TestHandleLineBlankInput verifies empty and whitespace-only lines are
// treated as unknown commands rather than crashing or being parsed.
func TestHandleLineBlankInput(t *testing.T) {
	for _, line := range []string{"", "   "} {
		out := captureStdout(t, func() {
			handleLine(line)
		})
		if !strings.Contains(out, "unknown command") {
			t.Errorf("handleLine(%q) expected %q, got %q", line, "unknown command", out)
		}
	}
}

// TestHandleLineNoAnsiLeak verifies that when decorations are disabled (as
// TestMain guarantees) neither success nor error paths emit escape sequences.
func TestHandleLineNoAnsiLeak(t *testing.T) {
	original := ui.Enabled()
	ui.SetEnabled(false)
	defer ui.SetEnabled(original)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`new req "%s"`, server.URL))
		handleLine("foo")
	})
	if strings.Contains(out, "\x1b") {
		t.Errorf("handleLine leaked ANSI escape sequences: %q", out)
	}
}

// failingTransport is a http.RoundTripper that always fails so a network
// failure can be simulated deterministically in main-level tests.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("simulated network failure")
}

// TestHandleLineNoLogOnNetworkFailure verifies that when the request itself
// fails, nothing is appended to the history log.
func TestHandleLineNoLogOnNetworkFailure(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	original := network.DefaultClient
	network.DefaultClient = &http.Client{Transport: failingTransport{}}
	t.Cleanup(func() { network.DefaultClient = original })

	out := captureStdout(t, func() {
		handleLine(`new req "http://example.com"`)
	})
	if !strings.Contains(out, "request failed") {
		t.Errorf("expected %q in output, got %q", "request failed", out)
	}

	content, err := os.ReadFile(logPath)
	if err == nil {
		if strings.Contains(string(content), "new req") {
			t.Errorf("history log should not contain new req on failure, got %q", content)
		}
	}
}
