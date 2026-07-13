package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nandub/dockyard/internal/oci"
	"go.yaml.in/yaml/v4"
)

const (
	DefaultCatalogRef = "oci://ghcr.io/nandub/dockyard-packages/catalog:latest"
	EnvCatalog        = "DOCKYARD_CATALOG"
	APIVersion        = "dockyard.dev/catalog/v1alpha1"
)

type Index struct {
	APIVersion string    `json:"apiVersion" yaml:"apiVersion"`
	Registry   string    `json:"registry" yaml:"registry"`
	Packages   []Package `json:"packages" yaml:"packages"`
}

type Package struct {
	Name        string   `json:"name" yaml:"name"`
	Latest      string   `json:"latest" yaml:"latest"`
	Description string   `json:"description" yaml:"description"`
	Source      string   `json:"source,omitempty" yaml:"source,omitempty"`
	Versions    []string `json:"versions,omitempty" yaml:"versions,omitempty"`
}

var packageNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Reference() string {
	ref := strings.TrimSpace(os.Getenv(EnvCatalog))
	if ref == "" {
		return DefaultCatalogRef
	}
	if strings.HasPrefix(ref, "oci://") || strings.HasPrefix(ref, "file://") {
		return strings.TrimRight(ref, "/")
	}
	if strings.HasSuffix(ref, ".yaml") || strings.HasSuffix(ref, ".yml") {
		return filepath.Clean(ref)
	}
	return "oci://" + strings.TrimRight(ref, "/") + "/catalog:latest"
}

func List() ([]Package, error) {
	idx, err := Load(context.Background())
	if err != nil {
		return nil, err
	}
	out := append([]Package(nil), idx.Packages...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func Get(name string) (Package, bool, error) {
	idx, err := Load(context.Background())
	if err != nil {
		return Package{}, false, err
	}
	return idx.Package(name)
}

func Resolve(input string) (string, bool, error) {
	idx, err := Load(context.Background())
	if err != nil {
		if strings.HasPrefix(input, "catalog://") {
			return "", true, err
		}
		return input, false, err
	}
	return ResolveWithIndex(idx, input)
}

func ResolveName(name string, version string) (string, error) {
	idx, err := Load(context.Background())
	if err != nil {
		return "", err
	}
	return idx.ResolveName(name, version)
}

func Load(ctx context.Context) (Index, error) {
	return LoadReference(ctx, Reference())
}

func LoadReference(ctx context.Context, ref string) (Index, error) {
	if strings.HasPrefix(ref, "file://") {
		data, err := os.ReadFile(strings.TrimPrefix(ref, "file://"))
		if err != nil {
			return Index{}, fmt.Errorf("read catalog index: %w", err)
		}
		return LoadBytes(data)
	}
	if strings.HasSuffix(ref, ".yaml") || strings.HasSuffix(ref, ".yml") {
		data, err := os.ReadFile(ref)
		if err != nil {
			return Index{}, fmt.Errorf("read catalog index: %w", err)
		}
		return LoadBytes(data)
	}
	if !strings.HasPrefix(ref, "oci://") {
		return Index{}, fmt.Errorf("catalog reference must start with oci://")
	}
	cached, ok := readCached(ref)
	if ok {
		return cached, nil
	}
	idx, err := pullIndex(ctx, ref)
	if err != nil {
		return Index{}, err
	}
	_ = writeCached(ref, idx)
	return idx, nil
}

func LoadBytes(data []byte) (Index, error) {
	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return Index{}, fmt.Errorf("parse catalog index: %w", err)
	}
	if err := idx.Validate(); err != nil {
		return Index{}, err
	}
	return idx, nil
}

func ResolveWithIndex(idx Index, input string) (string, bool, error) {
	if strings.HasPrefix(input, "catalog://") {
		ref := strings.TrimPrefix(input, "catalog://")
		name, version, err := parseRef(ref)
		if err != nil {
			return "", true, err
		}
		resolved, err := idx.ResolveName(name, version)
		return resolved, true, err
	}
	if _, ok, _ := idx.Package(input); ok {
		resolved, err := idx.ResolveName(input, "")
		return resolved, true, err
	}
	return input, false, nil
}

func (idx Index) Validate() error {
	if idx.APIVersion != APIVersion {
		return fmt.Errorf("unsupported catalog apiVersion %q; expected %q", idx.APIVersion, APIVersion)
	}
	if strings.TrimSpace(idx.Registry) == "" {
		return errors.New("catalog registry is required")
	}
	seen := map[string]struct{}{}
	for _, pkg := range idx.Packages {
		if !packageNamePattern.MatchString(pkg.Name) {
			return fmt.Errorf("invalid catalog package name %q", pkg.Name)
		}
		if _, ok := seen[pkg.Name]; ok {
			return fmt.Errorf("duplicate catalog package %q", pkg.Name)
		}
		seen[pkg.Name] = struct{}{}
		if strings.TrimSpace(pkg.Latest) == "" {
			return fmt.Errorf("catalog package %q is missing latest version", pkg.Name)
		}
		if strings.TrimSpace(pkg.Description) == "" {
			return fmt.Errorf("catalog package %q is missing description", pkg.Name)
		}
	}
	return nil
}

func (idx Index) Package(name string) (Package, bool, error) {
	if !packageNamePattern.MatchString(name) {
		return Package{}, false, fmt.Errorf("invalid catalog package name %q", name)
	}
	for _, pkg := range idx.Packages {
		if pkg.Name == name {
			return pkg, true, nil
		}
	}
	return Package{}, false, nil
}

func (idx Index) ResolveName(name string, version string) (string, error) {
	pkg, ok, err := idx.Package(name)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("package %q was not found in the configured catalog", name)
	}
	if version == "" {
		version = pkg.Latest
	}
	if version == "" {
		return "", fmt.Errorf("package %q has no default version; specify catalog://%s:VERSION", name, name)
	}
	if len(pkg.Versions) > 0 && !contains(pkg.Versions, version) {
		return "", fmt.Errorf("package %q version %q was not found in the configured catalog", name, version)
	}
	source := strings.TrimSpace(pkg.Source)
	if source == "" {
		source = "oci://" + strings.TrimRight(idx.Registry, "/") + "/" + name
	}
	source = strings.TrimRight(source, "/")
	if strings.Contains(source[strings.LastIndex(source, "/")+1:], ":") {
		return source, nil
	}
	return fmt.Sprintf("%s:%s", source, version), nil
}

func parseRef(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("catalog source is missing a package name")
	}
	name := ref
	version := ""
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		name = ref[:i]
		version = ref[i+1:]
		if version == "" {
			return "", "", fmt.Errorf("catalog source %q is missing a version after ':'", ref)
		}
	}
	if !packageNamePattern.MatchString(name) {
		return "", "", fmt.Errorf("invalid catalog package name %q", name)
	}
	return name, version, nil
}

func pullIndex(ctx context.Context, ref string) (Index, error) {
	tempRoot, err := os.MkdirTemp("", "dockyard-catalog-*")
	if err != nil {
		return Index{}, fmt.Errorf("create catalog temp dir: %w", err)
	}
	defer os.RemoveAll(tempRoot)

	pullCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := oci.PullFiles(pullCtx, ref, tempRoot, oci.PullOptions{Quiet: true}); err != nil {
		return Index{}, fmt.Errorf("pull catalog %q: %w", ref, err)
	}
	path, err := findIndexFile(tempRoot)
	if err != nil {
		return Index{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Index{}, fmt.Errorf("read catalog index: %w", err)
	}
	return LoadBytes(data)
}

func findIndexFile(dir string) (string, error) {
	candidates := []string{"catalog.yaml", "catalog.yml"}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("catalog artifact did not contain catalog.yaml")
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".dockyard", "cache", "catalogs"), nil
}

func cachePath(ref string) (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(ref))
	return filepath.Join(dir, hex.EncodeToString(sum[:])+".yaml"), nil
}

func readCached(ref string) (Index, bool) {
	path, err := cachePath(ref)
	if err != nil {
		return Index{}, false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || time.Since(info.ModTime()) > 5*time.Minute {
		return Index{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Index{}, false
	}
	idx, err := LoadBytes(data)
	if err != nil {
		return Index{}, false
	}
	return idx, true
}

func writeCached(ref string, idx Index) error {
	path, err := cachePath(ref)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
