// Package cli defines the Cobra command hierarchy for skern.
package cli

import (
	"errors"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root skern command with default dependencies.
func NewRootCmd() *cobra.Command {
	return newRootCmd(nil)
}

func newRootCmd(cc *CommandContext) *cobra.Command {
	if cc == nil {
		cc = &CommandContext{
			NewRegistry: defaultNewRegistry,
			NewDetector: defaultNewDetector,
		}
	}

	var (
		jsonFlag  bool
		quietFlag bool
	)

	cmd := &cobra.Command{
		Use:   "skern",
		Short: "Agent-first CLI for managing Agent Skills",
		Long:  "skern is a minimal, agent-first CLI tool for managing Agent Skills across agentic development platforms.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cc.Printer = output.NewPrinter(jsonFlag, quietFlag)
			cc.Printer.SetOut(cmd.OutOrStdout())
			cc.Printer.SetErrOut(cmd.ErrOrStderr())
			setContext(cmd, cc)
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
	cc := &CommandContext{
		NewRegistry: defaultNewRegistry,
		NewDetector: defaultNewDetector,
	}
	cmd := newRootCmd(cc)
	if err := cmd.Execute(); err != nil {
		if cc.Printer == nil {
			cc.Printer = output.NewPrinter(false, false)
		}
		cc.Printer.PrintErrorResult(err)

		var ve *ValidationError
		if errors.As(err, &ve) {
			return 2
		}
		return 1
	}
	return 0
}
