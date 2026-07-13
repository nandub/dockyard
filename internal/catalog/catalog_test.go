package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testIndex = `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: redis
    latest: 0.1.0
    description: Redis-compatible in-memory data store.
    source: oci://ghcr.io/example/packages/redis
    versions:
      - 0.1.0
  - name: postgres
    latest: 0.2.0
    description: PostgreSQL relational database.
    versions:
      - 0.1.0
      - 0.2.0
`

func TestReferenceDefaultsWhenEnvUnset(t *testing.T) {
	t.Setenv(EnvCatalog, "")

	if got := Reference(); got != DefaultCatalogRef {
		t.Fatalf("got %q want %q", got, DefaultCatalogRef)
	}
}

func TestReferenceNormalizesRegistryPrefix(t *testing.T) {
	t.Setenv(EnvCatalog, "ghcr.io/example/packages/")

	want := "oci://ghcr.io/example/packages/catalog:latest"
	if got := Reference(); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReferenceCleansLocalYAMLPath(t *testing.T) {
	dirty := filepath.Join(t.TempDir(), "nested", "..", "catalog.yaml")
	t.Setenv(EnvCatalog, dirty)

	want := filepath.Clean(dirty)
	if got := Reference(); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveCatalogURLDefaultVersion(t *testing.T) {
	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := ResolveWithIndex(idx, "catalog://redis")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog source")
	}
	want := "oci://ghcr.io/example/packages/redis:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveBareCatalogName(t *testing.T) {
	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := ResolveWithIndex(idx, "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog source")
	}
	want := "oci://ghcr.io/example/packages/postgres:0.2.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveUnknownBareNameIsUnchanged(t *testing.T) {
	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := ResolveWithIndex(idx, "not-a-package")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("unexpected catalog source")
	}
	if got != "not-a-package" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveUnknownCatalogURLFails(t *testing.T) {
	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err := ResolveWithIndex(idx, "catalog://not-a-package")
	if !ok {
		t.Fatal("expected catalog source")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRejectsUnknownVersion(t *testing.T) {
	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	_, err = idx.ResolveName("postgres", "9.9.9")
	if err == nil {
		t.Fatal("expected version error")
	}
}

func TestLoadBytesRejectsInvalidCatalogs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name: "unsupported api version",
			input: `apiVersion: dockyard.dev/catalog/v9
registry: ghcr.io/example/packages
packages: []
`,
			wantErr: "unsupported catalog apiVersion",
		},
		{
			name: "missing registry",
			input: `apiVersion: dockyard.dev/catalog/v1alpha1
packages: []
`,
			wantErr: "catalog registry is required",
		},
		{
			name: "duplicate package",
			input: `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: redis
    latest: 0.1.0
    description: Redis one.
  - name: redis
    latest: 0.2.0
    description: Redis two.
`,
			wantErr: `duplicate catalog package "redis"`,
		},
		{
			name: "invalid package name",
			input: `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: Redis
    latest: 0.1.0
    description: Redis.
`,
			wantErr: `invalid catalog package name "Redis"`,
		},
		{
			name: "missing latest",
			input: `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: redis
    description: Redis.
`,
			wantErr: `catalog package "redis" is missing latest version`,
		},
		{
			name: "missing description",
			input: `apiVersion: dockyard.dev/catalog/v1alpha1
registry: ghcr.io/example/packages
packages:
  - name: redis
    latest: 0.1.0
`,
			wantErr: `catalog package "redis" is missing description`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadBytes([]byte(tt.input))
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got error %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestListSortedFromFile(t *testing.T) {
	path := t.TempDir() + "/catalog.yaml"
	if err := os.WriteFile(path, []byte(testIndex), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCatalog, path)
	pkgs, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages", len(pkgs))
	}
	for i := 1; i < len(pkgs); i++ {
		if pkgs[i-1].Name > pkgs[i].Name {
			t.Fatalf("packages are not sorted: %s before %s", pkgs[i-1].Name, pkgs[i].Name)
		}
	}
}

func TestGetAndResolveNameFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.yaml")
	if err := os.WriteFile(path, []byte(testIndex), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCatalog, "file://"+path)

	pkg, ok, err := Get("redis")
	if err != nil {
		t.Fatalf("get catalog package: %v", err)
	}
	if !ok || pkg.Name != "redis" {
		t.Fatalf("unexpected package lookup: ok=%v pkg=%#v", ok, pkg)
	}
	source, err := ResolveName("postgres", "0.1.0")
	if err != nil {
		t.Fatalf("resolve catalog package: %v", err)
	}
	if source != "oci://ghcr.io/example/packages/postgres:0.1.0" {
		t.Fatalf("unexpected source: %s", source)
	}
}

func TestLoadReferenceRejectsInvalidReferenceAndFindsIndexFile(t *testing.T) {
	if _, err := LoadReference(nil, "not-oci"); err == nil {
		t.Fatal("expected invalid catalog reference error")
	}

	dir := t.TempDir()
	if _, err := findIndexFile(dir); err == nil {
		t.Fatal("expected missing catalog index error")
	}
	path := filepath.Join(dir, "catalog.yml")
	if err := os.WriteFile(path, []byte(testIndex), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := findIndexFile(dir)
	if err != nil {
		t.Fatalf("find catalog index: %v", err)
	}
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func TestCatalogCacheRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	idx, err := LoadBytes([]byte(testIndex))
	if err != nil {
		t.Fatal(err)
	}
	ref := "oci://ghcr.io/example/packages/catalog:latest"
	if _, ok := readCached(ref); ok {
		t.Fatal("did not expect cache before write")
	}
	if err := writeCached(ref, idx); err != nil {
		t.Fatalf("write cached catalog: %v", err)
	}
	cached, ok := readCached(ref)
	if !ok {
		t.Fatal("expected cached catalog")
	}
	if len(cached.Packages) != len(idx.Packages) {
		t.Fatalf("unexpected cached catalog: %#v", cached)
	}
}
