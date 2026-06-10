package cli

import (
	"encoding/json"
	"fmt"

	"github.com/nandub/dockyard/internal/policy"
	"github.com/spf13/cobra"
)

func newPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Inspect and run Dockyard security policy checks",
	}

	cmd.AddCommand(newPolicyListCommand())
	cmd.AddCommand(newPolicyCheckCommand())

	return cmd
}

func newPolicyListCommand() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List built-in Dockyard policy checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog := policy.Catalog()
			if jsonOutput {
				data, err := json.MarshalIndent(catalog, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			for _, finding := range catalog {
				fmt.Printf("%s: %s\n", finding.Severity, finding.Message)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print policy catalog as JSON")
	return cmd
}

func newPolicyCheckCommand() *cobra.Command {
	var valuesFile string
	var overlay string
	var allowRisk bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "check PACKAGE_SOURCE",
		Short: "Render a package source and run Dockyard security policy checks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := preparePackageSource(args[0], true)
			if err != nil {
				return err
			}
			defer src.cleanup()

			_, _, _, findings, err := buildPackage(src.Dir, "", packageBuildOptions{
				valuesFile: valuesFile,
				overlay:    overlay,
				allowRisk:  true,
			})
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
				fmt.Println("OK: no policy findings")
			} else {
				for _, finding := range findings {
					if finding.Service != "" {
						fmt.Printf("%s: service %q: %s\n", finding.Severity, finding.Service, finding.Message)
						continue
					}
					fmt.Printf("%s: %s\n", finding.Severity, finding.Message)
				}
			}

			if policy.HasHighFindings(findings) && !allowRisk {
				return fmt.Errorf("policy check found HIGH severity finding(s); re-run with --allow-risk to allow the result")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&overlay, "overlay", "", "compose overlay name")
	cmd.Flags().BoolVar(&allowRisk, "allow-risk", false, "return success even when HIGH policy findings are present")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print findings as JSON")

	return cmd
}
