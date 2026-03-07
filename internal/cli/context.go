package cli

import (
	"context"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/devrimcavusoglu/skern/internal/platform"
	"github.com/devrimcavusoglu/skern/internal/registry"
	"github.com/spf13/cobra"
)

type contextKey struct{}

// CommandContext holds injectable dependencies for CLI commands.
type CommandContext struct {
	Printer     *output.Printer
	NewRegistry func() (*registry.Registry, error)
	NewDetector func() (*platform.Detector, error)
}

func setContext(cmd *cobra.Command, cc *CommandContext) {
	cmd.SetContext(context.WithValue(cmd.Context(), contextKey{}, cc))
}

func getContext(cmd *cobra.Command) *CommandContext {
	return cmd.Context().Value(contextKey{}).(*CommandContext)
}
