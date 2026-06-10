package cli

import (
	"fmt"
	"path/filepath"

	"github.com/nandub/dockyard/internal/dockpkg"
	docklock "github.com/nandub/dockyard/internal/lock"
	"github.com/nandub/dockyard/internal/render"
	"github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

func newLockCommand() *cobra.Command {
	var valuesFile string
	var overlay string
	var output string

	cmd := &cobra.Command{
		Use:   "lock PACKAGE_DIR",
		Short: "Create or update dockyard.lock for a package render",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageDir := args[0]
			manifest, err := dockpkg.LoadManifest(packageDir)
			if err != nil {
				return err
			}
			vals, err := values.LoadValues(packageDir, valuesFile)
			if err != nil {
				return err
			}
			if err := values.ValidateAgainstSchema(packageDir, vals); err != nil {
				return err
			}
			rendered, err := render.RenderCompose(packageDir, manifest, vals, overlay)
			if err != nil {
				return err
			}
			lockfile, err := docklock.New(packageDir, manifest, vals, rendered, overlay)
			if err != nil {
				return err
			}
			if output == "" {
				output = filepath.Join(packageDir, docklock.FileName)
			}
			if err := docklock.Write(output, lockfile); err != nil {
				return err
			}
			fmt.Printf("lockfile written: %s\n", output)
			for _, image := range lockfile.Images {
				if image.Digest == "" {
					fmt.Printf("WARN: service %q image %q is not digest-pinned\n", image.Service, image.Image)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&overlay, "overlay", "", "compose overlay name")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output lockfile path; defaults to PACKAGE_DIR/dockyard.lock")
	return cmd
}
