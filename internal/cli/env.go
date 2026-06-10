package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/envfile"
	"github.com/spf13/cobra"
)

func newEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Generate and validate environment files for Dockyard values",
	}
	cmd.AddCommand(newEnvTemplateCommand())
	cmd.AddCommand(newEnvCheckCommand())
	return cmd
}

func newEnvTemplateCommand() *cobra.Command {
	var valuesFile string
	var outputFile string
	var prefix string
	var sensitiveOnly bool
	var force bool

	cmd := &cobra.Command{
		Use:   "template PACKAGE_DIR",
		Short: "Generate a .env template from Dockyard values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := envfile.GenerateTemplate(args[0], envfile.TemplateOptions{
				ValuesFile:    valuesFile,
				Prefix:        prefix,
				SensitiveOnly: sensitiveOnly,
			})
			if err != nil {
				return err
			}
			if outputFile == "" {
				fmt.Print(string(out))
				return nil
			}
			cleanOutput := filepath.Clean(outputFile)
			if !force {
				if _, err := os.Stat(cleanOutput); err == nil {
					return fmt.Errorf("refusing to overwrite existing file %q; use --force", cleanOutput)
				}
			}
			if err := os.MkdirAll(filepath.Dir(cleanOutput), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(cleanOutput, out, 0o600); err != nil {
				return err
			}
			fmt.Printf("env template written: %s\n", cleanOutput)
			return nil
		},
	}

	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output .env file; defaults to stdout")
	cmd.Flags().StringVar(&prefix, "prefix", "", "prefix generated environment variable names")
	cmd.Flags().BoolVar(&sensitiveOnly, "sensitive-only", false, "generate variables only for sensitive-looking values")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite output file when it already exists")

	return cmd
}

func newEnvCheckCommand() *cobra.Command {
	var jsonOutput bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "check ENV_FILE",
		Short: "Check a .env file for duplicate keys, invalid syntax, and populated secret-like values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			findings, err := envfile.CheckFile(args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				data, err := json.MarshalIndent(findings, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else if len(findings) == 0 {
				fmt.Println("OK: env file passed checks")
			} else {
				for _, finding := range findings {
					location := fmt.Sprintf("line %d", finding.Line)
					if finding.Key != "" {
						location += " " + finding.Key
					}
					fmt.Printf("MEDIUM: %s: %s\n", location, finding.Message)
				}
			}
			if strict && len(findings) > 0 {
				return fmt.Errorf("env check found %d finding(s)", len(findings))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print findings as JSON")
	cmd.Flags().BoolVar(&strict, "strict", false, "fail when findings are present")

	return cmd
}
