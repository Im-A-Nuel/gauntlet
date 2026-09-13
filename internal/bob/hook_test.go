package bob

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallHookOnFreshRepo(t *testing.T) {
	dir := t.TempDir()
	changed, err := InstallHook(dir)
	if err != nil {
		t.Fatalf("InstallHook: %v", err)
	}
	if !changed {
		t.Fatalf("changed = false on fresh install, want true")
	}

	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(SettingsPath)))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var root struct {
		Hooks struct {
			Stop []hookGroup `json:"Stop"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	if len(root.Hooks.Stop) != 1 || len(root.Hooks.Stop[0].Hooks) != 1 {
		t.Fatalf("Stop hooks = %+v, want exactly one command entry", root.Hooks.Stop)
	}
	h := root.Hooks.Stop[0].Hooks[0]
	if h.Type != "command" || h.Command != HookCommand {
		t.Fatalf("hook entry = %+v, want type=command command=%q", h, HookCommand)
	}
}

func TestInstallHookIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if _, err := InstallHook(dir); err != nil {
		t.Fatalf("first InstallHook: %v", err)
	}
	changed, err := InstallHook(dir)
	if err != nil {
		t.Fatalf("second InstallHook: %v", err)
	}
	if changed {
		t.Fatalf("changed = true on second install, want false (idempotent)")
	}
}

func TestInstallHookPreservesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".bob")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `{
  "hooks": {
    "PreToolUse": [{"matcher": "^write_file$", "hooks": [{"type": "command", "command": "sh check.sh"}]}]
  },
  "someOtherUserSetting": {"nested": true}
}`
	path := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := InstallHook(dir); err != nil {
		t.Fatalf("InstallHook: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse patched settings: %v", err)
	}
	if _, ok := root["someOtherUserSetting"]; !ok {
		t.Fatalf("someOtherUserSetting was dropped by InstallHook")
	}

	var hooksMap map[string]json.RawMessage
	if err := json.Unmarshal(root["hooks"], &hooksMap); err != nil {
		t.Fatalf("parse hooks: %v", err)
	}
	var preToolUse []hookGroup
	if err := json.Unmarshal(hooksMap["PreToolUse"], &preToolUse); err != nil {
		t.Fatalf("parse PreToolUse: %v", err)
	}
	if len(preToolUse) != 1 || preToolUse[0].Matcher != "^write_file$" {
		t.Fatalf("PreToolUse hooks = %+v, want the original untouched entry", preToolUse)
	}

	var stop []hookGroup
	if err := json.Unmarshal(hooksMap[StopEvent], &stop); err != nil {
		t.Fatalf("parse Stop: %v", err)
	}
	if len(stop) != 1 || stop[0].Hooks[0].Command != HookCommand {
		t.Fatalf("Stop hooks = %+v, want Gauntlet's hook added", stop)
	}
}

func TestInstallHookRejectsUnparsableSettings(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".bob")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallHook(dir); err == nil {
		t.Fatalf("InstallHook on unparsable settings.json = nil error, want refusal")
	}
}
