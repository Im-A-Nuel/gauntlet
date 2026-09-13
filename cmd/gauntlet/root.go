// Command gauntlet is the CLI entrypoint: adversarial mutation-testing
// verification for AI-written code (docs/README.md, docs/ARCHITECTURE.md).
package main

import (
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
func die(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		die(ExitConfigOrEnvError, "Error: %v", err)
	}
}
