package dockpkg

import (
	"os"
	"testing"
)

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

func TestSafeJoinRejectsEmptyAndAbsolutePaths(t *testing.T) {
	if _, err := SafeJoin("/tmp/pkg", ""); err == nil {
		t.Fatal("expected empty path to be rejected")
	}
	if _, err := SafeJoin("/tmp/pkg", os.TempDir()); err == nil {
		t.Fatal("expected absolute path to be rejected")
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

func TestManifestValidateRejectsMissingVersionBaseAndUnsupportedType(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		want     string
	}{
		{
			name: "missing version",
			manifest: Manifest{
				APIVersion: "dockyard.dev/v1alpha1",
				Name:       "app",
				Compose:    ComposeConfig{Base: "compose.yaml"},
			},
			want: "manifest version is required",
		},
		{
			name: "missing compose base",
			manifest: Manifest{
				APIVersion: "dockyard.dev/v1alpha1",
				Name:       "app",
				Version:    "0.1.0",
			},
			want: "compose.base is required",
		},
		{
			name: "unsupported type",
			manifest: Manifest{
				APIVersion: "dockyard.dev/v1alpha1",
				Name:       "app",
				Version:    "0.1.0",
				Type:       "service",
				Compose:    ComposeConfig{Base: "compose.yaml"},
			},
			want: `unsupported package type "service"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if err.Error() != tt.want {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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

func TestManifestValidateRejectsDuplicateDependencyNames(t *testing.T) {
	manifest := validManifest()
	manifest.Dependencies = []Dependency{
		{Name: "postgres", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
		{Name: "postgres", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.2.0"},
	}

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected duplicate dependency name validation error")
	}
	if err.Error() != `duplicate dependency name "postgres"` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifestValidateRejectsInvalidDependencyAlias(t *testing.T) {
	manifest := validManifest()
	manifest.Dependencies = []Dependency{
		{Name: "postgres", Alias: "../db", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
	}

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected invalid dependency alias validation error")
	}
	if err.Error() != "dependencies[0].alias must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifestValidateRejectsDependencySourceWhitespace(t *testing.T) {
	manifest := validManifest()
	manifest.Dependencies = []Dependency{
		{Name: "postgres", Source: " oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
	}

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected dependency source whitespace validation error")
	}
	if err.Error() != "dependencies[0].source must not contain surrounding or embedded whitespace" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifestValidateRejectsMissingAndInvalidDependencyName(t *testing.T) {
	tests := []struct {
		name string
		dep  Dependency
		want string
	}{
		{
			name: "missing source",
			dep:  Dependency{Name: "postgres"},
			want: "dependencies[0].source is required",
		},
		{
			name: "invalid name",
			dep:  Dependency{Name: "../postgres", Source: "oci://ghcr.io/nandub/dockyard/postgres:0.1.0"},
			want: "dependencies[0].name must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifest()
			manifest.Dependencies = []Dependency{tt.dep}
			err := manifest.Validate()
			if err == nil {
				t.Fatal("expected dependency validation error")
			}
			if err.Error() != tt.want {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func validManifest() Manifest {
	return Manifest{
		APIVersion: "dockyard.dev/v1alpha1",
		Name:       "app",
		Version:    "0.1.0",
		Compose: ComposeConfig{
			Base: "compose.yaml",
		},
	}
}
