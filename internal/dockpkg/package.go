package dockpkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"go.yaml.in/yaml/v4"
)

const ManifestFileName = "Dockyard.yaml"

var packageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

type Manifest struct {
	APIVersion  string         `yaml:"apiVersion" json:"apiVersion"`
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Version     string         `yaml:"version" json:"version"`
	AppVersion  string         `yaml:"appVersion" json:"appVersion"`
	Type        string         `yaml:"type" json:"type"`
	Compose     ComposeConfig  `yaml:"compose" json:"compose"`
	Security    SecurityPolicy `yaml:"security" json:"security"`
}

type ComposeConfig struct {
	Base     string            `yaml:"base" json:"base"`
	Overlays map[string]string `yaml:"overlays" json:"overlays"`
}

type SecurityPolicy struct {
	RequireNonRoot                bool `yaml:"requireNonRoot" json:"requireNonRoot"`
	RequireHealthchecks           bool `yaml:"requireHealthchecks" json:"requireHealthchecks"`
	RequireReadOnlyRootFilesystem bool `yaml:"requireReadOnlyRootFilesystem" json:"requireReadOnlyRootFilesystem"`
	RequireNoNewPrivileges        bool `yaml:"requireNoNewPrivileges" json:"requireNoNewPrivileges"`
	RequireCapDropAll             bool `yaml:"requireCapDropAll" json:"requireCapDropAll"`
	DisallowPrivileged            bool `yaml:"disallowPrivileged" json:"disallowPrivileged"`
	DisallowHostNetwork           bool `yaml:"disallowHostNetwork" json:"disallowHostNetwork"`
	DisallowDockerSocketMount     bool `yaml:"disallowDockerSocketMount" json:"disallowDockerSocketMount"`
	DisallowHostPathMounts        bool `yaml:"disallowHostPathMounts" json:"disallowHostPathMounts"`
	DisallowLatestTag             bool `yaml:"disallowLatestTag" json:"disallowLatestTag"`
}

func LoadManifest(packageDir string) (*Manifest, error) {
	manifestPath, err := SafeJoin(packageDir, ManifestFileName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (m Manifest) Validate() error {
	if m.APIVersion == "" {
		return errors.New("Dockyard.yaml is missing apiVersion; expected dockyard.dev/v1alpha1")
	}
	if m.APIVersion != "dockyard.dev/v1alpha1" {
		return fmt.Errorf("unsupported apiVersion %q; expected dockyard.dev/v1alpha1", m.APIVersion)
	}
	if !packageNamePattern.MatchString(m.Name) {
		return errors.New("manifest name must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$")
	}
	if m.Version == "" {
		return errors.New("manifest version is required")
	}
	if m.Compose.Base == "" {
		return errors.New("compose.base is required")
	}
	if m.Type != "" && m.Type != "application" && m.Type != "library" {
		return fmt.Errorf("unsupported package type %q", m.Type)
	}
	return nil
}

func SafeJoin(baseDir string, relPath string) (string, error) {
	if relPath == "" {
		return "", errors.New("path must not be empty")
	}
	if filepath.IsAbs(relPath) {
		return "", fmt.Errorf("absolute package path %q is not allowed", relPath)
	}
	cleanBase, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", fmt.Errorf("resolve base path: %w", err)
	}
	candidate := filepath.Join(cleanBase, filepath.Clean(relPath))
	cleanCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve candidate path: %w", err)
	}
	rel, err := filepath.Rel(cleanBase, cleanCandidate)
	if err != nil {
		return "", fmt.Errorf("check path containment: %w", err)
	}
	if rel == ".." || relHasParentPrefix(rel) {
		return "", fmt.Errorf("package path %q escapes package directory", relPath)
	}
	return cleanCandidate, nil
}

func relHasParentPrefix(rel string) bool {
	return len(rel) > 3 && rel[:2] == ".." && os.IsPathSeparator(rel[2])
}
