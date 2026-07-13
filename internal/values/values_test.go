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

func writeValuesTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
