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

package network

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestSendMethods verifies Send performs a request with each supported HTTP
// method and returns the exact status line produced by the server. The
// status is the value the console prints to the user, so it must be asserted
// rather than discarded.
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

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	wantMethods := []string{"POST", "GET", "PUT", "DELETE"}
	for _, m := range wantMethods {
		status, err := Send(m, server.URL)
		if err != nil {
			t.Fatalf("Send(%q) unexpected error: %v", m, err)
		}
		if status != "200 OK" {
			t.Errorf("Send(%q) status = %q ; want %q", m, status, "200 OK")
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

// failingTransport is a http.RoundTripper that always fails. It makes error
// tests deterministic instead of relying on the state of a local port.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("simulated network failure")
}

// TestSendNetworkFailure verifies Send surfaces the error wrapped with the
// "perform request" context so callers can tell a request-execution failure
// apart from a request-build failure.
func TestSendNetworkFailure(t *testing.T) {
	original := DefaultClient
	DefaultClient = &http.Client{Transport: failingTransport{}}
	t.Cleanup(func() { DefaultClient = original })

	_, err := Send("get", "http://example.com/x")
	if err == nil {
		t.Fatal("Send() expected error for failing transport, got nil")
	}
	if !strings.Contains(err.Error(), "perform request") {
		t.Errorf("Send() error = %q ; want it to mention %q", err, "perform request")
	}
}

// TestSendWithBodySendsJSON verifies the JSON body is serialized verbatim on
// the wire and that Content-Type is set to application/json.
func TestSendWithBodySendsJSON(t *testing.T) {
	const wantBody = `{"name":"user write new name here","description":"user write description"}`

	var (
		mu      sync.Mutex
		gotBody []byte
		gotCT   string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		gotBody = body
		gotCT = r.Header.Get("Content-Type")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	resp, err := SendWithBody("post", server.URL, nil, []byte(wantBody))
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error: %v", err)
	}
	status, code := resp.Status, resp.Code
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error: %v", err)
	}
	if status != "200 OK" {
		t.Errorf("SendWithBody() status = %q ; want %q", status, "200 OK")
	}
	if code != http.StatusOK {
		t.Errorf("SendWithBody() code = %d ; want %d", code, http.StatusOK)
	}

	mu.Lock()
	defer mu.Unlock()
	if string(gotBody) != wantBody {
		t.Errorf("SendWithBody() received body %q ; want %q", gotBody, wantBody)
	}
	if gotCT != "application/json" {
		t.Errorf("SendWithBody() Content-Type = %q ; want %q", gotCT, "application/json")
	}
}

// TestSendWithBodyReturnsNumericCode verifies the numeric status code is
// returned alongside the printable status line for a non-2xx response too,
// without being reported as a transport error.
func TestSendWithBodyReturnsNumericCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/not-found" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	resp, err := SendWithBody("get", server.URL+"/not-found", nil, nil)
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error on 404 response: %v", err)
	}
	status, code := resp.Status, resp.Code
	if status != "404 Not Found" {
		t.Errorf("SendWithBody() status = %q ; want %q", status, "404 Not Found")
	}
	if code != http.StatusNotFound {
		t.Errorf("SendWithBody() code = %d ; want %d", code, http.StatusNotFound)
	}

	resp, err = SendWithBody("post", server.URL, nil, []byte(`{"name":"a"}`))
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error on 200 response: %v", err)
	}
	status, code = resp.Status, resp.Code
	if status != "200 OK" || code != http.StatusOK {
		t.Errorf("SendWithBody() = (%q, %d) ; want (\"200 OK\", %d)", status, code, http.StatusOK)
	}
}

// TestSendWithBodyReturnsBodyAndHeaders verifies the full response is
// surfaced: raw body bytes and the response headers alongside the status.
func TestSendWithBodyReturnsBodyAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "yes")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	resp, err := SendWithBody("get", server.URL, nil, nil)
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error: %v", err)
	}
	if string(resp.Body) != `{"ok":true}` {
		t.Errorf("SendWithBody() body = %q ; want %q", resp.Body, `{"ok":true}`)
	}
	if got := resp.Header.Get("X-Test"); got != "yes" {
		t.Errorf("SendWithBody() X-Test header = %q ; want %q", got, "yes")
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("SendWithBody() Content-Type = %q ; want %q", got, "application/json")
	}
}

// TestSendWithBodyCustomHeaders verifies custom headers are placed on the
// outgoing request and that an explicit caller Content-Type wins over the
// default application/json.
func TestSendWithBodyCustomHeaders(t *testing.T) {
	var (
		mu       sync.Mutex
		gotAuth  string
		gotCT    string
		gotExtra string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		gotExtra = r.Header.Get("X-Custom")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	headers := map[string]string{
		"Authorization": "Bearer token123",
		"Content-Type":  "application/merge-patch+json",
		"X-Custom":      "42",
	}
	resp, err := SendWithBody("patch", server.URL, headers, []byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("SendWithBody() unexpected error: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Errorf("SendWithBody() code = %d ; want %d", resp.Code, http.StatusOK)
	}

	mu.Lock()
	defer mu.Unlock()
	if gotAuth != "Bearer token123" {
		t.Errorf("server Authorization = %q ; want %q", gotAuth, "Bearer token123")
	}
	if gotCT != "application/merge-patch+json" {
		t.Errorf("server Content-Type = %q ; want caller value to win", gotCT)
	}
	if gotExtra != "42" {
		t.Errorf("server X-Custom = %q ; want %q", gotExtra, "42")
	}
}

// TestSendMethodsExtended verifies the newly supported methods reach the
// server with their exact wire names.
func TestSendMethodsExtended(t *testing.T) {
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

	original := DefaultClient
	DefaultClient = &http.Client{}
	t.Cleanup(func() { DefaultClient = original })

	wantMethods := []string{"PATCH", "HEAD", "OPTIONS"}
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
