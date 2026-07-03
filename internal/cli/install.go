package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/state"
	"github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

func newInstallCommand(global *globalOptions) *cobra.Command {
	var opts packageBuildOptions
	var dryRun bool
	var jsonOut bool
	var withDependencies bool

	cmd := &cobra.Command{
		Use:   "install RELEASE [PACKAGE_SOURCE]",
		Short: "Install a Dockyard package from a path, archive, OCI reference, or catalog",
		Long: `Render, validate, record, and deploy a Dockyard package from a local directory,
archive, OCI reference, catalog reference, or configured catalog shorthand.

By default, install deploys only the root package. Use --with-dependencies to
install declared package dependencies before the root package. Catalog shorthand is supported:
dockyard install redis resolves to the configured catalog package named redis.
Dependency
installation is conservative: existing deployed dependency releases are reused,
uninstalled dependency releases are reinstalled, and dependencies are not
automatically removed if a later step fails or the root release is uninstalled.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			releaseName := args[0]
			source := args[0]
			if len(args) == 2 {
				source = args[1]
			}

			if len(args) == 1 && !sourcePathExists(source) {
				if _, ok, err := resolveCatalogPackageSource(source); err != nil {
					return err
				} else if !ok {
					return fmt.Errorf("package %q was not found in the configured catalog", source)
				}
			}

			if err := state.ValidateReleaseName(releaseName); err != nil {
				return err
			}
			if jsonOut && !dryRun {
				return errors.New("--json can only be used with --dry-run")
			}
			if dryRun {
				report, err := buildInstallDryRunPlan(global, releaseName, source, jsonOut)
				if err != nil {
					return err
				}
				if jsonOut {
					return printJSON(report)
				}
				printInstallPlan(report)
				return nil
			}

			if withDependencies {
				return installWithDependencies(global, releaseName, source, opts)
			}

			release, err := installSingleRelease(global, releaseName, source, opts, nil, releaseRelationshipMetadata{})
			if err != nil {
				return err
			}
			fmt.Printf("installed %s revision %d\n", release.Name, release.Revision)
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.envFile, "env-file", "", "dotenv file to pass to docker compose without mutating the shell environment")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show the dependency-aware install plan without deploying")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON when used with --dry-run")
	cmd.Flags().BoolVar(&withDependencies, "with-dependencies", false, "install declared package dependencies before the root package")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.skipComposeConfig, "skip-compose-config", false, "skip docker compose config validation")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")
	return cmd
}

func installWithDependencies(global *globalOptions, releaseName string, source string, opts packageBuildOptions) error {
	report, err := buildInstallPlan(global, releaseName, source)
	if err != nil {
		return err
	}

	for _, step := range report.Steps {
		if step.Type == planStepRoot && (step.Action == "exists" || step.Action == "blocked") {
			return fmt.Errorf("release %q already exists and is %s; use upgrade instead", step.PlannedRelease, step.ExistingReleaseStatus)
		}
	}

	dependencyRefs := make([]state.ReleaseDependency, 0, len(report.Steps))
	for _, step := range report.Steps {
		if step.Type != planStepDependency {
			continue
		}
		var depRelease *state.Release
		switch step.Action {
		case "exists":
			existing, _, err := readCurrentRelease(report.Home, step.PlannedRelease)
			if err != nil {
				return fmt.Errorf("read dependency release %q: %w", step.PlannedRelease, err)
			}
			depRelease = existing
			fmt.Printf("dependency %s already exists as release %s; leaving unchanged\n", step.Name, step.PlannedRelease)
		case "install", "reinstall":
			depOpts := dependencyInstallOptions(opts)
			release, err := installSingleRelease(global, step.PlannedRelease, step.Source, depOpts, step.Values, releaseRelationshipMetadata{
				parent: &state.ReleaseParent{
					Name:           releaseName,
					DependencyName: step.Name,
					Alias:          step.Alias,
				},
			})
			if err != nil {
				return fmt.Errorf("install dependency %q as release %q: %w", step.Name, step.PlannedRelease, err)
			}
			depRelease = release
			fmt.Printf("installed dependency %s as %s revision %d\n", step.Name, release.Name, release.Revision)
		case "blocked":
			return fmt.Errorf("dependency release %q already exists and is %s; resolve it before using --with-dependencies", step.PlannedRelease, step.ExistingReleaseStatus)
		default:
			return fmt.Errorf("unsupported dependency action %q for release %q", step.Action, step.PlannedRelease)
		}
		dependencyRefs = append(dependencyRefs, state.ReleaseDependency{
			Name:           step.Name,
			Alias:          step.Alias,
			Release:        step.PlannedRelease,
			PackageName:    depRelease.PackageName,
			PackageVersion: depRelease.PackageVersion,
			Source:         step.Source,
			Status:         depRelease.Status,
		})
	}

	release, err := installSingleRelease(global, releaseName, source, opts, nil, releaseRelationshipMetadata{dependencies: dependencyRefs})
	if err != nil {
		return err
	}
	fmt.Printf("installed %s revision %d\n", release.Name, release.Revision)
	return nil
}

func dependencyInstallOptions(opts packageBuildOptions) packageBuildOptions {
	depOpts := opts
	depOpts.valuesFile = ""
	depOpts.overlay = ""
	return depOpts
}

func installSingleRelease(global *globalOptions, releaseName string, source string, opts packageBuildOptions, overrideValues map[string]any, relationships releaseRelationshipMetadata) (*state.Release, error) {
	if err := state.ValidateReleaseName(releaseName); err != nil {
		return nil, err
	}

	cleanupValues, err := applyInlineValues(&opts, overrideValues)
	if err != nil {
		return nil, err
	}
	defer cleanupValues()

	src, err := preparePackageSource(source, true)
	if err != nil {
		return nil, err
	}
	defer src.cleanup()

	home, err := state.Home(global.home)
	if err != nil {
		return nil, err
	}
	revision := 1
	if currentRevision, err := state.ReadCurrentRevision(home, releaseName); err == nil {
		currentDir := state.RevisionDir(home, releaseName, currentRevision)
		currentRelease, readErr := state.ReadRelease(currentDir)
		if readErr != nil {
			return nil, readErr
		}
		if currentRelease.Status != "uninstalled" {
			return nil, fmt.Errorf("release %q already exists and is %s; use upgrade instead", releaseName, currentRelease.Status)
		}
		nextRevision, err := state.NextRevision(home, releaseName)
		if err != nil {
			return nil, err
		}
		revision = nextRevision
	}
	envEntries, err := loadCommandEnv(opts.envFile)
	if err != nil {
		return nil, err
	}
	manifest, vals, rendered, _, err := buildPackage(src.Dir, releaseName, opts)
	if err != nil {
		return nil, err
	}
	release, composePath, err := writeRevision(home, releaseName, revision, manifest, vals, rendered, src.Dir, src.Source, "pending", opts.envFile, relationships)
	if err != nil {
		return nil, err
	}
	if !opts.skipComposeConfig {
		ctx, cancel := context10m()
		defer cancel()
		if err := (runner.DockerComposeRunner{WorkDir: filepath.Dir(composePath), Project: releaseName, Env: envEntries}).ValidateConfig(ctx, composePath); err != nil {
			release.Status = "failed"
			_ = state.WriteRelease(filepath.Dir(composePath), *release)
			return nil, err
		}
	}
	ctx, cancel := context10m()
	defer cancel()
	if err := dockerRunnerWithEnv(releaseName, filepath.Dir(composePath), envEntries).Up(ctx, composePath); err != nil {
		release.Status = "failed"
		_ = state.WriteRelease(filepath.Dir(composePath), *release)
		return nil, err
	}
	release.Status = "deployed"
	if err := state.WriteRelease(filepath.Dir(composePath), *release); err != nil {
		return nil, err
	}
	if err := state.SetCurrentRevision(home, releaseName, revision); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		return nil, err
	}
	return release, nil
}

func applyInlineValues(opts *packageBuildOptions, inline map[string]any) (func(), error) {
	if len(inline) == 0 {
		return func() {}, nil
	}
	if opts.valuesFile != "" {
		return nil, errors.New("dependency inline values cannot be combined with --values")
	}
	tempDir, err := os.MkdirTemp("", "dockyard-dependency-values-*")
	if err != nil {
		return nil, fmt.Errorf("create dependency values temp dir: %w", err)
	}
	valuesPath := filepath.Join(tempDir, "values.yaml")
	if err := values.WriteValues(valuesPath, inline); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}
	opts.valuesFile = valuesPath
	return func() {
		_ = os.RemoveAll(tempDir)
	}, nil
}
