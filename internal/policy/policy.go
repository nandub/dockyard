package policy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nandub/dockyard/internal/dockpkg"
	"go.yaml.in/yaml/v4"
)

type FindingSeverity string

const (
	SeverityHigh   FindingSeverity = "HIGH"
	SeverityMedium FindingSeverity = "MEDIUM"
	SeverityLow    FindingSeverity = "LOW"
)

type Finding struct {
	Severity FindingSeverity `json:"severity"`
	Service  string          `json:"service,omitempty"`
	Message  string          `json:"message"`
}

func HasHighFindings(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Severity == SeverityHigh {
			return true
		}
	}
	return false
}

func LintCompose(composeBytes []byte, security dockpkg.SecurityPolicy) ([]Finding, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(composeBytes, &doc); err != nil {
		return nil, fmt.Errorf("parse compose for policy linting: %w", err)
	}

	rawServices, ok := doc["services"].(map[string]any)
	if !ok {
		return []Finding{{Severity: SeverityHigh, Message: "compose file must define services"}}, nil
	}

	var findings []Finding
	for serviceName, rawService := range rawServices {
		service, ok := rawService.(map[string]any)
		if !ok {
			findings = append(findings, Finding{Severity: SeverityHigh, Service: serviceName, Message: "service definition must be an object"})
			continue
		}
		findings = append(findings, lintService(serviceName, service, security)...)
	}

	return findings, nil
}

func lintService(serviceName string, service map[string]any, security dockpkg.SecurityPolicy) []Finding {
	var findings []Finding

	if security.RequireNonRoot {
		user, ok := service["user"].(string)
		if !ok || user == "" || user == "root" || strings.HasPrefix(user, "0") {
			findings = append(findings, Finding{Severity: SeverityHigh, Service: serviceName, Message: "service should run as a non-root user"})
		}
	}

	if security.RequireHealthchecks {
		if _, ok := service["healthcheck"]; !ok {
			findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "service should define a healthcheck"})
		}
	}

	if security.RequireReadOnlyRootFilesystem {
		if readOnly, ok := service["read_only"].(bool); !ok || !readOnly {
			findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "service should set read_only: true"})
		}
	}

	if security.RequireNoNewPrivileges {
		if !hasStringValue(service["security_opt"], "no-new-privileges:true") {
			findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "service should set security_opt: no-new-privileges:true"})
		}
	}

	if security.RequireCapDropAll {
		if !hasStringValue(service["cap_drop"], "ALL") {
			findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "service should drop all Linux capabilities with cap_drop: [ALL]"})
		}
	}

	if security.DisallowPrivileged {
		if privileged, ok := service["privileged"].(bool); ok && privileged {
			findings = append(findings, Finding{Severity: SeverityHigh, Service: serviceName, Message: "privileged mode is not allowed"})
		}
	}

	if security.DisallowHostNetwork {
		if networkMode, ok := service["network_mode"].(string); ok && networkMode == "host" {
			findings = append(findings, Finding{Severity: SeverityHigh, Service: serviceName, Message: "host network mode is not allowed"})
		}
	}

	if security.DisallowDockerSocketMount {
		for _, volume := range volumeStrings(service["volumes"]) {
			if strings.Contains(volume, "/var/run/docker.sock") {
				findings = append(findings, Finding{Severity: SeverityHigh, Service: serviceName, Message: "mounting the Docker socket is not allowed"})
			}
		}
	}

	if security.DisallowHostPathMounts {
		for _, volume := range volumeStrings(service["volumes"]) {
			if isLikelyHostPathVolume(volume) {
				findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "host path mounts are discouraged; prefer named volumes"})
			}
		}
	}

	if security.DisallowLatestTag {
		image, ok := service["image"].(string)
		if ok && usesLatestTag(image) {
			findings = append(findings, Finding{Severity: SeverityMedium, Service: serviceName, Message: "image should not use the latest tag"})
		}
	}

	return findings
}

func hasStringValue(value any, expected string) bool {
	for _, item := range stringSlice(value) {
		if item == expected {
			return true
		}
	}
	return false
}

func stringSlice(value any) []string {
	rawItems, ok := value.([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(rawItems))
	for _, rawItem := range rawItems {
		item, ok := rawItem.(string)
		if ok {
			items = append(items, item)
		}
	}
	return items
}

func volumeStrings(value any) []string {
	var out []string
	for _, item := range stringSlice(value) {
		out = append(out, item)
	}
	rawItems, ok := value.([]any)
	if !ok {
		return out
	}
	for _, rawItem := range rawItems {
		volumeMap, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		if typ, _ := volumeMap["type"].(string); typ == "bind" {
			source, _ := volumeMap["source"].(string)
			target, _ := volumeMap["target"].(string)
			if source != "" || target != "" {
				out = append(out, source+":"+target)
			}
		}
	}
	return out
}

func isLikelyHostPathVolume(volume string) bool {
	source := volume
	if idx := strings.Index(volume, ":"); idx >= 0 {
		source = volume[:idx]
	}
	if source == "" {
		return false
	}
	if filepath.IsAbs(source) {
		return true
	}
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return true
	}
	return false
}

func usesLatestTag(image string) bool {
	if image == "" {
		return false
	}
	parts := strings.Split(image, "/")
	last := parts[len(parts)-1]
	if !strings.Contains(last, ":") {
		return true
	}
	return strings.HasSuffix(last, ":latest")
}

func Catalog() []Finding {
	return []Finding{
		{Severity: SeverityHigh, Message: "privileged mode is not allowed"},
		{Severity: SeverityHigh, Message: "host network mode is not allowed"},
		{Severity: SeverityHigh, Message: "mounting the Docker socket is not allowed"},
		{Severity: SeverityMedium, Message: "image should not use the latest tag"},
		{Severity: SeverityMedium, Message: "service should define a healthcheck"},
		{Severity: SeverityMedium, Message: "service should set read_only: true"},
		{Severity: SeverityMedium, Message: "service should set security_opt: no-new-privileges:true"},
		{Severity: SeverityMedium, Message: "service should drop all Linux capabilities with cap_drop: [ALL]"},
		{Severity: SeverityMedium, Message: "host path mounts are discouraged; prefer named volumes"},
	}
}
