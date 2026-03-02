// Package cli defines the Cobra command hierarchy for skern.
package cli

import (
	"errors"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

var (
	jsonFlag  bool
	quietFlag bool
	printer   *output.Printer
)

// NewRootCmd creates the root skern command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skern",
		Short: "Agent-first CLI for managing Agent Skills",
		Long:  "skern is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			printer = output.NewPrinter(jsonFlag, quietFlag)
			printer.SetOut(cmd.OutOrStdout())
			printer.SetErrOut(cmd.ErrOrStderr())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")
	cmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "suppress non-essential output")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newSkillCmd())
	cmd.AddCommand(newPlatformCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCompletionCmd())

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

		var ve *ValidationError
		if errors.As(err, &ve) {
			return 2
		}
		return 1
	}
	return 0
}
