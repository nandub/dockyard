package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

type listOptions struct {
	all    bool
	status string
}

func newListCommand(global *globalOptions) *cobra.Command {
	opts := listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dockyard releases",
		Long: `List Dockyard releases.

By default, uninstalled releases are hidden so the output focuses on active or
operator-attention releases. Use --all to include historical uninstalled
releases, or --status to show a specific status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			rows, err := collectReleaseListRows(home, opts)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				if opts.status != "" {
					fmt.Printf("no releases with status %q\n", opts.status)
					return nil
				}
				if opts.all {
					fmt.Println("no releases")
					return nil
				}
				fmt.Println("no active releases (use --all to include uninstalled releases)")
				return nil
			}
			printReleaseListRows(rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "include uninstalled releases")
	cmd.Flags().StringVar(&opts.status, "status", "", "show only releases with the selected status")
	return cmd
}

type releaseListRow struct {
	Name           string
	Status         string
	Revision       int
	PackageName    string
	PackageVersion string
	Relation       string
}

func collectReleaseListRows(home string, opts listOptions) ([]releaseListRow, error) {
	releasesDir := filepath.Join(home, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	rows := make([]releaseListRow, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		release, _, err := readCurrentRelease(home, entry.Name())
		if err != nil {
			if opts.status == "" && opts.all {
				rows = append(rows, releaseListRow{Name: entry.Name(), Status: "unknown"})
			}
			continue
		}
		if !shouldIncludeReleaseInList(release.Status, opts) {
			continue
		}
		rows = append(rows, releaseListRow{
			Name:           release.Name,
			Status:         release.Status,
			Revision:       release.Revision,
			PackageName:    release.PackageName,
			PackageVersion: release.PackageVersion,
			Relation:       releaseRelationSummary(release),
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})
	return rows, nil
}

func shouldIncludeReleaseInList(status string, opts listOptions) bool {
	if opts.status != "" {
		return status == opts.status
	}
	if opts.all {
		return true
	}
	return status != "uninstalled"
}

func releaseRelationSummary(release *state.Release) string {
	if release.Parent != nil {
		return "child-of=" + release.Parent.Name
	}
	if len(release.Dependencies) > 0 {
		return fmt.Sprintf("deps=%d", len(release.Dependencies))
	}
	return "-"
}

func printReleaseListRows(rows []releaseListRow) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "NAME\tSTATUS\tREVISION\tPACKAGE\tRELATION")
	for _, row := range rows {
		_, _ = fmt.Fprintf(writer, "%s\t%s\t%d\t%s@%s\t%s\n", row.Name, row.Status, row.Revision, row.PackageName, row.PackageVersion, row.Relation)
	}
	_ = writer.Flush()
}
