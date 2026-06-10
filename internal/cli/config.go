package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/runner"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	var opts packageBuildOptions
	var outputFile string

	cmd := &cobra.Command{
		Use:   "config PACKAGE_SOURCE",
		Short: "Render a Dockyard package and validate it with docker compose config",
		Long: `Render a Dockyard package directory, archive, or OCI reference and run
docker compose config against the rendered Compose YAML.

This command validates Docker Compose compatibility without installing a release.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := preparePackageSource(args[0], true)
			if err != nil {
				return err
			}
			defer src.cleanup()

			envEntries, err := loadCommandEnv(opts.envFile)
			if err != nil {
				return err
			}

			_, _, rendered, _, err := buildPackage(src.Dir, "dockyard-config", opts)
			if err != nil {
				return err
			}

			tempDir, err := os.MkdirTemp("", "dockyard-config-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			defer os.RemoveAll(tempDir)

			composePath := filepath.Join(tempDir, "compose.rendered.yaml")
			if err := os.WriteFile(composePath, rendered, 0o600); err != nil {
				return fmt.Errorf("write rendered compose: %w", err)
			}

			ctx, cancel := context10m()
			defer cancel()

			if err := (runner.DockerComposeRunner{
				WorkDir: tempDir,
				Project: "dockyard-config",
				Env:     envEntries,
			}).Config(ctx, composePath); err != nil {
				return err
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, rendered, 0o600); err != nil {
					return fmt.Errorf("write output file: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.envFile, "env-file", "", "dotenv file to pass to docker compose without mutating the shell environment")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "optional path to write rendered Compose YAML after validation")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")

	return cmd
}
