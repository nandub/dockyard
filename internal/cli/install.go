package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newInstallCommand(global *globalOptions) *cobra.Command {
	var opts packageBuildOptions
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "install RELEASE PACKAGE_SOURCE",
		Short: "Render, validate, record, and deploy a Dockyard package directory or archive",
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
			revision := 1
			if currentRevision, err := state.ReadCurrentRevision(home, releaseName); err == nil {
				currentDir := state.RevisionDir(home, releaseName, currentRevision)
				currentRelease, readErr := state.ReadRelease(currentDir)
				if readErr != nil {
					return readErr
				}
				if currentRelease.Status != "uninstalled" {
					return fmt.Errorf("release %q already exists and is %s; use upgrade instead", releaseName, currentRelease.Status)
				}
				nextRevision, err := state.NextRevision(home, releaseName)
				if err != nil {
					return err
				}
				revision = nextRevision
			}
			manifest, vals, rendered, _, err := buildPackage(src.Dir, releaseName, opts)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Printf("Would install release %q into %s from %s source\n", releaseName, home, src.Source.Type)
				return nil
			}
			release, composePath, err := writeRevision(home, releaseName, revision, manifest, vals, rendered, src.Dir, src.Source, "pending")
			if err != nil {
				return err
			}
			if !opts.skipComposeConfig {
				ctx, cancel := context10m()
				defer cancel()
				if err := (runner.DockerComposeRunner{WorkDir: filepath.Dir(composePath), Project: releaseName}).ValidateConfig(ctx, composePath); err != nil {
					release.Status = "failed"
					_ = state.WriteRelease(filepath.Dir(composePath), *release)
					return err
				}
			}
			ctx, cancel := context10m()
			defer cancel()
			if err := dockerRunner(releaseName, filepath.Dir(composePath)).Up(ctx, composePath); err != nil {
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
			if err := os.MkdirAll(home, 0o700); err != nil {
				return err
			}
			fmt.Printf("installed %s revision %d\n", release.Name, release.Revision)
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without deploying")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.skipComposeConfig, "skip-compose-config", false, "skip docker compose config validation")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")
	return cmd
}
