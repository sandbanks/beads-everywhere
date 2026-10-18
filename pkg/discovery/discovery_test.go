package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandbanks/beads-everywhere/pkg/config"
)

func TestFindRepositoriesStopsAtTopLevelBeads(t *testing.T) {
	root := t.TempDir()

	// Real project: top-level .beads plus nested .beads folders
	// (test fixtures, sample db files, temp dirs) that must be ignored.
	repo := filepath.Join(root, "beads_rust")
	mustMkdirAll(t, filepath.Join(repo, ".beads"))
	mustMkdirAll(t, filepath.Join(repo, "tests", "fixtures", "workspace_failures", "corrupt_db_text", ".beads"))
	mustMkdirAll(t, filepath.Join(repo, "tests", "fixtures", "workspace_failures", "db_jsonl_disagreement", ".beads"))
	mustMkdirAll(t, filepath.Join(repo, "sample_beads_db_files", "repro_x", ".beads"))
	mustMkdirAll(t, filepath.Join(repo, "temp_test", ".beads"))

	// Sibling project nested deeper under the root must still be found.
	sibling := filepath.Join(root, "rust", "agentic_ssh")
	mustMkdirAll(t, filepath.Join(sibling, ".beads"))

	d := NewDiscoverer(&config.Config{
		ScanRoots:   []string{root},
		IgnoredDirs: []string{".git", "node_modules", "target", "vendor"},
	})

	projects, err := d.FindRepositories()
	if err != nil {
		t.Fatalf("FindRepositories: %v", err)
	}

	got := make(map[string]bool)
	for _, p := range projects {
		got[p.Path] = true
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d: %v", len(projects), projects)
	}
	if !got[repo] {
		t.Errorf("missing top-level project %s", repo)
	}
	if !got[sibling] {
		t.Errorf("missing sibling project %s", sibling)
	}

	for _, nested := range []string{
		filepath.Join(repo, "tests", "fixtures", "workspace_failures", "corrupt_db_text"),
		filepath.Join(repo, "tests", "fixtures", "workspace_failures", "db_jsonl_disagreement"),
		filepath.Join(repo, "sample_beads_db_files", "repro_x"),
		filepath.Join(repo, "temp_test"),
	} {
		if got[nested] {
			t.Errorf("nested .beads under %s must not be registered as a project", nested)
		}
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}