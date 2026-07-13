package lock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/format"
)

func TestExtractImages(t *testing.T) {
	images, err := ExtractImages([]byte(`services:
  web:
    image: nginx:1.27
  db:
    image: postgres@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	if images[0].Service != "db" {
		t.Fatalf("expected sorted services")
	}
	if images[0].Digest == "" {
		t.Fatalf("expected digest extraction")
	}
}

func TestLockDependenciesSortsByAliasOrName(t *testing.T) {
	deps := lockDependencies([]dockpkg.Dependency{
		{Name: "redis", Source: "oci://example.test/redis:0.1.0"},
		{Name: "postgres", Alias: "db", Source: "oci://example.test/postgres:0.1.0", Version: "0.1.0"},
	})
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(deps))
	}
	if deps[0].Alias != "db" || deps[0].Name != "postgres" {
		t.Fatalf("expected dependency with alias db first, got %#v", deps[0])
	}
	if deps[1].Name != "redis" {
		t.Fatalf("expected redis second, got %#v", deps[1])
	}
}

func TestExtractImagesIgnoresServicesWithoutImages(t *testing.T) {
	images, err := ExtractImages([]byte(`services:
  worker:
    build: .
  api:
    image: ghcr.io/example/api:1.0.0
  db:
    image: ""
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d: %#v", len(images), images)
	}
	if images[0].Service != "api" || images[0].Image != "ghcr.io/example/api:1.0.0" {
		t.Fatalf("unexpected image extraction: %#v", images[0])
	}
}

func TestVerifyDetectsDigestMismatches(t *testing.T) {
	packageDir := t.TempDir()
	writeLockTestFile(t, filepath.Join(packageDir, "Dockyard.yaml"), "name: app\n")
	values := map[string]any{"image": "nginx"}
	rendered := []byte("services: {}\n")

	lf := &Lockfile{
		APIVersion:            format.LockfileAPIVersion,
		GeneratedAt:           time.Unix(0, 0).UTC(),
		PackageName:           "app",
		PackageVersion:        "0.1.0",
		ValuesSHA256:          sha256Bytes([]byte("wrong values")),
		RenderedComposeSHA256: sha256Bytes(rendered),
		Files: []FileDigest{
			{Path: "Dockyard.yaml", SHA256: sha256Bytes([]byte("name: app\n"))},
		},
	}
	if err := Write(filepath.Join(packageDir, FileName), lf); err != nil {
		t.Fatal(err)
	}

	err := Verify(packageDir, rendered, values, "")
	if err == nil {
		t.Fatal("expected values digest mismatch")
	}
	if !strings.Contains(err.Error(), "values digest mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}

	valuesDigest, err := ValuesDigest(values)
	if err != nil {
		t.Fatal(err)
	}
	lf.ValuesSHA256 = valuesDigest
	lf.RenderedComposeSHA256 = sha256Bytes([]byte("wrong compose"))
	if err := Write(filepath.Join(packageDir, FileName), lf); err != nil {
		t.Fatal(err)
	}

	err = Verify(packageDir, rendered, values, "")
	if err == nil {
		t.Fatal("expected rendered compose digest mismatch")
	}
	if !strings.Contains(err.Error(), "rendered compose digest mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyDetectsOverlayAndFileDigestMismatches(t *testing.T) {
	packageDir := t.TempDir()
	writeLockTestFile(t, filepath.Join(packageDir, "Dockyard.yaml"), "name: app\n")
	values := map[string]any{"image": "nginx"}
	rendered := []byte("services: {}\n")
	valuesDigest, err := ValuesDigest(values)
	if err != nil {
		t.Fatal(err)
	}

	lf := &Lockfile{
		APIVersion:            format.LockfileAPIVersion,
		GeneratedAt:           time.Unix(0, 0).UTC(),
		PackageName:           "app",
		PackageVersion:        "0.1.0",
		Overlay:               "prod",
		ValuesSHA256:          valuesDigest,
		RenderedComposeSHA256: sha256Bytes(rendered),
		Files: []FileDigest{
			{Path: "Dockyard.yaml", SHA256: sha256Bytes([]byte("name: app\n"))},
		},
	}
	if err := Write(filepath.Join(packageDir, FileName), lf); err != nil {
		t.Fatal(err)
	}

	err = Verify(packageDir, rendered, values, "dev")
	if err == nil {
		t.Fatal("expected overlay mismatch")
	}
	if !strings.Contains(err.Error(), "overlay mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}

	lf.Overlay = "dev"
	lf.Files[0].SHA256 = sha256Bytes([]byte("wrong file"))
	if err := Write(filepath.Join(packageDir, FileName), lf); err != nil {
		t.Fatal(err)
	}

	err = Verify(packageDir, rendered, values, "dev")
	if err == nil {
		t.Fatal("expected file digest mismatch")
	}
	if !strings.Contains(err.Error(), "file digest mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeLockTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
