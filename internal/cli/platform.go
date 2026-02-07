package cli

import (
	"github.com/spf13/cobra"
)

func newPlatformCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "platform",
		Short: "Manage platform adapters",
		Long:  "List detected platforms and check skill installation status across platforms.",
	}

	cmd.AddCommand(newPlatformListCmd())
	cmd.AddCommand(newPlatformStatusCmd())

	return cmd
}
