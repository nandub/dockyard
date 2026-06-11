package quality

import (
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
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
