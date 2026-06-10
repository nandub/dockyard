package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/format"
	"github.com/nandub/dockyard/internal/lock"
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

type compatResult struct {
	Formats []format.Format `json:"formats,omitempty"`
	Checks  []compatCheck   `json:"checks,omitempty"`
}

type compatCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Details string `json:"details,omitempty"`
}

func newCompatCommand(global *globalOptions) *cobra.Command {
	var jsonOut bool
	var releaseName string

	cmd := &cobra.Command{
		Use:   "compat [PACKAGE_SOURCE]",
		Short: "Show Dockyard format compatibility or check a package/release",
		Long: `Show supported Dockyard file format versions.

With PACKAGE_SOURCE, Dockyard checks package manifest, optional lockfile,
values schema, and package archive structure when applicable.

With --release, Dockyard checks the current release metadata format.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if releaseName != "" {
				result, err := checkReleaseCompatibility(global, releaseName)
				if err != nil {
					return err
				}
				return printCompatResult(result, jsonOut)
			}

			if len(args) == 0 {
				result := compatResult{Formats: format.SupportedFormats()}
				return printCompatResult(result, jsonOut)
			}

			src, err := preparePackageSource(args[0], true)
			if err != nil {
				return err
			}
			defer src.cleanup()

			result := checkPackageCompatibility(src.Dir, args[0])
			return printCompatResult(result, jsonOut)
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().StringVar(&releaseName, "release", "", "check the current release metadata format")
	return cmd
}

func checkPackageCompatibility(packageDir string, originalSource string) compatResult {
	checks := []compatCheck{}

	manifest, err := dockpkg.LoadManifest(packageDir)
	if err != nil {
		checks = append(checks, compatCheck{Name: "Dockyard.yaml", Status: "FAIL", Details: err.Error()})
	} else {
		checks = append(checks, compatCheck{
			Name:    "Dockyard.yaml",
			Status:  "OK",
			Details: fmt.Sprintf("%s package=%s version=%s", manifest.APIVersion, manifest.Name, manifest.Version),
		})
	}

	valuesPath, err := dockpkg.SafeJoin(packageDir, "values.yaml")
	if err != nil {
		checks = append(checks, compatCheck{Name: "values.yaml", Status: "FAIL", Details: err.Error()})
	} else if _, err := os.Stat(valuesPath); err != nil {
		checks = append(checks, compatCheck{Name: "values.yaml", Status: "FAIL", Details: err.Error()})
	} else {
		checks = append(checks, compatCheck{Name: "values.yaml", Status: "OK"})
	}

	schemaPath, err := dockpkg.SafeJoin(packageDir, "values.schema.json")
	if err != nil {
		checks = append(checks, compatCheck{Name: "values.schema.json", Status: "FAIL", Details: err.Error()})
	} else if _, err := os.Stat(schemaPath); err != nil {
		if os.IsNotExist(err) {
			checks = append(checks, compatCheck{Name: "values.schema.json", Status: "WARN", Details: "schema is optional but recommended"})
		} else {
			checks = append(checks, compatCheck{Name: "values.schema.json", Status: "FAIL", Details: err.Error()})
		}
	} else {
		checks = append(checks, compatCheck{Name: "values.schema.json", Status: "OK"})
	}

	lockPath := filepath.Join(packageDir, lock.FileName)
	if _, err := os.Stat(lockPath); err == nil {
		lf, err := lock.Read(lockPath)
		if err != nil {
			checks = append(checks, compatCheck{Name: lock.FileName, Status: "FAIL", Details: err.Error()})
		} else {
			checks = append(checks, compatCheck{Name: lock.FileName, Status: "OK", Details: lf.APIVersion})
		}
	} else if os.IsNotExist(err) {
		checks = append(checks, compatCheck{Name: lock.FileName, Status: "WARN", Details: "lockfile is optional unless --require-lock is used"})
	} else {
		checks = append(checks, compatCheck{Name: lock.FileName, Status: "FAIL", Details: err.Error()})
	}

	if isArchivePath(originalSource) {
		if err := archive.VerifyArchive(originalSource, nil); err != nil {
			checks = append(checks, compatCheck{Name: "package archive", Status: "FAIL", Details: err.Error()})
		} else {
			checks = append(checks, compatCheck{Name: "package archive", Status: "OK", Details: "archive structure and SHA256SUMS verified"})
		}
	}

	return compatResult{Checks: checks}
}

func checkReleaseCompatibility(global *globalOptions, releaseName string) (compatResult, error) {
	if err := state.ValidateReleaseName(releaseName); err != nil {
		return compatResult{}, err
	}
	home, err := state.Home(global.home)
	if err != nil {
		return compatResult{}, err
	}
	release, _, err := readCurrentRelease(home, releaseName)
	if err != nil {
		return compatResult{}, err
	}
	check := compatCheck{
		Name:   "release.json",
		Status: "OK",
	}
	if release.APIVersion == "" {
		check.Status = "WARN"
		check.Details = "legacy release state without apiVersion; v0.11 can read it but new revisions write " + format.ReleaseAPIVersion
	} else {
		check.Details = release.APIVersion
	}
	return compatResult{Checks: []compatCheck{check}}, nil
}

func printCompatResult(result compatResult, jsonOut bool) error {
	if jsonOut {
		return printJSON(result)
	}

	if len(result.Formats) > 0 {
		fmt.Println("Supported Dockyard formats:")
		for _, f := range result.Formats {
			fmt.Printf("- %s: %s (%s)\n  %s\n", f.Name, f.APIVersion, f.Stability, f.Notes)
		}
		return nil
	}

	for _, check := range result.Checks {
		if check.Details != "" {
			fmt.Printf("%s: %s - %s\n", check.Status, check.Name, check.Details)
			continue
		}
		fmt.Printf("%s: %s\n", check.Status, check.Name)
	}
	return nil
}
