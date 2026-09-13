package bob

import (
	"os"
	"path/filepath"
)

// SkillPath is where `gauntlet init` installs the strengthen-tests Skill.
const SkillPath = ".bob/skills/strengthen-tests/SKILL.md"

// skillBody is the exact content from docs/SCHEMA.md §4.
const skillBody = `---
name: strengthen-tests
description: Kill surviving mutants by writing sharper tests. Trigger when asked
  to strengthen tests or when .gauntlet/survivors.md is referenced.
---
Read .gauntlet/survivors.md. For each surviving mutant, write the smallest test
that fails on the mutated code and passes on the original. Rules:
- Do not modify source files; only test files.
- Do not delete or weaken existing assertions.
- Prefer boundary-value cases; name tests after the behavior, not the mutant id.
- Analyze survivors file by file; delegate per-file analysis to subagents when
  more than 3 files are affected.
- Run the full test suite before finishing; all tests must pass on original code.
`

// InstallSkill writes the strengthen-tests Skill file. Wholly Gauntlet-owned
// content (unlike settings.json, there is no user content to preserve), so
// this always (re)writes the canonical body.
func InstallSkill(repoRoot string) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(SkillPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(skillBody), 0o644)
}
