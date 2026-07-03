package cli

import (
	"fmt"

	"github.com/nandub/dockyard/internal/catalog"
	"github.com/spf13/cobra"
)

func newCatalogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Inspect the configured Dockyard package catalog",
		Long: `Inspect Dockyard catalog packages.

The default catalog registry is ghcr.io/nandub/dockyard-packages. Override it by
setting DOCKYARD_CATALOG to another OCI registry prefix, for example:
DOCKYARD_CATALOG=ghcr.io/my-org/my-packages.`,
	}
	cmd.AddCommand(newCatalogListCommand())
	cmd.AddCommand(newCatalogInfoCommand())
	return cmd
}

func newCatalogListCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List packages in the configured catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgs := catalog.List()
			if jsonOut {
				return printJSON(struct {
					Registry string            `json:"registry"`
					Packages []catalog.Package `json:"packages"`
				}{Registry: catalog.Registry(), Packages: pkgs})
			}
			fmt.Printf("REGISTRY  %s\n\n", catalog.Registry())
			fmt.Println("NAME          LATEST  DESCRIPTION")
			for _, pkg := range pkgs {
				fmt.Printf("%-13s %-7s %s\n", pkg.Name, pkg.Latest, pkg.Description)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}

func newCatalogInfoCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "info PACKAGE",
		Short: "Show details for a catalog package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, ok := catalog.Get(args[0])
			if !ok {
				return fmt.Errorf("package %q was not found in the configured catalog", args[0])
			}
			source, err := catalog.ResolveName(pkg.Name, pkg.Latest)
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(struct {
					catalog.Package
					Registry string `json:"registry"`
					Source   string `json:"source"`
				}{Package: pkg, Registry: catalog.Registry(), Source: source})
			}
			fmt.Printf("Name: %s\n", pkg.Name)
			fmt.Printf("Latest: %s\n", pkg.Latest)
			fmt.Printf("Source: %s\n", source)
			fmt.Printf("Description: %s\n", pkg.Description)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}
