package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nandub/dockyard/internal/oci"
	"github.com/nandub/dockyard/internal/runner"
	"github.com/nandub/dockyard/internal/state"
	"github.com/spf13/cobra"
)

func newDoctorCommand(global *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local Dockyard and Docker prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := state.Home(global.home)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(home, 0o700); err != nil {
				return fmt.Errorf("Dockyard home is not writable: %w", err)
			}
			fmt.Printf("OK: Dockyard home writable: %s\n", home)
			if !runner.CommandExists("docker") {
				return fmt.Errorf("docker CLI was not found in PATH")
			}
			fmt.Println("OK: docker found")
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := runner.DockerVersion(ctx); err != nil {
				return err
			}
			fmt.Println("OK: docker CLI usable")
			if err := runner.ComposeAvailable(ctx); err != nil {
				return err
			}
			fmt.Println("OK: docker compose available")
			if err := runner.DaemonReachable(ctx); err != nil {
				return err
			}
			fmt.Println("OK: Docker daemon reachable")
			if oci.CommandAvailable() {
				fmt.Println("OK: oras found for OCI registry operations")
			} else {
				fmt.Println("WARN: oras not found; OCI push/pull commands will not work")
			}
			return nil
		},
	}
	return cmd
}
