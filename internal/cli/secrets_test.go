package cli

import "testing"

func TestScanSecretValuesFindsNestedPopulatedSecrets(t *testing.T) {
	findings := scanSecretValues("", map[string]any{
		"database": map[string]any{
			"password": "super-secret-value",
			"username": "dockyard",
		},
		"api": map[string]any{
			"token": "${API_TOKEN}",
		},
	})

	if len(findings) != 1 {
		t.Fatalf("expected one secret finding, got %#v", findings)
	}
	if findings[0].Path != "database.password" {
		t.Fatalf("unexpected finding path %q", findings[0].Path)
	}
	if findings[0].Preview != "su******ue" {
		t.Fatalf("unexpected masked preview %q", findings[0].Preview)
	}
}

func TestScanSecretValuesIgnoresEmptyAndPlaceholderSecrets(t *testing.T) {
	findings := scanSecretValues("", map[string]any{
		"password":    "changeme",
		"secret":      "",
		"api_key":     "replace-with-key",
		"private_key": "example-private-key",
		"token":       "${TOKEN}",
	})

	if len(findings) != 0 {
		t.Fatalf("expected no placeholder findings, got %#v", findings)
	}
}

func TestSensitivePathAndMaskPreview(t *testing.T) {
	for _, path := range []string{"db.passwd", "apiKey", "service.credentials"} {
		if !isSensitivePath(path) {
			t.Fatalf("expected sensitive path %q", path)
		}
	}
	if isSensitivePath("service.username") {
		t.Fatal("did not expect username-only path to be sensitive")
	}
	if got := maskPreview("abcd"); got != "****" {
		t.Fatalf("unexpected short mask %q", got)
	}
	if got := maskPreview("abcdef"); got != "ab******ef" {
		t.Fatalf("unexpected long mask %q", got)
	}
}
