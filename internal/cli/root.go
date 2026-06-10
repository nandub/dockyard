package cli

import (
	"github.com/spf13/cobra"
)

type globalOptions struct {
	home string
}

func NewRootCommand() *cobra.Command {
	opts := &globalOptions{}

	cmd := &cobra.Command{
		Use:   "dockyard",
		Short: "Package manager and security linter for Docker Compose applications",
	}

	cmd.PersistentFlags().StringVar(&opts.home, "home", "", "Dockyard home directory; overrides DOCKYARD_HOME")

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newLintCommand())
	cmd.AddCommand(newRenderCommand())
	cmd.AddCommand(newConfigCommand())
	cmd.AddCommand(newInstallCommand(opts))
	cmd.AddCommand(newUninstallCommand(opts))
	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newStatusCommand(opts))
	cmd.AddCommand(newInspectCommand(opts))
	cmd.AddCommand(newDiffCommand(opts))
	cmd.AddCommand(newUpgradeCommand(opts))
	cmd.AddCommand(newRollbackCommand(opts))
	cmd.AddCommand(newDoctorCommand(opts))
	cmd.AddCommand(newLockCommand())
	cmd.AddCommand(newValuesCommand())
	cmd.AddCommand(newPackageCommand())
	cmd.AddCommand(newVerifyCommand())
	cmd.AddCommand(newPushCommand())
	cmd.AddCommand(newPullCommand())
	cmd.AddCommand(newPolicyCommand())
	cmd.AddCommand(newSecretsCommand())
	cmd.AddCommand(newEnvCommand())

	return cmd
}
