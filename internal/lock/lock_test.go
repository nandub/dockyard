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

func TestNewWriteAndReadLockfile(t *testing.T) {
	packageDir := t.TempDir()
	writeLockTestFile(t, filepath.Join(packageDir, "Dockyard.yaml"), "name: app\n")
	writeLockTestFile(t, filepath.Join(packageDir, "compose.yaml"), "services: {}\n")
	writeLockTestFile(t, filepath.Join(packageDir, "values.yaml"), "{}\n")
	writeLockTestFile(t, filepath.Join(packageDir, "SHA256SUMS"), "ignored\n")
	writeLockTestFile(t, filepath.Join(packageDir, "package.provenance.json"), "ignored\n")
	if err := os.MkdirAll(filepath.Join(packageDir, ".git"), 0o700); err != nil {
		t.Fatalf("create .git dir: %v", err)
	}
	writeLockTestFile(t, filepath.Join(packageDir, ".git", "config"), "ignored\n")

	manifest := &dockpkg.Manifest{
		Name:       "app",
		Version:    "0.1.0",
		AppVersion: "1.0.0",
		Dependencies: []dockpkg.Dependency{
			{Name: "postgres", Alias: "db", Source: "oci://example.test/postgres:0.1.0", Version: "0.1.0"},
		},
	}
	rendered := []byte(`services:
  web:
    image: nginx@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`)

	lf, err := New(packageDir, manifest, map[string]any{"replicas": 2}, rendered, "prod")
	if err != nil {
		t.Fatalf("new lockfile: %v", err)
	}
	if lf.APIVersion != format.LockfileAPIVersion || lf.PackageName != "app" || lf.Overlay != "prod" {
		t.Fatalf("unexpected lockfile metadata: %#v", lf)
	}
	if len(lf.Images) != 1 || lf.Images[0].Digest != "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected image locks: %#v", lf.Images)
	}
	if len(lf.Dependencies) != 1 || lf.Dependencies[0].Alias != "db" {
		t.Fatalf("unexpected dependency locks: %#v", lf.Dependencies)
	}
	for _, file := range lf.Files {
		if file.Path == "SHA256SUMS" || file.Path == "package.provenance.json" || strings.HasPrefix(file.Path, ".git/") {
			t.Fatalf("generated or ignored file was locked: %#v", file)
		}
	}

	path := filepath.Join(packageDir, FileName)
	if err := Write(path, lf); err != nil {
		t.Fatalf("write lockfile: %v", err)
	}
	read, err := Read(path)
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	if read.PackageName != lf.PackageName || read.ValuesSHA256 != lf.ValuesSHA256 {
		t.Fatalf("unexpected lockfile round trip: %#v", read)
	}
}

func TestReadRejectsInvalidLockfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	writeLockTestFile(t, path, `{"apiVersion":"dockyard.dev/lockfile/v9"}`)
	_, err := Read(path)
	if err == nil {
		t.Fatal("expected unsupported lockfile version")
	}
	if !strings.Contains(err.Error(), "unsupported lockfile apiVersion") {
		t.Fatalf("unexpected error: %v", err)
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

func TestMergeValuesForLockAppliesOverride(t *testing.T) {
	packageDir := t.TempDir()
	writeLockTestFile(t, filepath.Join(packageDir, "values.yaml"), `app:
  tag: "1.0"
  replicas: 1
`)
	override := filepath.Join(packageDir, "override.yaml")
	writeLockTestFile(t, override, `app:
  replicas: 3
`)

	vals, err := MergeValuesForLock(packageDir, override)
	if err != nil {
		t.Fatalf("merge values for lock: %v", err)
	}
	app := vals["app"].(map[string]any)
	if app["tag"] != "1.0" || app["replicas"] != 3 {
		t.Fatalf("unexpected merged values: %#v", vals)
	}
}

func writeLockTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
