package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nandub/dockyard/internal/dockpkg"
	valpkg "github.com/nandub/dockyard/internal/values"
	"github.com/spf13/cobra"
)

func newValuesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "values",
		Short: "Work with Dockyard package values",
	}

	cmd.AddCommand(newValuesTemplateCommand())
	cmd.AddCommand(newValuesValidateCommand())
	cmd.AddCommand(newValuesSchemaCommand())

	return cmd
}

func newValuesTemplateCommand() *cobra.Command {
	var output string
	var force bool

	cmd := &cobra.Command{
		Use:   "template PACKAGE_DIR",
		Short: "Generate an operator-friendly values override template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageDir := args[0]
			if _, err := dockpkg.LoadManifest(packageDir); err != nil {
				return err
			}
			data, err := valpkg.GenerateTemplate(packageDir)
			if err != nil {
				return err
			}
			if output == "" {
				fmt.Print(string(data))
				return nil
			}
			cleanOutput := filepath.Clean(output)
			if _, err := os.Stat(cleanOutput); err == nil && !force {
				return fmt.Errorf("refusing to overwrite existing file %q; use --force", cleanOutput)
			}
			if err := os.MkdirAll(filepath.Dir(cleanOutput), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(cleanOutput, data, 0o600); err != nil {
				return err
			}
			fmt.Printf("values template written: %s\n", cleanOutput)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output values file; defaults to stdout")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite output file when it already exists")
	return cmd
}

func newValuesValidateCommand() *cobra.Command {
	var valuesFile string

	cmd := &cobra.Command{
		Use:   "validate PACKAGE_DIR",
		Short: "Validate values against a package values.schema.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageDir := args[0]
			if _, err := dockpkg.LoadManifest(packageDir); err != nil {
				return err
			}
			vals, err := valpkg.LoadValues(packageDir, valuesFile)
			if err != nil {
				return err
			}
			if err := valpkg.ValidateAgainstSchema(packageDir, vals); err != nil {
				return err
			}
			fmt.Println("OK: values are valid")
			return nil
		},
	}

	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "values override file")
	return cmd
}

func newValuesSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema PACKAGE_DIR",
		Short: "Print a package values.schema.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageDir := args[0]
			if _, err := dockpkg.LoadManifest(packageDir); err != nil {
				return err
			}
			schemaPath, err := dockpkg.SafeJoin(packageDir, valpkg.SchemaFile)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(schemaPath)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}
	return cmd
}
