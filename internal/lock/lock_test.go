package lock

import (
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
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
