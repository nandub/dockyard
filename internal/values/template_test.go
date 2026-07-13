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

func TestGenerateTemplateWithoutSchemaSortsKeysAndFormatsScalars(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(`zeta: last
app:
  enabled: true
  replicas: 2
alpha: first
`), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := GenerateTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	alphaIndex := strings.Index(out, `alpha: "first"`)
	appIndex := strings.Index(out, "app:")
	zetaIndex := strings.Index(out, `zeta: "last"`)
	if alphaIndex < 0 || appIndex < 0 || zetaIndex < 0 {
		t.Fatalf("expected formatted keys in template:\n%s", out)
	}
	if !(alphaIndex < appIndex && appIndex < zetaIndex) {
		t.Fatalf("expected top-level keys to be sorted:\n%s", out)
	}
	for _, want := range []string{"enabled: true", "replicas: 2"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in template:\n%s", want, out)
		}
	}
}

func TestGenerateTemplateMasksSecretLikeKeysWithoutSchema(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(`api:
  token: production-token
service:
  port: 8080
`), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := GenerateTemplate(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if strings.Contains(out, "production-token") {
		t.Fatalf("expected secret-like value to be masked:\n%s", out)
	}
	if !strings.Contains(out, "# Sensitive value.") || !strings.Contains(out, `token: ""`) {
		t.Fatalf("expected sensitive guidance and empty token placeholder:\n%s", out)
	}
	if !strings.Contains(out, "port: 8080") {
		t.Fatalf("expected non-sensitive value to remain:\n%s", out)
	}
}
