package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/policy"
	"github.com/nandub/dockyard/internal/render"
	"github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

func newLintCommand() *cobra.Command {
	var valuesFile string
	var overlay string

	cmd := &cobra.Command{
		Use:   "lint PACKAGE_DIR",
		Short: "Validate a Dockyard package and check Compose security policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageDir := args[0]
			manifest, err := dockpkg.LoadManifest(packageDir)
			if err != nil {
				return err
			}
			vals, err := values.LoadValues(packageDir, valuesFile)
			if err != nil {
				return err
			}
			if err := values.ValidateAgainstSchema(packageDir, vals); err != nil {
				return err
			}
			rendered, err := render.RenderCompose(packageDir, manifest, vals, overlay)
			if err != nil {
				return err
			}
			findings, err := policy.LintCompose(rendered, manifest.Security)
			if err != nil {
				return err
			}
			if len(findings) == 0 {
				fmt.Println("OK: package passed validation and policy checks")
				return nil
			}
			for _, finding := range findings {
				if finding.Service != "" {
					fmt.Printf("%s: service %q: %s\n", finding.Severity, finding.Service, finding.Message)
					continue
				}
				fmt.Printf("%s: %s\n", finding.Severity, finding.Message)
			}
			return fmt.Errorf("lint failed with %d finding(s)", len(findings))
		},
	}
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&overlay, "overlay", "", "compose overlay name")
	return cmd
}
