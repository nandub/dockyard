package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/spf13/cobra"
)

func newVerifyCommand() *cobra.Command {
	var opts packageBuildOptions

	cmd := &cobra.Command{
		Use:   "verify PACKAGE_ARCHIVE",
		Short: "Verify a local Dockyard package archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := archive.VerifyArchive(args[0], func(tempDir string) error {
				_, _, _, _, err := buildPackage(tempDir, "verify", opts)
				return err
			})
			if err != nil {
				return err
			}
			fmt.Println("OK: package verified")
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")
	return cmd
}
