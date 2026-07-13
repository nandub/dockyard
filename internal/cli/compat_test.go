package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/format"
	"github.com/nandub/dockyard/internal/state"
)

func TestCheckPackageCompatibilityReportsExpectedChecks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Dockyard.yaml"), `apiVersion: `+format.ManifestAPIVersion+`
name: nginx
version: 0.1.0
appVersion: "1.27"
compose:
  base: compose.yaml
dependencies:
  - name: postgres
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
`)
	writeFile(t, filepath.Join(dir, "values.yaml"), "{}\n")
	writeFile(t, filepath.Join(dir, "compose.yaml"), "services:\n  web:\n    image: nginx:1.27\n")

	result := checkPackageCompatibility(dir, dir)

	assertCompatCheck(t, result, "Dockyard.yaml", "OK")
	assertCompatCheck(t, result, "dependencies", "OK")
	assertCompatCheck(t, result, "values.yaml", "OK")
	assertCompatCheck(t, result, "values.schema.json", "WARN")
	assertCompatCheck(t, result, "dockyard.lock", "WARN")
	if !compatHasProblems(result) {
		t.Fatal("expected missing optional schema and lock warnings to count as problems")
	}
}

func TestCheckPackageCompatibilityReportsManifestFailure(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Dockyard.yaml"), `apiVersion: wrong
name: bad
version: 0.1.0
compose:
  base: compose.yaml
`)
	writeFile(t, filepath.Join(dir, "values.yaml"), "{}\n")

	result := checkPackageCompatibility(dir, dir)
	assertCompatCheck(t, result, "Dockyard.yaml", "FAIL")
	if !compatHasProblems(result) {
		t.Fatal("expected manifest failure to count as problem")
	}
}

func TestCheckReleaseCompatibilityReportsLegacyAndCurrentAPIVersions(t *testing.T) {
	home := t.TempDir()
	writeCurrentReleaseForCompat(t, home, "legacy", "")
	writeCurrentReleaseForCompat(t, home, "current", format.ReleaseAPIVersion)

	legacy, err := checkReleaseCompatibility(&globalOptions{home: home}, "legacy")
	if err != nil {
		t.Fatalf("check legacy release compatibility: %v", err)
	}
	assertCompatCheck(t, legacy, "release.json", "WARN")

	current, err := checkReleaseCompatibility(&globalOptions{home: home}, "current")
	if err != nil {
		t.Fatalf("check current release compatibility: %v", err)
	}
	assertCompatCheck(t, current, "release.json", "OK")
	if compatHasProblems(current) {
		t.Fatal("did not expect current release compatibility to have problems")
	}
}

func TestCompatHasProblemsIgnoresFormatOnlyResult(t *testing.T) {
	result := compatResult{Formats: format.SupportedFormats()}
	if compatHasProblems(result) {
		t.Fatal("supported format listing should not count as a problem")
	}
}

func assertCompatCheck(t *testing.T, result compatResult, name string, status string) {
	t.Helper()
	for _, check := range result.Checks {
		if check.Name == name {
			if check.Status != status {
				t.Fatalf("expected %s check status %s, got %s: %#v", name, status, check.Status, check)
			}
			return
		}
	}
	t.Fatalf("missing compatibility check %q in %#v", name, result.Checks)
}

func writeCurrentReleaseForCompat(t *testing.T, home string, releaseName string, apiVersion string) {
	t.Helper()
	revisionDir := state.RevisionDir(home, releaseName, 1)
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		t.Fatalf("create revision dir: %v", err)
	}
	release := state.Release{
		APIVersion:     apiVersion,
		Name:           releaseName,
		PackageName:    releaseName,
		PackageVersion: "0.1.0",
		Revision:       1,
		Status:         "deployed",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		ComposeProject: releaseName,
		Source:         state.Source{Type: "local", Path: "."},
	}
	if err := state.WriteRelease(revisionDir, release); err != nil {
		t.Fatalf("write release: %v", err)
	}
	if apiVersion == "" {
		path := filepath.Join(revisionDir, "release.json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read legacy release: %v", err)
		}
		withoutAPIVersion := []byte{}
		for _, line := range splitLines(string(data)) {
			if len(line) >= len(`  "apiVersion"`) && line[:len(`  "apiVersion"`)] == `  "apiVersion"` {
				continue
			}
			withoutAPIVersion = append(withoutAPIVersion, line...)
			withoutAPIVersion = append(withoutAPIVersion, '\n')
		}
		if err := os.WriteFile(path, withoutAPIVersion, 0o600); err != nil {
			t.Fatalf("write legacy release: %v", err)
		}
	}
	if err := state.SetCurrentRevision(home, releaseName, 1); err != nil {
		t.Fatalf("write current revision: %v", err)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for idx, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:idx])
			start = idx + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
