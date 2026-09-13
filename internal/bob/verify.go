package bob

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/bmatcuk/doublestar/v4"
)

// fileState is one tracked file's content fingerprint at a point in time.
type fileState struct {
	Hash           string
	AssertionCount int
}

// Snapshot is a before/after picture of the source and test files a
// strengthen invocation is allowed (and not allowed) to touch.
type Snapshot struct {
	Source map[string]fileState // path -> state, for files matching include globs
	Test   map[string]fileState // path -> state, for files matching exclude globs (the test-file patterns)
}

var assertionPattern = regexp.MustCompile(`\bexpect\s*\(`)

// TakeSnapshot walks repoRoot and fingerprints every file matching include
// (treated as "source", must not change) or exclude (treated as "test",
// tracked for deletions/weakened assertions). A file is only ever counted
// once, under include if it matches include and not exclude, else under
// exclude if it matches exclude.
func TakeSnapshot(repoRoot string, include, exclude []string) (Snapshot, error) {
	snap := Snapshot{Source: map[string]fileState{}, Test: map[string]fileState{}}
	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		isTest := matchesAny(exclude, rel)
		isSource := !isTest && matchesAny(include, rel)
		if !isTest && !isSource {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		state := fileState{Hash: hex.EncodeToString(sum[:])}
		if isTest {
			state.AssertionCount = len(assertionPattern.FindAll(data, -1))
			snap.Test[rel] = state
		} else {
			snap.Source[rel] = state
		}
		return nil
	})
	if err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}

func matchesAny(patterns []string, path string) bool {
	for _, p := range patterns {
		if ok, _ := doublestar.Match(p, path); ok {
			return true
		}
	}
	return false
}

// CheckResult reports whether a strengthen invocation respected the
// strengthen-tests Skill's rules: source untouched, no test file deleted,
// no test file's assertion count reduced. This is a mechanical check
// (content hash + `expect(` count), not full AST analysis — it catches
// deletion and gross weakening, not a semantically-equivalent-but-narrower
// rewrite of an assertion. Documented as a known limitation, not oversold.
type CheckResult struct {
	OK         bool
	Violations []string
}

// Compare diffs a before/after Snapshot pair and reports any violation of
// the strengthen-tests Skill's rules.
func Compare(before, after Snapshot) CheckResult {
	var violations []string

	paths := make([]string, 0, len(before.Source))
	for p := range before.Source {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		b := before.Source[p]
		a, stillExists := after.Source[p]
		switch {
		case !stillExists:
			violations = append(violations, "source file deleted: "+p)
		case a.Hash != b.Hash:
			violations = append(violations, "source file modified: "+p)
		}
	}
	for p := range after.Source {
		if _, existedBefore := before.Source[p]; !existedBefore {
			violations = append(violations, "new source file added: "+p)
		}
	}

	testPaths := make([]string, 0, len(before.Test))
	for p := range before.Test {
		testPaths = append(testPaths, p)
	}
	sort.Strings(testPaths)
	for _, p := range testPaths {
		b := before.Test[p]
		a, stillExists := after.Test[p]
		if !stillExists {
			violations = append(violations, "test file deleted: "+p)
			continue
		}
		if a.AssertionCount < b.AssertionCount {
			violations = append(violations, formatAssertionDrop(p, b.AssertionCount, a.AssertionCount))
		}
	}

	return CheckResult{OK: len(violations) == 0, Violations: violations}
}

func formatAssertionDrop(path string, before, after int) string {
	return "assertion count dropped in " + path + ": " + strconv.Itoa(before) + " -> " + strconv.Itoa(after)
}
