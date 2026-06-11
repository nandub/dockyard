package dockpkg

import "testing"

func TestManifestValidateRequiresValidName(t *testing.T) {
	manifest := Manifest{
		APIVersion: "dockyard.dev/v1alpha1",
		Name:       "../bad",
		Version:    "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
	}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected validation error for invalid name")
	}
}

func TestSafeJoinRejectsEscape(t *testing.T) {
	if _, err := SafeJoin("/tmp/pkg", "../secret"); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}

func TestSafeJoinAcceptsNestedPath(t *testing.T) {
	joined, err := SafeJoin("/tmp/pkg", "nested/compose.yaml")
	if err != nil {
		t.Fatalf("expected nested path to be accepted: %v", err)
	}
	if joined == "" {
		t.Fatal("expected joined path")
	}
}

func TestManifestValidateMissingAPIVersion(t *testing.T) {
	manifest := Manifest{
		Name:    "app",
		Version: "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
	}
	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing apiVersion")
	}
	if err.Error() != "dockyard.yaml is missing apiVersion; expected dockyard.dev/v1alpha1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifestValidateAcceptsDependencies(t *testing.T) {
	manifest := Manifest{
		APIVersion: "dockyard.dev/v1alpha1",
		Name:       "app",
		Version:    "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
		Dependencies: []Dependency{
			{
				Name:   "postgres",
				Alias:  "db",
				Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0",
				Values: map[string]any{
					"database": "app",
				},
			},
		},
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("expected dependencies to validate: %v", err)
	}
}

func TestManifestValidateRejectsDependencyWithoutPinnedOCIReference(t *testing.T) {
	manifest := Manifest{
		APIVersion: "dockyard.dev/v1alpha1",
		Name:       "app",
		Version:    "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
		Dependencies: []Dependency{
			{
				Name:   "postgres",
				Source: "oci://ghcr.io/nandub/dockyard/postgres",
			},
		},
	}
	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected dependency validation error")
	}
	if err.Error() != "dependencies[0].source OCI reference must include an explicit tag or digest" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifestValidateRejectsDuplicateDependencyAliases(t *testing.T) {
	manifest := Manifest{
		APIVersion: "dockyard.dev/v1alpha1",
		Name:       "app",
		Version:    "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
		Dependencies: []Dependency{
			{Name: "postgres", Alias: "db", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
			{Name: "mysql", Alias: "db", Source: "oci://ghcr.io/nandub/dockyard/mysql:0.1.0"},
		},
	}
	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected duplicate alias validation error")
	}
	if err.Error() != `duplicate dependency alias "db"` {
		t.Fatalf("unexpected error: %v", err)
	}
}
