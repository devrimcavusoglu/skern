// Package cli defines the Cobra command hierarchy for scribe.
package cli

import (
	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonFlag  bool
	quietFlag bool
	printer   *output.Printer
)

// NewRootCmd creates the root scribe command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scribe",
		Short: "Agent-first CLI for managing Agent Skills",
		Long:  "scribe is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			printer = output.NewPrinter(jsonFlag, quietFlag)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")
	cmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "suppress non-essential output")

	cmd.AddCommand(newVersionCmd())

	return cmd
}

// Execute runs the root command.
func Execute() int {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		if printer == nil {
			printer = output.NewPrinter(false, false)
		}
		printer.PrintErrorResult(err)
		return 1
	}
	return 0
}
