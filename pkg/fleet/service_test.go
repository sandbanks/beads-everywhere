package fleet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandbanks/beads-everywhere/pkg/config"
)

func TestGetIssue(t *testing.T) {
	// Create a temporary directory structure mimicking a beads repo
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, "sample-repo", ".beads")
	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("failed to create temp beads dir: %v", err)
	}

	jsonlContent := `{"id":"test-123","title":"Test Issue Title","description":"Full test description text","status":"open","priority":2,"issue_type":"task","created_at":"2026-09-14T10:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(beadsDir, "issues.jsonl"), []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("failed to write issues.jsonl: %v", err)
	}

	cfg := &config.Config{
		ScanRoots:    []string{tmpDir},
		ArchiveRoots: []string{},
		IgnoredDirs:  []string{".git"},
		Port:         "8425",
	}

	svc := NewService(cfg)

	// Test GetIssue
	iss, err := svc.GetIssue("test-123")
	if err != nil {
		t.Fatalf("GetIssue failed: %v", err)
	}
	if iss.Title != "Test Issue Title" {
		t.Errorf("expected title 'Test Issue Title', got %q", iss.Title)
	}
	if iss.Description != "Full test description text" {
		t.Errorf("expected description 'Full test description text', got %q", iss.Description)
	}
	if iss.Project != "sample-repo" {
		t.Errorf("expected project 'sample-repo', got %q", iss.Project)
	}

	// Test Nonexistent issue
	_, err = svc.GetIssue("non-existent-999")
	if err == nil {
		t.Errorf("expected error for nonexistent issue, got nil")
	}
}
