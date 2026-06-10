package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/oci"
	"github.com/spf13/cobra"
)

func newPushCommand() *cobra.Command {
	var skipVerify bool

	cmd := &cobra.Command{
		Use:   "push PACKAGE_ARCHIVE OCI_REFERENCE",
		Short: "Push a Dockyard package archive to an OCI registry using the oras CLI",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			archivePath := args[0]
			ref := args[1]
			if !skipVerify {
				if err := archive.VerifyArchive(archivePath, nil); err != nil {
					return err
				}
			}
			ctx, cancel := context10m()
			defer cancel()
			if err := oci.Push(ctx, archivePath, ref); err != nil {
				return err
			}
			fmt.Printf("pushed %s to %s\n", archivePath, ref)
			return nil
		},
	}
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "skip local package verification before push")
	return cmd
}
