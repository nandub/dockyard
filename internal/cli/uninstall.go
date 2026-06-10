package cli

import (
	"fmt"
	"os"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newUninstallCommand(global *globalOptions) *cobra.Command {
	var removeVolumes bool
	var purge bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "uninstall RELEASE",
		Short: "Remove a Dockyard release using Docker Compose",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			release, composePath, err := readCurrentRelease(home, releaseName)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Printf("Would run: docker compose -p %s -f %s down", releaseName, composePath)
				if removeVolumes {
					fmt.Print(" --volumes")
				}
				fmt.Println()
				return nil
			}
			ctx, cancel := context10m()
			defer cancel()
			if err := dockerRunner(releaseName, state.RevisionDir(home, releaseName, release.Revision)).Down(ctx, composePath, removeVolumes); err != nil {
				return err
			}
			release.Status = "uninstalled"
			if err := state.WriteRelease(state.RevisionDir(home, releaseName, release.Revision), *release); err != nil {
				return err
			}
			if purge {
				if err := os.RemoveAll(state.ReleaseDir(home, releaseName)); err != nil {
					return err
				}
			}
			fmt.Printf("uninstalled %s\n", releaseName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&removeVolumes, "volumes", false, "also remove named volumes")
	cmd.Flags().BoolVar(&purge, "purge", false, "remove Dockyard release metadata after uninstall")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without changing anything")
	return cmd
}
