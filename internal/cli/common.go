package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/catalog"
	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/envfile"
	"github.com/nandub/dockyard/internal/format"
	"github.com/nandub/dockyard/internal/lock"
	"github.com/nandub/dockyard/internal/oci"
	"github.com/nandub/dockyard/internal/policy"
	"github.com/nandub/dockyard/internal/render"
	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/state"
	"github.com/nandub/dockyard/internal/values"
	"github.com/nandub/dockyard/internal/version"
)

type packageBuildOptions struct {
	valuesFile        string
	overlay           string
	skipPolicy        bool
	allowRisk         bool
	skipComposeConfig bool
	requireLock       bool
	envFile           string
}

type preparedSource struct {
	Dir     string
	Source  state.Source
	cleanup func()
}

type preparePackageSourceOptions struct {
	QuietOCI bool
}

func preparePackageSource(input string, verifyArchive bool) (*preparedSource, error) {
	return preparePackageSourceWithOptions(input, verifyArchive, preparePackageSourceOptions{})
}

func preparePackageSourceWithOptions(input string, verifyArchive bool, opts preparePackageSourceOptions) (*preparedSource, error) {
	resolvedInput, _, err := resolveCatalogPackageSource(input)
	if err != nil {
		return nil, err
	}
	input = resolvedInput

	if oci.IsReference(input) {
		tempRoot, err := os.MkdirTemp("", "dockyard-oci-*")
		if err != nil {
			return nil, fmt.Errorf("create OCI temp dir: %w", err)
		}
		pullDir := filepath.Join(tempRoot, "pulled")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		archivePath, err := oci.PullWithOptions(ctx, input, pullDir, oci.PullOptions{Quiet: opts.QuietOCI})
		if err != nil {
			_ = os.RemoveAll(tempRoot)
			return nil, err
		}
		if verifyArchive {
			if err := archive.VerifyArchive(archivePath, nil); err != nil {
				_ = os.RemoveAll(tempRoot)
				return nil, err
			}
		}
		extractDir := filepath.Join(tempRoot, "src")
		if err := os.MkdirAll(extractDir, 0o700); err != nil {
			_ = os.RemoveAll(tempRoot)
			return nil, err
		}
		if err := archive.ExtractArchive(archivePath, extractDir); err != nil {
			_ = os.RemoveAll(tempRoot)
			return nil, err
		}
		return &preparedSource{
			Dir:    extractDir,
			Source: state.Source{Type: "oci", Path: input},
			cleanup: func() {
				_ = os.RemoveAll(tempRoot)
			},
		}, nil
	}

	cleanInput := filepath.Clean(input)
	if isArchivePath(cleanInput) {
		if verifyArchive {
			if err := archive.VerifyArchive(cleanInput, nil); err != nil {
				return nil, err
			}
		}
		tempDir, err := os.MkdirTemp("", "dockyard-src-*")
		if err != nil {
			return nil, fmt.Errorf("create package extraction temp dir: %w", err)
		}
		if err := archive.ExtractArchive(cleanInput, tempDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return nil, err
		}
		absArchive, _ := filepath.Abs(cleanInput)
		return &preparedSource{
			Dir:    tempDir,
			Source: state.Source{Type: "archive", Path: absArchive},
			cleanup: func() {
				_ = os.RemoveAll(tempDir)
			},
		}, nil
	}
	absSource, _ := filepath.Abs(cleanInput)
	return &preparedSource{
		Dir:     cleanInput,
		Source:  state.Source{Type: "local", Path: absSource},
		cleanup: func() {},
	}, nil
}

func resolveCatalogPackageSource(input string) (string, bool, error) {
	if strings.HasPrefix(input, "catalog://") {
		resolved, _, err := catalog.Resolve(input)
		return resolved, true, err
	}
	if sourcePathExists(input) || isArchivePath(input) || oci.IsReference(input) {
		return input, false, nil
	}
	resolved, ok, err := catalog.Resolve(input)
	return resolved, ok, err
}

func sourcePathExists(input string) bool {
	if input == "" {
		return false
	}
	_, err := os.Stat(input)
	return err == nil
}

func isArchivePath(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".dockyard.tgz") || strings.HasSuffix(lower, ".tgz") || strings.HasSuffix(lower, ".tar.gz")
}

func buildPackage(packageDir string, releaseName string, opts packageBuildOptions) (*dockpkg.Manifest, map[string]any, []byte, []policy.Finding, error) {
	manifest, err := dockpkg.LoadManifest(packageDir)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	vals, err := values.LoadValues(packageDir, opts.valuesFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := values.ValidateAgainstSchema(packageDir, vals); err != nil {
		return nil, nil, nil, nil, err
	}
	rendered, err := render.RenderCompose(packageDir, manifest, vals, opts.overlay)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if opts.requireLock {
		if err := lock.Verify(packageDir, rendered, vals, opts.overlay); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	var findings []policy.Finding
	if !opts.skipPolicy {
		findings, err = policy.LintCompose(rendered, manifest.Security)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if policy.HasHighFindings(findings) && !opts.allowRisk {
			return nil, nil, nil, findings, fmt.Errorf("policy check found HIGH severity finding(s); re-run with --allow-risk to continue")
		}
	}
	return manifest, vals, rendered, findings, nil
}

type releaseRelationshipMetadata struct {
	parent       *state.ReleaseParent
	dependencies []state.ReleaseDependency
}

func writeRevision(home string, releaseName string, revision int, manifest *dockpkg.Manifest, vals map[string]any, rendered []byte, packageDir string, src state.Source, statusValue string, envFile string, relationships releaseRelationshipMetadata) (*state.Release, string, error) {
	revisionDir := state.RevisionDir(home, releaseName, revision)
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		return nil, "", err
	}
	manifestSrc, err := dockpkg.SafeJoin(packageDir, dockpkg.ManifestFileName)
	if err != nil {
		return nil, "", err
	}
	if err := values.CopyFile(filepath.Join(revisionDir, dockpkg.ManifestFileName), manifestSrc, 0o600); err != nil {
		return nil, "", err
	}
	if err := values.WriteValues(filepath.Join(revisionDir, "values.yaml"), vals); err != nil {
		return nil, "", err
	}
	if _, err := os.Stat(filepath.Join(packageDir, lock.FileName)); err == nil {
		if err := values.CopyFile(filepath.Join(revisionDir, lock.FileName), filepath.Join(packageDir, lock.FileName), 0o600); err != nil {
			return nil, "", err
		}
	}
	composePath := filepath.Join(revisionDir, "compose.rendered.yaml")
	if err := os.WriteFile(composePath, rendered, 0o600); err != nil {
		return nil, "", err
	}
	now := time.Now().UTC()
	release := state.Release{
		APIVersion:      format.ReleaseAPIVersion,
		DockyardVersion: version.Version,
		Name:            releaseName,
		PackageName:     manifest.Name,
		PackageVersion:  manifest.Version,
		AppVersion:      manifest.AppVersion,
		Revision:        revision,
		Status:          statusValue,
		CreatedAt:       now,
		UpdatedAt:       now,
		ComposeProject:  releaseName,
		Source:          src,
		EnvFile:         envFile,
		Parent:          relationships.parent,
		Dependencies:    relationships.dependencies,
	}
	if err := state.WriteRelease(revisionDir, release); err != nil {
		return nil, "", err
	}
	return &release, composePath, nil
}

func readCurrentRelease(home string, releaseName string) (*state.Release, string, error) {
	if err := state.ValidateReleaseName(releaseName); err != nil {
		return nil, "", err
	}
	revision, err := state.ReadCurrentRevision(home, releaseName)
	if err != nil {
		return nil, "", err
	}
	revisionDir := state.RevisionDir(home, releaseName, revision)
	release, err := state.ReadRelease(revisionDir)
	if err != nil {
		return nil, "", err
	}
	return release, filepath.Join(revisionDir, "compose.rendered.yaml"), nil
}

func printJSON(v any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func loadCommandEnv(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	return envfile.LoadForProcess(path)
}

func dockerRunnerWithEnv(releaseName string, workDir string, env []string) runner.DockerComposeRunner {
	return runner.DockerComposeRunner{WorkDir: workDir, Project: releaseName, Env: env}
}

func dockerRunner(releaseName string, workDir string) runner.DockerComposeRunner {
	return runner.DockerComposeRunner{WorkDir: workDir, Project: releaseName}
}

func context10m() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Minute)
}
