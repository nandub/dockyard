package cli

import (
	"fmt"
	"strconv"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newRollbackCommand(global *globalOptions) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "rollback RELEASE REVISION",
		Short: "Redeploy a previous release revision",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			revision, err := strconv.Atoi(args[1])
			if err != nil || revision < 1 {
				return fmt.Errorf("revision must be a positive integer")
			}
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			revisionDir := state.RevisionDir(home, releaseName, revision)
			release, err := state.ReadRelease(revisionDir)
			if err != nil {
				return err
			}
			composePath := revisionDir + "/compose.rendered.yaml"
			if dryRun {
				fmt.Printf("Would rollback %s to revision %d using %s\n", releaseName, revision, composePath)
				return nil
			}
			ctx, cancel := context10m()
			defer cancel()
			if err := dockerRunner(releaseName, revisionDir).Up(ctx, composePath); err != nil {
				return err
			}
			release.Status = "deployed"
			if err := state.WriteRelease(revisionDir, *release); err != nil {
				return err
			}
			if err := state.SetCurrentRevision(home, releaseName, revision); err != nil {
				return err
			}
			fmt.Printf("rolled back %s to revision %d\n", releaseName, revision)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without deploying")
	return cmd
}
