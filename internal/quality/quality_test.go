package quality

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/format"
)

func TestInspectSchemaQualityFindsMissingSensitiveMarker(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"database": map[string]any{
				"properties": map[string]any{
					"password": map[string]any{
						"type":        "string",
						"description": "Database password.",
					},
				},
			},
		},
	}

	missingDesc, missingSensitive := inspectSchemaQuality(schema, "")

	if len(missingDesc) != 0 {
		t.Fatalf("expected no missing descriptions, got %v", missingDesc)
	}
	if len(missingSensitive) != 1 || missingSensitive[0] != "database.password" {
		t.Fatalf("expected database.password to require sensitive marker, got %v", missingSensitive)
	}
}

func TestInspectSchemaQualityAcceptsDescriptionsAndSensitiveMarker(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"database": map[string]any{
				"properties": map[string]any{
					"password": map[string]any{
						"type":                 "string",
						"description":          "Database password.",
						"x-dockyard-sensitive": true,
					},
				},
			},
		},
	}

	missingDesc, missingSensitive := inspectSchemaQuality(schema, "")

	if len(missingDesc) != 0 {
		t.Fatalf("expected no missing descriptions, got %v", missingDesc)
	}
	if len(missingSensitive) != 0 {
		t.Fatalf("expected no missing sensitive markers, got %v", missingSensitive)
	}
}

func TestHasBlockingFindingsStrictFailsWarnings(t *testing.T) {
	report := Report{
		Checks: []Check{
			{Name: "LICENSE", Severity: SeverityWarn, Message: "missing"},
		},
	}

	if !HasBlockingFindings(report, Options{Strict: true}) {
		t.Fatal("expected strict mode to fail on warnings")
	}
}

func TestHasBlockingFindingsAllowAdvisorySkipsAdvisoryWarnings(t *testing.T) {
	report := Report{
		Checks: []Check{
			{Name: "LICENSE", Severity: SeverityWarn, Message: "missing", Advisory: true},
		},
	}

	if HasBlockingFindings(report, Options{Strict: true, AllowAdvisory: true}) {
		t.Fatal("expected advisory warning to be allowed")
	}
}

func TestHasBlockingFindingsAllowAdvisoryStillFailsNonAdvisoryWarnings(t *testing.T) {
	report := Report{
		Checks: []Check{
			{Name: "values.schema.json", Severity: SeverityWarn, Message: "missing"},
		},
	}

	if !HasBlockingFindings(report, Options{Strict: true, AllowAdvisory: true}) {
		t.Fatal("expected non-advisory warning to fail even when advisory warnings are allowed")
	}
}

func TestCheckDependenciesReportsDeclaredDependencies(t *testing.T) {
	check := checkDependencies(&dockpkg.Manifest{
		Dependencies: []dockpkg.Dependency{
			{Name: "postgres", Alias: "db", Version: "0.1.0", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
		},
	})
	if check.Severity != SeverityOK {
		t.Fatalf("expected OK dependency check, got %s", check.Severity)
	}
	if check.Message != "dependency metadata is valid; automatic dependency installation is not enabled" {
		t.Fatalf("unexpected message: %s", check.Message)
	}
	if len(check.Details) != 1 || check.Details[0] != "postgres as db@0.1.0 from oci://ghcr.io/nandub/dockyard/postgres:0.1.0" {
		t.Fatalf("unexpected dependency details: %v", check.Details)
	}
}

func TestLintPackageReportsMissingStrictFilesAndForbiddenFiles(t *testing.T) {
	dir := t.TempDir()
	writeQualityFile(t, dir, dockpkg.ManifestFileName, `apiVersion: `+format.ManifestAPIVersion+`
name: nginx
version: 0.1.0
appVersion: "1.27"
description: test package
compose:
  base: compose.yaml
`)
	writeQualityFile(t, dir, "compose.yaml", `services:
  web:
    image: nginx:1.27
`)
	writeQualityFile(t, dir, "values.yaml", "{}\n")
	writeQualityFile(t, dir, ".env", "PASSWORD=secret\n")

	report, err := LintPackage(dir, Options{Strict: true})
	if err != nil {
		t.Fatalf("lint package: %v", err)
	}

	assertCheck(t, report, "Dockyard.yaml", SeverityOK)
	assertCheck(t, report, "README.md", SeverityFail)
	assertCheck(t, report, "SECURITY.md", SeverityFail)
	assertCheck(t, report, "LICENSE", SeverityWarn)
	assertCheck(t, report, "values.schema.json", SeverityFail)
	assertCheck(t, report, "forbidden files", SeverityFail)
	if !HasFailures(report) {
		t.Fatal("expected strict package lint to have failures")
	}
}

func TestLintPackageAllowsAdvisoryLicenseWarning(t *testing.T) {
	report := Report{
		Checks: []Check{
			{Name: "LICENSE", Severity: SeverityWarn, Message: "missing", Advisory: true},
		},
	}
	if !HasBlockingFindings(report, Options{Strict: true}) {
		t.Fatal("expected advisory warning to block when not allowed")
	}
	if HasBlockingFindings(report, Options{Strict: true, AllowAdvisory: true}) {
		t.Fatal("expected advisory warning to be allowed")
	}
}

func TestCheckValuesAndSchemaReportsSchemaQuality(t *testing.T) {
	dir := t.TempDir()
	writeQualityFile(t, dir, "values.yaml", `database:
  password: changeme
app:
  port: 8080
`)
	writeQualityFile(t, dir, "values.schema.json", `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "database": {
      "type": "object",
      "properties": {
        "password": {
          "type": "string"
        }
      }
    },
    "app": {
      "type": "object",
      "properties": {
        "port": {
          "type": "integer"
        }
      }
    }
  }
}`)

	checks := checkValuesAndSchema(dir, true)
	assertChecksContain(t, checks, "values.yaml", SeverityOK)
	assertChecksContain(t, checks, "values.schema.json", SeverityOK)
	assertChecksContain(t, checks, "schema descriptions", SeverityFail)
	assertChecksContain(t, checks, "schema sensitive markers", SeverityFail)
}

func TestCheckDefaultRenderReportsPolicyFindings(t *testing.T) {
	dir := t.TempDir()
	writeQualityFile(t, dir, "values.yaml", "{}\n")
	writeQualityFile(t, dir, "compose.yaml", `services:
  web:
    image: nginx:latest
    privileged: true
`)
	manifest := &dockpkg.Manifest{
		Compose: dockpkg.ComposeConfig{Base: "compose.yaml"},
		Security: dockpkg.SecurityPolicy{
			DisallowPrivileged: true,
			DisallowLatestTag:  true,
		},
	}

	checks := checkDefaultRender(dir, manifest)
	assertChecksContain(t, checks, "default render", SeverityOK)
	assertChecksContain(t, checks, "policy lint", SeverityFail)
}

func assertChecksContain(t *testing.T, checks []Check, name string, severity Severity) {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			if check.Severity != severity {
				t.Fatalf("expected check %q severity %s, got %s: %#v", name, severity, check.Severity, check)
			}
			return
		}
	}
	t.Fatalf("missing check %q in %#v", name, checks)
}

func assertCheck(t *testing.T, report Report, name string, severity Severity) {
	t.Helper()
	for _, check := range report.Checks {
		if check.Name == name {
			if check.Severity != severity {
				t.Fatalf("expected check %q severity %s, got %s: %#v", name, severity, check.Severity, check)
			}
			return
		}
	}
	t.Fatalf("missing check %q in %#v", name, report.Checks)
}

func writeQualityFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create parent for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
