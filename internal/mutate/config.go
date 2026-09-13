// Package mutate wraps StrykerJS: it generates a scoped stryker.conf.json,
// runs `npx stryker` as a child process, and normalizes Stryker's own
// mutation.json report into Gauntlet's report.Artifact shape. Gauntlet never
// implements mutation operators itself (docs/AGENTS.md).
package mutate

import (
	"encoding/json"
	"fmt"
	"os"
)

// strykerConfig is the subset of stryker.conf.json fields Gauntlet controls.
// The "command" test runner shells out to the project's own test command per
// mutant, so Gauntlet needs no language-specific runner plugin.
type strykerConfig struct {
	Mutate        []string      `json:"mutate"`
	TestRunner    string        `json:"testRunner"`
	CommandRunner commandRunner `json:"commandRunner"`
	Concurrency   int           `json:"concurrency"`
	Reporters     []string      `json:"reporters"`
	// A generated config always scopes "mutate" to an explicit, non-empty
	// file list (internal/scope); GenerateConfig refuses to write a config
	// with an empty list so a caller cannot accidentally trigger a
	// whole-repo run by omitting the scope.
}

type commandRunner struct {
	Command string `json:"command"`
}

// GenerateConfig writes a scoped stryker.conf.json to a fresh temp file and
// returns its absolute path plus a cleanup func. files must be non-empty and
// must already be filtered/relative to the target repo root (internal/scope
// output) — this function does not itself apply include/exclude filtering,
// it only refuses to run mutation-free.
func GenerateConfig(files []string, testCommand string, concurrency int) (path string, cleanup func(), err error) {
	if len(files) == 0 {
		return "", nil, fmt.Errorf("refusing to generate a Stryker config with an empty mutate list (no whole-repo fallback)")
	}
	if testCommand == "" {
		return "", nil, fmt.Errorf("testCommand must not be empty")
	}
	if concurrency < 1 {
		concurrency = 1
	}

	cfg := strykerConfig{
		Mutate:        files,
		TestRunner:    "command",
		CommandRunner: commandRunner{Command: testCommand},
		Concurrency:   concurrency,
		Reporters:     []string{"json"},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("encode stryker config: %w", err)
	}

	f, err := os.CreateTemp("", "gauntlet-stryker-*.json")
	if err != nil {
		return "", nil, fmt.Errorf("create temp stryker config: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("write temp stryker config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("close temp stryker config: %w", err)
	}
	name := f.Name()
	return name, func() { os.Remove(name) }, nil
}
