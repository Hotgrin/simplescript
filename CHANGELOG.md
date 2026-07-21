# Changelog

All notable changes to hotgrin are recorded here. This project follows
[Semantic Versioning](https://semver.org/) loosely while it is pre-1.0.

## [0.5.7] - 2026-07-21

### Added
- **Day Zero** (`docs/day-zero.md`): a five-minute, no-computer pre-lesson
  for people who have never coded before — recipes, grocery lists, and
  labelled jars mapped onto algorithms, loops, decisions, variables, and
  actions, plus a "break it on purpose" exercise showing why step order
  matters, and honest answers to the fears that stop beginners before they
  start (breaking the computer, forgetting everything, "too old to learn").
  Linked from the main README, `docs/getting-started.md`, and as lesson
  "00" at the top of `examples/learn/README.md`. First piece of a wider
  push to make hotgrin approachable for absolute beginners, not just
  people who already think of themselves as technical.

## [0.5.6] - 2026-07-09

### Fixed
- **Version const now matches its release tag.** The [0.5.5] work below was
  meant to ship as v0.5.4, but that tag number had already been used by a
  same-day housekeeping commit, so the release automation tagged it v0.5.5
  instead — while `cmd/hotgrin/main.go` still read `0.5.4`. `hotgrin
  version` was reporting one release behind. Fixed, and the entry below is
  relabeled to match what's actually in the v0.5.5 tag.
- This is exactly the failure mode house rule 6 exists to catch — the rule
  stays, the process just missed a step this once.

## [0.5.5] - 2026-07-09

### Added
- **The learn path**: `examples/learn/` — 27 numbered, heavily-commented
  lesson programs from first `say` to two mini-projects, every one
  machine-verified, with an indexed README.
- **Record field writes** — `set price of order to 249` now works (it had
  been documented since v0.1 but never implemented). Checked by the Watcher
  (unknown fields flagged; fallible values must be named first).

### Changed (breaking, alpha)
- std/text: `starts with` / `ends with` are now **`has prefix` /
  `has suffix`** — the old spoken names contained the reserved word `with`
  and were uncallable.
- std/random: `random up to` is now **`random below`** — `to` is reserved.
- Library guide now warns Go-bridge authors to pick names whose spoken form
  avoids connector words.

## [0.5.3] - 2026-07-08

### Added
- **Invoice Maker** flagship project (`examples/projects/invoice-maker/`):
  ~200 heavily-commented lines with a numbered section index, plus an
  extensive beginner TUTORIAL.md. Produces a text invoice and an HTML page.
- README: dynamic release badge (auto-updates with every tag), downloads
  badge, Go Report Card badge; Status section rewritten version-agnostically.

### Fixed
- Two sequential `try` blocks in one scope no longer collide on an internal
  error name.
- The `version` command now tracks releases again.

## [0.5.2] - 2026-07-08

### Added
- **Examples gallery**: nine categories (seo, api, math, finance, science,
  text-files, html, email, games), twenty machine-verified programs, indexed
  in `examples/README.md` and the README's "What can you build?" table.
- `std/text`: fallible `text between` (substring extraction).
- **String escapes**: `\n`, `\t`, `\"`, `\\` now work in text literals.
- **Rate division for units**: dividing across dimensions yields the plain
  base-unit rate (`5 km divided by 25 min` is metres per second).
- New Watcher rule: a fallible call nested inside a larger expression is
  caught kindly ("set it into a name first").

### Fixed
- Unused variables set from fallible calls inside `try` get their `_ =` guard.

## [0.5.1] - 2026-07-07

### Added
- **AI prompt pack** (`docs/ai-prompt-pack.md`): a complete language spec for
  AI assistants, with a growing GOTCHAS list; plus `llms.txt`, `AGENTS.md`,
  and `CLAUDE.md`.

### Fixed
- Repeated fallible sets of the same variable (including in nested tries) no
  longer emit duplicate Go declarations.

## [0.5.0] - 2026-07-07

The headline release: measurements are part of the language.

### Added
- **Units of measure** — `set weight to 129 kg`. Measurements print
  themselves (`129 kg`), join text naturally, convert with `in`
  (`weight in g`), and combine across units of the same dimension
  (`2 km plus 500 m` is `2.5 km`; `90 min is greater than 1 h` works).
  Mass, length, time, and volume ship first.
- **Dimension safety** — adding kg to metres, or a bare number to a
  measurement, is caught kindly *before* the program runs.
- **`std/web`** — `fetch text with <url>` and
  `json value with <doc>, "dotted.path"`, both fallible, verified against a
  live API.
- **New Watcher rule** — using an action that gives nothing back as a value
  is a provable mistake, now explained in both languages.

### Deferred
- Optional type annotations move out of v0.5: the syntax has to coexist with
  multi-word names, and that deserves an unhurried design.

## [0.4.0] - 2026-07-06

The ecosystem door: hotgrin programs can now use each other's code — and all
of Go's.

### Added
- **Remote libraries** — `use tools from "github.com/user/repo"` fetches a
  library with git, caches it under `~/.hotgrin/cache/`, and compiles it in.
  Subpaths and `@tag` version pinning work. Verified against a live repo.
- **The `use go` escape hatch** — embed Go between `use go` and `end go`.
  Declared functions become hotgrin actions (`shoutCase` reads as
  `shout case`); imports are merged; functions returning `(T, error)` are
  fallible and integrate with `try / if it fails`.
- **Standard libraries**, embedded in the binary:
  `std/text` (case, trim, replace, length), `std/data` (read/write files,
  fallible), `std/random` (random numbers).
- **[Library-authoring guide](docs/library-guide.md)**.
- Bare action names are zero-argument calls: `say lucky number`.

## [0.3.0] - 2026-07-04

Interactive programs, tidy numbers, and smarter inference — every item driven
by real use.

### Added
- **`ask`** — interactive prompts: `ask "What is your name?" into name`.
  Answers arrive as text, trimmed. (`hotgrin run` now passes stdin through.)
- **`stop with error "message"`** — end the program with a message on stderr
  and exit code 1. The Watcher knows code after it can never run.
- **`rounded to`** — number formatting at last: `payment rounded to 2` gives
  `2666.07`. Binds looser than arithmetic, so `a plus b rounded to 2` rounds
  the sum. (To round a call's result, wrap it in parentheses first.)
- **Variable list indexing** — `item i of scores` now works, including
  multi-word index names (`item current pos of scores`).

### Improved
- **Deeper type inference** — an action's local variables now drive parameter
  and return inference for the actions it calls. The loan-calculator example
  gets its `growth factor` helper action back thanks to this.
- **Cleaner failures** — `hotgrin run` executes a compiled binary directly, so
  a `stop with error` shows only *your* message (no more `exit status 1`
  noise), and your program's exit code is passed through faithfully.

### Fixed
- The `version` command now reports the right version.

## [0.2.0] - 2026-07-02

**The language has a new name: hotgrin** (formerly SimpleScript) — the language
that makes you grin. Same language, same promises, new identity.

### Changed
- Project renamed **SimpleScript → hotgrin**; repository is now
  `github.com/Hotgrin/hotgrin` (the old URL redirects).
- File extension **`.ss` → `.hot`** (`hello.ss` becomes `hello.hot`).
- The CLI is now the `hotgrin` command (`hotgrin run hello.hot`).
- Developer tools renamed: `hotlex`, `hotparse`, `hotrun`, `hotcheck`.
- Browser playground rebranded; now at `hotgrin.github.io/hotgrin/playground/`.

### Why
The old name collided with several existing projects; "hotgrin" is unique,
matches the maintainer's domain (hotgrin.com) and GitHub handle, and fits the
language's friendly identity.

## [0.1.0] - 2026-06-30

The first public release: a clean rebuild from a full specification. The whole
pipeline works end to end — near-English source compiles to a real native
program (or a Windows `.exe`).

### Language
- Variables (`set`), output (`say`), arithmetic, comparisons, and boolean logic.
- `plus` concatenates when either side is text; `divided by` always gives a decimal.
- Conditionals (`if` / `else`) and loops (`repeat N times`, `repeat while`,
  `repeat for each`).
- Actions with inferred parameter and return types (`action ... give back ...`).
- Records via `describe ... end describe`, with `field of record` access.
- Lists with `list of ...`, `put X into list`, `item N of list`, `count of list`.
- Concurrency: `at the same time` (with safe collection) and `start` (background).
- Error handling: `give back problem`, `try` / `if it fails`, and `the problem`.
- First-class tests: `test "..." ... end test` with `expect` assertions.
- Command-line inputs: `input name as text default "..."`, with an auto `--help`.
- Local libraries: `use "path"` merges another file's actions (whole-program).
- Unicode identifiers and strings; English/Afrikaans messages.

### Tooling
- `SimpleScript` command: `run`, `test`, `build` (`--windows` for a `.exe`),
  `check` (`--af` for Afrikaans), `reveal`, `help`, `version`.
- The **Watcher**: a deterministic checker that reports only provable problems
  (unknown names, wrong argument counts, divide-by-zero, unreachable code,
  duplicate definitions, unhandled fallible calls, and more) — with the iron
  rule that a flag always means a real issue.
- Developer tools: `sslex`, `ssparse`, `ssrun`, `sscheck`.

### Notes
- Requires Go 1.22+ installed; the toolchain is used under the hood.
- Known limitations: tiny standard library, no remote libraries yet, no
  interactive `ask` yet, cross-file line numbers in some messages are per-file.
