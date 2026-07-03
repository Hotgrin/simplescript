# hotgrin roadmap

Where hotgrin is going, in order. This is a living document — building real
programs keeps teaching us what matters next. Dates are deliberately absent:
each release ships when it's ready and tested, and small honest releases beat
big late ones.

## v0.3 — everyday completeness

The gaps you hit in your first week of real use:

- **`ask`** — interactive prompts (`ask "What is your name?" into name`), so
  programs can talk to the person running them.
- **`stop with error`** — exit a program cleanly with a message.
- **Number formatting** — a way to say `R2666.07` instead of
  `R2666.0740787821514`. The loan-calculator example is the motivating case.
- **List index by variable** — today `item 0 of scores` works but
  `item i of scores` does not (the words `item i` read as one name). Known
  gap; needs a small grammar decision.
- **Deeper type inference** — action-local variables passed as arguments to
  other actions don't yet drive parameter inference (top-level values do).
- **More cookbook recipes**, grown from real questions.

## v0.4 — the ecosystem door

- **Remote libraries** — `use tools from "github.com/someone/hotgrin-tools"`,
  fetched and cached, so sharing hotgrin code becomes one line. (Local
  libraries already work.)
- **The `use go` escape hatch** — embed Go directly for the rare gap, which
  makes the entire Go package ecosystem reachable from hotgrin.
- **Seed standard libraries** — small, first-party: `text` (formatting,
  casing), `data` (read/write files, CSV), `web` (fetch JSON from an API).
- **A library-authoring guide**, so community libraries feel native.

## v0.5 — the headline features

- **Units of measure** — `set weight to 129 kg`, with unit-aware maths and
  conversions. Designed but deliberately parked until it can be done properly.
- **Type annotations** — optional stated types; needs a syntax that survives
  multi-word names (open design question).
- **More Watcher rules** — always provable-only, never a false alarm.

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
