package cli

import (
	"fmt"
	"path/filepath"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newInspectCommand(global *globalOptions) *cobra.Command {
	var revision int
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "inspect RELEASE",
		Short: "Inspect release metadata and file locations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			if revision == 0 {
				revision, err = state.ReadCurrentRevision(home, releaseName)
				if err != nil {
					return err
				}
			}
			revisionDir := state.RevisionDir(home, releaseName, revision)
			release, err := state.ReadRelease(revisionDir)
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(release)
			}
			fmt.Printf("Release: %s\nRevision: %d\nStatus: %s\nRevision Dir: %s\nCompose File: %s\nValues File: %s\nManifest File: %s\n", release.Name, release.Revision, release.Status, revisionDir, filepath.Join(revisionDir, "compose.rendered.yaml"), filepath.Join(revisionDir, "values.yaml"), filepath.Join(revisionDir, "Dockyard.yaml"))
			return nil
		},
	}
	cmd.Flags().IntVar(&revision, "revision", 0, "revision to inspect; defaults to current")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}
