// Package scope resolves the set of source files a Gauntlet run should
// mutate: the union of the base-ref-to-HEAD diff, working tree/index
// changes, and untracked files, filtered by config include/exclude globs
// (docs/SCHEMA.md §2, "MVP implementation clarifications").
package scope

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Result is the resolved changed-file scope for one run.
type Result struct {
	HeadSha string
	Files   []string // repo-root-relative (relative to the --repo target), sorted, deduplicated
}

func git(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}

// HeadSha resolves the current HEAD commit of repoRoot.
func HeadSha(repoRoot string) (string, error) {
	out, err := git(repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve HEAD (is %s a git repo with at least one commit?): %w", repoRoot, err)
	}
	return strings.TrimSpace(out), nil
}

// IsGitRepo reports whether repoRoot is inside a Git working tree.
func IsGitRepo(repoRoot string) bool {
	out, err := git(repoRoot, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// parseNameStatusZ parses `git diff -z --name-status` output: NUL-delimited
// tokens rather than newline-delimited lines, so filenames containing
// spaces, unicode, or literal newlines are never mis-split (git only quotes
// "unusual" characters in the default newline-terminated format; -z
// disables that quoting and guarantees one token per NUL). Pure deletions
// are dropped; renames/copies (R###/C### followed by two paths) resolve to
// the new path. Everything else keeps its path as-is.
func parseNameStatusZ(out string) []string {
	tokens := strings.Split(out, "\x00")
	// git terminates every record with NUL, so the final split element is "".
	if len(tokens) > 0 && tokens[len(tokens)-1] == "" {
		tokens = tokens[:len(tokens)-1]
	}

	var files []string
	for i := 0; i < len(tokens); i++ {
		status := tokens[i]
		if status == "" {
			continue
		}
		switch {
		case strings.HasPrefix(status, "D"):
			i++ // consume the deleted path, drop it
		case strings.HasPrefix(status, "R"), strings.HasPrefix(status, "C"):
			// R###/C### is followed by old-path, new-path.
			if i+2 < len(tokens) {
				files = append(files, tokens[i+2])
			}
			i += 2
		default:
			if i+1 < len(tokens) {
				files = append(files, tokens[i+1])
			}
			i++
		}
	}
	return files
}

// parseLinesZ splits NUL-delimited git output (e.g. `ls-files -z`) into
// non-empty tokens.
func parseLinesZ(out string) []string {
	var files []string
	for _, tok := range strings.Split(out, "\x00") {
		if tok != "" {
			files = append(files, tok)
		}
	}
	return files
}

// matchesAny reports whether path matches any of the given doublestar globs.
func matchesAny(patterns []string, path string) bool {
	for _, p := range patterns {
		if ok, _ := doublestar.Match(p, path); ok {
			return true
		}
	}
	return false
}

// Resolve computes the changed-file scope for repoRoot against baseRef,
// filtered by include/exclude globs. All git invocations are scoped to
// repoRoot with `--relative -- .` so a repoRoot that is a subdirectory of a
// larger repository (e.g. demo-repo/ inside the Gauntlet project repo)
// yields paths relative to repoRoot itself, not the outer repo's top level.
//
// A repo with no matching changes returns an explicit empty Files slice
// (never nil-vs-empty ambiguity, never a whole-repo fallback).
func Resolve(repoRoot, baseRef string, include, exclude []string) (Result, error) {
	if !IsGitRepo(repoRoot) {
		return Result{}, fmt.Errorf("%s is not inside a git working tree", repoRoot)
	}
	headSha, err := HeadSha(repoRoot)
	if err != nil {
		return Result{}, err
	}

	candidates := map[string]bool{}

	// 1. Committed changes since the merge-base with baseRef.
	if baseOut, err := git(repoRoot, "diff", "-z", "--relative", "--merge-base", baseRef, "HEAD", "--name-status", "--", "."); err == nil {
		for _, f := range parseNameStatusZ(baseOut) {
			candidates[f] = true
		}
	} else if !isUnknownRevision(err) {
		return Result{}, fmt.Errorf("resolve diff against baseRef %q: %w", baseRef, err)
	}
	// An unknown/unrelated baseRef (e.g. a fresh repo where baseRef has no
	// common history yet) is not a hard failure: worktree + untracked scope
	// still applies below.

	// 2. Working tree + index changes vs HEAD.
	if wtOut, err := git(repoRoot, "diff", "-z", "--relative", "HEAD", "--name-status", "--", "."); err == nil {
		for _, f := range parseNameStatusZ(wtOut) {
			candidates[f] = true
		}
	} else {
		return Result{}, fmt.Errorf("resolve worktree diff: %w", err)
	}

	// 3. Untracked files (respecting .gitignore).
	if untOut, err := git(repoRoot, "ls-files", "-z", "--others", "--exclude-standard", "--", "."); err == nil {
		for _, f := range parseLinesZ(untOut) {
			candidates[f] = true
		}
	} else {
		return Result{}, fmt.Errorf("resolve untracked files: %w", err)
	}

	files := make([]string, 0, len(candidates))
	for f := range candidates {
		fwd := strings.ReplaceAll(f, "\\", "/")
		if matchesAny(include, fwd) && !matchesAny(exclude, fwd) {
			files = append(files, fwd)
		}
	}
	sort.Strings(files)

	if err := rejectEscapingSymlinks(repoRoot, files); err != nil {
		return Result{}, err
	}

	return Result{HeadSha: headSha, Files: files}, nil
}

// rejectEscapingSymlinks fails closed if any candidate file is a symlink
// (at any path component) that resolves outside repoRoot. Stryker mutates
// candidate files in place; a symlink pointing outside the project could
// otherwise let a mutation run write to an arbitrary path on disk.
func rejectEscapingSymlinks(repoRoot string, relFiles []string) error {
	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root %q: %w", repoRoot, err)
	}
	for _, rel := range relFiles {
		full := filepath.Join(repoRoot, filepath.FromSlash(rel))
		if info, err := os.Lstat(full); err != nil || info.Mode()&fs.ModeSymlink == 0 {
			continue // missing or not a symlink itself; still need target check below
		}
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil {
			return fmt.Errorf("resolve symlink %q: %w", rel, err)
		}
		relToRoot, err := filepath.Rel(root, resolved)
		if err != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
			return fmt.Errorf("candidate file %q is a symlink that resolves outside the project root; refusing to mutate it", rel)
		}
	}
	return nil
}

func isUnknownRevision(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unknown revision") || strings.Contains(msg, "bad revision") || strings.Contains(msg, "ambiguous argument")
}
