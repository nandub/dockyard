package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newUninstallCommand(global *globalOptions) *cobra.Command {
	var removeVolumes bool
	var purge bool
	var dryRun bool
	var force bool

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
			dependents, err := activeDependentsForRelease(home, releaseName)
			if err != nil {
				return err
			}
			if len(dependents) > 0 && !force {
				return dependencyUninstallBlockedError(releaseName, dependents)
			}
			if dryRun {
				if len(dependents) > 0 && force {
					fmt.Printf("Warning: %s is still required by active release(s): %s\n", releaseName, strings.Join(dependents, ", "))
				}
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
	cmd.Flags().BoolVar(&force, "force", false, "uninstall even when active releases still depend on this release")
	return cmd
}

type activeDependent struct {
	release    string
	dependency string
	alias      string
}

func activeDependentsForRelease(home string, releaseName string) ([]string, error) {
	dependents, err := collectActiveDependentsForRelease(home, releaseName)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(dependents))
	for _, dependent := range dependents {
		label := dependent.release
		if dependent.dependency != "" {
			label += " (" + dependent.dependency
			if dependent.alias != "" {
				label += " as " + dependent.alias
			}
			label += ")"
		}
		out = append(out, label)
	}
	sort.Strings(out)
	return out, nil
}

func collectActiveDependentsForRelease(home string, releaseName string) ([]activeDependent, error) {
	releasesDir := filepath.Join(home, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read releases directory: %w", err)
	}

	var dependents []activeDependent
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == releaseName {
			continue
		}
		release, _, err := readCurrentRelease(home, entry.Name())
		if err != nil {
			continue
		}
		if release.Status == "uninstalled" {
			continue
		}
		for _, dep := range release.Dependencies {
			if dep.Release == releaseName {
				dependents = append(dependents, activeDependent{
					release:    release.Name,
					dependency: dep.Name,
					alias:      dep.Alias,
				})
			}
		}
	}
	return dependents, nil
}

func dependencyUninstallBlockedError(releaseName string, dependents []string) error {
	if len(dependents) == 0 {
		return nil
	}
	return fmt.Errorf("release %q is still required by active release(s): %s; uninstall dependent release(s) first or re-run with --force", releaseName, strings.Join(dependents, ", "))
}
