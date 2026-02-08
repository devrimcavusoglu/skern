package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/scribe/internal/output"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .scribe project directory",
		Long:  "Creates .scribe/ and .scribe/skills/ directories in the current project. Idempotent — safe to run multiple times.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsDir := filepath.Join(".", ".scribe", "skills")

			// Check if already initialized
			if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
				result := output.InitResult{
					Path:    filepath.Join(".", ".scribe"),
					Created: false,
				}
				text := fmt.Sprintf("Already initialized: %s\n", filepath.Join(".", ".scribe"))
				printer.PrintResult(result, text)
				return nil
			}

			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return fmt.Errorf("creating .scribe directory: %w", err)
			}

			result := output.InitResult{
				Path:    filepath.Join(".", ".scribe"),
				Created: true,
			}
			text := fmt.Sprintf("Initialized scribe project at %s\n", filepath.Join(".", ".scribe"))
			printer.PrintResult(result, text)
			return nil
		},
	}

	return cmd
}
