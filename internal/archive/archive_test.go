package archive

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractArchiveRejectsPathTraversal(t *testing.T) {
	archivePath := writeTestArchive(t, map[string]string{
		"../escape.txt": "nope",
	})

	err := ExtractArchive(archivePath, filepath.Join(t.TempDir(), "extract"))
	if err == nil {
		t.Fatal("expected path traversal archive to be rejected")
	}
	if !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func TestVerifyArchiveRejectsExtraFileNotListedInSHA256SUMS(t *testing.T) {
	files := map[string]string{
		"Dockyard.yaml": testManifest(),
		"compose.yaml":  "services: {}\n",
		"extra.txt":     "unexpected\n",
	}
	sums := sha256Bytes([]byte(files["Dockyard.yaml"])) + "  Dockyard.yaml\n" +
		sha256Bytes([]byte(files["compose.yaml"])) + "  compose.yaml\n"
	files["SHA256SUMS"] = sums

	err := VerifyArchive(writeTestArchive(t, files), nil)
	if err == nil {
		t.Fatal("expected archive with unlisted file to be rejected")
	}
	if !strings.Contains(err.Error(), "not listed in SHA256SUMS") {
		t.Fatalf("expected unlisted file error, got %v", err)
	}
}

func TestPackageDirRejectsForbiddenSensitiveFiles(t *testing.T) {
	tests := []string{
		".env",
		"id_rsa",
		"tls.key",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			packageDir := writePackageDir(t)
			writeFile(t, filepath.Join(packageDir, name), "secret\n")

			_, err := PackageDir(packageDir, filepath.Join(t.TempDir(), "out.dockyard.tgz"), false)
			if err == nil {
				t.Fatal("expected forbidden file to be rejected")
			}
			if !strings.Contains(err.Error(), "forbidden") {
				t.Fatalf("expected forbidden file error, got %v", err)
			}
		})
	}
}

func TestPackageDirIncludesGeneratedMetadataAndFiltersStaleGeneratedFiles(t *testing.T) {
	packageDir := writePackageDir(t)
	writeFile(t, filepath.Join(packageDir, "values.yaml"), "{}\n")
	writeFile(t, filepath.Join(packageDir, "SHA256SUMS"), "stale\n")
	writeFile(t, filepath.Join(packageDir, "package.provenance.json"), "stale\n")

	archivePath, err := PackageDir(packageDir, filepath.Join(t.TempDir(), "out.dockyard.tgz"), false)
	if err != nil {
		t.Fatalf("package dir: %v", err)
	}
	extractDir := t.TempDir()
	if err := ExtractArchive(archivePath, extractDir); err != nil {
		t.Fatalf("extract package: %v", err)
	}

	sums, err := os.ReadFile(filepath.Join(extractDir, "SHA256SUMS"))
	if err != nil {
		t.Fatalf("read generated sums: %v", err)
	}
	if strings.Contains(string(sums), "stale") || !strings.Contains(string(sums), "package.provenance.json") {
		t.Fatalf("unexpected generated sums:\n%s", sums)
	}

	data, err := os.ReadFile(filepath.Join(extractDir, "package.provenance.json"))
	if err != nil {
		t.Fatalf("read provenance: %v", err)
	}
	var provenance Provenance
	if err := json.Unmarshal(data, &provenance); err != nil {
		t.Fatalf("parse provenance: %v", err)
	}
	if provenance.PackageName != "test-package" || provenance.PackageVersion != "0.1.0" || provenance.Builder != "dockyard" {
		t.Fatalf("unexpected provenance: %#v", provenance)
	}
}

func TestPackageDirRequireLockRejectsMissingLockfile(t *testing.T) {
	_, err := PackageDir(writePackageDir(t), filepath.Join(t.TempDir(), "out.dockyard.tgz"), true)
	if err == nil {
		t.Fatal("expected missing lockfile error")
	}
	if !strings.Contains(err.Error(), "--locked requires dockyard.lock") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writePackageDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Dockyard.yaml"), testManifest())
	writeFile(t, filepath.Join(dir, "compose.yaml"), "services: {}\n")
	return dir
}

func writeTestArchive(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "package.dockyard.tgz")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for name, content := range files {
		data := []byte(content)
		header := &tar.Header{
			Name:    name,
			Mode:    0o600,
			Size:    int64(len(data)),
			ModTime: time.Unix(0, 0).UTC(),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatalf("write tar body: %v", err)
		}
	}
	return path
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func testManifest() string {
	return `apiVersion: dockyard.dev/v1alpha1
name: test-package
version: 0.1.0
type: application
compose:
  base: compose.yaml
  overlays: {}
`
}
