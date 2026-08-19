package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/devrimcavusoglu/skern/internal/cli/instructions"
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		writeInstr  bool
		noInstr     bool
		toolForming bool
		printInstr  bool
		targetPaths []string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .skern project directory",
		Long: `Creates .skern/ and .skern/skills/ directories in the current project.
Optionally writes a skern usage snippet into agent instruction files
(AGENTS.md, CLAUDE.md, .claude/CLAUDE.md) so the agent uses skern for
all skill-related tasks.

Idempotent — safe to run multiple times. The instruction snippet is
wrapped in start/end markers so re-running updates the block in place.

Interactivity: when no instruction flag is given, init asks whether to write
the snippet only if stdin is a terminal and --json is not set. When stdin is
not a TTY (installers, CI, piped input) or --json is set, skern never prompts
and never blocks on input — the answer defaults to "no". Pass
--no-instructions to state that opt-out explicitly instead of relying on the
non-TTY default, or --instructions to opt in.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, runInitOpts{
				writeInstr:  writeInstr,
				noInstr:     noInstr,
				toolForming: toolForming,
				printInstr:  printInstr,
				targetPaths: targetPaths,
			})
		},
	}

	cmd.Flags().BoolVar(&writeInstr, "instructions", false,
		"write the skern usage snippet to agent instruction files (AGENTS.md, CLAUDE.md, .claude/CLAUDE.md by default)")
	cmd.Flags().BoolVar(&noInstr, "no-instructions", false,
		"do not write or offer the instruction snippet; never prompts (explicit opt-out for installers and CI)")
	cmd.Flags().BoolVar(&toolForming, "tool-forming-loop", false,
		"include the tool-forming-loop section in the instruction snippet (search-before-create workflow)")
	cmd.Flags().BoolVar(&printInstr, "print-instructions", false,
		"print the rendered instruction snippet to stdout instead of writing files")
	cmd.Flags().StringSliceVar(&targetPaths, "target", nil,
		"explicit instruction file path to write to; repeatable. Disables auto-discovery when set.")

	return cmd
}

type runInitOpts struct {
	writeInstr  bool
	noInstr     bool
	toolForming bool
	printInstr  bool
	targetPaths []string
}

func runInit(cmd *cobra.Command, opts runInitOpts) error {
	cc := getContext(cmd)

	skillsDir := filepath.Join(".", ".skern", "skills")
	skernPath := filepath.Join(".", ".skern")
	created := true
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		created = false
	} else if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("creating .skern directory: %w", err)
	}

	instrResult, err := handleInstructions(cmd, cc, opts)
	if err != nil {
		return err
	}

	result := output.InitResult{
		Path:         skernPath,
		Created:      created,
		Instructions: instrResult,
	}
	text := initTextSummary(skernPath, created, instrResult)
	cc.Printer.PrintResult(result, text)
	return nil
}

// handleInstructions resolves the user's choices (flags + interactive
// prompts), writes or prints the snippet, and returns a structured result.
// Returns nil when the user did not opt in to writing instructions.
func handleInstructions(cmd *cobra.Command, cc *CommandContext, opts runInitOpts) (*output.InstructionsResult, error) {
	wantInstr, wantToolForming, err := resolveInstructionChoices(cmd, cc, opts)
	if err != nil {
		return nil, err
	}
	if !wantInstr {
		return nil, nil
	}

	rendered := instructions.Render(wantToolForming)

	if opts.printInstr {
		_, _ = io.WriteString(cmd.OutOrStdout(), rendered)
		return &output.InstructionsResult{
			ToolForming: wantToolForming,
			Targets:     nil,
			Writes:      nil,
			Printed:     true,
		}, nil
	}

	targets, err := resolveTargets(opts)
	if err != nil {
		return nil, err
	}

	res := &output.InstructionsResult{
		ToolForming: wantToolForming,
		Targets:     targets,
		Writes:      []output.InstructionWriteResult{},
	}
	for _, t := range targets {
		w, werr := instructions.Write(t, rendered)
		if werr != nil {
			return nil, werr
		}
		res.Writes = append(res.Writes, output.InstructionWriteResult{
			Path:    w.Path,
			Action:  w.Action,
			Created: w.Created,
		})
	}
	return res, nil
}

// resolveInstructionChoices folds flag values + TTY interactivity into the
// final (writeInstructions, toolFormingLoop) decision.
func resolveInstructionChoices(cmd *cobra.Command, cc *CommandContext, opts runInitOpts) (bool, bool, error) {
	flags := cmd.Flags()

	// Explicit opt-out (#104): nothing is written and neither prompt runs,
	// regardless of TTY state. Combining it with an opt-in flag is a
	// contradiction, reported as a validation error (exit 2).
	if opts.noInstr {
		for _, name := range []string{"instructions", "print-instructions", "target", "tool-forming-loop"} {
			if flags.Changed(name) {
				return false, false, &ValidationError{Message: fmt.Sprintf("--no-instructions cannot be combined with --%s", name)}
			}
		}
		return false, false, nil
	}

	wantInstr := opts.writeInstr || opts.printInstr || len(opts.targetPaths) > 0
	wantToolForming := opts.toolForming

	// Skip prompting when JSON mode (machine-driven) or when stdin is not a
	// terminal (CI, scripts, redirected input, tests). This is the documented
	// non-interactive contract: without a TTY skern never blocks on input and
	// both questions resolve to their default, "no".
	in := cmd.InOrStdin()
	canPrompt := !cc.Printer.IsJSON() && isTerminal(in)

	// Prompts go to stderr so they never collide with --print-instructions
	// output on stdout when scripts pipe init through.
	promptOut := cmd.ErrOrStderr()

	if !wantInstr && canPrompt && !flags.Changed("instructions") &&
		!flags.Changed("print-instructions") && len(opts.targetPaths) == 0 {
		yes, err := promptYesNo(in, promptOut,
			"Append skern usage instructions to agent config files (AGENTS.md, CLAUDE.md, .claude/CLAUDE.md)?", false)
		if err != nil {
			return false, false, err
		}
		wantInstr = yes
	}

	if wantInstr && !wantToolForming && canPrompt && !flags.Changed("tool-forming-loop") {
		yes, err := promptYesNo(in, promptOut,
			"Include tool-forming-loop section (instructs the agent to search before creating)?", false)
		if err != nil {
			return false, false, err
		}
		wantToolForming = yes
	}

	return wantInstr, wantToolForming, nil
}

// resolveTargets returns the list of files to write to. Explicit --target
// paths win; otherwise auto-discovery probes CandidateProjectFiles in cwd.
func resolveTargets(opts runInitOpts) ([]string, error) {
	if len(opts.targetPaths) > 0 {
		return opts.targetPaths, nil
	}
	return instructions.DiscoverTargets(".")
}

// isTerminal reports whether r is a *os.File backed by a character device
// (terminal). Returns false for non-file readers (e.g. test injectees) so
// tests never trigger interactive prompts.
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

// promptYesNo writes prompt to w and reads a y/n answer from r. The default
// (returned when the user just hits enter) is controlled by defaultYes.
func promptYesNo(r io.Reader, w io.Writer, prompt string, defaultYes bool) (bool, error) {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}
	if _, err := fmt.Fprint(w, prompt+suffix); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		// EOF / no input — fall back to default.
		return defaultYes, nil
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	switch answer {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return defaultYes, nil
	default:
		return defaultYes, nil
	}
}

func initTextSummary(skernPath string, created bool, instr *output.InstructionsResult) string {
	var b strings.Builder
	if created {
		fmt.Fprintf(&b, "Initialized skern project at %s\n", skernPath)
	} else {
		fmt.Fprintf(&b, "Already initialized: %s\n", skernPath)
	}
	if instr == nil {
		return b.String()
	}
	if instr.Printed {
		return b.String() // snippet already streamed to stdout above
	}
	if len(instr.Writes) == 0 {
		fmt.Fprintln(&b, "No agent instruction files found (looked for AGENTS.md, CLAUDE.md, .claude/CLAUDE.md).")
		fmt.Fprintln(&b, "Pass --target <path> to write to a specific file.")
		return b.String()
	}
	tfTag := ""
	if instr.ToolForming {
		tfTag = " (with tool-forming-loop)"
	}
	fmt.Fprintf(&b, "Wrote skern usage snippet%s to:\n", tfTag)
	for _, w := range instr.Writes {
		fmt.Fprintf(&b, "  %s [%s]\n", w.Path, w.Action)
	}
	return b.String()
}
