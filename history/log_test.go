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

package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogActionFormat verifies that LogAction writes a line containing
// the action in the format "YYYY/MM/DD action".
func TestLogActionFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath = filepath.Join(tmpDir, historyFileName)

	action := "test_action"
	if err := LogAction(action); err != nil {
		t.Fatalf("LogAction() unexpected error: %v", err)
	}

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	entry := strings.TrimSpace(string(content))
	if !strings.HasSuffix(entry, action) {
		t.Errorf("expected entry ending with %q, got %q", action, entry)
	}
	if len(entry) != len("2006/01/02 ")+len(action) {
		t.Errorf("expected date+space+action format, got %q", entry)
	}
}

// TestLogNewReq verifies that LogNewReq records a "new req" action.
func TestLogNewReq(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath = filepath.Join(tmpDir, historyFileName)

	if err := LogNewReq(); err != nil {
		t.Fatalf("LogNewReq() unexpected error: %v", err)
	}

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "new req") {
		t.Errorf("expected log entry to contain %q, got %q", "new req", string(content))
	}
}

// TestLogLogin verifies that LogLogin records a "login" action.
func TestLogLogin(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath = filepath.Join(tmpDir, historyFileName)

	if err := LogLogin(); err != nil {
		t.Fatalf("LogLogin() unexpected error: %v", err)
	}

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "login") {
		t.Errorf("expected log entry to contain %q, got %q", "login", string(content))
	}
}

// TestSetLogFilePath verifies that SetLogFilePath redirects logging, and
// that an empty path restores the default location.
func TestSetLogFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	override := filepath.Join(tmpDir, historyFileName)
	SetLogFilePath(override)

	if err := LogLogin(); err != nil {
		t.Fatalf("LogLogin() unexpected error: %v", err)
	}

	content, err := os.ReadFile(override)
	if err != nil {
		t.Fatalf("failed to read redirected log file: %v", err)
	}
	if !strings.Contains(string(content), "login") {
		t.Errorf("expected redirected log entry to contain %q, got %q", "login", string(content))
	}

	SetLogFilePath("")
	if logFilePath != defaultLogFilePath {
		t.Errorf("SetLogFilePath(\"\") did not restore default path: got %q", logFilePath)
	}
}
