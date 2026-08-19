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
	"golang.org/x/term"
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

Interactivity: init asks its two questions (write the snippet? include the
tool-forming loop?) only when no instruction flag is given, stdin is a
terminal, and --json is not set. Any instruction flag (--instructions,
--no-instructions, --print-instructions, --target, --tool-forming-loop)
disables both prompts. When stdin is not a TTY (installers, CI, piped or
redirected input, /dev/null) or --json is set, skern never prompts and never
blocks on input — both answers default to "no". Pass --no-instructions to
state that opt-out explicitly instead of relying on the non-TTY default, or
--instructions to opt in.`,
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

	// Flag contradictions are usage errors; reject them before creating
	// anything on disk so a failed run leaves no trace.
	if err := validateInitFlags(opts); err != nil {
		return err
	}

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

// validateInitFlags rejects contradictory flag combinations (#104): the
// explicit opt-out cannot be combined with any opt-in. Values, not
// "changed" state, are compared, so `--instructions=false --no-instructions`
// is accepted as the consistent statement it is.
func validateInitFlags(opts runInitOpts) error {
	if !opts.noInstr {
		return nil
	}
	switch {
	case opts.writeInstr:
		return &ValidationError{Message: "--no-instructions cannot be combined with --instructions"}
	case opts.printInstr:
		return &ValidationError{Message: "--no-instructions cannot be combined with --print-instructions"}
	case len(opts.targetPaths) > 0:
		return &ValidationError{Message: "--no-instructions cannot be combined with --target"}
	case opts.toolForming:
		return &ValidationError{Message: "--no-instructions cannot be combined with --tool-forming-loop"}
	}
	return nil
}

// resolveInstructionChoices folds flag values + TTY interactivity into the
// final (writeInstructions, toolFormingLoop) decision.
//
// The contract: prompts appear only when no instruction flag was given,
// stdin is a terminal, and output is not JSON. Any instruction flag —
// including the explicit opt-out — silences both prompts, and a
// non-terminal stdin (pipe, file, /dev/null, CI) never prompts and never
// blocks, resolving both questions to "no".
func resolveInstructionChoices(cmd *cobra.Command, cc *CommandContext, opts runInitOpts) (bool, bool, error) {
	flags := cmd.Flags()

	// Explicit opt-out (#104): nothing is written and neither prompt runs,
	// regardless of TTY state (flag conflicts were rejected up front).
	if opts.noInstr {
		return false, false, nil
	}

	wantInstr := opts.writeInstr || opts.printInstr || len(opts.targetPaths) > 0
	wantToolForming := opts.toolForming

	// Any instruction flag means the caller chose flags over prompts; only
	// the no-flag, interactive case asks.
	flagged := flags.Changed("instructions") || flags.Changed("print-instructions") ||
		flags.Changed("target") || flags.Changed("tool-forming-loop")
	in := cmd.InOrStdin()
	canPrompt := !flagged && !cc.Printer.IsJSON() && isTerminalFn(in)
	if !canPrompt {
		return wantInstr, wantToolForming, nil
	}

	// Prompts go to stderr so they never collide with --print-instructions
	// output on stdout when scripts pipe init through.
	promptOut := cmd.ErrOrStderr()

	yes, err := promptYesNo(in, promptOut,
		"Append skern usage instructions to agent config files (AGENTS.md, CLAUDE.md, .claude/CLAUDE.md)?", false)
	if err != nil {
		return false, false, err
	}
	wantInstr = yes

	if wantInstr {
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

// isTerminalFn decides whether stdin is interactive. A package variable so
// tests can simulate a terminal without a pty.
var isTerminalFn = isTerminal

// isTerminal reports whether r is a *os.File attached to a terminal, using a
// real isatty check. A character-device test is not enough: /dev/null (and
// NUL on Windows) is a character device but not a terminal, and an installer
// running `skern init < /dev/null` must not see a prompt. Non-file readers
// (test injectees) are never terminals.
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
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
