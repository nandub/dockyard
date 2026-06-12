package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/state"
)

func TestActiveDependentsForReleaseFindsActiveParents(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForUninstall(t, home, "team-dashboard-db", "deployed", nil)
	writeCurrentReleaseForUninstall(t, home, "team-dashboard", "deployed", []state.ReleaseDependency{
		{Name: "postgres", Alias: "db", Release: "team-dashboard-db"},
	})

	dependents, err := activeDependentsForRelease(home, "team-dashboard-db")
	if err != nil {
		t.Fatalf("find active dependents: %v", err)
	}

	if len(dependents) != 1 {
		t.Fatalf("expected 1 dependent, got %d: %#v", len(dependents), dependents)
	}
	if dependents[0] != "team-dashboard (postgres as db)" {
		t.Fatalf("unexpected dependent label: %q", dependents[0])
	}
}

func TestActiveDependentsForReleaseIgnoresUninstalledParents(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForUninstall(t, home, "team-dashboard-db", "deployed", nil)
	writeCurrentReleaseForUninstall(t, home, "team-dashboard", "uninstalled", []state.ReleaseDependency{
		{Name: "postgres", Alias: "db", Release: "team-dashboard-db"},
	})

	dependents, err := activeDependentsForRelease(home, "team-dashboard-db")
	if err != nil {
		t.Fatalf("find active dependents: %v", err)
	}

	if len(dependents) != 0 {
		t.Fatalf("expected no active dependents, got %#v", dependents)
	}
}

func TestDependencyUninstallBlockedError(t *testing.T) {
	err := dependencyUninstallBlockedError("team-dashboard-db", []string{"team-dashboard (postgres as db)"})
	if err == nil {
		t.Fatal("expected dependency uninstall error")
	}
	msg := err.Error()
	for _, want := range []string{
		`release "team-dashboard-db" is still required`,
		"team-dashboard (postgres as db)",
		"uninstall dependent release(s) first",
		"--force",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected error %q to contain %q", msg, want)
		}
	}
}

func writeCurrentReleaseForUninstall(t *testing.T, home string, releaseName string, status string, deps []state.ReleaseDependency) {
	t.Helper()
	revisionDir := state.RevisionDir(home, releaseName, 1)
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		t.Fatalf("create revision dir: %v", err)
	}
	release := state.Release{
		Name:           releaseName,
		PackageName:    releaseName,
		PackageVersion: "0.1.0",
		Revision:       1,
		Status:         status,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		ComposeProject: releaseName,
		Source:         state.Source{Type: "local", Path: filepath.Join("examples", releaseName)},
		Dependencies:   deps,
	}
	if err := state.WriteRelease(revisionDir, release); err != nil {
		t.Fatalf("write release: %v", err)
	}
	if err := state.SetCurrentRevision(home, releaseName, 1); err != nil {
		t.Fatalf("write current revision: %v", err)
	}
}
