package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
)

func TestFlattenValues(t *testing.T) {
	vals := map[string]any{"image": map[string]any{"repository": "nginx", "tag": "1.27"}}
	flat := FlattenValues("", vals)
	if flat["image.repository"] != "nginx" {
		t.Fatalf("expected image.repository to be nginx")
	}
	if flat["image.tag"] != "1.27" {
		t.Fatalf("expected image.tag to be 1.27")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	if !IsSensitiveKey("database.password") {
		t.Fatal("expected password key to be sensitive")
	}
	if IsSensitiveKey("service.port") {
		t.Fatal("did not expect service.port to be sensitive")
	}
}

func TestRenderComposeWithDiagnosticsMasksSensitiveValues(t *testing.T) {
	dir := t.TempDir()
	writeRenderTestFile(t, filepath.Join(dir, "compose.yaml"), `services:
  web:
    image: ${image.repository}:${image.tag}
    environment:
      DATABASE_PASSWORD: ${database.password}
`)
	manifest := &dockpkg.Manifest{
		Compose: dockpkg.ComposeConfig{Base: "compose.yaml"},
	}

	result, err := RenderComposeWithDiagnostics(dir, manifest, map[string]any{
		"image":    map[string]any{"repository": "nginx", "tag": "1.27"},
		"database": map[string]any{"password": "supersecret"},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result.YAML), "supersecret") {
		t.Fatalf("expected rendered YAML to contain actual value:\n%s", string(result.YAML))
	}
	var foundMasked bool
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Key == "database.password" {
			foundMasked = true
			if !diagnostic.Masked || diagnostic.Value != "********" {
				t.Fatalf("expected masked diagnostic, got %#v", diagnostic)
			}
		}
	}
	if !foundMasked {
		t.Fatalf("expected database.password diagnostic, got %#v", result.Diagnostics)
	}
}

func TestRenderComposeReportsInvalidAndMissingPlaceholders(t *testing.T) {
	dir := t.TempDir()
	writeRenderTestFile(t, filepath.Join(dir, "compose.yaml"), `services:
  web:
    image: ${image.repository}:${missing.tag}
    command: ${bad ref}
`)
	manifest := &dockpkg.Manifest{
		Compose: dockpkg.ComposeConfig{Base: "compose.yaml"},
	}

	_, err := RenderCompose(dir, manifest, map[string]any{
		"image": map[string]any{"repository": "nginx"},
	}, "")
	if err == nil {
		t.Fatal("expected render error")
	}
	message := err.Error()
	if !strings.Contains(message, "invalid placeholders: bad ref") {
		t.Fatalf("expected invalid placeholder in error, got %q", message)
	}
	if !strings.Contains(message, "unresolved values: missing.tag") {
		t.Fatalf("expected missing value in error, got %q", message)
	}
}

func TestRenderComposeMergesOverlay(t *testing.T) {
	dir := t.TempDir()
	writeRenderTestFile(t, filepath.Join(dir, "compose.yaml"), `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
`)
	writeRenderTestFile(t, filepath.Join(dir, "compose.prod.yaml"), `services:
  web:
    ports:
      - "80:80"
    environment:
      MODE: prod
`)
	manifest := &dockpkg.Manifest{
		Compose: dockpkg.ComposeConfig{
			Base:     "compose.yaml",
			Overlays: map[string]string{"prod": "compose.prod.yaml"},
		},
	}

	rendered, err := RenderCompose(dir, manifest, nil, "prod")
	if err != nil {
		t.Fatal(err)
	}
	out := string(rendered)
	if !strings.Contains(out, "MODE: prod") || !strings.Contains(out, "80:80") {
		t.Fatalf("expected overlay content in rendered YAML:\n%s", out)
	}
	if strings.Contains(out, "8080:80") {
		t.Fatalf("expected overlay to replace ports:\n%s", out)
	}
}

func TestRenderComposeRejectsUnknownOverlay(t *testing.T) {
	dir := t.TempDir()
	writeRenderTestFile(t, filepath.Join(dir, "compose.yaml"), "services: {}\n")
	manifest := &dockpkg.Manifest{
		Compose: dockpkg.ComposeConfig{
			Base:     "compose.yaml",
			Overlays: map[string]string{"prod": "compose.prod.yaml"},
		},
	}

	_, err := RenderCompose(dir, manifest, nil, "staging")
	if err == nil {
		t.Fatal("expected unknown overlay error")
	}
	if !strings.Contains(err.Error(), `unknown overlay "staging"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeYAMLProducesCanonicalYAML(t *testing.T) {
	got, err := NormalizeYAML([]byte("services:\n  web:\n    image: nginx\n"))
	if err != nil {
		t.Fatalf("normalize YAML: %v", err)
	}
	if !strings.Contains(string(got), "services:") || !strings.Contains(string(got), "image: nginx") {
		t.Fatalf("unexpected normalized YAML:\n%s", got)
	}
}

func TestNormalizeYAMLRejectsInvalidYAML(t *testing.T) {
	if _, err := NormalizeYAML([]byte("services:\n  - [")); err == nil {
		t.Fatal("expected invalid YAML error")
	}
}

func writeRenderTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
