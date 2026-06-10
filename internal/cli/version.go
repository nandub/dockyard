package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print Dockyard version and build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			if jsonOut {
				return printJSON(info)
			}
			fmt.Println(info.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "print version metadata as JSON")
	return cmd
}
