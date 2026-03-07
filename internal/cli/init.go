package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .skern project directory",
		Long:  "Creates .skern/ and .skern/skills/ directories in the current project. Idempotent — safe to run multiple times.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			skillsDir := filepath.Join(".", ".skern", "skills")

			// Check if already initialized
			if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
				result := output.InitResult{
					Path:    filepath.Join(".", ".skern"),
					Created: false,
				}
				text := fmt.Sprintf("Already initialized: %s\n", filepath.Join(".", ".skern"))
				getContext(cmd).Printer.PrintResult(result, text)
				return nil
			}

			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return fmt.Errorf("creating .skern directory: %w", err)
			}

			result := output.InitResult{
				Path:    filepath.Join(".", ".skern"),
				Created: true,
			}
			text := fmt.Sprintf("Initialized skern project at %s\n", filepath.Join(".", ".skern"))
			getContext(cmd).Printer.PrintResult(result, text)
			return nil
		},
	}

	return cmd
}
