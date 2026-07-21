# hotgrin

**The language that makes you grin.** A programming language that reads like plain
English — and compiles to a real, fast program.

[![build](https://github.com/Hotgrin/hotgrin/actions/workflows/test.yml/badge.svg)](https://github.com/Hotgrin/hotgrin/actions/workflows/test.yml)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev/dl/)
[![release](https://img.shields.io/github/v/release/Hotgrin/hotgrin?color=B8374F)](https://github.com/Hotgrin/hotgrin/releases/latest)
[![downloads](https://img.shields.io/github/downloads/Hotgrin/hotgrin/total)](https://github.com/Hotgrin/hotgrin/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/hotgrin/hotgrin)](https://goreportcard.com/report/github.com/hotgrin/hotgrin)
![status: alpha](https://img.shields.io/badge/status-alpha-orange.svg)

> hotgrin is for people who want to *make* something without first learning
> punctuation-heavy syntax. You write near-English; it transpiles to **Go** and
> builds a real native executable (or a Windows `.exe`). Small language, real
> performance, and a checker that explains mistakes kindly — in English **or**
> Afrikaans.

**[▶ Try it in your browser — nothing to install](https://hotgrin.github.io/hotgrin/playground/)**

---

## See it

```
action discount with price, percent
    give back price minus (price times percent divided by 100)
end action

set total to 897
say "Total: R" plus total
say "After 10% off: R" plus discount with total, 10
```

```
Total: R897
After 10% off: R807.3
```

No semicolons, no curly braces, no `public static void`. Names can have spaces
(`cart total` is one name), and `divided by` always gives you a decimal — because
that is what a beginner expects.

**Curious what it becomes?** `hotgrin reveal hello.hot` shows you the Go:

```
say "Hello, world"
```

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world")
}
```

Nothing is hidden. The whole pipeline is: your words → tokens → a tree → Go → a
real program.

---

## The Watcher — mistakes explained kindly

hotgrin ships with an always-on checker called the **Watcher**. Its iron
rule: *if it flags something, it is genuinely wrong* — it never raises a false
alarm. And it speaks your language:

```
say discountt        # a typo
set x to 10 divided by 0
```

```
error  line 1: there is no value called 'discountt' here — is it a typo, or did you forget to set it?
error  line 2: this divides by zero, which a computer cannot do
```

Prefer Afrikaans? Add `--af`:

```
fout   reël 1: daar is geen waarde genaamd 'discountt' hier nie — is dit 'n tikfout, of het jy vergeet om dit te stel?
fout   reël 2: hierdie deel deur nul, wat 'n rekenaar nie kan doen nie
```

`run` and `build` run the Watcher first, so a beginner sees friendly hotgrin
messages — **never a raw Go error**.

---

## Quick start

You need **[Go 1.22+](https://go.dev/dl/)** (hotgrin uses the Go toolchain
under the hood). Then one line installs the command:

```bash
go install github.com/hotgrin/hotgrin/cmd/hotgrin@latest
```

Make a file called `hello.hot` containing `say "Hello, world"`, and:

```bash
hotgrin run    hello.hot            # run it
hotgrin check  hello.hot            # check for problems (--af for Afrikaans)
hotgrin build  --windows hello.hot  # make a Windows .exe you can share
hotgrin reveal hello.hot            # show the Go it becomes
```

Prebuilt binaries are on the
[releases page](https://github.com/Hotgrin/hotgrin/releases/latest); full
options (including Docker and building from source) are in
**[Getting started](docs/getting-started.md)**.

**New to programming?** Start with the [gentle tutorial](docs/tutorial.md),
then raid the **[cookbook](docs/cookbook.md)** — 21 copy-paste recipes that all
run as shown. **Coming from another language?** The
[language reference](docs/language-reference.md) covers every construct and its
Go mapping. **Want a real example?** The
[invoice maker](examples/projects/invoice-maker/) is a complete worked
project — heavily commented, with a full beginner walkthrough and tests. And the **[roadmap](ROADMAP.md)** shows where hotgrin is
going.

---

## A quick tour

**Variables, math, and text.** `plus` joins text or adds numbers; it converts for
you.

```
set learner to "Adriaan"
set score to 82
say learner plus " scored " plus score
```

**Decisions and loops** read the way you would say them.

```
if score is at least 50
    say "passed"
else
    say "try again"
end if

repeat 3 times
    say "hi"
end repeat
```

**Records** describe a thing; read fields with `of`.

```
describe order
    item is "Wireless mouse"
    price is 299
end describe

say item of order
```

**Handle things that can fail** with `try` / `if it fails`. The Watcher makes sure
a call that can fail is always handled.

```
action safe divide with a, b
    if b is 0
        give back problem "cannot divide by zero"
    end if
    give back a divided by b
end action

try
    set answer to safe divide with 10, 0
    say answer
if it fails
    say "Could not: " plus the problem
end try
```

**Tests are part of the language.** `hotgrin test` runs them.

```
test "ten percent off 100 is 90"
    expect discount with 100, 10 to be 90
end test
```

**Do things at the same time**, safely (hotgrin guards the shared list for
you):

```
at the same time
    do download report
    do download invoices
end at the same time
```

**Take command-line inputs** (you get a `--help` for free), and **reuse code**
from other files:

```
input name as text default "world"
use "lib/textutils"
say greet with name
```

```bash
hotgrin run greet.hot --name AJ
```

---

## Status

**Early alpha** (see the [current release](https://github.com/Hotgrin/hotgrin/releases/latest)
and the [changelog](CHANGELOG.md)). The language works end to end and is well
tested — 90+ tests across eight packages, CI on every commit, and every
cookbook recipe and example is machine-verified — but it is young. Expect
rough edges, and please
[open an issue](https://github.com/Hotgrin/hotgrin/issues) when you find one.

**On the roadmap:** an interpreter mode (no Go install at all), optional type
annotations, more message languages, and the Assessor / AI-mentor layers —
the full list lives in [ROADMAP.md](ROADMAP.md).

## Learn it in 27 small steps

Never coded before — not even a little? Start with
**[Day Zero](docs/day-zero.md)**, a five-minute, no-computer page that shows
you've already been thinking like a programmer your whole life. Then come
back here.

[`examples/learn/`](examples/learn/) is a numbered lesson path — every file a
tiny, runnable, heavily-commented program, from your first `say "hello"` to a
playable word game. Read, run, change, break, fix: that cycle is the course.

## What can you build?

Anything — but start where your problem lives. Each category has verified,
copy-paste examples:

| Category | Examples |
|---|---|
| SEO & websites | [examples/seo](examples/seo/) — page checks, title lengths, robots.txt |
| APIs & JSON | [examples/api](examples/api/) — query real APIs, uptime checker |
| Maths & formulas | [examples/math](examples/math/) — percentages, compound interest, stats |
| Money & investing | [examples/finance](examples/finance/) — savings goals, ROI, a [loan calculator](examples/projects/loan-calculator/) and the [flagship invoice maker](examples/projects/invoice-maker/) |
| Science & units | [examples/science](examples/science/) — speed, conversions, recipe scaling |
| Text & files | [examples/text-files](examples/text-files/) — word counts, find/replace, a journal |
| Web pages | [examples/html](examples/html/) — generate real HTML from data |
| Email | [examples/email](examples/email/) — SMTP through the escape hatch |
| Games & learning | [examples/games](examples/games/) — quiz with scores, dice battle |

More categories (AI, WordPress, Shopify, databases, PDFs) land as their
enablers ship — see the [roadmap](ROADMAP.md).

## hotgrin + AI assistants

hotgrin is newer than most AI models' training data. To have Claude, ChatGPT,
Gemini, or any other assistant write correct hotgrin, paste
**[the AI prompt pack](docs/ai-prompt-pack.md)** into the chat first — it's a
complete spec written for AIs, and every rule in it is machine-verified.
There's also an [llms.txt](llms.txt) and [AGENTS.md](AGENTS.md) for tools that
look for them.

## Contributing

Contributions are very welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). Two house
rules: every change ships with a test, and the Watcher never raises a false alarm.

## Licence

MIT — see [LICENSE](LICENSE).
