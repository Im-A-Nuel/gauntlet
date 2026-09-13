---
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
