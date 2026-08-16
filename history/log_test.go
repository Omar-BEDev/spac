// This file was generated with the assistance of an artificial
// intelligence coding agent (opencode). It contains unit tests for the
// history package, following the go test convention required by the
// project contributing guidelines.
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
