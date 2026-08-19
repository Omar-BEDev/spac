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
	"spac/template"
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

// okServer returns a minimal httptest server that answers every request with
// 200 OK. Tests use it so no real network is involved.
func okServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server
}

// setTemplatePath points the shared body-template path at content in a temp
// file and restores the previous value afterwards.
func setTemplatePath(t *testing.T, content string) {
	t.Helper()
	previous := template.DefaultPath
	path := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write template file: %v", err)
	}
	template.DefaultPath = path
	t.Cleanup(func() { template.DefaultPath = previous })
}

// TemplateBodyForTest is a valid struct-tagged body template used by the
// body-visibility tests.
const templateBodyForTest = `{"struct": {"product": {"name": "user write new name here", "price": 0}}}`

// TestHandleLinePostAndPutPrintBody verifies the data-driven body structure is
// shown for POST and PUT requests.
func TestHandleLinePostAndPutPrintBody(t *testing.T) {
	setTemplatePath(t, templateBodyForTest)

	server := okServer(t)
	for _, method := range []string{"post", "put"} {
		out := captureStdout(t, func() {
			handleLine(fmt.Sprintf(`new req "%s" -method(%s)`, server.URL, method))
		})
		if !strings.Contains(out, "body structure") {
			t.Errorf("%s: expected %q in output, got %q", method, "body structure", out)
		}
		if !strings.Contains(out, "user write new name here") {
			t.Errorf("%s: expected template placeholder in output, got %q", method, out)
		}
	}
}

// TestHandleLineGetAndDeleteNoBody verifies GET and DELETE print no body
// structure even when a template exists.
func TestHandleLineGetAndDeleteNoBody(t *testing.T) {
	setTemplatePath(t, templateBodyForTest)

	server := okServer(t)
	for _, method := range []string{"get", "delete"} {
		out := captureStdout(t, func() {
			handleLine(fmt.Sprintf(`new req "%s" -method(%s)`, server.URL, method))
		})
		if strings.Contains(out, "body structure") {
			t.Errorf("%s: did not expect body structure in output, got %q", method, out)
		}
	}
}

// TestHandleLineRunTestsSample runs a real sample tests file (generated in a
// temp dir against a local server) and verifies every case passes and is
// logged.
func TestHandleLineRunTestsSample(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	server := okServer(t)
	sample := fmt.Sprintf(`{
		"tests": [
			{"method": "get", "url": "%s/health"},
			{"method": "post", "url": "%s/users", "body": {"name": "a"}},
			{"method": "put", "url": "%s/users/1", "body": {"name": "b"}},
			{"method": "delete", "url": "%s/users/1"}
		]
	}`, server.URL, server.URL, server.URL, server.URL)

	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})

	if got := strings.Count(out, "PASS"); got != 4 {
		t.Errorf("expected 4 PASS lines, got %d in %q", got, out)
	}
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		if !strings.Contains(out, method) {
			t.Errorf("expected %s case in output, got %q", method, out)
		}
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read history log: %v", err)
	}
	for _, action := range []string{"run tests get", "run tests post", "run tests put", "run tests delete"} {
		if !strings.Contains(string(content), action) {
			t.Errorf("history log missing %q ; got %q", action, content)
		}
	}
}

// TestHandleLineRunTestsMissingPath verifies a clear error when the path is
// omitted.
func TestHandleLineRunTestsMissingPath(t *testing.T) {
	out := captureStdout(t, func() {
		handleLine("run -tests")
	})
	if !strings.Contains(out, "missing tests file path") {
		t.Errorf("expected missing-path error, got %q", out)
	}
}

// TestHandleLineRunTestsBadPath verifies a clear error when the file does not
// exist.
func TestHandleLineRunTestsBadPath(t *testing.T) {
	out := captureStdout(t, func() {
		handleLine(`run -tests "no-such-tests.json"`)
	})
	if !strings.Contains(out, "read tests file") {
		t.Errorf("expected read error, got %q", out)
	}
}

// TestHandleLineRunTestsBadJSON verifies a clear error for malformed test
// content.
func TestHandleLineRunTestsBadJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte("this is not json"), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "parse tests file") {
		t.Errorf("expected parse error, got %q", out)
	}
}

// TestHandleLineRunTestsMissingTag verifies the explicit "not a tests file"
// error surfaces for a JSON document without a top-level tests tag.
func TestHandleLineRunTestsMissingTag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "file is not a tests file") {
		t.Errorf("expected not-a-tests-file error, got %q", out)
	}
}

// TestHandleLinePostMissingStructTag verifies the explicit "not a body
// template" error surfaces for a POST request with a template lacking the
// struct header.
func TestHandleLinePostMissingStructTag(t *testing.T) {
	setTemplatePath(t, `{"name": "user write new name here"}`)

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`new req "%s" -method(post)`, okServer(t).URL))
	})
	if !strings.Contains(out, "not a body template") {
		t.Errorf("expected not-a-body-template error, got %q", out)
	}
}

// TestHandleLineRunTestsFailure verifies failing cases print FAIL and are not
// recorded in history.
func TestHandleLineRunTestsFailure(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	original := network.DefaultClient
	network.DefaultClient = &http.Client{Transport: failingTransport{}}
	t.Cleanup(func() { network.DefaultClient = original })

	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(`{"tests": [{"method": "get", "url": "http://x.invalid/"}]}`), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL in output, got %q", out)
	}

	content, err := os.ReadFile(logPath)
	if err == nil && strings.Contains(string(content), "run tests") {
		t.Errorf("history should be empty for failed cases, got %q", content)
	}
}

// statusServer returns a minimal httptest server that answers every request
// with the given status code, so non-2xx responses can be produced without a
// real network call.
func statusServer(t *testing.T, code int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}))
	t.Cleanup(server.Close)
	return server
}

// TestHandleLineRunTestsWrongStatus verifies a case without expected_status
// that receives a non-2xx response (404) prints FAIL, shows the actual code,
// and is not counted in the PASS total.
func TestHandleLineRunTestsWrongStatus(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	server := statusServer(t, http.StatusNotFound)
	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(fmt.Sprintf(`{"tests": [{"method": "get", "url": "%s/not-found"}]}`, server.URL)), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL for non-2xx status, got %q", out)
	}
	if !strings.Contains(out, "404") {
		t.Errorf("expected actual status 404 in output, got %q", out)
	}
	if got := strings.Count(out, "PASS"); got != 0 {
		t.Errorf("wrong-status case must not count in PASS total, got %d PASS in %q", got, out)
	}

	content, err := os.ReadFile(logPath)
	if err == nil && strings.Contains(string(content), "run tests") {
		t.Errorf("history should be empty for a wrong-status case, got %q", content)
	}
}

// TestHandleLineRunTestsExpectedStatusMatches verifies a case with an
// explicit expected_status set to a non-2xx value passes when the server
// returns exactly that status, proving intentional non-2xx checks work.
func TestHandleLineRunTestsExpectedStatusMatches(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "history.log")
	history.SetLogFilePath(logPath)

	server := statusServer(t, http.StatusNotFound)
	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(fmt.Sprintf(`{"tests": [{"method": "get", "url": "%s/should-not-exist", "expected_status": 404}]}`, server.URL)), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "PASS") {
		t.Errorf("expected PASS for intentional 404 check, got %q", out)
	}
	if !strings.Contains(out, "404") {
		t.Errorf("expected 404 in output, got %q", out)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read history log: %v", err)
	}
	if !strings.Contains(string(content), "run tests get") {
		t.Errorf("history log missing passing case, got %q", content)
	}
}

// TestHandleLineRunTestsExpectedStatusMismatch verifies a case whose explicit
// expected_status does not match the response prints a FAIL showing both the
// wanted and the actual code.
func TestHandleLineRunTestsExpectedStatusMismatch(t *testing.T) {
	server := okServer(t)
	path := filepath.Join(t.TempDir(), "tests.json")
	if err := os.WriteFile(path, []byte(fmt.Sprintf(`{"tests": [{"method": "get", "url": "%s/x", "expected_status": 404}]}`, server.URL)), 0o644); err != nil {
		t.Fatalf("write tests file: %v", err)
	}

	out := captureStdout(t, func() {
		handleLine(fmt.Sprintf(`run -tests "%s"`, path))
	})
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL on expected_status mismatch, got %q", out)
	}
	if !strings.Contains(out, "want status 404 got 200") {
		t.Errorf("expected want/got mismatch message, got %q", out)
	}
}
