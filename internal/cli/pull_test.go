package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/format"
)

func TestDefaultPulledArchiveNameUsesManifestIdentity(t *testing.T) {
	packageDir := t.TempDir()
	writeCLITestFile(t, packageDir, "Dockyard.yaml", `apiVersion: `+format.ManifestAPIVersion+`
name: redis
version: 0.3.0
appVersion: "7"
compose:
  base: compose.yaml
`)
	writeCLITestFile(t, packageDir, "compose.yaml", "services:\n  redis:\n    image: redis:7\n")
	writeCLITestFile(t, packageDir, "values.yaml", "{}\n")

	archivePath := filepath.Join(t.TempDir(), "download.tgz")
	if _, err := archive.PackageDir(packageDir, archivePath, false); err != nil {
		t.Fatalf("package fixture: %v", err)
	}

	got, err := defaultPulledArchiveName(archivePath)
	if err != nil {
		t.Fatalf("default pulled archive name: %v", err)
	}
	if got != "redis-0.3.0.dockyard.tgz" {
		t.Fatalf("unexpected archive name %q", got)
	}
}

func TestCopyPulledArchiveRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.tgz")
	dst := filepath.Join(dir, "dst.tgz")
	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatalf("write source archive: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o600); err != nil {
		t.Fatalf("write destination archive: %v", err)
	}

	if err := copyPulledArchive(dst, src); err == nil {
		t.Fatal("expected overwrite refusal")
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read destination archive: %v", err)
	}
	if string(data) != "old" {
		t.Fatalf("destination was overwritten: %q", data)
	}
}

func TestCopyPulledArchiveWritesNewFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.tgz")
	dst := filepath.Join(dir, "dst.tgz")
	if err := os.WriteFile(src, []byte("archive"), 0o600); err != nil {
		t.Fatalf("write source archive: %v", err)
	}

	if err := copyPulledArchive(dst, src); err != nil {
		t.Fatalf("copy pulled archive: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read destination archive: %v", err)
	}
	if string(data) != "archive" {
		t.Fatalf("unexpected copied archive content %q", data)
	}
}

func writeCLITestFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create parent for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
