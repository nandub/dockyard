package oci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	Scheme         = "oci://"
	ArtifactType   = "application/vnd.dockyard.package.v1+gzip"
	LayerMediaType = "application/vnd.dockyard.package.archive.v1+gzip"
)

func IsReference(input string) bool {
	return strings.HasPrefix(strings.ToLower(input), Scheme)
}

func NormalizeReference(input string) (string, error) {
	if !IsReference(input) {
		return "", fmt.Errorf("OCI reference must start with %s", Scheme)
	}
	ref := strings.TrimSpace(input[len(Scheme):])
	if ref == "" {
		return "", errors.New("OCI reference is empty")
	}
	if strings.ContainsAny(ref, " \t\n\r") {
		return "", errors.New("OCI reference must not contain whitespace")
	}
	lastSlash := strings.LastIndex(ref, "/")
	tagIndex := strings.LastIndex(ref, ":")
	hasTag := tagIndex > lastSlash
	hasDigest := strings.Contains(ref[lastSlash+1:], "@sha256:")
	if !hasTag && !hasDigest {
		return "", errors.New("OCI reference must include an explicit tag or digest")
	}
	return ref, nil
}

func Push(ctx context.Context, archivePath string, ref string) error {
	return PushNamedFile(ctx, archivePath, ref, filepath.Base(filepath.Clean(archivePath)), ArtifactType, LayerMediaType)
}

func PushFile(ctx context.Context, filePath string, ref string, artifactType string, layerMediaType string) error {
	return PushNamedFile(ctx, filePath, ref, filepath.Base(filepath.Clean(filePath)), artifactType, layerMediaType)
}

func PushNamedFile(ctx context.Context, filePath string, ref string, artifactName string, artifactType string, layerMediaType string) error {
	normalized, err := NormalizeReference(ref)
	if err != nil {
		return err
	}
	cleanPath := filepath.Clean(filePath)
	if strings.TrimSpace(artifactName) == "" {
		return errors.New("OCI artifact file name is empty")
	}
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("stat OCI artifact file: %w", err)
	}
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("resolve OCI artifact file: %w", err)
	}
	repo, err := newRepository(normalized)
	if err != nil {
		return err
	}
	if err := pushFileToTarget(ctx, absPath, repo.Reference.Reference, repo, artifactName, artifactType, layerMediaType); err != nil {
		return fmt.Errorf("push OCI artifact: %w", err)
	}
	return nil
}

type PullOptions struct {
	Quiet bool
}

func Pull(ctx context.Context, ref string, outputDir string) (string, error) {
	return PullWithOptions(ctx, ref, outputDir, PullOptions{})
}

func PullWithOptions(ctx context.Context, ref string, outputDir string, opts PullOptions) (string, error) {
	if err := PullFiles(ctx, ref, outputDir, opts); err != nil {
		return "", err
	}
	archivePath, err := findPulledArchive(filepath.Clean(outputDir))
	if err != nil {
		return "", err
	}
	return archivePath, nil
}

func PullFiles(ctx context.Context, ref string, outputDir string, _ PullOptions) error {
	normalized, err := NormalizeReference(ref)
	if err != nil {
		return err
	}
	cleanOutput := filepath.Clean(outputDir)
	if err := os.MkdirAll(cleanOutput, 0o700); err != nil {
		return fmt.Errorf("create OCI pull directory: %w", err)
	}
	repo, err := newRepository(normalized)
	if err != nil {
		return err
	}
	if err := pullReferenceToDir(ctx, repo, repo.Reference.Reference, cleanOutput); err != nil {
		return fmt.Errorf("pull OCI package: %w", err)
	}
	return nil
}

func newRepository(normalizedRef string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(normalizedRef)
	if err != nil {
		return nil, fmt.Errorf("parse OCI reference: %w", err)
	}
	client, err := newAuthClient()
	if err != nil {
		return nil, err
	}
	repo.Client = client
	return repo, nil
}

func newAuthClient() (*auth.Client, error) {
	client := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
	}
	client.SetUserAgent("dockyard")
	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return nil, fmt.Errorf("load Docker registry credentials: %w", err)
	}
	client.Credential = credentials.Credential(store)
	return client, nil
}

func pushArchiveToTarget(ctx context.Context, archivePath string, ref string, target oras.Target) error {
	archiveName := filepath.Base(archivePath)
	return pushFileToTarget(ctx, archivePath, ref, target, archiveName, ArtifactType, LayerMediaType)
}

func pushFileToTarget(ctx context.Context, filePath string, ref string, target oras.Target, artifactName string, artifactType string, layerMediaType string) error {
	fileDir := filepath.Dir(filePath)
	store, err := file.New(fileDir)
	if err != nil {
		return fmt.Errorf("create OCI file store: %w", err)
	}
	defer store.Close()
	layer, err := store.Add(ctx, artifactName, layerMediaType, filePath)
	if err != nil {
		return fmt.Errorf("add OCI artifact layer: %w", err)
	}
	manifest, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, artifactType, oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{layer},
	})
	if err != nil {
		return fmt.Errorf("pack OCI artifact manifest: %w", err)
	}
	if err := store.Tag(ctx, manifest, ref); err != nil {
		return fmt.Errorf("tag OCI artifact manifest: %w", err)
	}
	if _, err := oras.Copy(ctx, store, ref, target, ref, oras.DefaultCopyOptions); err != nil {
		return err
	}
	return nil
}

func pullReferenceToDir(ctx context.Context, source oras.ReadOnlyTarget, ref string, outputDir string) error {
	store, err := file.New(outputDir)
	if err != nil {
		return fmt.Errorf("create OCI output store: %w", err)
	}
	defer store.Close()
	store.DisableOverwrite = true
	if _, err := oras.Copy(ctx, source, ref, store, ref, oras.DefaultCopyOptions); err != nil {
		return err
	}
	return nil
}

func findPulledArchive(dir string) (string, error) {
	var matches []string
	if err := filepath.WalkDir(filepath.Clean(dir), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		lower := strings.ToLower(entry.Name())
		if strings.HasSuffix(lower, ".dockyard.tgz") || strings.HasSuffix(lower, ".tgz") || strings.HasSuffix(lower, ".tar.gz") {
			matches = append(matches, path)
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("scan OCI pull output: %w", err)
	}
	if len(matches) == 0 {
		return "", errors.New("OCI artifact did not contain a Dockyard archive")
	}
	if len(matches) > 1 {
		return "", errors.New("OCI artifact contained multiple archive files; expected exactly one")
	}
	return matches[0], nil
}
