package lock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/values"
	"go.yaml.in/yaml/v4"
)

const FileName = "dockyard.lock"

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ImageRef struct {
	Service string `json:"service"`
	Image   string `json:"image"`
	Digest  string `json:"digest,omitempty"`
}

type Lockfile struct {
	APIVersion            string       `json:"apiVersion"`
	GeneratedAt           time.Time    `json:"generatedAt"`
	PackageName           string       `json:"packageName"`
	PackageVersion        string       `json:"packageVersion"`
	AppVersion            string       `json:"appVersion"`
	Overlay               string       `json:"overlay,omitempty"`
	ValuesSHA256          string       `json:"valuesSha256"`
	RenderedComposeSHA256 string       `json:"renderedComposeSha256"`
	Images                []ImageRef   `json:"images"`
	Files                 []FileDigest `json:"files"`
}

func New(packageDir string, manifest *dockpkg.Manifest, vals map[string]any, rendered []byte, overlay string) (*Lockfile, error) {
	valuesBytes, err := yaml.Marshal(vals)
	if err != nil {
		return nil, fmt.Errorf("marshal values for lockfile: %w", err)
	}
	files, err := collectPackageFiles(packageDir)
	if err != nil {
		return nil, err
	}
	fileDigests := make([]FileDigest, 0, len(files))
	for _, rel := range files {
		path, err := dockpkg.SafeJoin(packageDir, rel)
		if err != nil {
			return nil, err
		}
		digest, err := sha256File(path)
		if err != nil {
			return nil, err
		}
		fileDigests = append(fileDigests, FileDigest{Path: rel, SHA256: digest})
	}
	images, err := ExtractImages(rendered)
	if err != nil {
		return nil, err
	}
	return &Lockfile{
		APIVersion:            "dockyard.dev/lockfile/v1alpha1",
		GeneratedAt:           time.Now().UTC(),
		PackageName:           manifest.Name,
		PackageVersion:        manifest.Version,
		AppVersion:            manifest.AppVersion,
		Overlay:               overlay,
		ValuesSHA256:          sha256Bytes(valuesBytes),
		RenderedComposeSHA256: sha256Bytes(rendered),
		Images:                images,
		Files:                 fileDigests,
	}, nil
}

func Write(path string, lf *Lockfile) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}
	return os.WriteFile(filepath.Clean(path), append(data, '\n'), 0o600)
}

func Read(path string) (*Lockfile, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read lockfile: %w", err)
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse lockfile: %w", err)
	}
	if lf.APIVersion != "dockyard.dev/lockfile/v1alpha1" {
		return nil, fmt.Errorf("unsupported lockfile apiVersion %q", lf.APIVersion)
	}
	return &lf, nil
}

func Verify(packageDir string, rendered []byte, vals map[string]any, overlay string) error {
	path := filepath.Join(packageDir, FileName)
	lf, err := Read(path)
	if err != nil {
		return err
	}
	valuesBytes, err := yaml.Marshal(vals)
	if err != nil {
		return err
	}
	if lf.ValuesSHA256 != sha256Bytes(valuesBytes) {
		return fmt.Errorf("lockfile values digest mismatch; run dockyard lock")
	}
	if lf.RenderedComposeSHA256 != sha256Bytes(rendered) {
		return fmt.Errorf("lockfile rendered compose digest mismatch; run dockyard lock")
	}
	if lf.Overlay != overlay {
		return fmt.Errorf("lockfile overlay mismatch; expected %q got %q", lf.Overlay, overlay)
	}
	for _, f := range lf.Files {
		if f.Path == FileName {
			continue
		}
		path, err := dockpkg.SafeJoin(packageDir, f.Path)
		if err != nil {
			return err
		}
		digest, err := sha256File(path)
		if err != nil {
			return err
		}
		if digest != f.SHA256 {
			return fmt.Errorf("lockfile file digest mismatch for %s", f.Path)
		}
	}
	return nil
}

func ExtractImages(rendered []byte) ([]ImageRef, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(rendered, &doc); err != nil {
		return nil, fmt.Errorf("parse rendered compose for image extraction: %w", err)
	}
	rawServices, ok := doc["services"].(map[string]any)
	if !ok {
		return nil, nil
	}
	images := make([]ImageRef, 0, len(rawServices))
	for serviceName, rawService := range rawServices {
		service, ok := rawService.(map[string]any)
		if !ok {
			continue
		}
		image, ok := service["image"].(string)
		if !ok || strings.TrimSpace(image) == "" {
			continue
		}
		images = append(images, ImageRef{
			Service: serviceName,
			Image:   image,
			Digest:  digestFromImageRef(image),
		})
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i].Service < images[j].Service
	})
	return images, nil
}

var digestPattern = regexp.MustCompile(`@sha256:([a-fA-F0-9]{64})`)

func digestFromImageRef(image string) string {
	match := digestPattern.FindStringSubmatch(image)
	if len(match) != 2 {
		return ""
	}
	return "sha256:" + strings.ToLower(match[1])
}

func collectPackageFiles(root string) ([]string, error) {
	var files []string
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == ".dockyard" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if name == FileName || name == "SHA256SUMS" || name == "package.provenance.json" {
			return nil
		}
		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("read file for digest: %w", err)
	}
	return sha256Bytes(data), nil
}

func sha256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func ValuesDigest(vals map[string]any) (string, error) {
	data, err := yaml.Marshal(vals)
	if err != nil {
		return "", err
	}
	return sha256Bytes(data), nil
}

func MergeValuesForLock(packageDir, overrideFile string) (map[string]any, error) {
	return values.LoadValues(packageDir, overrideFile)
}
