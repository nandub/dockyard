package oci

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
)

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

func TestNormalizeReferenceTrimsOuterWhitespaceAndRejectsInnerWhitespace(t *testing.T) {
	got, err := NormalizeReference("oci:// ghcr.io/nandub/dockyard/app:0.1.0 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ghcr.io/nandub/dockyard/app:0.1.0" {
		t.Fatalf("unexpected normalized reference %q", got)
	}

	if _, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app:0.1.0 extra"); err == nil {
		t.Fatal("expected whitespace error")
	}
}

func TestNormalizeReferenceAcceptsDigest(t *testing.T) {
	_, err := NormalizeReference("oci://ghcr.io/nandub/dockyard/app@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindPulledArchiveFindsSingleSupportedArchive(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	archive := filepath.Join(nested, "nginx-0.1.0.dockyard.tgz")
	if err := os.WriteFile(archive, []byte("archive"), 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}

	got, err := findPulledArchive(dir)
	if err != nil {
		t.Fatalf("unexpected archive scan error: %v", err)
	}
	if got != archive {
		t.Fatalf("expected archive %q, got %q", archive, got)
	}
}

func TestFindPulledArchiveRequiresExactlyOneArchive(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o600); err != nil {
			t.Fatalf("write ignored file: %v", err)
		}
		if _, err := findPulledArchive(dir); err == nil {
			t.Fatal("expected missing archive error")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range []string{"a.tgz", "b.tar.gz"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("archive"), 0o600); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
		if _, err := findPulledArchive(dir); err == nil {
			t.Fatal("expected multiple archive error")
		}
	})
}

func TestPushArchiveToTargetAndPullReferenceToDirRoundTrip(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	archiveName := "nginx-0.1.0.dockyard.tgz"
	archivePath := filepath.Join(dir, archiveName)
	archiveData := []byte("archive")
	if err := os.WriteFile(archivePath, archiveData, 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}

	target := memory.New()
	const ref = "0.1.0"
	if err := pushArchiveToTarget(ctx, archivePath, ref, target); err != nil {
		t.Fatalf("push archive to target: %v", err)
	}

	desc, err := target.Resolve(ctx, ref)
	if err != nil {
		t.Fatalf("resolve pushed reference: %v", err)
	}
	manifestData, err := content.FetchAll(ctx, target, desc)
	if err != nil {
		t.Fatalf("fetch manifest: %v", err)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.ArtifactType != ArtifactType {
		t.Fatalf("expected artifact type %q, got %q", ArtifactType, manifest.ArtifactType)
	}
	if len(manifest.Layers) != 1 {
		t.Fatalf("expected one archive layer, got %d", len(manifest.Layers))
	}
	layer := manifest.Layers[0]
	if layer.MediaType != LayerMediaType {
		t.Fatalf("expected layer media type %q, got %q", LayerMediaType, layer.MediaType)
	}
	if layer.Annotations[ocispec.AnnotationTitle] != archiveName {
		t.Fatalf("expected layer title %q, got %q", archiveName, layer.Annotations[ocispec.AnnotationTitle])
	}

	outDir := t.TempDir()
	if err := pullReferenceToDir(ctx, target, ref, outDir); err != nil {
		t.Fatalf("pull reference to dir: %v", err)
	}
	pulledPath := filepath.Join(outDir, archiveName)
	pulledData, err := os.ReadFile(pulledPath)
	if err != nil {
		t.Fatalf("read pulled archive: %v", err)
	}
	if string(pulledData) != string(archiveData) {
		t.Fatalf("pulled archive data mismatch: got %q", pulledData)
	}
}

func TestPushFileToTargetUsesProvidedNameAndMediaTypes(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "custom-index.yml")
	sourceData := []byte("catalog")
	if err := os.WriteFile(sourcePath, sourceData, 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	target := memory.New()
	const (
		ref            = "latest"
		artifactName   = "catalog.yaml"
		artifactType   = "application/vnd.dockyard.catalog.v1+yaml"
		layerMediaType = "application/vnd.dockyard.catalog.index.v1+yaml"
	)
	if err := pushFileToTarget(ctx, sourcePath, ref, target, artifactName, artifactType, layerMediaType); err != nil {
		t.Fatalf("push file to target: %v", err)
	}

	desc, err := target.Resolve(ctx, ref)
	if err != nil {
		t.Fatalf("resolve pushed reference: %v", err)
	}
	manifestData, err := content.FetchAll(ctx, target, desc)
	if err != nil {
		t.Fatalf("fetch manifest: %v", err)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.ArtifactType != artifactType {
		t.Fatalf("expected artifact type %q, got %q", artifactType, manifest.ArtifactType)
	}
	if len(manifest.Layers) != 1 {
		t.Fatalf("expected one layer, got %d", len(manifest.Layers))
	}
	layer := manifest.Layers[0]
	if layer.MediaType != layerMediaType {
		t.Fatalf("expected layer media type %q, got %q", layerMediaType, layer.MediaType)
	}
	if layer.Annotations[ocispec.AnnotationTitle] != artifactName {
		t.Fatalf("expected title %q, got %q", artifactName, layer.Annotations[ocispec.AnnotationTitle])
	}

	outDir := t.TempDir()
	if err := pullReferenceToDir(ctx, target, ref, outDir); err != nil {
		t.Fatalf("pull reference to dir: %v", err)
	}
	pulledData, err := os.ReadFile(filepath.Join(outDir, artifactName))
	if err != nil {
		t.Fatalf("read pulled file: %v", err)
	}
	if string(pulledData) != string(sourceData) {
		t.Fatalf("pulled file data mismatch: got %q", pulledData)
	}
}
