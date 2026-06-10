package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/quality"
	"github.com/spf13/cobra"
)

func newPackageCommand() *cobra.Command {
	var output string
	var locked bool
	var opts packageBuildOptions

	cmd := &cobra.Command{
		Use:   "package PACKAGE_DIR",
		Short: "Create a local Dockyard package archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if locked {
				opts.requireLock = true
				if _, _, _, _, err := buildPackage(args[0], "package", opts); err != nil {
					return err
				}
			}
			path, err := archive.PackageDir(args[0], output, locked)
			if err != nil {
				return err
			}
			fmt.Printf("package created: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path")
	cmd.Flags().BoolVar(&locked, "locked", false, "require dockyard.lock and include lock provenance")
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file used to verify dockyard.lock when --locked is set")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name used to verify dockyard.lock when --locked is set")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings during --locked verification")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip policy checks during --locked verification")
	cmd.AddCommand(newPackageLintCommand())
	return cmd
}

func newPackageLintCommand() *cobra.Command {
	var strict bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "lint PACKAGE_DIR",
		Short: "Run package quality checks for Dockyard package authors",
		Long: `Run package quality checks that go beyond basic format compatibility.

This command checks package documentation files, forbidden local artifacts,
values schema quality, sensitive schema markers, default rendering, and policy
linting. Use --strict before publishing package examples or registry-ready
packages.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := quality.LintPackage(args[0], quality.Options{Strict: strict})
			if err != nil {
				return err
			}
			if jsonOut {
				if quality.HasFailures(report) {
					_ = printJSON(report)
					return fmt.Errorf("package quality checks failed")
				}
				return printJSON(report)
			}
			printQualityReport(report)
			if quality.HasFailures(report) {
				return fmt.Errorf("package quality checks failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "treat package quality warnings as failures")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}

func printQualityReport(report quality.Report) {
	if report.PackageName != "" {
		fmt.Printf("Package: %s@%s\n", report.PackageName, report.PackageVersion)
	}
	for _, check := range report.Checks {
		fmt.Printf("%s: %s - %s\n", check.Severity, check.Name, check.Message)
		for _, detail := range check.Details {
			fmt.Printf("  - %s\n", detail)
		}
	}
}
