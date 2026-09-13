package main

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"

	"github.com/Im-A-Nuel/gauntlet/internal/gate"
	"github.com/Im-A-Nuel/gauntlet/internal/report"
	"github.com/Im-A-Nuel/gauntlet/internal/scope"
)

func newGateCmd() *cobra.Command {
	var minScore float64
	var runID string

	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Exit non-zero if the latest run artifact's Trust Score is below the threshold",
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

			threshold := a.Threshold
			if cmd.Flags().Changed("min-score") {
				threshold = minScore
			}
			if math.IsNaN(threshold) || math.IsInf(threshold, 0) || threshold < 0 || threshold > 100 {
				die(ExitConfigOrEnvError, "--min-score %v is invalid: must be a finite number between 0 and 100", threshold)
			}

			currentHead, err := scope.HeadSha(repoRoot)
			if err != nil {
				die(ExitConfigOrEnvError, "resolve current HEAD for freshness check: %v", err)
			}

			result := gate.Evaluate(a, threshold, currentHead)
			fmt.Println(result.Message)
			if !result.Pass {
				die(ExitGateFailed, "gate: FAIL (%s)", result.Reason)
			}
			fmt.Println("gate: PASS")
			return nil
		},
	}
	cmd.Flags().Float64Var(&minScore, "min-score", 0, "minimum Trust Score required (default: the threshold recorded in the run artifact)")
	cmd.Flags().StringVar(&runID, "run-id", "", "gate a specific run instead of the latest")
	return cmd
}
