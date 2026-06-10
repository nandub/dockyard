package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newPruneCommand(global *globalOptions) *cobra.Command {
	var releaseName string
	var keep int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove old Dockyard release revisions while keeping current revision history",
		RunE: func(cmd *cobra.Command, args []string) error {
			if keep < 1 {
				return fmt.Errorf("--keep must be at least 1")
			}
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			if releaseName != "" {
				if err := state.ValidateReleaseName(releaseName); err != nil {
					return err
				}
				return pruneRelease(home, releaseName, keep, dryRun)
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
				if err := pruneRelease(home, entry.Name(), keep, dryRun); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&releaseName, "release", "", "release to prune; defaults to all releases")
	cmd.Flags().IntVar(&keep, "keep", 3, "number of newest revisions to keep per release")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show revisions that would be removed")
	return cmd
}

func pruneRelease(home string, releaseName string, keep int, dryRun bool) error {
	revisionsDir := filepath.Join(state.ReleaseDir(home, releaseName), "revisions")
	entries, err := os.ReadDir(revisionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	currentRevision, _ := state.ReadCurrentRevision(home, releaseName)
	revisions := make([]int, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		revision, err := strconv.Atoi(entry.Name())
		if err == nil {
			revisions = append(revisions, revision)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(revisions)))

	keepSet := map[int]bool{}
	for idx, revision := range revisions {
		if idx < keep {
			keepSet[revision] = true
		}
	}
	if currentRevision > 0 {
		keepSet[currentRevision] = true
	}

	removed := 0
	for _, revision := range revisions {
		if keepSet[revision] {
			continue
		}
		revisionDir := state.RevisionDir(home, releaseName, revision)
		if dryRun {
			fmt.Printf("Would remove %s revision %d: %s\n", releaseName, revision, revisionDir)
			removed++
			continue
		}
		if err := os.RemoveAll(revisionDir); err != nil {
			return err
		}
		fmt.Printf("removed %s revision %d\n", releaseName, revision)
		removed++
	}
	if removed == 0 {
		fmt.Printf("%s: nothing to prune\n", releaseName)
	}
	return nil
}
