package cli

import (
	"os"
	"path/filepath"
	"reflect"
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

	dryRunPlan, err := buildInstallDryRunPlan(&globalOptions{home: home}, "team-dashboard", packageDir)
	if err != nil {
		t.Fatalf("build dry-run install plan: %v", err)
	}

	if !reflect.DeepEqual(dryRunPlan, installPlan) {
		t.Fatalf("dry-run plan does not match install-plan\ninstall-plan: %#v\ndry-run: %#v", installPlan, dryRunPlan)
	}
}

func TestDependencyReleaseNameUsesNameWhenAliasMissing(t *testing.T) {
	got := dependencyReleaseName("app", dockpkg.Dependency{Name: "postgres"})
	if got != "app-postgres" {
		t.Fatalf("unexpected release name: %s", got)
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
