package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/state"
)

func TestCollectReleaseListRowsHidesUninstalledByDefault(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForList(t, home, "active", "deployed")
	writeCurrentReleaseForList(t, home, "old", "uninstalled")

	rows, err := collectReleaseListRows(home, listOptions{})
	if err != nil {
		t.Fatalf("collect release list rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 active release, got %d: %#v", len(rows), rows)
	}
	if rows[0].Name != "active" {
		t.Fatalf("unexpected release row: %#v", rows[0])
	}
}

func TestCollectReleaseListRowsAllIncludesUninstalled(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForList(t, home, "active", "deployed")
	writeCurrentReleaseForList(t, home, "old", "uninstalled")

	rows, err := collectReleaseListRows(home, listOptions{all: true})
	if err != nil {
		t.Fatalf("collect release list rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 releases, got %d: %#v", len(rows), rows)
	}
	if rows[0].Name != "active" || rows[1].Name != "old" {
		t.Fatalf("expected sorted release rows, got %#v", rows)
	}
}

func TestCollectReleaseListRowsStatusFilter(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForList(t, home, "active", "deployed")
	writeCurrentReleaseForList(t, home, "old", "uninstalled")

	rows, err := collectReleaseListRows(home, listOptions{status: "uninstalled"})
	if err != nil {
		t.Fatalf("collect release list rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 uninstalled release, got %d: %#v", len(rows), rows)
	}
	if rows[0].Name != "old" || rows[0].Status != "uninstalled" {
		t.Fatalf("unexpected release row: %#v", rows[0])
	}
}

func writeCurrentReleaseForList(t *testing.T, home string, releaseName string, status string) {
	t.Helper()
	revisionDir := state.RevisionDir(home, releaseName, 1)
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		t.Fatalf("create revision dir: %v", err)
	}
	release := state.Release{
		Name:           releaseName,
		PackageName:    "example-app",
		PackageVersion: "0.1.0",
		Revision:       1,
		Status:         status,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		ComposeProject: releaseName,
		Source:         state.Source{Type: "local", Path: filepath.Join("examples", "example-app")},
	}
	if err := state.WriteRelease(revisionDir, release); err != nil {
		t.Fatalf("write release: %v", err)
	}
	if err := state.SetCurrentRevision(home, releaseName, 1); err != nil {
		t.Fatalf("write current revision: %v", err)
	}
}
