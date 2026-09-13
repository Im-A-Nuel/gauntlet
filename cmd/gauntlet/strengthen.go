package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/Im-A-Nuel/gauntlet/internal/bob"
	"github.com/Im-A-Nuel/gauntlet/internal/config"
	"github.com/Im-A-Nuel/gauntlet/internal/orchestrate"
	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

func newStrengthenCmd() *cobra.Command {
	var prepareOnly bool
	var bobCmd string
	var bobArgs []string

	cmd := &cobra.Command{
		Use:   "strengthen",
		Short: "Hand surviving mutants to Bob, then re-run and report the score delta",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := rootRepo()
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}

			before, ok, err := report.Latest(repoRoot)
			if err != nil {
				die(ExitConfigOrEnvError, "read latest run: %v", err)
			}
			if !ok {
				die(ExitConfigOrEnvError, "no runs found in %s; run `gauntlet run` first", report.RunsDir(repoRoot))
			}

			path, count, err := bob.WriteSurvivors(repoRoot, before, before.Threshold)
			if err != nil {
				die(ExitConfigOrEnvError, "write survivor handoff: %v", err)
			}
			fmt.Printf("wrote %s (%d surviving mutant(s))\n", path, count)

			if count == 0 {
				fmt.Println("nothing to strengthen")
				return nil
			}

			if prepareOnly {
				fmt.Println("--prepare-only: handoff prepared, no Bob invocation was made")
				return nil
			}

			cfg, err := config.Load(repoRoot)
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}

			release, err := bob.Acquire(repoRoot)
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}
			defer release()

			beforeSnap, err := bob.TakeSnapshot(repoRoot, cfg.Include, cfg.Exclude)
			if err != nil {
				die(ExitConfigOrEnvError, "snapshot before Bob invocation: %v", err)
			}

			if len(bobArgs) == 0 {
				bobArgs = bob.DefaultArgs(bob.SurvivorsPath)
			}
			adapter := bob.Adapter{Executable: bobCmd, Args: bobArgs}
			fmt.Printf("invoking Bob headlessly: %s %v\n", bobCmd, bobArgs)
			if err := adapter.Invoke(repoRoot, os.Stdout, os.Stderr); err != nil {
				fmt.Printf("Bob invocation failed: %v\n", err)
				fmt.Printf("the survivor handoff is still available at %s for manual use\n", path)
				die(ExitRunFailed, "strengthen aborted")
			}

			afterSnap, err := bob.TakeSnapshot(repoRoot, cfg.Include, cfg.Exclude)
			if err != nil {
				die(ExitConfigOrEnvError, "snapshot after Bob invocation: %v", err)
			}

			check := bob.Compare(beforeSnap, afterSnap)
			if !check.OK {
				fmt.Println("Bob's changes violated the strengthen-tests rules; NOT re-running or reporting a score:")
				for _, v := range check.Violations {
					fmt.Printf("  - %s\n", v)
				}
				die(ExitRunFailed, "strengthen aborted")
			}

			after, err := orchestrate.Run(orchestrate.Options{
				RepoRoot:        repoRoot,
				Trigger:         report.TriggerStrengthen,
				BaseRefOverride: before.BaseRef,
				Stdout:          os.Stdout,
				Stderr:          os.Stderr,
				SkipLock:        true,
			})
			if err != nil {
				die(ExitRunFailed, "re-run after strengthen failed: %v", err)
			}

			if !reflect.DeepEqual(before.ChangedFiles, after.ChangedFiles) {
				die(ExitRunFailed,
					"strengthen aborted: mutation scope changed during the Bob invocation (before: %v, after: %v); the delta would not be comparable",
					before.ChangedFiles, after.ChangedFiles)
			}

			fmt.Println("\nstrengthen delta:")
			fmt.Printf("  before: trustScore %s (run %s)\n", formatPct(before.Totals.TrustScore), before.RunID)
			fmt.Printf("  after:  trustScore %s (run %s)\n", formatPct(after.Totals.TrustScore), after.RunID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&prepareOnly, "prepare-only", false, "write the survivor handoff without invoking Bob")
	cmd.Flags().StringVar(&bobCmd, "bob-cmd", "bob", "Bob executable to invoke headlessly")
	cmd.Flags().StringArrayVar(&bobArgs, "bob-arg", nil, "argument to pass to the Bob executable (repeatable); default builds the documented `run --format json \"@survivors.md ...\"` invocation")
	return cmd
}
