package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/format"
	"github.com/nandub/dockyard/internal/state"
)

func TestIsArchivePathRecognizesSupportedExtensions(t *testing.T) {
	valid := []string{
		"nginx.dockyard.tgz",
		"nginx.TGZ",
		"nginx.tar.gz",
	}
	for _, path := range valid {
		if !isArchivePath(path) {
			t.Fatalf("expected %q to be treated as archive path", path)
		}
	}
	if isArchivePath("nginx.zip") {
		t.Fatal("did not expect .zip to be treated as archive path")
	}
}

func TestSourcePathExistsRejectsEmptyAndAcceptsExistingPath(t *testing.T) {
	if sourcePathExists("") {
		t.Fatal("empty source path should not exist")
	}
	path := filepath.Join(t.TempDir(), "Dockyard.yaml")
	if err := os.WriteFile(path, []byte("metadata"), 0o600); err != nil {
		t.Fatalf("write source path: %v", err)
	}
	if !sourcePathExists(path) {
		t.Fatalf("expected source path %q to exist", path)
	}
}

func TestWriteRevisionAndReadCurrentRelease(t *testing.T) {
	home := t.TempDir()
	packageDir := t.TempDir()
	writeFile(t, filepath.Join(packageDir, dockpkg.ManifestFileName), `apiVersion: `+format.ManifestAPIVersion+`
name: nginx
version: 0.1.0
appVersion: "1.27"
compose:
  base: compose.yaml
`)
	manifest := &dockpkg.Manifest{
		Name:       "nginx",
		Version:    "0.1.0",
		AppVersion: "1.27",
	}

	release, composePath, err := writeRevision(
		home,
		"web",
		1,
		manifest,
		map[string]any{"replicas": 2},
		[]byte("services:\n  web:\n    image: nginx:1.27\n"),
		packageDir,
		state.Source{Type: "local", Path: packageDir},
		"deployed",
		".env",
		releaseRelationshipMetadata{
			parent: &state.ReleaseParent{Name: "parent", DependencyName: "nginx"},
			dependencies: []state.ReleaseDependency{
				{Name: "postgres", Release: "web-postgres", Source: "oci://example/postgres:0.1.0"},
			},
		},
	)
	if err != nil {
		t.Fatalf("write revision: %v", err)
	}
	if release.Name != "web" || release.EnvFile != ".env" || release.Parent == nil || len(release.Dependencies) != 1 {
		t.Fatalf("unexpected release metadata: %#v", release)
	}
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected rendered compose file: %v", err)
	}
	if err := state.SetCurrentRevision(home, "web", 1); err != nil {
		t.Fatalf("set current revision: %v", err)
	}

	current, currentComposePath, err := readCurrentRelease(home, "web")
	if err != nil {
		t.Fatalf("read current release: %v", err)
	}
	if current.Name != release.Name || current.PackageName != release.PackageName {
		t.Fatalf("unexpected current release: %#v", current)
	}
	if currentComposePath != composePath {
		t.Fatalf("expected compose path %q, got %q", composePath, currentComposePath)
	}
}

func TestLoadCommandEnvAndDockerRunnerHelpers(t *testing.T) {
	if env, err := loadCommandEnv(""); err != nil || env != nil {
		t.Fatalf("expected empty env path to return nil env, got %#v, err %v", env, err)
	}
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("B=2\nA=1\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	env, err := loadCommandEnv(path)
	if err != nil {
		t.Fatalf("load command env: %v", err)
	}
	if len(env) != 2 || env[0] != "A=1" || env[1] != "B=2" {
		t.Fatalf("unexpected env entries: %#v", env)
	}

	runner := dockerRunnerWithEnv("web", "work", env)
	if runner.Project != "web" || runner.WorkDir != "work" || len(runner.Env) != 2 {
		t.Fatalf("unexpected runner with env: %#v", runner)
	}
	plain := dockerRunner("api", "dir")
	if plain.Project != "api" || plain.WorkDir != "dir" || plain.Env != nil {
		t.Fatalf("unexpected plain runner: %#v", plain)
	}
}
