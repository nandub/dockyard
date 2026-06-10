package values

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"go.yaml.in/yaml/v4"
)

const (
	DefaultValuesFile = "values.yaml"
	SchemaFile        = "values.schema.json"
)

func LoadValues(packageDir string, overrideFile string) (map[string]any, error) {
	basePath, err := dockpkg.SafeJoin(packageDir, DefaultValuesFile)
	if err != nil {
		return nil, err
	}
	baseValues, err := LoadYAMLMap(basePath)
	if err != nil {
		return nil, err
	}
	if overrideFile == "" {
		return baseValues, nil
	}
	overrideValues, err := LoadYAMLMap(overrideFile)
	if err != nil {
		return nil, err
	}
	return MergeMaps(baseValues, overrideValues), nil
}

func LoadYAMLMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read values file %q: %w", path, err)
	}
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("parse values file %q: %w", path, err)
	}
	if parsed == nil {
		return map[string]any{}, nil
	}
	return parsed, nil
}

func ValidateAgainstSchema(packageDir string, vals map[string]any) error {
	schemaPath, err := dockpkg.SafeJoin(packageDir, SchemaFile)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filepath.Clean(schemaPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read schema: %w", err)
	}

	var schemaDoc any
	if err := json.Unmarshal(data, &schemaDoc); err != nil {
		return fmt.Errorf("parse schema JSON: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schemaDoc); err != nil {
		return fmt.Errorf("load schema: %w", err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	jsonBytes, err := json.Marshal(vals)
	if err != nil {
		return fmt.Errorf("marshal values for schema validation: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		return fmt.Errorf("decode values for schema validation: %w", err)
	}
	if err := schema.Validate(decoded); err != nil {
		return fmt.Errorf("values validation failed: %w", err)
	}
	return nil
}

func WriteValues(path string, vals map[string]any) error {
	data, err := yaml.Marshal(vals)
	if err != nil {
		return fmt.Errorf("marshal values: %w", err)
	}
	return os.WriteFile(filepath.Clean(path), data, 0o600)
}

func MergeMaps(base map[string]any, override map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for key, value := range base {
		result[key] = value
	}
	for key, overrideValue := range override {
		if baseNested, ok := result[key].(map[string]any); ok {
			if overrideNested, ok := overrideValue.(map[string]any); ok {
				result[key] = MergeMaps(baseNested, overrideNested)
				continue
			}
		}
		result[key] = overrideValue
	}
	return result
}

func CopyFile(dst string, src string, mode os.FileMode) error {
	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile(filepath.Clean(dst), os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	return nil
}
