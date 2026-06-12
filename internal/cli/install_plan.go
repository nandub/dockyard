package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

const (
	planStepDependency = "dependency"
	planStepRoot       = "root"
)

type installPlanReport struct {
	ReleaseName    string            `json:"releaseName"`
	Home           string            `json:"home"`
	PackageName    string            `json:"packageName"`
	PackageVersion string            `json:"packageVersion"`
	Source         state.Source      `json:"source"`
	Steps          []installPlanStep `json:"steps"`
	ReadOnly       bool              `json:"readOnly"`
}

type installPlanStep struct {
	Order                 int            `json:"order"`
	Type                  string         `json:"type"`
	Name                  string         `json:"name"`
	Alias                 string         `json:"alias,omitempty"`
	Version               string         `json:"version,omitempty"`
	Source                string         `json:"source"`
	PlannedRelease        string         `json:"plannedRelease"`
	Action                string         `json:"action"`
	ExistingReleaseStatus string         `json:"existingReleaseStatus,omitempty"`
	Values                map[string]any `json:"values,omitempty"`
}

func newInstallPlanCommand(global *globalOptions) *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "install-plan RELEASE PACKAGE_SOURCE",
		Short: "Preview a dependency-aware Dockyard install plan",
		Long: `Preview how Dockyard would install a package and its declared dependencies.

This command is intentionally read-only. It validates the root release name,
loads dependency metadata, generates deterministic dependency release names,
checks for existing releases, and prints the planned install order. It does not
install, upgrade, uninstall, or modify release state.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := buildInstallPlan(global, args[0], args[1])
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(report)
			}
			printInstallPlan(report)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}

func buildInstallPlan(global *globalOptions, releaseName string, source string) (installPlanReport, error) {
	if err := state.ValidateReleaseName(releaseName); err != nil {
		return installPlanReport{}, err
	}

	prepared, err := preparePackageSource(source, true)
	if err != nil {
		return installPlanReport{}, err
	}
	defer prepared.cleanup()

	manifest, err := dockpkg.LoadManifest(prepared.Dir)
	if err != nil {
		return installPlanReport{}, err
	}

	home, err := state.Home(global.home)
	if err != nil {
		return installPlanReport{}, err
	}

	steps := make([]installPlanStep, 0, len(manifest.Dependencies)+1)
	plannedReleases := map[string]struct{}{}

	for _, dep := range manifest.Dependencies {
		plannedRelease := dependencyReleaseName(releaseName, dep)
		if err := state.ValidateReleaseName(plannedRelease); err != nil {
			return installPlanReport{}, fmt.Errorf("dependency %q planned release %q is invalid: %w", dep.Name, plannedRelease, err)
		}
		if _, exists := plannedReleases[plannedRelease]; exists {
			return installPlanReport{}, fmt.Errorf("duplicate planned release name %q", plannedRelease)
		}
		plannedReleases[plannedRelease] = struct{}{}

		status, err := currentReleaseStatus(home, plannedRelease)
		if err != nil {
			return installPlanReport{}, err
		}
		steps = append(steps, installPlanStep{
			Order:                 len(steps) + 1,
			Type:                  planStepDependency,
			Name:                  dep.Name,
			Alias:                 dep.Alias,
			Version:               dep.Version,
			Source:                dep.Source,
			PlannedRelease:        plannedRelease,
			Action:                plannedInstallAction(status),
			ExistingReleaseStatus: status,
			Values:                dep.Values,
		})
	}

	if _, exists := plannedReleases[releaseName]; exists {
		return installPlanReport{}, fmt.Errorf("root release name %q conflicts with a dependency planned release", releaseName)
	}
	rootStatus, err := currentReleaseStatus(home, releaseName)
	if err != nil {
		return installPlanReport{}, err
	}
	steps = append(steps, installPlanStep{
		Order:                 len(steps) + 1,
		Type:                  planStepRoot,
		Name:                  manifest.Name,
		Version:               manifest.Version,
		Source:                source,
		PlannedRelease:        releaseName,
		Action:                plannedInstallAction(rootStatus),
		ExistingReleaseStatus: rootStatus,
	})

	return installPlanReport{
		ReleaseName:    releaseName,
		Home:           home,
		PackageName:    manifest.Name,
		PackageVersion: manifest.Version,
		Source:         prepared.Source,
		Steps:          steps,
		ReadOnly:       true,
	}, nil
}

func buildInstallDryRunPlan(global *globalOptions, releaseName string, source string) (installPlanReport, error) {
	return buildInstallPlan(global, releaseName, source)
}

func dependencyReleaseName(rootRelease string, dep dockpkg.Dependency) string {
	suffix := dep.Alias
	if suffix == "" {
		suffix = dep.Name
	}
	return rootRelease + "-" + suffix
}

func currentReleaseStatus(home string, releaseName string) (string, error) {
	revision, err := state.ReadCurrentRevision(home, releaseName)
	if err != nil {
		if isNotExistError(err) {
			return "", nil
		}
		return "", err
	}
	release, err := state.ReadRelease(state.RevisionDir(home, releaseName, revision))
	if err != nil {
		if isNotExistError(err) {
			return "", nil
		}
		return "", err
	}
	return release.Status, nil
}

func isNotExistError(err error) bool {
	if err == nil {
		return false
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return os.IsNotExist(pathErr)
	}
	return os.IsNotExist(err)
}

func plannedInstallAction(existingStatus string) string {
	switch existingStatus {
	case "":
		return "install"
	case "uninstalled":
		return "reinstall"
	case "deployed":
		return "exists"
	default:
		return "blocked"
	}
}

func printInstallPlan(report installPlanReport) {
	fmt.Printf("Install plan for release %s\n\n", report.ReleaseName)
	for _, step := range report.Steps {
		switch step.Type {
		case planStepDependency:
			label := step.Name
			if step.Alias != "" {
				label += " as " + step.Alias
			}
			if step.Version != "" {
				label += "@" + step.Version
			}
			fmt.Printf("%d. dependency: %s\n", step.Order, label)
			fmt.Printf("   source: %s\n", step.Source)
			fmt.Printf("   planned release: %s\n", step.PlannedRelease)
			fmt.Printf("   action: %s\n", step.Action)
			if step.ExistingReleaseStatus != "" {
				fmt.Printf("   existing status: %s\n", step.ExistingReleaseStatus)
			}
			fmt.Println("   automatic install: use `dockyard install --with-dependencies`")
		default:
			fmt.Printf("%d. root package: %s@%s\n", step.Order, step.Name, step.Version)
			fmt.Printf("   source: %s\n", step.Source)
			fmt.Printf("   planned release: %s\n", step.PlannedRelease)
			fmt.Printf("   action: %s\n", step.Action)
			if step.ExistingReleaseStatus != "" {
				fmt.Printf("   existing status: %s\n", step.ExistingReleaseStatus)
			}
		}
		if step.Order != len(report.Steps) {
			fmt.Println()
		}
	}
	fmt.Println()
	fmt.Println("Read-only: no releases were installed, upgraded, uninstalled, or modified.")
}
