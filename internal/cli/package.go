package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/spf13/cobra"
)

func newPackageCommand() *cobra.Command {
	var output string
	var locked bool
	var opts packageBuildOptions

	cmd := &cobra.Command{
		Use:   "package PACKAGE_DIR",
		Short: "Create a local Dockyard package archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if locked {
				opts.requireLock = true
				if _, _, _, _, err := buildPackage(args[0], "package", opts); err != nil {
					return err
				}
			}
			path, err := archive.PackageDir(args[0], output, locked)
			if err != nil {
				return err
			}
			fmt.Printf("package created: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path")
	cmd.Flags().BoolVar(&locked, "locked", false, "require dockyard.lock and include lock provenance")
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file used to verify dockyard.lock when --locked is set")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name used to verify dockyard.lock when --locked is set")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings during --locked verification")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip policy checks during --locked verification")
	return cmd
}
