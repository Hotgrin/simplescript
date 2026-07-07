# hotgrin roadmap

Where hotgrin is going, in order. This is a living document — building real
programs keeps teaching us what matters next. Dates are deliberately absent:
each release ships when it's ready and tested, and small honest releases beat
big late ones.

## v0.3 — everyday completeness ✅ shipped

`ask` · `stop with error` · `rounded to` number formatting · variable list
indexing · deeper type inference — all landed in v0.3.0 (see the
[changelog](CHANGELOG.md)). Still growing continuously: cookbook recipes and
Watcher rules.

## v0.4 — the ecosystem door ✅ shipped

Remote GitHub libraries (fetched + cached, `@tag` pinning) · the `use go`
escape hatch (including fallible `(T, error)` functions) · first standard
libraries (`std/text`, `std/data`, `std/random`) · the
[library-authoring guide](docs/library-guide.md). A `std/web` library (fetch
JSON from APIs) moves to v0.5.

## v0.5 — the headline features ✅ shipped

**Units of measure** (`set weight to 129 kg`, cross-unit maths, `in`
conversions, dimension mistakes caught before running) · **`std/web`**
(`fetch text`, `json value`) · a new Watcher rule (using a give-nothing-back
action as a value). See the [changelog](CHANGELOG.md).

**Deferred deliberately:** optional type annotations — the syntax must survive
multi-word names, and that design question deserves its own careful session
rather than a rushed answer.

## Further out

- **Interpreter mode** — run `.hot` files directly with no Go installation.
  This is the biggest beginner-freedom unlock and the largest single piece of
  work.
- **The Assessor** — an on-demand deeper review ("assess my code") with a
  catalogue of deterministic checks beyond the Watcher's always-on set.
- **The AI Mentor** — optional, bring-your-own-key explanations and guidance
  layered on top; never required, never phoning home by default.
- **More message languages** — the Watcher speaks English and Afrikaans today;
  the mechanism is general, so isiZulu, isiXhosa, and others are one
  translation file away. Native-speaker contributions welcome.
- **Playground upgrades** — sharable links, and (with the interpreter) running
  programs fully in the browser.

## How this list is kept honest

Everything above the "Further out" line is scoped small enough to ship as one
tested release. When real use exposes a sharper need (it already has — three
v0.3 items came from writing the cookbook and the loan calculator), the list
gets reordered. If something here matters to you,
[open an issue](https://github.com/Hotgrin/hotgrin/issues) and say so — that's
literally how this list is prioritised.

## Contributing

Good first contributions: a cookbook recipe, a message-language translation, a
new Watcher rule (with its no-false-alarm proof), or an example project. See
[CONTRIBUTING.md](CONTRIBUTING.md). Two house rules: every change ships with a
test, and the Watcher never raises a false alarm.
