package values

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValuesMergesNestedOverride(t *testing.T) {
	dir := t.TempDir()
	writeValuesTestFile(t, filepath.Join(dir, "values.yaml"), `app:
  image:
    repository: nginx
    tag: "1.27"
  replicas: 1
`)
	override := filepath.Join(dir, "override.yaml")
	writeValuesTestFile(t, override, `app:
  image:
    tag: "1.28"
`)

	vals, err := LoadValues(dir, override)
	if err != nil {
		t.Fatal(err)
	}
	app := vals["app"].(map[string]any)
	image := app["image"].(map[string]any)
	if image["repository"] != "nginx" {
		t.Fatalf("expected repository to be preserved, got %#v", image["repository"])
	}
	if image["tag"] != "1.28" {
		t.Fatalf("expected tag override, got %#v", image["tag"])
	}
	if app["replicas"] != 1 {
		t.Fatalf("expected replicas to be preserved, got %#v", app["replicas"])
	}
}

func TestValidateAgainstSchemaRejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	writeValuesTestFile(t, filepath.Join(dir, "values.schema.json"), `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "replicas": {
      "type": "integer",
      "minimum": 1
    }
  },
  "required": ["replicas"]
}`)

	err := ValidateAgainstSchema(dir, map[string]any{"replicas": 0})
	if err == nil {
		t.Fatal("expected schema validation error")
	}
	if !strings.Contains(err.Error(), "values validation failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAgainstSchemaAllowsMissingSchema(t *testing.T) {
	if err := ValidateAgainstSchema(t.TempDir(), map[string]any{"replicas": 1}); err != nil {
		t.Fatalf("expected missing schema to be allowed: %v", err)
	}
}

func TestLoadYAMLMapNullDocumentReturnsEmptyMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "values.yaml")
	writeValuesTestFile(t, path, "null\n")

	vals, err := LoadYAMLMap(path)
	if err != nil {
		t.Fatalf("load null values: %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty map, got %#v", vals)
	}
}

func TestWriteValuesAndCopyFile(t *testing.T) {
	dir := t.TempDir()
	valuesPath := filepath.Join(dir, "values.yaml")
	if err := WriteValues(valuesPath, map[string]any{
		"app": map[string]any{
			"replicas": 2,
		},
	}); err != nil {
		t.Fatalf("write values: %v", err)
	}

	vals, err := LoadYAMLMap(valuesPath)
	if err != nil {
		t.Fatalf("reload values: %v", err)
	}
	app := vals["app"].(map[string]any)
	if app["replicas"] != 2 {
		t.Fatalf("unexpected values after round trip: %#v", vals)
	}

	copyPath := filepath.Join(dir, "copy.yaml")
	if err := CopyFile(copyPath, valuesPath, 0o600); err != nil {
		t.Fatalf("copy values: %v", err)
	}
	if err := CopyFile(copyPath, valuesPath, 0o600); err == nil {
		t.Fatal("expected copy to reject existing destination")
	}
}

func TestMergeMapsReplacesNestedMapWithScalar(t *testing.T) {
	got := MergeMaps(
		map[string]any{"app": map[string]any{"image": "nginx", "replicas": 1}},
		map[string]any{"app": "disabled"},
	)
	if got["app"] != "disabled" {
		t.Fatalf("expected scalar override to replace nested map, got %#v", got["app"])
	}
}

func writeValuesTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
