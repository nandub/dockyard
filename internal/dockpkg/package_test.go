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
	if err.Error() != "Dockyard.yaml is missing apiVersion; expected dockyard.dev/v1alpha1" {
		t.Fatalf("unexpected error: %v", err)
	}
}
