# Project state

The single source of truth for "where are we right now." Update this at the
end of any session that changes it — a stale README section is a bug; a
stale `PROJECT-STATE.md` is a much worse one, because it's the thing meant
to prevent us re-discovering our own history by accident.

This file answers: what's actually shipped (verified against the real
GitHub remote, not a chat summary), what's mid-flight, what's decided, and
what's next. If a chat and this file disagree, **this file wins** — it's
checked against source; a chat summary might describe a sandbox that never
got pushed.

Last verified against remote: **2026-07-21**, commit `0d58b19`, tag `v0.5.6`.

---

## Shipped, confirmed on the real remote

- **v0.5.6** is the latest tag. Core language is stable: units of measure,
  `std/web`, built-in testing, the Watcher, bilingual (English/Afrikaans)
  errors, remote GitHub libraries, `use go` escape hatch.
- 27-lesson learn path (`examples/learn/`), 21-recipe cookbook, browser
  playground, AI prompt pack + `llms.txt` + `.cursorrules` +
  `copilot-instructions.md`.
- Four homepage showcase examples + the 283-line `site-report.hot` flagship
  program (commit `0d58b19`).
- Record-field-writes (`set price of order to 249`), std renames that
  removed reserved-word collisions (`starts with` → `has prefix`, etc).

## Mid-flight — needs a decision or re-verification, not assumed done

- **Day Zero** (`docs/day-zero.md`) — written, live-tested (every code
  snippet actually run through the built binary), wired into README,
  `getting-started.md`, and `examples/learn/README.md`. Version bump to
  **v0.5.7** prepared. **Not yet pushed to GitHub as of this writing** —
  applying the patch and pushing is the one open step.
- **Record-instantiation compiler fix** (records failing to instantiate
  inside `test`/`try` blocks) — worked on in two separate sessions
  (2026-07-16 and 2026-07-20). One session's summary claims it was fixed,
  tested, and tagged v0.5.7; the actual remote shows no such tag and no
  such commit. **Treat this bug as still open until someone re-verifies
  against a fresh clone and actually pushes.** Don't assume it's fixed
  because a past chat said so.
- **Known engine bugs from the 2026-07-16 audit**, drafted as GitHub issues
  but not confirmed filed (checking was blocked by API rate limits when
  this file was written — reconfirm next session):
  1. CRITICAL — `at the same time` hangs the compiler on any nested
     statement other than a literal `do` line.
  2. HIGH — `list of nothing` collapses the whole list's type to `any`,
     breaking field access on collected records.
  3. HIGH — record instantiation inside actions/test blocks (see above —
     possibly the same root cause already partially diagnosed).
  4. MEDIUM — `give back` inside `try` breaks the build while the Watcher
     says "All good."
  5. MEDIUM — writing to an unknown record field passes the Watcher, then
     fails at Go compile.
  6. LOW — two genuine Watcher false alarms on variables that are clearly
     used.
- **gobug** (`github.com/Hotgrin/gobug`) — a separate, real side project
  (friendly Go error explainer, Wails desktop app, v0.2.0 tagged). CI fix
  prepared (release permissions + macOS runner pinned off Tahoe/LC_UUID
  issue) but not yet applied/pushed as of this writing.

## Decided — house rules, don't relitigate

- **One source of truth for content.** Beginner-education copy lives in the
  repo (`docs/`), not duplicated into WordPress. hotgrin.com is a thin
  front door that *links to* the repo, never a second copy of it.
- **Day Zero is the canonical absolute-beginner entry point**, folding in
  the strongest idea from an earlier, independent draft called "Class 1:
  You Already Know How To Do This" (the "break it on purpose" exercise).
  That earlier draft is retired — don't resurrect it as a separate page.
- **Numbered comment system is now a formal teaching convention**, not
  just a style the Invoice Maker example happened to use: `[1]`–`[N]`
  section index at the top of a file, matching numbered dividers through
  the body. Next beginner lesson (working title "Class 2" / "Day One")
  teaches this explicitly, plus a "before you type anything" planning
  ritual and a comment-first-code-second workflow.
- **Every shipped change gets a version bump, a changelog entry, and a
  pushed tag.** No partial states on the remote.
- **Live verification only.** Nothing ships without being built, `go vet`,
  `go test`, and (for `.hot` code) actually run through the compiled
  binary — never eyeballed or assumed correct from reading source.

## Beginner-education initiative — status

Sequence agreed: **Day Zero → living glossary → first micro-lessons → AI
Mentor**, with community-building ("Study Stoep") and the hotgrin.com
homepage redesign running alongside whenever there's room. Day Zero is
built; homepage redesign was proposed but not started; glossary and AI
Mentor not started.

## Marketing / launch — status

dev.to article published, low traction (reported ~1 view, no external
distribution). Account restricted from r/learnprogramming and
r/ProgrammingLanguages for AI-assisted development disclosure — a
community-mood issue, not a verdict on hotgrin. ZATech Slack post was
drafted; whether it was actually posted is unconfirmed — **reconfirm with
AJ before assuming it's live.** Show HN not yet attempted. Current stance:
grow the real thing first (Day Zero, real beginner users), let institutions
and wider marketing follow evidence rather than chase them early.

## Chat hygiene

Six prior chats existed in this project as of 2026-07-21 covering: initial
build (v0.1–v0.5.4), a stale container-reset checkpoint (safe to delete),
a documentation/bug-fix audit + dev.to/Reddit fallout, an off-topic
portfolio-strategy detour (belongs in a different project, not here), the
Class-1/Day-Zero duplicate (retired, see above), and a compiler-bug session
that ended mid-verification. Going forward: **update this file instead of
relying on chat summaries to reconstruct state.** If a new session needs
history a search can't answer, that's a sign this file needs a better
entry, not a sign to go digging through old chats.

## Next up

1. Push the Day Zero patch (v0.5.7) — the one blocking step to close that
   loop.
2. Push the gobug CI fix.
3. Re-verify the record-instantiation bug against a fresh clone; ship it
   properly (as v0.5.8, since v0.5.7 is now Day Zero) if still broken.
4. Build the next beginner lesson: project planning, folder/file thinking,
   the numbered-comment convention, "before you type anything" ritual.
5. hotgrin.com homepage redesign (simple, plain-language nav, one button).
