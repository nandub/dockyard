package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/oci"
	"github.com/spf13/cobra"
)

func newPullCommand() *cobra.Command {
	var output string
	var skipVerify bool

	cmd := &cobra.Command{
		Use:   "pull OCI_REFERENCE",
		Short: "Pull a Dockyard package archive from an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tempDir, err := os.MkdirTemp("", "dockyard-pull-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempDir)

			ctx, cancel := context10m()
			defer cancel()
			pulledArchive, err := oci.Pull(ctx, args[0], tempDir)
			if err != nil {
				return err
			}
			if !skipVerify {
				if err := archive.VerifyArchive(pulledArchive, nil); err != nil {
					return err
				}
			}
			if output == "" {
				output, err = defaultPulledArchiveName(pulledArchive)
				if err != nil {
					return err
				}
			}
			cleanOutput := filepath.Clean(output)
			if err := copyPulledArchive(cleanOutput, pulledArchive); err != nil {
				return err
			}
			fmt.Printf("pulled %s\n", cleanOutput)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "skip package verification after pull")
	return cmd
}

func defaultPulledArchiveName(archivePath string) (string, error) {
	tempDir, err := os.MkdirTemp("", "dockyard-name-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)
	if err := archive.ExtractArchive(archivePath, tempDir); err != nil {
		return "", err
	}
	manifest, err := dockpkg.LoadManifest(tempDir)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s.dockyard.tgz", manifest.Name, manifest.Version), nil
}

func copyPulledArchive(dst string, src string) error {
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("refusing to overwrite existing file %q", dst)
	}
	input, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0o600)
}
