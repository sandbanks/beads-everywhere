package templates

import (
	"bytes"
	"testing"

	"github.com/sandbanks/beads-everywhere/pkg/models"
)

func TestRenderIssueDetailModal(t *testing.T) {
	issue := models.Issue{
		ID:          "test-detail-42",
		Project:     "test-project",
		ProjectPath: "/path/to/test-project",
		Title:       "Detailed Test Issue",
		Description: "A multi-line\ndescription with\ncontext.",
		Status:      "open",
		Priority:    1,
		IssueType:   "feature",
		CreatedAt:   "2026-09-14T12:00:00Z",
	}

	var buf bytes.Buffer
	err := Render(&buf, "issue_detail_modal.html", issue)
	if err != nil {
		t.Fatalf("failed to render issue_detail_modal.html: %v", err)
	}

	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Detailed Test Issue")) {
		t.Errorf("rendered output missing title, got:\n%s", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("test-detail-42")) {
		t.Errorf("rendered output missing ID, got:\n%s", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("A multi-line\ndescription with\ncontext.")) {
		t.Errorf("rendered output missing description, got:\n%s", out)
	}
}

func TestRenderIssueCard(t *testing.T) {
	issue := models.Issue{
		ID:          "test-card-1",
		Project:     "test-project",
		ProjectPath: "/path/to/test-project",
		Title:       "Card Test Issue",
		Description: "Card description text",
		Status:      "in_progress",
		Priority:    0,
		IssueType:   "bug",
	}

	var buf bytes.Buffer
	err := Render(&buf, "issue_card.html", issue)
	if err != nil {
		t.Fatalf("failed to render issue_card.html: %v", err)
	}

	out := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("hx-get=\"/issues/test-card-1\"")) {
		t.Errorf("rendered issue card missing detail link hx-get='/issues/test-card-1', got:\n%s", out)
	}
}
