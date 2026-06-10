package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nandub/dockyard/internal/archive"
	"github.com/nandub/dockyard/internal/quality"
	"github.com/nandub/dockyard/internal/runner"
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
	cmd.AddCommand(newPackageTestCommand())
	return cmd
}

func newPackageLintCommand() *cobra.Command {
	var strict bool
	var allowAdvisory bool
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
			qualityOpts := quality.Options{Strict: strict, AllowAdvisory: allowAdvisory}
			report, err := quality.LintPackage(args[0], qualityOpts)
			if err != nil {
				return err
			}
			if jsonOut {
				if quality.HasBlockingFindings(report, qualityOpts) {
					_ = printJSON(report)
					return fmt.Errorf("package quality checks failed")
				}
				return printJSON(report)
			}
			printQualityReport(report)
			if quality.HasBlockingFindings(report, qualityOpts) {
				return fmt.Errorf("package quality checks failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "treat package quality warnings as failures")
	cmd.Flags().BoolVar(&allowAdvisory, "allow-advisory", false, "allow advisory warnings such as a missing package-local LICENSE when --strict is used")
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

type packageTestReport struct {
	Source          string         `json:"source"`
	PackageName     string         `json:"packageName,omitempty"`
	PackageVersion  string         `json:"packageVersion,omitempty"`
	Quality         quality.Report `json:"quality"`
	Rendered        bool           `json:"rendered"`
	ComposeConfigOK bool           `json:"composeConfigOk"`
	SmokeTested     bool           `json:"smokeTested"`
}

func newPackageTestCommand() *cobra.Command {
	var strict bool
	var allowAdvisory bool
	var smoke bool
	var jsonOut bool
	var opts packageBuildOptions

	cmd := &cobra.Command{
		Use:   "test PACKAGE_SOURCE",
		Short: "Run a non-destructive package validation pipeline",
		Long: `Run a package author validation pipeline.

The default pipeline prepares the source, runs package quality checks, renders
Compose with the selected values, runs Dockyard policy checks, and validates the
rendered output with docker compose config.

With --smoke, Dockyard also runs docker compose up/down using a temporary
Compose project name. Smoke tests do not write Dockyard release state.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := runPackageTest(args[0], quality.Options{Strict: strict, AllowAdvisory: allowAdvisory}, smoke, opts)
			if jsonOut {
				if err != nil {
					_ = printJSON(report)
					return err
				}
				return printJSON(report)
			}
			printPackageTestReport(report)
			return err
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "run package quality checks in strict mode")
	cmd.Flags().BoolVar(&allowAdvisory, "allow-advisory", false, "allow advisory warnings such as a missing package-local LICENSE when --strict is used")
	cmd.Flags().BoolVar(&smoke, "smoke", false, "run docker compose up/down with a temporary Compose project")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().StringVarP(&opts.valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&opts.overlay, "overlay", "", "compose overlay name")
	cmd.Flags().StringVar(&opts.envFile, "env-file", "", "env file passed to docker compose subprocesses")
	cmd.Flags().BoolVar(&opts.requireLock, "require-lock", false, "require dockyard.lock to match rendered output")
	cmd.Flags().BoolVar(&opts.allowRisk, "allow-risk", false, "allow HIGH policy findings")
	cmd.Flags().BoolVar(&opts.skipPolicy, "skip-policy", false, "skip Dockyard policy checks")
	cmd.Flags().BoolVar(&opts.skipComposeConfig, "skip-compose-config", false, "skip docker compose config validation")
	return cmd
}

func runPackageTest(source string, qualityOpts quality.Options, smoke bool, opts packageBuildOptions) (packageTestReport, error) {
	report := packageTestReport{Source: source}

	prepared, err := preparePackageSource(source, true)
	if err != nil {
		return report, err
	}
	defer prepared.cleanup()

	qualityReport, err := quality.LintPackage(prepared.Dir, qualityOpts)
	report.Quality = qualityReport
	report.PackageName = qualityReport.PackageName
	report.PackageVersion = qualityReport.PackageVersion
	if err != nil {
		return report, err
	}
	if quality.HasBlockingFindings(qualityReport, qualityOpts) {
		return report, fmt.Errorf("package quality checks failed")
	}

	_, _, rendered, _, err := buildPackage(prepared.Dir, "dockyard-test", opts)
	if err != nil {
		return report, err
	}
	report.Rendered = true

	tempDir, err := os.MkdirTemp("", "dockyard-package-test-*")
	if err != nil {
		return report, fmt.Errorf("create package test temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	composePath := filepath.Join(tempDir, "compose.rendered.yaml")
	if err := os.WriteFile(composePath, rendered, 0o600); err != nil {
		return report, fmt.Errorf("write rendered compose for package test: %w", err)
	}

	env, err := loadCommandEnv(opts.envFile)
	if err != nil {
		return report, err
	}

	project := fmt.Sprintf("dockyard-test-%d", time.Now().UnixNano())
	runner := dockerRunnerWithEnv(project, tempDir, env)

	if !opts.skipComposeConfig {
		ctx, cancel := context10m()
		defer cancel()
		if err := runner.ValidateConfig(ctx, composePath); err != nil {
			return report, err
		}
		report.ComposeConfigOK = true
	}

	if smoke {
		ctx, cancel := context10m()
		defer cancel()
		if err := preflightSmokeTest(ctx); err != nil {
			return report, err
		}
		if err := runner.Up(ctx, composePath); err != nil {
			return report, fmt.Errorf("smoke test failed: docker compose up failed; run `dockyard doctor` and verify Docker Desktop or the Docker daemon is running: %w", err)
		}
		defer func() {
			downCtx, downCancel := context10m()
			defer downCancel()
			_ = runner.Down(downCtx, composePath, false)
		}()
		if err := runner.PS(ctx, composePath, true); err != nil {
			return report, fmt.Errorf("smoke test failed: docker compose ps failed: %w", err)
		}
		report.SmokeTested = true
	}

	return report, nil
}

func preflightSmokeTest(ctx context.Context) error {
	if !runner.CommandExists("docker") {
		return fmt.Errorf("smoke test requires Docker: docker CLI not found")
	}
	if err := runner.DockerVersion(ctx); err != nil {
		return fmt.Errorf("smoke test requires Docker: %w", err)
	}
	if err := runner.ComposeAvailable(ctx); err != nil {
		return fmt.Errorf("smoke test requires Docker Compose: %w", err)
	}
	if err := runner.DaemonReachable(ctx); err != nil {
		return fmt.Errorf("smoke test requires a reachable Docker daemon; start Docker Desktop or your Docker daemon and retry: %w", err)
	}
	return nil
}

func printPackageTestReport(report packageTestReport) {
	printQualityReport(report.Quality)
	if report.Rendered {
		fmt.Println("OK: render - compose rendered with selected values")
	}
	if report.ComposeConfigOK {
		fmt.Println("OK: compose config - rendered Compose passed docker compose config")
	}
	if report.SmokeTested {
		fmt.Println("OK: smoke - docker compose up/down smoke test completed")
	}
}
