package bob

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockPath guards the mutation-run critical section against concurrent or
// re-entrant execution in the same repo — e.g. `strengthen` invoking Bob
// headlessly, one of Bob's own subagents finishing and re-firing the Stop
// hook (-> a nested `gauntlet run --changed --trigger hook`) while the outer
// strengthen loop is mid-run.
const LockPath = ".gauntlet/.lock"

// Acquire takes an exclusive, non-blocking lock for repoRoot. It fails fast
// (does not wait/retry) so a re-entrant invocation gets a clear, immediate
// error rather than hanging or racing the original run.
//
// This is a simple existence-lock, not a liveness check: a lock left behind
// by a crashed process must be removed by hand (documented in
// docs/CLAUDE_PROGRESS.md as a known limitation, not silently auto-cleared,
// since guessing at a stale lock risks two runs clobbering the same Stryker
// report file).
func Acquire(repoRoot string) (release func(), err error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(LockPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf(
				"another Gauntlet run appears to be in progress in this repo (%s exists); "+
					"if a previous run crashed, delete that file and retry", LockPath,
			)
		}
		return nil, fmt.Errorf("create lock file: %w", err)
	}
	fmt.Fprintf(f, "pid=%d\nstarted=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	f.Close()

	return func() { os.Remove(path) }, nil
}
