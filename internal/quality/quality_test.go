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
