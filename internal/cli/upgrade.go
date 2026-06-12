package cli

import (
	"fmt"
	"path/filepath"

	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newUpgradeCommand(global *globalOptions) *cobra.Command {
	var opts packageBuildOptions
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "upgrade RELEASE PACKAGE_SOURCE",
		Short: "Render and deploy a new revision for an existing release",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			if err := state.ValidateReleaseName(releaseName); err != nil {
				return err
			}
			src, err := preparePackageSource(args[1], true)
			if err != nil {
				return err
			}
			defer src.cleanup()

			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			if _, err := state.ReadCurrentRevision(home, releaseName); err != nil {
				return fmt.Errorf("release %q is not installed", releaseName)
			}
			envEntries, err := loadCommandEnv(opts.envFile)
			if err != nil {
				return err
			}
			manifest, vals, rendered, _, err := buildPackage(src.Dir, releaseName, opts)
			if err != nil {
				return err
			}
			revision, err := state.NextRevision(home, releaseName)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Printf("Would upgrade release %q to revision %d from %s source\n", releaseName, revision, src.Source.Type)
				return nil
			}
			release, composePath, err := writeRevision(home, releaseName, revision, manifest, vals, rendered, src.Dir, src.Source, "pending", opts.envFile, releaseRelationshipMetadata{})
			if err != nil {
				return err
			}
			if !opts.skipComposeConfig {
				ctx, cancel := context10m()
				defer cancel()
				if err := (runner.DockerComposeRunner{WorkDir: filepath.Dir(composePath), Project: releaseName, Env: envEntries}).ValidateConfig(ctx, composePath); err != nil {
					release.Status = "failed"
					_ = state.WriteRelease(filepath.Dir(composePath), *release)
					return err
				}
			}
			ctx, cancel := context10m()
			defer cancel()
			if err := dockerRunnerWithEnv(releaseName, filepath.Dir(composePath), envEntries).Up(ctx, composePath); err != nil {
				release.Status = "failed"
				_ = state.WriteRelease(filepath.Dir(composePath), *release)
				return err
			}
			release.Status = "deployed"
			if err := state.WriteRelease(filepath.Dir(composePath), *release); err != nil {
				return err
			}
			if err := state.SetCurrentRevision(home, releaseName, revision); err != nil {
				return err
			}
			fmt.Printf("upgraded %s to revision %d\n", release.Name, release.Revision)
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.envFile, "env-file", "", "dotenv file to pass to docker compose without mutating the shell environment")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without deploying")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.skipComposeConfig, "skip-compose-config", false, "skip docker compose config validation")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")
	return cmd
}
