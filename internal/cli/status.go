package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newStatusCommand(global *globalOptions) *cobra.Command {
	var jsonOut bool
	var composePS bool
	var composePSAll bool

	cmd := &cobra.Command{
		Use:   "status RELEASE",
		Short: "Show release status",
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
			if jsonOut {
				return printJSON(release)
			}
			fmt.Printf("Name: %s\nStatus: %s\nRevision: %d\nPackage: %s@%s\nApp: %s\nCompose Project: %s\n", release.Name, release.Status, release.Revision, release.PackageName, release.PackageVersion, release.AppVersion, release.ComposeProject)
			if release.Parent != nil {
				if release.Parent.Alias != "" {
					fmt.Printf("Parent: %s (dependency %s as %s)\n", release.Parent.Name, release.Parent.DependencyName, release.Parent.Alias)
				} else {
					fmt.Printf("Parent: %s (dependency %s)\n", release.Parent.Name, release.Parent.DependencyName)
				}
			}
			if len(release.Dependencies) > 0 {
				fmt.Println("Dependencies:")
				for _, dep := range release.Dependencies {
					alias := ""
					if dep.Alias != "" {
						alias = " as " + dep.Alias
					}
					pkg := dep.PackageName
					if dep.PackageVersion != "" {
						pkg = pkg + "@" + dep.PackageVersion
					}
					fmt.Printf("  - %s%s -> %s (%s, %s)\n", dep.Name, alias, dep.Release, pkg, dep.Status)
				}
			}
			if release.DockyardVersion != "" {
				fmt.Printf("Dockyard Version: %s\n", release.DockyardVersion)
			}
			if release.EnvFile != "" {
				fmt.Printf("Env File: %s\n", release.EnvFile)
			}
			if composePS {
				ctx, cancel := context10m()
				defer cancel()
				return dockerRunner(releaseName, state.RevisionDir(home, releaseName, release.Revision)).PS(ctx, composePath, composePSAll)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().BoolVar(&composePS, "compose-ps", false, "also run docker compose ps")
	cmd.Flags().BoolVar(&composePSAll, "all", false, "include stopped containers when used with --compose-ps")
	return cmd
}
