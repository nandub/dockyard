package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newListCommand(global *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dockyard releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			releasesDir := filepath.Join(home, "releases")
			entries, err := os.ReadDir(releasesDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("no releases")
					return nil
				}
				return err
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				release, _, err := readCurrentRelease(home, entry.Name())
				if err != nil {
					fmt.Printf("%s\tunknown\n", entry.Name())
					continue
				}
				fmt.Printf("%s\t%s\trevision=%d\tpackage=%s@%s\n", release.Name, release.Status, release.Revision, release.PackageName, release.PackageVersion)
			}
			return nil
		},
	}
	return cmd
}
