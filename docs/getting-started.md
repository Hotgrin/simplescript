# Getting started with hotgrin

Three ways in, from easiest to most hands-on. Pick one.

> **Never written a line of code before?** Read
> **[Day Zero](day-zero.md)** first — five minutes, no computer needed,
> no installs. It'll show you that you already think like a programmer.
> Then come back here.

## Option 1 — Nothing to install (30 seconds)

Open the **[browser playground](https://hotgrin.github.io/hotgrin/playground/)**.
Type on the left, watch real Go appear on the right, and let the Watcher check
your work — in English or Afrikaans. Nothing leaves your browser.

The playground shows and checks your code; to *run* programs on your computer,
use option 2 or 3.

## Option 2 — One command (if you have Go)

hotgrin uses the [Go](https://go.dev/dl/) toolchain under the hood. If Go 1.22+
is installed, this one line installs the `hotgrin` command:

```bash
go install github.com/hotgrin/hotgrin/cmd/hotgrin@latest
```

That's it. Check it worked:

```bash
hotgrin version
```

> If your terminal says it can't find `hotgrin`, add Go's bin folder to your
> PATH: it's `%USERPROFILE%\go\bin` on Windows, `~/go/bin` on Mac/Linux.

## Option 3 — Download a binary

Grab the file for your system from the
[latest release](https://github.com/Hotgrin/hotgrin/releases/latest)
(`hotgrin-windows-amd64.zip`, `hotgrin-linux-amd64.tar.gz`, or a macOS build),
unzip it somewhere on your PATH — done. You still need
[Go installed](https://go.dev/dl/) for `run`/`build`/`test` (hotgrin compiles
your programs with it).

## Your first program (5 minutes)

Make a file called `hello.hot`:

```
say "Hello, world"

set name to "Adriaan"
say "Welcome, " plus name
```

Run it:

```bash
hotgrin run hello.hot
```

Check it for problems (try misspelling `name` and run this again):

```bash
hotgrin check hello.hot        # or:  hotgrin check --af hello.hot
```

Turn it into a program you can give to anyone:

```bash
hotgrin build hello.hot              # for this computer
hotgrin build --windows hello.hot    # a Windows .exe
```

And if you're curious what your program *becomes*:

```bash
hotgrin reveal hello.hot             # shows the generated Go
```

## Where to next?

- **[The gentle tutorial](tutorial.md)** — from zero to sharing your own `.exe`,
  step by step.
- **[The cookbook](cookbook.md)** — copy-paste recipes for everyday tasks.
- **[A complete worked project](../examples/projects/loan-calculator/)** — a real
  loan calculator with tests.
- **[The language reference](language-reference.md)** — every construct, with its
  Go mapping.
- **[The roadmap](../ROADMAP.md)** — where hotgrin is going.
