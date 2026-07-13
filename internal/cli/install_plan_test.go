package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/state"
)

func TestBuildInstallPlanWithDependency(t *testing.T) {
	packageDir := writeInstallPlanPackage(t)
	home := t.TempDir()

	report, err := buildInstallPlan(&globalOptions{home: home}, "team-dashboard", packageDir)
	if err != nil {
		t.Fatalf("build install plan: %v", err)
	}

	if !report.ReadOnly {
		t.Fatal("expected install plan to be read-only")
	}
	if report.ReleaseName != "team-dashboard" {
		t.Fatalf("unexpected release name: %s", report.ReleaseName)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 plan steps, got %d", len(report.Steps))
	}

	dep := report.Steps[0]
	if dep.Type != planStepDependency {
		t.Fatalf("expected first step to be dependency, got %s", dep.Type)
	}
	if dep.PlannedRelease != "team-dashboard-db" {
		t.Fatalf("unexpected dependency release name: %s", dep.PlannedRelease)
	}
	if dep.Action != "install" {
		t.Fatalf("unexpected dependency action: %s", dep.Action)
	}
	if dep.ExistingReleaseStatus != "" {
		t.Fatalf("unexpected dependency existing status: %s", dep.ExistingReleaseStatus)
	}

	root := report.Steps[1]
	if root.Type != planStepRoot {
		t.Fatalf("expected second step to be root, got %s", root.Type)
	}
	if root.PlannedRelease != "team-dashboard" {
		t.Fatalf("unexpected root release name: %s", root.PlannedRelease)
	}
	if root.Action != "install" {
		t.Fatalf("unexpected root action: %s", root.Action)
	}
}

func TestBuildInstallPlanDetectsExistingReleases(t *testing.T) {
	packageDir := writeInstallPlanPackage(t)
	home := t.TempDir()
	writeCurrentReleaseForPlan(t, home, "team-dashboard-db", "deployed")
	writeCurrentReleaseForPlan(t, home, "team-dashboard", "uninstalled")

	report, err := buildInstallPlan(&globalOptions{home: home}, "team-dashboard", packageDir)
	if err != nil {
		t.Fatalf("build install plan: %v", err)
	}
	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 plan steps, got %d", len(report.Steps))
	}
	if report.Steps[0].Action != "exists" || report.Steps[0].ExistingReleaseStatus != "deployed" {
		t.Fatalf("unexpected dependency existing state: action=%s status=%s", report.Steps[0].Action, report.Steps[0].ExistingReleaseStatus)
	}
	if report.Steps[1].Action != "reinstall" || report.Steps[1].ExistingReleaseStatus != "uninstalled" {
		t.Fatalf("unexpected root existing state: action=%s status=%s", report.Steps[1].Action, report.Steps[1].ExistingReleaseStatus)
	}
}

func TestBuildInstallDryRunPlanMatchesInstallPlan(t *testing.T) {
	packageDir := writeInstallPlanPackage(t)
	home := t.TempDir()

	installPlan, err := buildInstallPlan(&globalOptions{home: home}, "team-dashboard", packageDir)
	if err != nil {
		t.Fatalf("build install plan: %v", err)
	}

	dryRunPlan, err := buildInstallDryRunPlan(&globalOptions{home: home}, "team-dashboard", packageDir, false)
	if err != nil {
		t.Fatalf("build dry-run install plan: %v", err)
	}

	if !reflect.DeepEqual(dryRunPlan, installPlan) {
		t.Fatalf("dry-run plan does not match install-plan\ninstall-plan: %#v\ndry-run: %#v", installPlan, dryRunPlan)
	}
}

func TestBuildInstallPlanBlocksFailedDependency(t *testing.T) {
	packageDir := writeInstallPlanPackage(t)
	home := t.TempDir()
	writeCurrentReleaseForPlan(t, home, "team-dashboard-db", "failed")

	report, err := buildInstallPlan(&globalOptions{home: home}, "team-dashboard", packageDir)
	if err != nil {
		t.Fatalf("build install plan: %v", err)
	}
	if report.Steps[0].Action != "blocked" || report.Steps[0].ExistingReleaseStatus != "failed" {
		t.Fatalf("unexpected failed dependency state: action=%s status=%s", report.Steps[0].Action, report.Steps[0].ExistingReleaseStatus)
	}
}

func TestDependencyReleaseNameUsesNameWhenAliasMissing(t *testing.T) {
	got := dependencyReleaseName("app", dockpkg.Dependency{Name: "postgres"})
	if got != "app-postgres" {
		t.Fatalf("unexpected release name: %s", got)
	}
}

func TestPlannedInstallActionMapsReleaseStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{status: "", want: "install"},
		{status: "uninstalled", want: "reinstall"},
		{status: "deployed", want: "exists"},
		{status: "failed", want: "blocked"},
		{status: "pending", want: "blocked"},
	}
	for _, tt := range tests {
		if got := plannedInstallAction(tt.status); got != tt.want {
			t.Fatalf("plannedInstallAction(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestCurrentReleaseStatusMissingReleaseIsEmpty(t *testing.T) {
	got, err := currentReleaseStatus(t.TempDir(), "missing")
	if err != nil {
		t.Fatalf("current release status: %v", err)
	}
	if got != "" {
		t.Fatalf("expected missing release status to be empty, got %q", got)
	}
}

func TestBuildInstallPlanRejectsDuplicateDependencyPlannedRelease(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Dockyard.yaml"), `apiVersion: dockyard.dev/v1alpha1
name: team-dashboard
description: Team dashboard example
version: 0.2.0
type: application
compose:
  base: compose.yaml
dependencies:
  - name: postgres
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
  - name: redis
    alias: postgres
    source: oci://ghcr.io/nandub/dockyard/redis:0.1.0
`)
	writeFile(t, filepath.Join(dir, "compose.yaml"), "services:\n  web:\n    image: nginx:1.27\n")
	writeFile(t, filepath.Join(dir, "values.yaml"), "{}\n")

	_, err := buildInstallPlan(&globalOptions{home: t.TempDir()}, "team-dashboard", dir)
	if err == nil {
		t.Fatal("expected duplicate planned release name error")
	}
	if !strings.Contains(err.Error(), "duplicate planned release name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeInstallPlanPackage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Dockyard.yaml"), `apiVersion: dockyard.dev/v1alpha1
name: team-dashboard
description: Team dashboard example
version: 0.2.0
type: application
compose:
  base: compose.yaml
dependencies:
  - name: postgres
    alias: db
    version: 0.1.0
    source: oci://ghcr.io/nandub/dockyard/postgres:0.1.0
    values:
      database: dashboard
      username: dashboard
`)
	writeFile(t, filepath.Join(dir, "compose.yaml"), `services:
  web:
    image: nginx:1.27
`)
	writeFile(t, filepath.Join(dir, "values.yaml"), "{}\n")
	return dir
}

func writeCurrentReleaseForPlan(t *testing.T, home string, releaseName string, status string) {
	t.Helper()
	revisionDir := state.RevisionDir(home, releaseName, 1)
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		t.Fatalf("create revision dir: %v", err)
	}
	release := state.Release{
		Name:           releaseName,
		PackageName:    releaseName,
		PackageVersion: "0.1.0",
		Revision:       1,
		Status:         status,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		ComposeProject: releaseName,
		Source:         state.Source{Type: "local", Path: "."},
	}
	if err := state.WriteRelease(revisionDir, release); err != nil {
		t.Fatalf("write release: %v", err)
	}
	if err := state.SetCurrentRevision(home, releaseName, 1); err != nil {
		t.Fatalf("write current revision: %v", err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDependencyInstallOptionsDoNotReuseRootPackageSpecificFiles(t *testing.T) {
	opts := packageBuildOptions{
		valuesFile:        "root-values.yaml",
		overlay:           "prod",
		envFile:           ".env",
		skipPolicy:        true,
		allowRisk:         true,
		skipComposeConfig: true,
		requireLock:       true,
	}

	depOpts := dependencyInstallOptions(opts)

	if depOpts.valuesFile != "" {
		t.Fatalf("dependency values file should not reuse root values file: %q", depOpts.valuesFile)
	}
	if depOpts.overlay != "" {
		t.Fatalf("dependency overlay should not reuse root overlay: %q", depOpts.overlay)
	}
	if depOpts.envFile != opts.envFile {
		t.Fatalf("dependency env file should be preserved")
	}
	if !depOpts.skipPolicy || !depOpts.allowRisk || !depOpts.skipComposeConfig || !depOpts.requireLock {
		t.Fatalf("dependency safety flags were not preserved: %#v", depOpts)
	}
}

func TestApplyInlineValuesWritesTemporaryValuesFile(t *testing.T) {
	opts := packageBuildOptions{}
	cleanup, err := applyInlineValues(&opts, map[string]any{
		"database": "dashboard",
		"username": "dashboard",
	})
	if err != nil {
		t.Fatalf("apply inline values: %v", err)
	}
	defer cleanup()

	if opts.valuesFile == "" {
		t.Fatal("expected temporary values file to be configured")
	}
	data, err := os.ReadFile(opts.valuesFile)
	if err != nil {
		t.Fatalf("read temporary values file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "database: dashboard") || !strings.Contains(content, "username: dashboard") {
		t.Fatalf("temporary values file did not contain dependency values: %s", content)
	}
}

func TestApplyInlineValuesRejectsExistingValuesFile(t *testing.T) {
	opts := packageBuildOptions{valuesFile: "root-values.yaml"}
	_, err := applyInlineValues(&opts, map[string]any{"database": "dashboard"})
	if err == nil {
		t.Fatal("expected inline values to reject an existing values file")
	}
}
