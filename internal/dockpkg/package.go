package dockpkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nandub/dockyard/internal/format"
	"go.yaml.in/yaml/v4"
)

const ManifestFileName = "Dockyard.yaml"

var packageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

type Manifest struct {
	APIVersion   string         `yaml:"apiVersion" json:"apiVersion"`
	Name         string         `yaml:"name" json:"name"`
	Description  string         `yaml:"description" json:"description"`
	Version      string         `yaml:"version" json:"version"`
	AppVersion   string         `yaml:"appVersion" json:"appVersion"`
	Type         string         `yaml:"type" json:"type"`
	Compose      ComposeConfig  `yaml:"compose" json:"compose"`
	Security     SecurityPolicy `yaml:"security" json:"security"`
	Dependencies []Dependency   `yaml:"dependencies" json:"dependencies,omitempty"`
}

type Dependency struct {
	Name    string         `yaml:"name" json:"name"`
	Source  string         `yaml:"source" json:"source"`
	Alias   string         `yaml:"alias,omitempty" json:"alias,omitempty"`
	Version string         `yaml:"version,omitempty" json:"version,omitempty"`
	Values  map[string]any `yaml:"values,omitempty" json:"values,omitempty"`
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
		return fmt.Errorf("dockyard.yaml is missing apiVersion; expected %s", format.ManifestAPIVersion)
	}
	if m.APIVersion != format.ManifestAPIVersion {
		return fmt.Errorf("unsupported apiVersion %q; expected %s", m.APIVersion, format.ManifestAPIVersion)
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
	if err := validateDependencies(m.Dependencies); err != nil {
		return err
	}
	return nil
}

func validateDependencies(dependencies []Dependency) error {
	seenNames := map[string]struct{}{}
	seenAliases := map[string]struct{}{}
	for i, dep := range dependencies {
		position := fmt.Sprintf("dependencies[%d]", i)
		if !packageNamePattern.MatchString(dep.Name) {
			return fmt.Errorf("%s.name must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$", position)
		}
		if _, ok := seenNames[dep.Name]; ok {
			return fmt.Errorf("duplicate dependency name %q", dep.Name)
		}
		seenNames[dep.Name] = struct{}{}

		if dep.Alias != "" {
			if !packageNamePattern.MatchString(dep.Alias) {
				return fmt.Errorf("%s.alias must match ^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$", position)
			}
			if _, ok := seenAliases[dep.Alias]; ok {
				return fmt.Errorf("duplicate dependency alias %q", dep.Alias)
			}
			seenAliases[dep.Alias] = struct{}{}
		}

		if err := validateDependencySource(position, dep.Source); err != nil {
			return err
		}
	}
	return nil
}

func validateDependencySource(position string, source string) error {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return fmt.Errorf("%s.source is required", position)
	}
	if trimmed != source || strings.ContainsAny(source, " \t\n\r") {
		return fmt.Errorf("%s.source must not contain surrounding or embedded whitespace", position)
	}
	if strings.HasPrefix(strings.ToLower(source), "oci://") {
		ref := source[len("oci://"):]
		lastSlash := strings.LastIndex(ref, "/")
		tagIndex := strings.LastIndex(ref, ":")
		hasTag := tagIndex > lastSlash
		hasDigest := strings.Contains(ref[lastSlash+1:], "@sha256:")
		if !hasTag && !hasDigest {
			return fmt.Errorf("%s.source OCI reference must include an explicit tag or digest", position)
		}
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
