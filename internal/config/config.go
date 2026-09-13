// Package config loads and validates .gauntlet/config.yaml (docs/SCHEMA.md §1).
package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Path (relative to a repo root) of the Gauntlet config file.
const Path = ".gauntlet/config.yaml"

// Config mirrors docs/SCHEMA.md §1.
type Config struct {
	BaseRef     string   `yaml:"baseRef"`
	TestCommand string   `yaml:"testCommand"`
	MinScore    float64  `yaml:"minScore"`
	Concurrency int      `yaml:"concurrency"`
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
}

// Default returns the MVP default config (matches the example in
// docs/SCHEMA.md §1), used by `gauntlet init` and as a fallback when no
// config file exists yet.
func Default() Config {
	return Config{
		BaseRef:     "main",
		TestCommand: "npm test",
		MinScore:    80,
		Concurrency: 4,
		Include:     []string{"src/**/*.ts"},
		Exclude:     []string{"**/*.spec.ts", "**/*.test.ts"},
	}
}

// FilePath returns the absolute config path for a repo root.
func FilePath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(Path))
}

// Load reads and validates the config at repoRoot. A missing file is a
// config/environment error (exit code 3 per docs/SCHEMA.md §7): callers must
// run `gauntlet init` first rather than silently falling back to defaults,
// so a run's scope is always the config the user actually wrote.
func Load(repoRoot string) (Config, error) {
	data, err := os.ReadFile(FilePath(repoRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("%s not found: run `gauntlet init` first", Path)
		}
		return Config{}, fmt.Errorf("read %s: %w", Path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", Path, err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid %s: %w", Path, err)
	}
	return c, nil
}

// Validate rejects configs that would make a run ambiguous or unsafe:
// an empty test command, non-positive concurrency/minScore out of [0,100],
// or an include list wide enough to defeat changed-file scoping (empty
// include is rejected rather than silently treated as "everything").
func (c Config) Validate() error {
	if c.BaseRef == "" {
		return fmt.Errorf("baseRef must not be empty")
	}
	if c.TestCommand == "" {
		return fmt.Errorf("testCommand must not be empty")
	}
	if math.IsNaN(c.MinScore) || math.IsInf(c.MinScore, 0) || c.MinScore < 0 || c.MinScore > 100 {
		return fmt.Errorf("minScore must be a finite number between 0 and 100, got %v", c.MinScore)
	}
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1, got %d", c.Concurrency)
	}
	if len(c.Include) == 0 {
		return fmt.Errorf("include must list at least one glob (no implicit whole-repo scope)")
	}
	return nil
}

// Write serializes c to repoRoot's config file. Used by `gauntlet init`;
// callers must check for an existing file themselves (init must not clobber
// user edits).
func Write(repoRoot string, c Config) error {
	dir := filepath.Dir(FilePath(repoRoot))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s dir: %w", dir, err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(FilePath(repoRoot), data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", Path, err)
	}
	return nil
}

// Exists reports whether a config file is already present at repoRoot.
func Exists(repoRoot string) bool {
	_, err := os.Stat(FilePath(repoRoot))
	return err == nil
}
