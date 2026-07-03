package cli

import (
	"os"
	"path/filepath"
	"testing"
)

const testCatalogIndex = `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: redis
    latest: 0.1.0
    description: Redis-compatible in-memory data store.
    versions:
      - 0.1.0
  - name: postgres
    latest: 0.1.0
    description: PostgreSQL relational database.
    versions:
      - 0.1.0
`

func useTestCatalog(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog.yaml")
	if err := os.WriteFile(path, []byte(testCatalogIndex), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOCKYARD_CATALOG", path)
}

func TestResolveCatalogPackageSourceBareName(t *testing.T) {
	useTestCatalog(t)
	got, ok, err := resolveCatalogPackageSource("redis")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog shorthand to resolve")
	}
	want := "oci://ghcr.io/example/packages/redis:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveCatalogPackageSourceKeepsUnknownBareName(t *testing.T) {
	useTestCatalog(t)
	got, ok, err := resolveCatalogPackageSource("custom-local-name")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected catalog resolution")
	}
	if got != "custom-local-name" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveCatalogPackageSourceCatalogURL(t *testing.T) {
	useTestCatalog(t)
	got, ok, err := resolveCatalogPackageSource("catalog://postgres:0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog URL to resolve")
	}
	want := "oci://ghcr.io/example/packages/postgres:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
