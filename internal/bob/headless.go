package bob

import (
	"fmt"
	"io"
	"os/exec"
)

// Adapter configures the headless Bob invocation. Executable and Args are
// fully caller-supplied (CLI flags, not docs/SCHEMA.md config — that schema
// is frozen and owned elsewhere) precisely because IBM Bob is not installed
// in this build: exact flags are documented at
// https://bob.ibm.com/docs/shell/getting-started/start-bobshell-non-interactive
// (`bob run [options] [prompt...]`, e.g. --format/--max-turns/--mode/--workspace)
// but have not been exercised against a live binary. A configurable adapter
// lets a real Bob install be wired in without touching Go code.
type Adapter struct {
	Executable string
	Args       []string
}

// DefaultArgs builds the documented invocation shape for the strengthen
// loop: `bob run --format json "<prompt>"`. The prompt references the
// survivors file with Bob's documented `@path` file-reference syntax so Bob
// reads it, which — per the strengthen-tests Skill's own description
// ("Trigger when ... .gauntlet/survivors.md is referenced") — is what is
// expected to invoke the Skill; no CLI flag for explicitly selecting a Skill
// in headless mode is documented.
func DefaultArgs(survivorsRelPath string) []string {
	prompt := fmt.Sprintf(
		"@%s Use the strengthen-tests skill to write the smallest tests that kill each surviving mutant listed here, following that skill's rules exactly.",
		survivorsRelPath,
	)
	return []string{"run", "--format", "json", prompt}
}

// Invoke runs the configured Bob executable with Args as an argument array
// (exec.Command, no shell) — the same "no user-interpolated shell strings"
// discipline docs/ARCHITECTURE.md requires of the Stop hook applies here.
// cwd is repoRoot so `@relative/path` references resolve against the target
// project.
func (a Adapter) Invoke(repoRoot string, stdout, stderr io.Writer) error {
	if a.Executable == "" {
		return fmt.Errorf("no Bob executable configured (pass --bob-cmd, or use --prepare-only to skip invocation)")
	}
	cmd := exec.Command(a.Executable, a.Args...)
	cmd.Dir = repoRoot
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("headless Bob invocation failed (executable=%q): %w", a.Executable, err)
	}
	return nil
}
