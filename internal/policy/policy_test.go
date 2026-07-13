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

func TestLintDetectsDockerSocketHostPathAndLatestTag(t *testing.T) {
	compose := []byte(`services:
  web:
    image: nginx:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./data:/data
  db:
    image: postgres
    volumes:
      - type: bind
        source: ./pgdata
        target: /var/lib/postgres
`)
	findings, err := LintCompose(compose, dockpkg.SecurityPolicy{
		DisallowDockerSocketMount: true,
		DisallowHostPathMounts:    true,
		DisallowLatestTag:         true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, findings, SeverityHigh, "web", "mounting the Docker socket is not allowed")
	assertFinding(t, findings, SeverityMedium, "web", "host path mounts are discouraged; prefer named volumes")
	assertFinding(t, findings, SeverityMedium, "db", "host path mounts are discouraged; prefer named volumes")
	assertFinding(t, findings, SeverityMedium, "web", "image should not use the latest tag")
	assertFinding(t, findings, SeverityMedium, "db", "image should not use the latest tag")
}

func TestLintDetectsMissingHardeningSettings(t *testing.T) {
	compose := []byte(`services:
  web:
    image: nginx:1.27
    user: "0:0"
`)
	findings, err := LintCompose(compose, dockpkg.SecurityPolicy{
		RequireNonRoot:                true,
		RequireHealthchecks:           true,
		RequireReadOnlyRootFilesystem: true,
		RequireNoNewPrivileges:        true,
		RequireCapDropAll:             true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFinding(t, findings, SeverityHigh, "web", "service should run as a non-root user")
	assertFinding(t, findings, SeverityMedium, "web", "service should define a healthcheck")
	assertFinding(t, findings, SeverityMedium, "web", "service should set read_only: true")
	assertFinding(t, findings, SeverityMedium, "web", "service should set security_opt: no-new-privileges:true")
	assertFinding(t, findings, SeverityMedium, "web", "service should drop all Linux capabilities with cap_drop: [ALL]")
	if !HasHighFindings(findings) {
		t.Fatal("expected high findings")
	}
}

func assertFinding(t *testing.T, findings []Finding, severity FindingSeverity, service string, message string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Severity == severity && finding.Service == service && finding.Message == message {
			return
		}
	}
	t.Fatalf("missing finding severity=%s service=%s message=%q in %#v", severity, service, message, findings)
}
