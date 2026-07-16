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

The default catalog is oci://ghcr.io/nandub/dockyard-packages/catalog:latest.
Override it by setting DOCKYARD_CATALOG to another OCI catalog reference, for example:
DOCKYARD_CATALOG=oci://ghcr.io/my-org/my-packages/catalog:latest.

For compatibility, DOCKYARD_CATALOG may also be set to a registry prefix such as
ghcr.io/my-org/my-packages; Dockyard will resolve it to /catalog:latest.`,
	}
	cmd.AddCommand(newCatalogListCommand())
	cmd.AddCommand(newCatalogInfoCommand())
	cmd.AddCommand(newCatalogPublishCommand())
	return cmd
}

func newCatalogListCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List packages in the configured catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgs, err := catalog.List()
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(struct {
					Reference string            `json:"reference"`
					Packages  []catalog.Package `json:"packages"`
				}{Reference: catalog.Reference(), Packages: pkgs})
			}
			fmt.Printf("CATALOG  %s\n\n", catalog.Reference())
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

func newCatalogPublishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish CATALOG_YAML OCI_REFERENCE",
		Short: "Publish a catalog YAML file to an OCI registry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context10m()
			defer cancel()
			if err := catalog.Publish(ctx, args[0], args[1]); err != nil {
				return err
			}
			fmt.Printf("published %s to %s\n", args[0], args[1])
			return nil
		},
	}
	return cmd
}

func newCatalogInfoCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "info PACKAGE",
		Short: "Show details for a catalog package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, ok, err := catalog.Get(args[0])
			if err != nil {
				return err
			}
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
					Reference string `json:"reference"`
					Source    string `json:"source"`
				}{Package: pkg, Reference: catalog.Reference(), Source: source})
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
