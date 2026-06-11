package render

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/values"
	"go.yaml.in/yaml/v4"
)

var placeholderPattern = regexp.MustCompile(`\$\{([^}]+)\}`)
var allowedReferencePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

type Diagnostic struct {
	Key      string
	Value    string
	Masked   bool
	Location string
}

type Result struct {
	YAML        []byte
	Diagnostics []Diagnostic
}

func RenderCompose(packageDir string, manifest *dockpkg.Manifest, vals map[string]any, overlay string) ([]byte, error) {
	result, err := RenderComposeWithDiagnostics(packageDir, manifest, vals, overlay)
	if err != nil {
		return nil, err
	}
	return result.YAML, nil
}

func RenderComposeWithDiagnostics(packageDir string, manifest *dockpkg.Manifest, vals map[string]any, overlay string) (*Result, error) {
	basePath, err := dockpkg.SafeJoin(packageDir, manifest.Compose.Base)
	if err != nil {
		return nil, err
	}

	base, diagnostics, err := renderFile(basePath, vals)
	if err != nil {
		return nil, err
	}

	var rendered map[string]any
	if err := yaml.Unmarshal(base, &rendered); err != nil {
		return nil, fmt.Errorf("parse rendered compose base: %w", err)
	}

	if overlay != "" {
		overlayFile, ok := manifest.Compose.Overlays[overlay]
		if !ok {
			return nil, fmt.Errorf("unknown overlay %q", overlay)
		}

		overlayPath, err := dockpkg.SafeJoin(packageDir, overlayFile)
		if err != nil {
			return nil, err
		}

		overlayBytes, overlayDiagnostics, err := renderFile(overlayPath, vals)
		if err != nil {
			return nil, err
		}

		var overlayMap map[string]any
		if err := yaml.Unmarshal(overlayBytes, &overlayMap); err != nil {
			return nil, fmt.Errorf("parse rendered compose overlay: %w", err)
		}
		diagnostics = append(diagnostics, overlayDiagnostics...)
		rendered = values.MergeMaps(rendered, overlayMap)
	}

	out, err := yaml.Marshal(rendered)
	if err != nil {
		return nil, fmt.Errorf("marshal rendered compose: %w", err)
	}

	sort.Slice(diagnostics, func(i, j int) bool {
		return diagnostics[i].Key < diagnostics[j].Key
	})

	return &Result{YAML: out, Diagnostics: diagnostics}, nil
}

func renderFile(path string, vals map[string]any) ([]byte, []Diagnostic, error) {
	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read compose file %q: %w", cleanPath, err)
	}

	flat := FlattenValues("", vals)
	input := string(data)
	missing := map[string]struct{}{}
	invalid := map[string]struct{}{}
	diagnosticByKey := map[string]Diagnostic{}

	rendered := placeholderPattern.ReplaceAllStringFunc(input, func(token string) string {
		match := placeholderPattern.FindStringSubmatch(token)
		if len(match) != 2 {
			return token
		}
		key := strings.TrimSpace(match[1])
		if !allowedReferencePattern.MatchString(key) {
			invalid[key] = struct{}{}
			return token
		}
		value, ok := flat[key]
		if !ok {
			missing[key] = struct{}{}
			return token
		}
		display := fmt.Sprint(value)
		masked := IsSensitiveKey(key)
		if masked {
			display = "********"
		}
		diagnosticByKey[key] = Diagnostic{
			Key:      key,
			Value:    display,
			Masked:   masked,
			Location: cleanPath,
		}
		return fmt.Sprint(value)
	})

	if len(invalid) > 0 || len(missing) > 0 {
		var b strings.Builder
		b.WriteString("render failed")
		if len(invalid) > 0 {
			b.WriteString("; invalid placeholders: ")
			b.WriteString(strings.Join(sortedKeys(invalid), ", "))
		}
		if len(missing) > 0 {
			b.WriteString("; unresolved values: ")
			b.WriteString(strings.Join(sortedKeys(missing), ", "))
		}
		return nil, nil, errors.New(b.String())
	}

	diagnostics := make([]Diagnostic, 0, len(diagnosticByKey))
	for _, diagnostic := range diagnosticByKey {
		diagnostics = append(diagnostics, diagnostic)
	}

	return []byte(rendered), diagnostics, nil
}

func FlattenValues(prefix string, vals map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range vals {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		nested, ok := value.(map[string]any)
		if ok {
			for nestedKey, nestedValue := range FlattenValues(fullKey, nested) {
				result[nestedKey] = nestedValue
			}
			continue
		}
		result[fullKey] = value
	}
	return result
}

func IsSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitiveParts := []string{"password", "secret", "token", "credential", "privatekey", "private_key", "apikey", "api_key"}
	for _, part := range sensitiveParts {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}

func sortedKeys(items map[string]struct{}) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func NormalizeYAML(data []byte) ([]byte, error) {
	var parsed any
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false)
	if err := decoder.Decode(&parsed); err != nil {
		return nil, err
	}
	return yaml.Marshal(parsed)
}
