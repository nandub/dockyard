package cli

import (
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newDiffCommand(global *globalOptions) *cobra.Command {
	var opts packageBuildOptions

	cmd := &cobra.Command{
		Use:   "diff RELEASE PACKAGE_SOURCE",
		Short: "Show a simple line diff between the current rendered Compose file and a new render",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			src, err := preparePackageSource(args[1], true)
			if err != nil {
				return err
			}
			defer src.cleanup()

			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			_, currentCompose, err := readCurrentRelease(home, releaseName)
			if err != nil {
				return err
			}
			_, _, newRendered, _, err := buildPackage(src.Dir, releaseName, opts)
			if err != nil {
				return err
			}
			oldBytes, err := osReadFile(currentCompose)
			if err != nil {
				return err
			}
			printSimpleDiff(string(oldBytes), string(newRendered))
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
