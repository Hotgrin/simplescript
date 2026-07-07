# Instructions for AI coding agents

Two situations. Pick yours:

## A. You are WRITING hotgrin programs (.hot files)

Read `docs/ai-prompt-pack.md` first — it is the complete language spec written
for you, including a GOTCHAS list of the mistakes agents actually make.
hotgrin is from 2026 and is NOT in your training data; never guess syntax.

Verify your work with the CLI (never by eyeballing):

```
hotgrin check file.hot     # the Watcher: friendly, provable findings only
hotgrin run file.hot       # check + compile + execute
hotgrin test file.hot      # run the file's test blocks
hotgrin reveal file.hot    # show the generated Go
```

The Watcher's messages are precise and actionable — read them and fix exactly
what they say. If the Watcher is silent, the generated Go compiles.

## B. You are working ON the hotgrin compiler (this repo)

Go 1.22+. Build/test:

```
go build ./...
go test ./...            # must stay 100% green
gofmt -w . && go vet ./...
```

Architecture (a .hot file's journey): `internal/lexer` (line-oriented;
multi-word names bounded by reserved connector words, longest-match) →
`internal/parser` (recursive descent + precedence climbing) →
`internal/loader` (resolves `use`: local files, embedded std/*.hot, remote
git) → `internal/watcher` (bilingual findings; errors block execution) →
`internal/transpiler` (typed Go emission; whole program in one main.go, tests
in a _test.go) → the `hotgrin` CLI (`cmd/hotgrin`) shells out to the Go
toolchain. `internal/gobridge` parses `use go` blocks; `internal/units` is the
measurement table; `cmd/hotplayground` builds the WASM playground.

House rules (non-negotiable):
1. **Every change ships with a test.**
2. **The Watcher never raises a false alarm** — a finding must be provably
   correct. When unsure, don't flag.
3. **Users never see raw Go errors.** If the Watcher passes, the emitted Go
   must compile. Transpiler-detected problems go through friendly reporting.
4. **Docs claims must be true of the current version** — the cookbook's
   recipes are machine-verified; keep it that way (extract and run them).
5. gofmt clean, go vet clean, before any handoff.

Common tasks:
- New Watcher rule → `internal/watcher/watcher.go` (+ bilingual message pair
  + a test proving no false positive).
- New std library → `internal/loader/std/<name>.hot` (written in hotgrin,
  usually via `use go`; fallible funcs return `(T, error)`).
- New syntax → token in `internal/lexer/token.go`, AST node, parser case,
  transpiler emission, Watcher awareness, tests at every layer, then update
  `docs/language-reference.md` and `docs/ai-prompt-pack.md`.
