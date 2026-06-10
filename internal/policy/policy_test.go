package policy

import (
	"testing"

	"github.com/nandub/dockyard/internal/dockpkg"
)

func TestLintDetectsPrivilegedService(t *testing.T) {
	compose := []byte(`services:
  web:
    image: nginx:1.27
    privileged: true
`)
	findings, err := LintCompose(compose, dockpkg.SecurityPolicy{DisallowPrivileged: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != SeverityHigh {
		t.Fatalf("expected HIGH severity")
	}
}

func TestLintPassesSecureService(t *testing.T) {
	compose := []byte(`services:
  web:
    image: nginx:1.27
    user: "101:101"
    healthcheck:
      test: ["CMD", "nginx", "-t"]
`)
	findings, err := LintCompose(compose, dockpkg.SecurityPolicy{
		RequireNonRoot:      true,
		RequireHealthchecks: true,
		DisallowLatestTag:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}
