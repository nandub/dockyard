package quality

import "testing"

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
