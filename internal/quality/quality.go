package quality

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/policy"
	"github.com/nandub/dockyard/internal/render"
	"github.com/nandub/dockyard/internal/values"
)

type Severity string

const (
	SeverityFail Severity = "FAIL"
	SeverityWarn Severity = "WARN"
	SeverityOK   Severity = "OK"
)

type Check struct {
	Name     string   `json:"name"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Details  []string `json:"details,omitempty"`
	Advisory bool     `json:"advisory,omitempty"`
}

type Report struct {
	PackageName    string  `json:"packageName,omitempty"`
	PackageVersion string  `json:"packageVersion,omitempty"`
	Checks         []Check `json:"checks"`
}

type Options struct {
	Strict        bool
	AllowAdvisory bool
}

func LintPackage(packageDir string, opts Options) (Report, error) {
	var report Report

	manifest, err := dockpkg.LoadManifest(packageDir)
	if err != nil {
		report.Checks = append(report.Checks, Check{Name: "Dockyard.yaml", Severity: SeverityFail, Message: err.Error()})
		return report, nil
	}
	report.PackageName = manifest.Name
	report.PackageVersion = manifest.Version
	report.Checks = append(report.Checks, Check{Name: "Dockyard.yaml", Severity: SeverityOK, Message: "manifest is valid"})

	report.Checks = append(report.Checks, checkRequiredFiles(packageDir, opts.Strict)...)
	report.Checks = append(report.Checks, checkForbiddenFiles(packageDir)...)
	report.Checks = append(report.Checks, checkValuesAndSchema(packageDir, opts.Strict)...)
	report.Checks = append(report.Checks, checkDefaultRender(packageDir, manifest)...)

	return report, nil
}

func HasFailures(report Report) bool {
	for _, check := range report.Checks {
		if check.Severity == SeverityFail {
			return true
		}
	}
	return false
}

func HasBlockingFindings(report Report, opts Options) bool {
	for _, check := range report.Checks {
		if check.Severity == SeverityFail {
			return true
		}
		if opts.Strict && check.Severity == SeverityWarn {
			if opts.AllowAdvisory && check.Advisory {
				continue
			}
			return true
		}
	}
	return false
}

func checkRequiredFiles(packageDir string, strict bool) []Check {
	files := []struct {
		name         string
		strictFatal  bool
		missingNotes string
	}{
		{name: "README.md", strictFatal: true, missingNotes: "package README is missing"},
		{name: "SECURITY.md", strictFatal: true, missingNotes: "package security contact/instructions are missing"},
		{name: "LICENSE", strictFatal: false, missingNotes: "package license file is missing; recommended before public distribution"},
	}

	var checks []Check
	for _, file := range files {
		path, err := dockpkg.SafeJoin(packageDir, file.name)
		if err != nil {
			checks = append(checks, Check{Name: file.name, Severity: SeverityFail, Message: err.Error()})
			continue
		}
		if _, err := os.Stat(path); err != nil {
			severity := SeverityWarn
			if strict && file.strictFatal {
				severity = SeverityFail
			}
			checks = append(checks, Check{Name: file.name, Severity: severity, Message: file.missingNotes, Advisory: file.name == "LICENSE"})
			continue
		}
		checks = append(checks, Check{Name: file.name, Severity: SeverityOK, Message: "present"})
	}
	return checks
}

func checkForbiddenFiles(packageDir string) []Check {
	var findings []string
	forbiddenDirs := map[string]struct{}{
		".git": {}, ".dockyard": {}, "node_modules": {}, "vendor": {},
		"deploy-values": {}, "dockyard-work": {}, "dockyard-artifacts": {},
	}
	forbiddenExact := map[string]struct{}{
		".env": {}, "id_rsa": {}, "id_ed25519": {}, "docker.sock": {},
	}
	forbiddenSuffix := []string{".pem", ".key", ".p12", ".pfx"}

	err := filepath.WalkDir(filepath.Clean(packageDir), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if _, ok := forbiddenDirs[name]; ok {
				rel, _ := filepath.Rel(packageDir, path)
				findings = append(findings, filepath.ToSlash(rel)+"/")
				return filepath.SkipDir
			}
			return nil
		}
		if _, ok := forbiddenExact[name]; ok {
			rel, _ := filepath.Rel(packageDir, path)
			findings = append(findings, filepath.ToSlash(rel))
			return nil
		}
		lower := strings.ToLower(name)
		for _, suffix := range forbiddenSuffix {
			if strings.HasSuffix(lower, suffix) {
				rel, _ := filepath.Rel(packageDir, path)
				findings = append(findings, filepath.ToSlash(rel))
				break
			}
		}
		return nil
	})
	if err != nil {
		return []Check{{Name: "forbidden files", Severity: SeverityFail, Message: err.Error()}}
	}
	if len(findings) > 0 {
		return []Check{{Name: "forbidden files", Severity: SeverityFail, Message: "package contains files or directories that should not be packaged", Details: findings}}
	}
	return []Check{{Name: "forbidden files", Severity: SeverityOK, Message: "no forbidden files found"}}
}

func checkValuesAndSchema(packageDir string, strict bool) []Check {
	var checks []Check

	vals, err := values.LoadValues(packageDir, "")
	if err != nil {
		return []Check{{Name: "values.yaml", Severity: SeverityFail, Message: err.Error()}}
	}
	checks = append(checks, Check{Name: "values.yaml", Severity: SeverityOK, Message: "default values loaded"})

	schemaPath, err := dockpkg.SafeJoin(packageDir, values.SchemaFile)
	if err != nil {
		checks = append(checks, Check{Name: values.SchemaFile, Severity: SeverityFail, Message: err.Error()})
		return checks
	}
	data, err := os.ReadFile(filepath.Clean(schemaPath))
	if err != nil {
		severity := SeverityWarn
		if strict {
			severity = SeverityFail
		}
		checks = append(checks, Check{Name: values.SchemaFile, Severity: severity, Message: "schema is missing; schema is required for high-quality packages"})
		return checks
	}
	if err := values.ValidateAgainstSchema(packageDir, vals); err != nil {
		checks = append(checks, Check{Name: values.SchemaFile, Severity: SeverityFail, Message: err.Error()})
		return checks
	}
	checks = append(checks, Check{Name: values.SchemaFile, Severity: SeverityOK, Message: "schema validates default values"})

	var schema any
	if err := json.Unmarshal(data, &schema); err != nil {
		checks = append(checks, Check{Name: values.SchemaFile, Severity: SeverityFail, Message: "schema JSON could not be parsed: " + err.Error()})
		return checks
	}
	missingDescriptions, missingSensitiveMarkers := inspectSchemaQuality(schema, "")
	if len(missingDescriptions) > 0 {
		severity := SeverityWarn
		if strict {
			severity = SeverityFail
		}
		checks = append(checks, Check{Name: "schema descriptions", Severity: severity, Message: "public values should include description fields", Details: missingDescriptions})
	} else {
		checks = append(checks, Check{Name: "schema descriptions", Severity: SeverityOK, Message: "descriptions present for public values"})
	}
	if len(missingSensitiveMarkers) > 0 {
		severity := SeverityWarn
		if strict {
			severity = SeverityFail
		}
		checks = append(checks, Check{Name: "schema sensitive markers", Severity: severity, Message: "secret-like values should set x-dockyard-sensitive: true", Details: missingSensitiveMarkers})
	} else {
		checks = append(checks, Check{Name: "schema sensitive markers", Severity: SeverityOK, Message: "secret-like schema fields are marked or absent"})
	}
	return checks
}

func inspectSchemaQuality(schema any, prefix string) ([]string, []string) {
	obj, ok := schema.(map[string]any)
	if !ok {
		return nil, nil
	}

	props, _ := obj["properties"].(map[string]any)
	if len(props) == 0 {
		return nil, nil
	}

	var missingDescriptions []string
	var missingSensitiveMarkers []string
	for name, raw := range props {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		prop, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		childProps, _ := prop["properties"].(map[string]any)
		if len(childProps) > 0 {
			childMissingDesc, childMissingSensitive := inspectSchemaQuality(prop, path)
			missingDescriptions = append(missingDescriptions, childMissingDesc...)
			missingSensitiveMarkers = append(missingSensitiveMarkers, childMissingSensitive...)
			continue
		}
		if _, ok := prop["description"].(string); !ok {
			missingDescriptions = append(missingDescriptions, path)
		}
		if isSensitiveName(name) || isSensitiveName(path) {
			if marked, ok := prop["x-dockyard-sensitive"].(bool); !ok || !marked {
				missingSensitiveMarkers = append(missingSensitiveMarkers, path)
			}
		}
	}
	return missingDescriptions, missingSensitiveMarkers
}

func checkDefaultRender(packageDir string, manifest *dockpkg.Manifest) []Check {
	vals, err := values.LoadValues(packageDir, "")
	if err != nil {
		return []Check{{Name: "default render", Severity: SeverityFail, Message: err.Error()}}
	}
	rendered, err := render.RenderCompose(packageDir, manifest, vals, "")
	if err != nil {
		return []Check{{Name: "default render", Severity: SeverityFail, Message: err.Error()}}
	}
	checks := []Check{{Name: "default render", Severity: SeverityOK, Message: "compose.yaml renders with default values"}}
	findings, err := policy.LintCompose(rendered, manifest.Security)
	if err != nil {
		checks = append(checks, Check{Name: "policy lint", Severity: SeverityFail, Message: err.Error()})
		return checks
	}
	var high []string
	var nonHigh []string
	for _, finding := range findings {
		msg := finding.Message
		if finding.Service != "" {
			msg = finding.Service + ": " + msg
		}
		if finding.Severity == policy.SeverityHigh {
			high = append(high, msg)
		} else {
			nonHigh = append(nonHigh, string(finding.Severity)+": "+msg)
		}
	}
	if len(high) > 0 {
		checks = append(checks, Check{Name: "policy lint", Severity: SeverityFail, Message: "HIGH policy findings in default render", Details: high})
	} else if len(nonHigh) > 0 {
		checks = append(checks, Check{Name: "policy lint", Severity: SeverityWarn, Message: "non-blocking policy findings in default render", Details: nonHigh})
	} else {
		checks = append(checks, Check{Name: "policy lint", Severity: SeverityOK, Message: "default render passes configured policy"})
	}
	return checks
}

func isSensitiveName(name string) bool {
	lower := strings.ToLower(name)
	for _, marker := range []string{"password", "secret", "token", "credential", "privatekey", "private_key", "apikey", "api_key"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
