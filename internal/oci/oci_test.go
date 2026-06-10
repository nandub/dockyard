package oci

import "testing"

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

func TestNormalizeReferenceAcceptsDigest(t *testing.T) {
	_, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
