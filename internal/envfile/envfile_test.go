package envfile

import (
	"os"
	"strings"
	"testing"
)

func TestEnvName(t *testing.T) {
	got := EnvName("", "database.password")
	if got != "DATABASE_PASSWORD" {
		t.Fatalf("expected DATABASE_PASSWORD, got %q", got)
	}
	got = EnvName("dockyard", "app.image-tag")
	if got != "DOCKYARD_APP_IMAGE_TAG" {
		t.Fatalf("expected prefixed env name, got %q", got)
	}
}

func TestGenerateTemplateMasksSensitive(t *testing.T) {
	t.Parallel()
	out := formatEnvValue("hello world")
	if out != `"hello world"` {
		t.Fatalf("expected quoted value, got %q", out)
	}
	if !IsSensitiveKey("database.password") {
		t.Fatal("expected password path to be sensitive")
	}
	if IsSensitiveKey("service.port") {
		t.Fatal("did not expect service.port to be sensitive")
	}
}

func TestCheckFileDetectsDuplicateAndSecret(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/test.env"
	content := "DATABASE_PASSWORD=supersecret\nDATABASE_PASSWORD=other\nBAD LINE\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	findings, err := CheckFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) < 3 {
		t.Fatalf("expected at least 3 findings, got %d: %#v", len(findings), findings)
	}
	var joined strings.Builder
	for _, finding := range findings {
		joined.WriteString(finding.Message)
		joined.WriteString("\n")
	}
	text := joined.String()
	if !strings.Contains(text, "duplicate") || !strings.Contains(text, "KEY=VALUE") || !strings.Contains(text, "secret-like") {
		t.Fatalf("unexpected findings: %s", text)
	}
}

func TestLoadForProcessParsesQuotedValues(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/test.env"
	content := "APP_PORT=8080\nAPP_NAME=\"hello world\"\nexport TOKEN='abc123'\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	entries, err := LoadForProcess(path)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(entries, "\n")
	if !strings.Contains(joined, "APP_NAME=hello world") || !strings.Contains(joined, "TOKEN=abc123") {
		t.Fatalf("unexpected entries: %s", joined)
	}
}

func TestLoadForProcessRejectsDuplicate(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/test.env"
	if err := os.WriteFile(path, []byte("A=1\nA=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadForProcess(path); err == nil {
		t.Fatal("expected duplicate key to fail")
	}
}
