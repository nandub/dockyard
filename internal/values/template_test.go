package values

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTemplateUsesSchemaDescriptionsAndMasksSensitiveValues(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(`app:
  tag: "1.0.0"
database:
  password: "change-me-in-prod"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "values.schema.json"), []byte(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "app": {
      "type": "object",
      "properties": {
        "tag": {
          "type": "string",
          "description": "Application image tag."
        }
      }
    },
    "database": {
      "type": "object",
      "properties": {
        "password": {
          "type": "string",
          "description": "Database password.",
          "x-dockyard-sensitive": true
        }
      }
    }
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := GenerateTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}

	out := string(data)
	if !strings.Contains(out, "# Application image tag.") {
		t.Fatalf("expected schema description in template:\n%s", out)
	}
	if !strings.Contains(out, "# Database password.") {
		t.Fatalf("expected sensitive field description in template:\n%s", out)
	}
	if strings.Contains(out, "change-me-in-prod") {
		t.Fatalf("expected sensitive default to be masked:\n%s", out)
	}
	if !strings.Contains(out, `password: ""`) {
		t.Fatalf("expected empty sensitive placeholder:\n%s", out)
	}
}
