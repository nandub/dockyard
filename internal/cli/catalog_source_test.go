package cli

import "testing"

func TestResolveCatalogPackageSourceBareName(t *testing.T) {
	t.Setenv("DOCKYARD_CATALOG", "ghcr.io/example/packages")
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
	got, ok, err := resolveCatalogPackageSource("catalog://postgres:0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected catalog URL to resolve")
	}
	want := "oci://ghcr.io/nandub/dockyard-packages/postgres:0.1.0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
