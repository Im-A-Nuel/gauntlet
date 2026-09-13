package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RunsDirName is the directory (relative to a repo root) holding run artifacts.
const RunsDirName = ".gauntlet/runs"

// NewRunID builds a "<unix-ms>-<short-sha>" run ID. Millisecond precision
// (docs/SCHEMA.md §2: "runId may include millisecond timestamps to avoid
// collisions") keeps rapid successive runs (e.g. run -> strengthen -> run)
// from colliding on the same second.
func NewRunID(now time.Time, headSha string) string {
	short := headSha
	if len(short) > 7 {
		short = short[:7]
	}
	if short == "" {
		short = "nogit"
	}
	return fmt.Sprintf("%d-%s", now.UnixMilli(), short)
}

// RunsDir returns the absolute runs directory for a repo root.
func RunsDir(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(RunsDirName))
}

// pathFor returns the artifact path for a given run ID.
func pathFor(repoRoot, runID string) string {
	return filepath.Join(RunsDir(repoRoot), runID+".json")
}

// Write atomically persists an artifact: encode to a temp file in the same
// directory, then rename over the destination. A crash or concurrent reader
// never observes a partially-written or corrupt artifact (docs/SCHEMA.md §2:
// "Artifact writes use a temporary file plus rename").
func Write(repoRoot string, a Artifact) (string, error) {
	dir := RunsDir(repoRoot)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create runs dir: %w", err)
	}
	dest := pathFor(repoRoot, a.RunID)

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode artifact: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-"+a.RunID+"-*")
	if err != nil {
		return "", fmt.Errorf("create temp artifact: %w", err)
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup: on the success path the rename below removes the
	// temp name entirely, so this Remove is a no-op that only fires on error.
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write temp artifact: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp artifact: %w", err)
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		return "", fmt.Errorf("rename temp artifact into place: %w", err)
	}
	return dest, nil
}

// Read loads one artifact by run ID.
func Read(repoRoot, runID string) (Artifact, error) {
	var a Artifact
	data, err := os.ReadFile(pathFor(repoRoot, runID))
	if err != nil {
		return a, err
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return a, fmt.Errorf("parse artifact %s: %w", runID, err)
	}
	return a, nil
}

// List returns every run ID present in the runs directory, oldest first
// (NewRunID's unix-ms prefix sorts lexicographically = chronologically).
func List(repoRoot string) ([]string, error) {
	dir := RunsDir(repoRoot)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".tmp-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(ids)
	return ids, nil
}

// Latest returns the most recently written artifact, or ok=false if none exist.
func Latest(repoRoot string) (Artifact, bool, error) {
	ids, err := List(repoRoot)
	if err != nil {
		return Artifact{}, false, err
	}
	if len(ids) == 0 {
		return Artifact{}, false, nil
	}
	a, err := Read(repoRoot, ids[len(ids)-1])
	if err != nil {
		return Artifact{}, false, err
	}
	return a, true, nil
}
