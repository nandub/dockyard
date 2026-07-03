package catalog

import (
	"testing"
)

func TestResolveCatalogURLDefaultVersion(t *testing.T) {
	t.Setenv(EnvRegistry, "")
	got, ok, err := Resolve("catalog://redis")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog source")
	}
	want := "oci://ghcr.io/nandub/dockyard-packages/redis:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveBareCatalogName(t *testing.T) {
	t.Setenv(EnvRegistry, "ghcr.io/example/catalog")
	got, ok, err := Resolve("postgres")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog source")
	}
	want := "oci://ghcr.io/example/catalog/postgres:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveUnknownBareNameIsUnchanged(t *testing.T) {
	got, ok, err := Resolve("not-a-package")
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
	_, ok, err := Resolve("catalog://not-a-package")
	if !ok {
		t.Fatal("expected catalog source")
	}
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListSorted(t *testing.T) {
	pkgs := List()
	if len(pkgs) == 0 {
		t.Fatal("expected packages")
	}
	for i := 1; i < len(pkgs); i++ {
		if pkgs[i-1].Name > pkgs[i].Name {
			t.Fatalf("packages are not sorted: %s before %s", pkgs[i-1].Name, pkgs[i].Name)
		}
	}
}
