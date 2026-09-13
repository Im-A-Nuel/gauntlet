// Package bob integrates Gauntlet with IBM Bob 2.0's public extension
// points: the workspace settings.json lifecycle hook, a strengthen-tests
// Skill, and a headless "bob run" invocation for the self-healing loop.
//
// IBM Bob is not installed in this build environment. Everything in this
// package that talks to a real `bob` binary (Invoke, and the hook actually
// firing) is implemented against IBM's public documentation but has not been
// exercised against a live Bob install. Treat it as "implemented to spec,
// unverified in practice" until run against real Bob — see
// docs/CLAUDE_PROGRESS.md for exactly what was and wasn't tested.
//
// Verified against https://bob.ibm.com/docs/ide/configuration/lifecycle-hooks
// (fetched during this build): the session-end event is named "Stop", not
// "agentStop" as drafted in docs/SCHEMA.md §5 — that doc predates the
// verification pass and should be corrected by whoever owns docs/SCHEMA.md.
// The verified hook schema wraps each event's entries as
// `{"hooks": {"<Event>": [{"hooks": [{"type":"command","command":...,"timeout":...}]}]}}`,
// with "matcher" documented only for PreToolUse/PostToolUse (a plain "Stop"
// group has no matcher). Settings live at workspace-scope `.bob/settings.json`
// or global `~/.bob/settings/settings.json`; this package only ever touches
// the workspace file.
package bob

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SettingsPath is the workspace-scope Bob settings file, relative to a repo root.
const SettingsPath = ".bob/settings.json"

// StopEvent is the verified lifecycle event name fired when a Bob agent
// session ends (was drafted as "agentStop" before verification).
const StopEvent = "Stop"

// HookCommand is what Gauntlet installs into the Stop hook.
const HookCommand = "gauntlet run --changed --trigger hook"

// hookTimeoutSeconds overrides Bob's documented 10s hook default. A mutation
// run can legitimately take up to the ~3 minute budget in docs/REQUIREMENTS.md;
// leaving the default would kill the hook mid-run on anything but a trivial
// change set.
const hookTimeoutSeconds = 300

type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type hookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

// InstallHook patches (never overwrites) repoRoot's .bob/settings.json so the
// Stop event runs HookCommand. It is idempotent: calling it again when the
// command is already present is a no-op. All other keys in settings.json —
// including other hook events and any other Bob settings — are preserved
// byte-for-byte via json.RawMessage passthrough.
func InstallHook(repoRoot string) (changed bool, err error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(SettingsPath))

	root := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &root); err != nil {
			return false, fmt.Errorf("parse existing %s: %w (refusing to overwrite a file Gauntlet cannot understand)", SettingsPath, err)
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", SettingsPath, err)
	}

	hooksMap := map[string]json.RawMessage{}
	if raw, ok := root["hooks"]; ok {
		if err := json.Unmarshal(raw, &hooksMap); err != nil {
			return false, fmt.Errorf("parse %s .hooks: %w", SettingsPath, err)
		}
	}

	var stopGroups []hookGroup
	if raw, ok := hooksMap[StopEvent]; ok {
		if err := json.Unmarshal(raw, &stopGroups); err != nil {
			return false, fmt.Errorf("parse %s .hooks.%s: %w", SettingsPath, StopEvent, err)
		}
	}

	for _, g := range stopGroups {
		for _, h := range g.Hooks {
			if h.Type == "command" && h.Command == HookCommand {
				return false, nil // already installed
			}
		}
	}

	stopGroups = append(stopGroups, hookGroup{
		Hooks: []hookEntry{{Type: "command", Command: HookCommand, Timeout: hookTimeoutSeconds}},
	})

	stopRaw, err := json.Marshal(stopGroups)
	if err != nil {
		return false, fmt.Errorf("encode %s hooks: %w", StopEvent, err)
	}
	hooksMap[StopEvent] = stopRaw

	hooksRaw, err := json.Marshal(hooksMap)
	if err != nil {
		return false, fmt.Errorf("encode hooks: %w", err)
	}
	root["hooks"] = hooksRaw

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return false, fmt.Errorf("encode %s: %w", SettingsPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create %s dir: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", SettingsPath, err)
	}
	return true, nil
}
