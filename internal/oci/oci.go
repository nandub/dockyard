package oci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	Scheme    = "oci://"
	MediaType = "application/vnd.dockyard.package.v1+gzip"
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

func CommandAvailable() bool {
	_, err := exec.LookPath("oras")
	return err == nil
}

func Push(ctx context.Context, archivePath string, ref string) error {
	normalized, err := NormalizeReference(ref)
	if err != nil {
		return err
	}
	if !CommandAvailable() {
		return errors.New("oras CLI was not found in PATH; install oras to use OCI push/pull")
	}
	cleanArchive := filepath.Clean(archivePath)
	if _, err := os.Stat(cleanArchive); err != nil {
		return fmt.Errorf("stat package archive: %w", err)
	}
	absArchive, err := filepath.Abs(cleanArchive)
	if err != nil {
		return fmt.Errorf("resolve package archive: %w", err)
	}
	archiveDir := filepath.Dir(absArchive)
	archiveName := filepath.Base(absArchive)
	layer := archiveName + ":" + MediaType
	cmd := exec.CommandContext(ctx, "oras", "push", normalized, layer)
	cmd.Dir = archiveDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("oras push failed")
	}
	return nil
}

func Pull(ctx context.Context, ref string, outputDir string) (string, error) {
	normalized, err := NormalizeReference(ref)
	if err != nil {
		return "", err
	}
	if !CommandAvailable() {
		return "", errors.New("oras CLI was not found in PATH; install oras to use OCI push/pull")
	}
	cleanOutput := filepath.Clean(outputDir)
	if err := os.MkdirAll(cleanOutput, 0o700); err != nil {
		return "", fmt.Errorf("create OCI pull directory: %w", err)
	}
	cmd := exec.CommandContext(ctx, "oras", "pull", normalized, "-o", cleanOutput)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", errors.New("oras pull failed")
	}
	archivePath, err := findPulledArchive(cleanOutput)
	if err != nil {
		return "", err
	}
	return archivePath, nil
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
