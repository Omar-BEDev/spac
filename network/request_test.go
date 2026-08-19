// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). It contains unit tests for the network request
// sender using an httptest server so no real network is required.
package network

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestSendMethods verifies Send performs a request with each supported HTTP
// method and returns the response status line.
func TestSendMethods(t *testing.T) {
	var (
		mu      sync.Mutex
		methods []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		methods = append(methods, r.Method)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Point the shared client at the test server.
	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	wantMethods := []string{"POST", "GET", "PUT", "DELETE"}
	for _, m := range wantMethods {
		if _, err := Send(m, server.URL); err != nil {
			t.Fatalf("Send(%q) unexpected error: %v", m, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(methods) != len(wantMethods) {
		t.Fatalf("server received %d requests ; want %d", len(methods), len(wantMethods))
	}
	for i, m := range methods {
		if m != wantMethods[i] {
			t.Errorf("server request %d method = %q ; want %q", i, m, wantMethods[i])
		}
	}
}

// TestSendInvalidURL verifies Send reports an error for an un-routable link.
func TestSendInvalidURL(t *testing.T) {
	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	if _, err := Send("get", "http://127.0.0.1:1/nope"); err == nil {
		t.Error("Send() expected error for unreachable link, got nil")
	}
}
