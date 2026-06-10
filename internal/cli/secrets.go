package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

type secretFinding struct {
	Path    string `json:"path"`
	Reason  string `json:"reason"`
	Preview string `json:"preview,omitempty"`
}

func newSecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Scan Dockyard values for secret-like defaults",
	}
	cmd.AddCommand(newSecretsScanCommand())
	return cmd
}

func newSecretsScanCommand() *cobra.Command {
	var valuesFile string
	var jsonOutput bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "scan PACKAGE_DIR",
		Short: "Scan values for secret-like keys that contain non-empty values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := values.DefaultValuesFile
			if valuesFile != "" {
				target = valuesFile
			} else {
				target = filepath.Join(args[0], values.DefaultValuesFile)
			}

			vals, err := values.LoadYAMLMap(target)
			if err != nil {
				return err
			}
			findings := scanSecretValues("", vals)

			if jsonOutput {
				data, err := json.MarshalIndent(findings, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else if len(findings) == 0 {
				fmt.Println("OK: no secret-like populated values found")
			} else {
				for _, finding := range findings {
					fmt.Printf("MEDIUM: %s: %s", finding.Path, finding.Reason)
					if finding.Preview != "" {
						fmt.Printf(" (%s)", finding.Preview)
					}
					fmt.Println()
				}
			}

			if strict && len(findings) > 0 {
				return fmt.Errorf("secret scan found %d populated secret-like value(s)", len(findings))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "scan this values file instead of package defaults")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print findings as JSON")
	cmd.Flags().BoolVar(&strict, "strict", false, "fail when findings are present")

	return cmd
}

func scanSecretValues(prefix string, vals map[string]any) []secretFinding {
	var findings []secretFinding
	for key, value := range vals {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if nested, ok := value.(map[string]any); ok {
			findings = append(findings, scanSecretValues(path, nested)...)
			continue
		}
		if !isSensitivePath(path) {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			continue
		}
		if isPlaceholderValue(text) {
			continue
		}
		findings = append(findings, secretFinding{
			Path:    path,
			Reason:  "secret-like key has a populated value; keep production secrets outside packages",
			Preview: maskPreview(text),
		})
	}
	return findings
}

func isSensitivePath(path string) bool {
	lower := strings.ToLower(path)
	needles := []string{"password", "passwd", "secret", "token", "api_key", "apikey", "private_key", "credential"}
	for _, needle := range needles {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func isPlaceholderValue(value string) bool {
	lower := strings.ToLower(value)
	if lower == "changeme" || lower == "change-me" || strings.Contains(lower, "replace") || strings.Contains(lower, "example") {
		return true
	}
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		return true
	}
	return false
}

func maskPreview(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", 6) + value[len(value)-2:]
}
