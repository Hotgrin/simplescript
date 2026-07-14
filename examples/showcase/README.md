# Showcase examples

Small-to-medium example programs written for the hotgrin.com homepage —
kept here too so they stay machine-verified alongside the rest of the repo.

- `pack-weight-checker.hot` (~30 lines) — units of measure, lists, repeat for each
- `safe-greeting.hot` (~20 lines) — std/data, the fallible/try pattern, a test
- `class-grade-report.hot` (~65 lines) — records, lists of records, actions, tests

All three: `hotgrin check`, `hotgrin run`, and `hotgrin test` (where applicable)
clean against the current release. Re-verify after any lexer/parser/transpiler
change, same as the cookbook.
- `site-report.hot` (~280 lines) — the flagship: std/web, concurrency
  (`at the same time`), records, fallible calls, a tested primitive-only
  logic core, and three documented workarounds for real language
  limitations found while writing it (see the file's header comment and
  the accompanying bug reports).
