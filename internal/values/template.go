package values

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type schemaDoc struct {
	Description string
	Sensitive   bool
	Properties  map[string]schemaDoc
}

func GenerateTemplate(packageDir string) ([]byte, error) {
	vals, err := LoadValues(packageDir, "")
	if err != nil {
		return nil, err
	}
	doc, err := loadSchemaDocs(filepath.Join(packageDir, SchemaFile))
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	writeTemplateMap(&b, vals, doc, "", 0)
	return []byte(b.String()), nil
}

func loadSchemaDocs(path string) (schemaDoc, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return schemaDoc{}, nil
		}
		return schemaDoc{}, fmt.Errorf("read schema file: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return schemaDoc{}, fmt.Errorf("parse schema file: %w", err)
	}
	return parseSchemaDoc(raw), nil
}

func parseSchemaDoc(raw map[string]any) schemaDoc {
	doc := schemaDoc{
		Properties: map[string]schemaDoc{},
	}
	if description, ok := raw["description"].(string); ok {
		doc.Description = strings.TrimSpace(description)
	}
	if sensitive, ok := raw["x-dockyard-sensitive"].(bool); ok {
		doc.Sensitive = sensitive
	}
	if properties, ok := raw["properties"].(map[string]any); ok {
		for key, property := range properties {
			propertyMap, ok := property.(map[string]any)
			if !ok {
				continue
			}
			doc.Properties[key] = parseSchemaDoc(propertyMap)
		}
	}
	return doc
}

func writeTemplateMap(b *strings.Builder, vals map[string]any, doc schemaDoc, path string, indent int) {
	keys := make([]string, 0, len(vals))
	for key := range vals {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for idx, key := range keys {
		if idx > 0 {
			b.WriteString("\n")
		}
		value := vals[key]
		fullPath := key
		if path != "" {
			fullPath = path + "." + key
		}
		propertyDoc := doc.Properties[key]
		writeComment(b, propertyDoc.Description, indent)
		if propertyDoc.Sensitive || isSensitiveValueKey(fullPath) {
			writeComment(b, "Sensitive value. Keep this file private and do not commit production secrets.", indent)
		}
		padding := strings.Repeat(" ", indent)
		if nested, ok := value.(map[string]any); ok {
			b.WriteString(padding)
			b.WriteString(key)
			b.WriteString(":\n")
			writeTemplateMap(b, nested, propertyDoc, fullPath, indent+2)
			continue
		}
		b.WriteString(padding)
		b.WriteString(key)
		b.WriteString(": ")
		if propertyDoc.Sensitive || isSensitiveValueKey(fullPath) {
			b.WriteString(templateScalarForSensitive(value))
		} else {
			b.WriteString(formatTemplateScalar(value))
		}
		b.WriteString("\n")
	}
}

func writeComment(b *strings.Builder, description string, indent int) {
	description = strings.TrimSpace(description)
	if description == "" {
		return
	}
	padding := strings.Repeat(" ", indent)
	for _, line := range strings.Split(description, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString(padding)
		b.WriteString("# ")
		b.WriteString(line)
		b.WriteString("\n")
	}
}

func templateScalarForSensitive(value any) string {
	switch value.(type) {
	case int, int64, float64, bool:
		return formatTemplateScalar(value)
	default:
		return `""`
	}
}

func formatTemplateScalar(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case string:
		return strconv.Quote(typed)
	default:
		return strconv.Quote(fmt.Sprint(typed))
	}
}

func isSensitiveValueKey(key string) bool {
	lower := strings.ToLower(key)
	sensitiveParts := []string{"password", "secret", "token", "credential", "privatekey", "private_key", "apikey", "api_key"}
	for _, part := range sensitiveParts {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}
