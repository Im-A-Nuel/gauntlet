package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

func newReportCmd() *cobra.Command {
	var runID string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Print the latest (or a specific) run artifact as a terminal summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := rootRepo()
			if err != nil {
				die(ExitConfigOrEnvError, "%v", err)
			}

			var a report.Artifact
			if runID != "" {
				a, err = report.Read(repoRoot, runID)
				if err != nil {
					die(ExitConfigOrEnvError, "read run %s: %v", runID, err)
				}
			} else {
				var ok bool
				a, ok, err = report.Latest(repoRoot)
				if err != nil {
					die(ExitConfigOrEnvError, "read latest run: %v", err)
				}
				if !ok {
					die(ExitConfigOrEnvError, "no runs found in %s; run `gauntlet run` first", report.RunsDir(repoRoot))
				}
			}

			printRunSummary(a)
			fmt.Println()
			for _, f := range a.Files {
				fmt.Printf("  %-40s trustScore: %s  mutants: %d\n", f.Path, formatPct(f.TrustScore), len(f.Mutants))
				for _, m := range f.Mutants {
					fmt.Printf("    [%s] line %-4d %-22s %s\n", m.ID, m.Line, m.Mutator, m.Status)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&runID, "run-id", "", "print a specific run instead of the latest")
	return cmd
}
