package cli

import (
	"os"
	"testing"

	"github.com/nandub/dockyard/internal/state"
)

func TestPruneReleaseKeepsNewestAndCurrentRevision(t *testing.T) {
	home := t.TempDir()
	for _, revision := range []int{1, 2, 3, 4} {
		if err := os.MkdirAll(state.RevisionDir(home, "web", revision), 0o700); err != nil {
			t.Fatalf("create revision %d: %v", revision, err)
		}
	}
	if err := state.SetCurrentRevision(home, "web", 2); err != nil {
		t.Fatalf("set current revision: %v", err)
	}

	if err := pruneRelease(home, "web", 1, false); err != nil {
		t.Fatalf("prune release: %v", err)
	}

	assertRevisionExists(t, home, "web", 2)
	assertRevisionExists(t, home, "web", 4)
	assertRevisionMissing(t, home, "web", 1)
	assertRevisionMissing(t, home, "web", 3)
}

func TestPruneReleaseDryRunDoesNotRemoveRevisions(t *testing.T) {
	home := t.TempDir()
	for _, revision := range []int{1, 2, 3} {
		if err := os.MkdirAll(state.RevisionDir(home, "api", revision), 0o700); err != nil {
			t.Fatalf("create revision %d: %v", revision, err)
		}
	}

	if err := pruneRelease(home, "api", 1, true); err != nil {
		t.Fatalf("dry-run prune release: %v", err)
	}

	for _, revision := range []int{1, 2, 3} {
		assertRevisionExists(t, home, "api", revision)
	}
}

func assertRevisionExists(t *testing.T, home string, releaseName string, revision int) {
	t.Helper()
	if _, err := os.Stat(state.RevisionDir(home, releaseName, revision)); err != nil {
		t.Fatalf("expected revision %d to exist: %v", revision, err)
	}
}

func assertRevisionMissing(t *testing.T, home string, releaseName string, revision int) {
	t.Helper()
	if _, err := os.Stat(state.RevisionDir(home, releaseName, revision)); !os.IsNotExist(err) {
		t.Fatalf("expected revision %d to be removed, stat err: %v", revision, err)
	}
}
