package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/dockpkg"
	"github.com/nandub/dockyard/internal/render"
	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

func newRenderCommand() *cobra.Command {
	var valuesFile string
	var overlay string
	var outputFile string
	var explain bool
	var validateCompose bool
	var envFile string

	cmd := &cobra.Command{
		Use:   "render PACKAGE_DIR",
		Short: "Render a Dockyard package to Docker Compose YAML",
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
			result, err := render.RenderComposeWithDiagnostics(packageDir, manifest, vals, overlay)
			if err != nil {
				return err
			}
			if validateCompose {
				envEntries, err := loadCommandEnv(envFile)
				if err != nil {
					return err
				}
				tempDir, err := os.MkdirTemp("", "dockyard-render-*")
				if err != nil {
					return fmt.Errorf("create temp dir: %w", err)
				}
				defer os.RemoveAll(tempDir)
				composePath := filepath.Join(tempDir, "compose.rendered.yaml")
				if err := os.WriteFile(composePath, result.YAML, 0o600); err != nil {
					return fmt.Errorf("write temporary compose file: %w", err)
				}
				ctx, cancel := context10m()
				defer cancel()
				if err := (runner.DockerComposeRunner{WorkDir: tempDir, Project: "dockyard-render", Env: envEntries}).ValidateConfig(ctx, composePath); err != nil {
					return err
				}
			}
			if explain {
				for _, diagnostic := range result.Diagnostics {
					fmt.Fprintf(os.Stderr, "%s => %s\n", diagnostic.Key, diagnostic.Value)
				}
			}
			if outputFile == "" {
				fmt.Print(string(result.YAML))
				return nil
			}
			return os.WriteFile(outputFile, result.YAML, 0o600)
		},
	}
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	cmd.Flags().StringVar(&envFile, "env-file", "", "dotenv file to pass to docker compose when using --validate-compose")
	cmd.Flags().StringVar(&overlay, "overlay", "", "compose overlay name")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output compose file")
	cmd.Flags().BoolVar(&explain, "explain", false, "show resolved placeholders on stderr; sensitive values are masked")
	cmd.Flags().BoolVar(&validateCompose, "validate-compose", false, "run docker compose config against the rendered output")
	return cmd
}
