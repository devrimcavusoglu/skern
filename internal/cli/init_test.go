package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devrimcavusoglu/skern/internal/cli/instructions"
	"github.com/devrimcavusoglu/skern/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTempCwd switches the test process to a fresh temp dir and restores
// the original on cleanup. Init writes relative paths so tests must run
// from a clean cwd.
func withTempCwd(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return dir
}

func TestInit(t *testing.T) {
	dir := withTempCwd(t)

	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Initialized")

	info, err := os.Stat(filepath.Join(dir, ".skern", "skills"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInit_Idempotent(t *testing.T) {
	withTempCwd(t)

	_, err := runCmd(t, nil, "init")
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Already initialized")
}

func TestInit_JSON(t *testing.T) {
	withTempCwd(t)

	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.True(t, result.Created)
	assert.NotEmpty(t, result.Path)
	assert.Nil(t, result.Instructions, "Instructions should be omitted when not opted in")
}

func TestInit_JSON_AlreadyExists(t *testing.T) {
	withTempCwd(t)

	_, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.False(t, result.Created)
}

// --- Instruction-snippet flag flows ---

func TestInit_Instructions_NoTargetsFound(t *testing.T) {
	withTempCwd(t)

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	assert.False(t, result.Instructions.ToolForming)
	assert.Empty(t, result.Instructions.Targets)
	assert.Empty(t, result.Instructions.Writes)
	assert.False(t, result.Instructions.Printed)
}

func TestInit_Instructions_DiscoversAndWrites(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, filepath.Join("AGENTS.md"), result.Instructions.Writes[0].Path)
	assert.Equal(t, "appended", result.Instructions.Writes[0].Action)

	body, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "# Project")
	assert.Contains(t, bodyStr, instructions.StartMarker)
	assert.Contains(t, bodyStr, "Skern (skill management)")
	assert.NotContains(t, bodyStr, "Tool-forming loop", "tool-forming should be off by default")
}

func TestInit_Instructions_WithToolFormingLoop(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(""), 0o644))

	_, err := runCmd(t, nil, "init", "--instructions", "--tool-forming-loop", "--json")
	require.NoError(t, err)

	body, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "Tool-forming loop")
	assert.Contains(t, string(body), "skern skill search")
}

func TestInit_Instructions_TargetCreatesNewFile(t *testing.T) {
	dir := withTempCwd(t)
	target := filepath.Join("subdir", "MY_AGENT.md")

	_, err := runCmd(t, nil, "init", "--target", target, "--json")
	require.NoError(t, err)

	body, err := os.ReadFile(filepath.Join(dir, target))
	require.NoError(t, err)
	assert.Contains(t, string(body), instructions.StartMarker)
	assert.Contains(t, string(body), instructions.EndMarker)
}

func TestInit_Instructions_TargetSkipsAutoDiscovery(t *testing.T) {
	dir := withTempCwd(t)
	// Create both a candidate file and an explicit target. Only the target
	// should be touched.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("orig"), 0o644))
	target := "EXPLICIT.md"
	require.NoError(t, os.WriteFile(filepath.Join(dir, target), []byte("explicit-orig"), 0o644))

	out, err := runCmd(t, nil, "init", "--target", target, "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, target, result.Instructions.Writes[0].Path)

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "orig", string(agents), "auto-discovery candidate must not be touched when --target is set")
}

func TestInit_Instructions_PrintToStdout(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("untouched"), 0o644))

	out, err := runCmd(t, nil, "init", "--print-instructions", "--tool-forming-loop")
	require.NoError(t, err)

	assert.Contains(t, out, instructions.StartMarker)
	assert.Contains(t, out, "Tool-forming loop")

	// Discovered file must NOT be modified in print mode.
	body, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "untouched", string(body))
}

func TestInit_Instructions_IdempotentReRun(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# proj\n"), 0o644))

	_, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	first, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)

	out, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	second, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, string(first), string(second), "second run must be a no-op when content unchanged")

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	require.NotNil(t, result.Instructions)
	require.Len(t, result.Instructions.Writes, 1)
	assert.Equal(t, "unchanged", result.Instructions.Writes[0].Action)
}

func TestInit_Instructions_UpdatesBlockOnToggleChange(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(""), 0o644))

	// Round 1: opt out of tool-forming.
	_, err := runCmd(t, nil, "init", "--instructions", "--json")
	require.NoError(t, err)

	round1, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.NotContains(t, string(round1), "Tool-forming loop")

	// Round 2: opt in. The block should be replaced in place, not appended.
	_, err = runCmd(t, nil, "init", "--instructions", "--tool-forming-loop", "--json")
	require.NoError(t, err)

	round2, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	round2Str := string(round2)
	assert.Contains(t, round2Str, "Tool-forming loop")
	assert.Equal(t, 1, strings.Count(round2Str, instructions.StartMarker),
		"block must be replaced, not duplicated, on re-run")
}

func TestInit_Instructions_NoPromptInJSONMode(t *testing.T) {
	withTempCwd(t)
	// Run --json with no instruction flags. Should not prompt (no stdin
	// input given, and with --json the writer never reaches the prompt
	// branch). Result.Instructions must be nil.
	out, err := runCmd(t, nil, "init", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Nil(t, result.Instructions)
}

// #104: --no-instructions is the explicit non-interactive opt-out. It must
// write nothing, prompt nothing, and report no instructions result — even
// when an instruction file is present to be discovered.
func TestInit_NoInstructions_WritesNothing(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))

	out, err := runCmd(t, nil, "init", "--no-instructions", "--json")
	require.NoError(t, err)

	var result output.InitResult
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Nil(t, result.Instructions)
	assert.True(t, result.Created)

	got, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Project\n", string(got), "AGENTS.md must be untouched")
	assert.NotContains(t, out, "Append skern usage instructions", "no prompt text may be emitted")
}

func TestInit_NoInstructions_TextMode(t *testing.T) {
	withTempCwd(t)

	out, err := runCmd(t, nil, "init", "--no-instructions")
	require.NoError(t, err)
	assert.Contains(t, out, "Initialized")
	assert.NotContains(t, out, "instruction")
}

// Opting out and opting in at once is a contradiction: validation error
// (exit 2), nothing written — not even .skern/.
func TestInit_NoInstructions_ConflictsWithOptIn(t *testing.T) {
	for _, optIn := range [][]string{
		{"--instructions"},
		{"--print-instructions"},
		{"--target", "AGENTS.md"},
		{"--tool-forming-loop"},
	} {
		t.Run(optIn[0], func(t *testing.T) {
			dir := withTempCwd(t)
			args := append([]string{"init", "--no-instructions"}, optIn...)
			_, err := runCmd(t, nil, args...)
			require.Error(t, err)
			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Contains(t, err.Error(), "--no-instructions cannot be combined with "+optIn[0])
			_, statErr := os.Stat(filepath.Join(dir, "AGENTS.md"))
			assert.True(t, os.IsNotExist(statErr), "nothing should be written on a flag conflict")
			_, statErr = os.Stat(filepath.Join(dir, ".skern"))
			assert.True(t, os.IsNotExist(statErr), ".skern/ must not be created when flags are rejected")
		})
	}

	// Values are compared, not "changed" state: an explicit false is consistent.
	t.Run("explicit false is allowed", func(t *testing.T) {
		dir := withTempCwd(t)
		_, err := runCmd(t, nil, "init", "--no-instructions", "--instructions=false", "--tool-forming-loop=false")
		require.NoError(t, err)
		_, statErr := os.Stat(filepath.Join(dir, ".skern", "skills"))
		require.NoError(t, statErr)
	})
}

// lineReader hands out at most one line per Read, the way a terminal in
// canonical mode does, so successive bufio.Scanners over the same stdin each
// see exactly one answer (a plain strings.Reader would be slurped whole by
// the first scanner).
type lineReader struct{ rest string }

func (l *lineReader) Read(p []byte) (int, error) {
	if l.rest == "" {
		return 0, io.EOF
	}
	n := strings.IndexByte(l.rest, '\n') + 1
	if n == 0 {
		n = len(l.rest)
	}
	n = copy(p, l.rest[:n])
	l.rest = l.rest[n:]
	return n, nil
}

// runInitWithTTY runs init with a simulated terminal on stdin feeding `input`.
// It returns the combined stdout+stderr, the error, and whatever of `input`
// was left unread (so tests can prove no prompt consumed it).
func runInitWithTTY(t *testing.T, input string, args ...string) (out string, rest string, err error) {
	t.Helper()
	orig := isTerminalFn
	isTerminalFn = func(io.Reader) bool { return true }
	t.Cleanup(func() { isTerminalFn = orig })

	in := &lineReader{rest: input}
	cmd := newRootCmd(nil)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetIn(in)
	cmd.SetArgs(append([]string{"init"}, args...))
	err = cmd.Execute()
	return buf.String(), in.rest, err
}

const promptInstr = "Append skern usage instructions"
const promptToolForming = "Include tool-forming-loop section"

// With a terminal and no flags, both prompts fire and are honored — this
// proves the TTY simulation reaches the prompt path, so the negative tests
// below are meaningful.
func TestInit_TTY_NoFlags_Prompts(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))

	out, rest, err := runInitWithTTY(t, "y\ny\n")
	require.NoError(t, err)
	assert.Contains(t, out, promptInstr)
	assert.Contains(t, out, promptToolForming)
	assert.Empty(t, rest, "both answers should have been consumed")
	got, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	assert.Contains(t, string(got), instructions.StartMarker)
	assert.Contains(t, string(got), "Tool-forming loop")

	// Declining the first question skips the second.
	withTempCwd(t)
	out, rest, err = runInitWithTTY(t, "n\ny\n")
	require.NoError(t, err)
	assert.Contains(t, out, promptInstr)
	assert.NotContains(t, out, promptToolForming)
	assert.Equal(t, "y\n", rest)
}

// --no-instructions on a real terminal: no prompt text, stdin untouched,
// nothing written. This is the guarantee #104 asks for.
func TestInit_TTY_NoInstructions_NeverPrompts(t *testing.T) {
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))

	out, rest, err := runInitWithTTY(t, "y\ny\n", "--no-instructions")
	require.NoError(t, err)
	assert.NotContains(t, out, promptInstr)
	assert.NotContains(t, out, promptToolForming)
	assert.Equal(t, "y\ny\n", rest, "stdin must not be read")
	got, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	assert.Equal(t, "# Project\n", string(got))
}

// Any instruction flag silences both prompts, including the second one: a
// setup script running `skern init --instructions` in a terminal must not
// hang on the tool-forming question.
func TestInit_TTY_InstructionFlagsSilencePrompts(t *testing.T) {
	for _, args := range [][]string{
		{"--instructions"},
		{"--print-instructions"},
		{"--target", "AGENTS.md"},
		{"--tool-forming-loop"},
		{"--instructions=false"},
		{"--json"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			dir := withTempCwd(t)
			require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))
			out, rest, err := runInitWithTTY(t, "y\ny\n", args...)
			require.NoError(t, err)
			assert.NotContains(t, out, promptInstr)
			assert.NotContains(t, out, promptToolForming)
			assert.Equal(t, "y\ny\n", rest, "stdin must not be read")
		})
	}

	// And --instructions alone writes the snippet without the tool-forming
	// section (the unasked question defaults to no).
	dir := withTempCwd(t)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Project\n"), 0o644))
	_, _, err := runInitWithTTY(t, "", "--instructions")
	require.NoError(t, err)
	got, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	assert.Contains(t, string(got), instructions.StartMarker)
	assert.NotContains(t, string(got), "Tool-forming loop")
}

// The real isTerminal must say "not a terminal" for the non-TTY inputs an
// installer is likely to hand us: /dev/null, a pipe, a regular file, a
// non-file reader. (A real pty is not available under go test.)
func TestIsTerminal_NonTTYInputs(t *testing.T) {
	devnull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer func() { _ = devnull.Close() }()
	assert.False(t, isTerminal(devnull), os.DevNull+" is a character device but not a terminal")

	f, err := os.CreateTemp(t.TempDir(), "stdin")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	assert.False(t, isTerminal(f))

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = r.Close(); _ = w.Close() }()
	assert.False(t, isTerminal(r))

	assert.False(t, isTerminal(strings.NewReader("")))
}
