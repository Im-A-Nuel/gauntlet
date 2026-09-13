package mutate

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGenerateConfigRefusesEmptyFileList(t *testing.T) {
	if _, _, err := GenerateConfig(nil, "npm test", 4); err == nil {
		t.Fatalf("GenerateConfig(nil, ...) = nil error, want refusal (no whole-repo fallback)")
	}
}

func TestGenerateConfigRefusesEmptyTestCommand(t *testing.T) {
	if _, _, err := GenerateConfig([]string{"src/a.ts"}, "", 4); err == nil {
		t.Fatalf("GenerateConfig with empty testCommand = nil error, want error")
	}
}

func TestGenerateConfigWritesScopedMutateList(t *testing.T) {
	files := []string{"src/a.ts", "src/b.ts"}
	path, cleanup, err := GenerateConfig(files, "npm test", 4)
	if err != nil {
		t.Fatalf("GenerateConfig: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	var cfg strykerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse generated config: %v", err)
	}
	if len(cfg.Mutate) != 2 || cfg.Mutate[0] != "src/a.ts" || cfg.Mutate[1] != "src/b.ts" {
		t.Fatalf("cfg.Mutate = %v, want %v", cfg.Mutate, files)
	}
	if cfg.TestRunner != "command" || cfg.CommandRunner.Command != "npm test" {
		t.Fatalf("cfg runner = %q/%q, want command/npm test", cfg.TestRunner, cfg.CommandRunner.Command)
	}
	if cfg.Concurrency != 4 {
		t.Fatalf("cfg.Concurrency = %d, want 4", cfg.Concurrency)
	}

	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("cleanup() did not remove temp config file %s", path)
	}
}

func TestGenerateConfigClampsNonPositiveConcurrency(t *testing.T) {
	path, cleanup, err := GenerateConfig([]string{"src/a.ts"}, "npm test", 0)
	if err != nil {
		t.Fatalf("GenerateConfig: %v", err)
	}
	defer cleanup()
	data, _ := os.ReadFile(path)
	var cfg strykerConfig
	json.Unmarshal(data, &cfg)
	if cfg.Concurrency != 1 {
		t.Fatalf("cfg.Concurrency = %d, want clamped to 1", cfg.Concurrency)
	}
}
