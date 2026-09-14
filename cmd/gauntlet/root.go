// Command gauntlet is the CLI entrypoint: adversarial mutation-testing
// verification for AI-written code (docs/README.md, docs/ARCHITECTURE.md).
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Exit codes per docs/SCHEMA.md §7.
const (
	ExitOK               = 0
	ExitGateFailed       = 1
	ExitRunFailed        = 2
	ExitConfigOrEnvError = 3
)

var repoFlag string

func rootRepo() (string, error) {
	abs, err := filepath.Abs(repoFlag)
	if err != nil {
		return "", fmt.Errorf("resolve --repo %q: %w", repoFlag, err)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("--repo %q is not a directory", repoFlag)
	}
	return abs, nil
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "gauntlet",
		Short:         "Adversarial mutation-testing verification for AI-written code",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&repoFlag, "repo", ".", "target repository root")

	root.AddCommand(newInitCmd())
	root.AddCommand(newRunCmd())
	root.AddCommand(newReportCmd())
	root.AddCommand(newGateCmd())
	root.AddCommand(newStrengthenCmd())
	return root
}

// die prints a message to stderr and exits with code — used for the
// documented exit-code contract (docs/SCHEMA.md §7), which plain cobra
// error-returns can't express (cobra always exits 1 on a returned error).
//
// die calls os.Exit directly, which runs no deferred functions anywhere in
// the process. It is therefore only safe to call from a command handler at a
// point where that handler has no pending `defer` (e.g. a held lock) still
// on its stack. A handler that acquires a resource needing cleanup (see
// cmd/gauntlet/strengthen.go) must return an *exitError instead of calling
// die after acquiring it, so the deferred release runs during the normal
// Go return path before main maps the error to an exit code.
func die(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

// exitError pairs an error with the process exit code it should map to
// (docs/SCHEMA.md §7). Returning one from a command's RunE — rather than
// calling die — lets any pending `defer` in that handler run to completion
// as part of normal Go function return before main ever calls os.Exit.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

// exitErrorf builds an *exitError with a formatted message, for command
// handlers that need a specific exit code (docs/SCHEMA.md §7) but must
// return through pending defers rather than calling die directly.
func exitErrorf(code int, format string, args ...any) error {
	return &exitError{code: code, err: fmt.Errorf(format, args...)}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		code := ExitConfigOrEnvError
		var ee *exitError
		if errors.As(err, &ee) {
			code = ee.code
		}
		die(code, "Error: %v", err)
	}
}
