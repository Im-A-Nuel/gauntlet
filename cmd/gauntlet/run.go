package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Im-A-Nuel/gauntlet/internal/orchestrate"
	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

var validTriggers = map[string]report.Trigger{
	"manual":     report.TriggerManual,
	"hook":       report.TriggerHook,
	"strengthen": report.TriggerStrengthen,
	"ci":         report.TriggerCI,
}

func newRunCmd() *cobra.Command {
	var changed bool
	var trigger string
	var baseRef string
	var testCommand string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a scoped mutation pass against changed files and write a run artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = changed // scope is always changed-files only in this MVP; the flag is accepted for interface compatibility with docs/README.md's quickstart.

			trig, ok := validTriggers[trigger]
			if !ok {
				die(ExitConfigOrEnvError, "invalid --trigger %q (want one of: manual, hook, strengthen, ci)", trigger)
			}
			repoRoot, err := rootRepo()
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}

			a, err := orchestrate.Run(orchestrate.Options{
				RepoRoot:            repoRoot,
				Trigger:             trig,
				BaseRefOverride:     baseRef,
				TestCommandOverride: testCommand,
				Stdout:              os.Stdout,
				Stderr:              os.Stderr,
			})
			if err != nil {
				var runFailed *orchestrate.RunFailedError
				if errors.As(err, &runFailed) {
					die(ExitRunFailed, "run failed: %v", runFailed)
				}
				die(ExitConfigOrEnvError, "%v", err)
			}

			printRunSummary(a)
			return nil
		},
	}
	cmd.Flags().BoolVar(&changed, "changed", true, "scope the run to changed files (always true in this MVP; no whole-repo mode exists)")
	cmd.Flags().StringVar(&trigger, "trigger", "manual", "what triggered this run: manual, hook, strengthen, ci")
	cmd.Flags().StringVar(&baseRef, "base-ref", "", "override config.yaml baseRef")
	cmd.Flags().StringVar(&testCommand, "test-command", "", "override config.yaml testCommand (e.g. to run an alternate/strengthened test suite)")
	return cmd
}

func printRunSummary(a report.Artifact) {
	fmt.Printf("\nrun %s (%s)\n", a.RunID, a.Trigger)
	fmt.Printf("  base: %s  head: %s\n", a.BaseRef, shortSha(a.HeadSha))
	fmt.Printf("  changed files: %d\n", len(a.ChangedFiles))
	for _, f := range a.ChangedFiles {
		fmt.Printf("    %s\n", f)
	}
	fmt.Printf("  mutants: %d  killed: %d  survived: %d  timeout: %d  noCoverage: %d\n",
		a.Totals.Mutants, a.Totals.Killed, a.Totals.Survived, a.Totals.Timeout, a.Totals.NoCoverage)
	fmt.Printf("  trustScore: %s  lineCoverage: %s  duration: %dms\n",
		formatPct(a.Totals.TrustScore), formatPct(a.Totals.LineCoverage), a.Totals.DurationMs)
	if a.ComparedTo != nil {
		fmt.Printf("  comparedTo: %s\n", *a.ComparedTo)
	}
}

func formatPct(v *float64) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%.1f%%", *v)
}

func shortSha(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return "(none)"
	}
	return sha
}
