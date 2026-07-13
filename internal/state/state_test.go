package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/format"
)

func TestValidateReleaseName(t *testing.T) {
	if err := ValidateReleaseName("myapp-1"); err != nil {
		t.Fatalf("expected valid release name: %v", err)
	}
	if err := ValidateReleaseName("../bad"); err == nil {
		t.Fatal("expected invalid release name")
	}
}

func TestHomeUsesExplicitThenEnvironment(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), "explicit", "..", "dockyard")
	got, err := Home(explicit)
	if err != nil {
		t.Fatalf("unexpected explicit home error: %v", err)
	}
	want, err := filepath.Abs(filepath.Clean(explicit))
	if err != nil {
		t.Fatalf("resolve expected explicit home: %v", err)
	}
	if got != want {
		t.Fatalf("expected explicit home %q, got %q", want, got)
	}

	envHome := filepath.Join(t.TempDir(), "env-home")
	t.Setenv(EnvDockyardHome, envHome)
	got, err = Home("")
	if err != nil {
		t.Fatalf("unexpected env home error: %v", err)
	}
	want, err = filepath.Abs(filepath.Clean(envHome))
	if err != nil {
		t.Fatalf("resolve expected env home: %v", err)
	}
	if got != want {
		t.Fatalf("expected env home %q, got %q", want, got)
	}
}

func TestWriteAndReadReleaseDefaultsAPIVersion(t *testing.T) {
	dir := t.TempDir()
	release := Release{
		Name:           "web",
		PackageName:    "nginx",
		PackageVersion: "0.1.0",
		AppVersion:     "1.27",
		Revision:       2,
		Status:         "deployed",
		CreatedAt:      time.Unix(10, 0).UTC(),
		UpdatedAt:      time.Unix(20, 0).UTC(),
		ComposeProject: "web",
		Source:         Source{Type: "local", Path: "."},
	}

	if err := WriteRelease(dir, release); err != nil {
		t.Fatalf("write release: %v", err)
	}
	got, err := ReadRelease(dir)
	if err != nil {
		t.Fatalf("read release: %v", err)
	}

	if got.APIVersion != format.ReleaseAPIVersion {
		t.Fatalf("expected apiVersion %q, got %q", format.ReleaseAPIVersion, got.APIVersion)
	}
	if got.Name != release.Name || got.Revision != release.Revision || got.Source.Path != release.Source.Path {
		t.Fatalf("unexpected release after round trip: %#v", got)
	}
}

func TestReadReleaseRejectsUnsupportedAPIVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "release.json")
	if err := os.WriteFile(path, []byte(`{"apiVersion":"dockyard.io/v99","name":"web"}`), 0o600); err != nil {
		t.Fatalf("write release metadata: %v", err)
	}

	if _, err := ReadRelease(dir); err == nil {
		t.Fatal("expected unsupported apiVersion error")
	}
}

func TestCurrentAndNextRevision(t *testing.T) {
	home := t.TempDir()
	if got, err := NextRevision(home, "web"); err != nil || got != 1 {
		t.Fatalf("expected first revision 1, got %d, err %v", got, err)
	}
	if err := os.MkdirAll(RevisionDir(home, "web", 1), 0o700); err != nil {
		t.Fatalf("create revision 1: %v", err)
	}
	if err := os.MkdirAll(RevisionDir(home, "web", 3), 0o700); err != nil {
		t.Fatalf("create revision 3: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ReleaseDir(home, "web"), "revisions", "note.txt"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write non-directory revision entry: %v", err)
	}

	if got, err := NextRevision(home, "web"); err != nil || got != 4 {
		t.Fatalf("expected next revision 4, got %d, err %v", got, err)
	}
	if err := SetCurrentRevision(home, "web", 3); err != nil {
		t.Fatalf("set current revision: %v", err)
	}
	if got, err := ReadCurrentRevision(home, "web"); err != nil || got != 3 {
		t.Fatalf("expected current revision 3, got %d, err %v", got, err)
	}
}
