package oci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeReferenceRequiresScheme(t *testing.T) {
	if _, err := NormalizeReference("ghcr.io/nandub/dockyard/app:0.1.0"); err == nil {
		t.Fatal("expected missing scheme error")
	}
}

func TestNormalizeReferenceRequiresTagOrDigest(t *testing.T) {
	if _, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app"); err == nil {
		t.Fatal("expected missing tag or digest error")
	}
}

func TestNormalizeReferenceAcceptsTag(t *testing.T) {
	got, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app:0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ghcr.io/nandub/dockyard/app:0.1.0" {
		t.Fatalf("unexpected normalized reference %q", got)
	}
}

func TestNormalizeReferenceTrimsOuterWhitespaceAndRejectsInnerWhitespace(t *testing.T) {
	got, err := NormalizeReference("oci:// ghcr.io/nandub/dockyard/app:0.1.0 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ghcr.io/nandub/dockyard/app:0.1.0" {
		t.Fatalf("unexpected normalized reference %q", got)
	}

	if _, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app:0.1.0 extra"); err == nil {
		t.Fatal("expected whitespace error")
	}
}

func TestNormalizeReferenceAcceptsDigest(t *testing.T) {
	_, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPushArgsUsesDockyardArtifactType(t *testing.T) {
	got := PushArgs("ghcr.io/nandub/dockyard/nginx:0.1.0", "nginx-0.1.0.dockyard.tgz")
	want := []string{
		"push",
		"--artifact-type",
		ArtifactType,
		"ghcr.io/nandub/dockyard/nginx:0.1.0",
		"nginx-0.1.0.dockyard.tgz:" + LayerMediaType,
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d args, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestFindPulledArchiveFindsSingleSupportedArchive(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	archive := filepath.Join(nested, "nginx-0.1.0.dockyard.tgz")
	if err := os.WriteFile(archive, []byte("archive"), 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}

	got, err := findPulledArchive(dir)
	if err != nil {
		t.Fatalf("unexpected archive scan error: %v", err)
	}
	if got != archive {
		t.Fatalf("expected archive %q, got %q", archive, got)
	}
}

func TestFindPulledArchiveRequiresExactlyOneArchive(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o600); err != nil {
			t.Fatalf("write ignored file: %v", err)
		}
		if _, err := findPulledArchive(dir); err == nil {
			t.Fatal("expected missing archive error")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"a.tgz", "b.tar.gz"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("archive"), 0o600); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		if _, err := findPulledArchive(dir); err == nil {
			t.Fatal("expected multiple archive error")
		}
	})
}
