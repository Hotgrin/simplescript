# The Invoice Maker — a complete walkthrough for absolute beginners

This is the manual for [`invoice.hot`](invoice.hot): a real, useful program in
under 200 lines of hotgrin, heavily commented, that builds a professional
invoice with VAT and saves it as a text file *and* a web page.

You don't need any programming experience. We'll walk through the program
**section by section** — the same numbered sections `[1]` to `[8]` you'll find
in the code's comment index, so you can read this manual and the code side by
side.

## First, run it

```bash
hotgrin run invoice.hot
```

It will ask *"Who is this invoice for?"* — type any name and press Enter.
You'll see a finished invoice on screen, and two new files appear next to the
program: `invoice.txt` and `invoice.html`. Double-click `invoice.html` — that's
your invoice, in a browser, made by a program you can read.

Now try it with your own details, without touching the code:

```bash
hotgrin run invoice.hot --company "My Shop" --vat 15
hotgrin run invoice.hot --vat 0
hotgrin run invoice.hot --help
```

That `--help` came free. Keep it in mind when we reach Section [1].

## The two lines before Section [1]

```
use "std/data"
use "std/text"
```

`use` borrows abilities from a **library**. These two ship inside hotgrin:
`std/data` knows how to read and write files (we'll need `write file` in
Section [7]), and `std/text` knows text tricks like `upper case` and
`trim spaces`. A `use` line is like taking a toolbox off the shelf before you
start working.

## [1] Settings — the `input` lines

```
input company as text default "Hotgrin Trading"
input vat as decimal default 15
input invoice number as text default "INV-2026-001"
```

Each `input` line creates a **setting** that lives outside the code:

- `company` holds **text** (words in quotes),
- `vat` holds a **decimal** (a number that may have a comma... well, a point),
- `invoice number` holds text — and notice the name has a **space in it**.
  That's normal in hotgrin: `invoice number` is one name, exactly the way
  you'd say it out loud.

The `default` part is what's used when you don't say otherwise. And every
`input` automatically becomes a `--flag`, which is why
`--company "My Shop"` worked above without editing anything.

**Try it:** change the default company name, run again.

## [2] The items — three lists that travel together

```
set item names to list of "Wireless mouse", "Mechanical keyboard", "USB-C cable"
set item prices to list of 299.0, 1450.0, 89.5
set item quantities to list of 2, 1, 3
```

`set <name> to <value>` stores a value under a name. A `list of` holds several
values in order. Here we keep three lists side by side, like three columns of
a table: position 0 of each list describes the mouse, position 1 the
keyboard, position 2 the cable.

**Why does counting start at 0?** Computers count the *distance from the
start*: the first item is 0 steps in. Every programming language you'll ever
meet does this — you're learning the real thing.

**Try it:** add a fourth product — one new entry in *each* of the three lists.
Run it. The invoice grows by itself; nothing else needs changing. (If you add
to only one list, the program will complain — the loop in Section [5] walks by
position, so the lists must stay the same length.)

## [3] Money actions — teach it once, use it often

```
action line total with price, quantity
    give back price times quantity
end action
```

An **action** is your own command: a small recipe with a name. `with price,
quantity` means it needs two ingredients handed to it; `give back` hands the
answer to whoever asked. Once taught, you use it by name:
`line total with 299.0, 2` is `598`.

Look at the second action closely:

```
action vat portion with amount, rate
    give back amount times rate divided by 100
end action
```

You might wonder: the VAT rate already exists up in Section [1] — why hand it
in again as `rate`? Because of an important hotgrin rule worth learning early:

> **An action only sees what you hand it.** Values created at the top of the
> program are not visible inside actions — everything an action needs must
> travel in through its parameters.

This looks strict, but it's a gift: you can read any action on its own and
know *exactly* what it depends on. Nothing sneaks in from outside.

The third action, `rand`, dresses a number as money: `rand with 89.456` gives
`"R89.46"`. The rounding is done by `rounded to 2` — two decimal places,
because that's how money looks.

## [4] Ask the client — talking to the person at the keyboard

```
ask "Who is this invoice for?" into client name
set client name to trim spaces with client name
```

`ask` prints the question, waits for typing, and stores the answer in
`client name`. One honest detail: whatever is typed arrives as **text** —
even if someone types `42`, it's the *text* "42", not the number. For numbers,
use the `input` flags from Section [1] instead; that's the pattern this
program follows.

The second line is politeness: `trim spaces` (from `std/text`) removes
accidental spaces around the answer.

## [5] Build the lines — the loop

This is the engine room. Read it slowly; it's the most "programmy" part, and
once it clicks, you can write loops forever.

```
set lines to ""
set subtotal to 0.0
set i to 0

repeat while i is less than count of item names
    set name to item i of item names
    set price to item i of item prices
    set quantity to item i of item quantities
    ...
    increase i by 1
end repeat
```

Three running values start empty: `lines` (the invoice text we're building),
`subtotal` (money so far), and `i` (which position we're on — programmers
love the name `i` for *index*).

`repeat while <something is true>` keeps going as long as the condition
holds. Each time around: `item i of item names` reads position `i` from the
list — first 0, then 1, then 2. We compute the line's cost with our Section
[3] action, grow the subtotal, and glue a human-readable line onto `lines`.
The `"\n"` at the end of each glued line means "start a new line" — it's how
text files and screens know where lines break.

And then the single most important line for a beginner to understand:

```
    increase i by 1
end repeat
```

Without it, `i` stays 0 forever, the condition stays true forever, and the
program loops until you stop it. Every programmer has written that bug.
Now you know the cure before the disease.

**Try it:** change `is less than` to `is at most` and run. One item repeats?
Off-by-one errors are the other classic — see how `count of` (which is 3)
and positions (0, 1, 2) fit together.

## [6] The totals

```
set vat amount to vat portion with subtotal, vat
set grand total to subtotal plus vat amount
```

Plain arithmetic with our own action. Notice `vat` — the Section [1] setting —
being handed in as the rate, exactly as rule [3] taught us.

## [7] Show and save — screen, then files

The `say` lines print the invoice. `upper case with company` shouts the
company name — a `std/text` ability.

Then the interesting part: saving. In the real world, **writing a file can
fail** — a full disk, a protected folder. hotgrin refuses to let you ignore
that possibility:

```
try
    write file with "invoice.txt", text copy
    say "Saved invoice.txt"
if it fails
    say "Could not save the text copy: " plus the problem
end try
```

Everything inside `try` is attempted; if any step fails, the program jumps to
`if it fails`, where `the problem` holds a plain-English reason. If you delete
the `try` and call `write file` bare, hotgrin's checker (the Watcher) will
stop you *before running* and explain — a failure can never slip past
silently. That's not nagging; that's the language keeping your program
honest.

The HTML section right after looks fancy but isn't: a web page is **just
text** with markers like `<h1>` (a heading) and `<b>` (bold). We build that
text the same way we built the invoice lines, and save it the same way.
That's the demystifying secret of this whole section: *files and web pages
are text, and you already know how to build text.*

## [8] The proof — tests

```
test "a line total multiplies price by quantity"
    expect line total with 100.0, 3 to be 300
end test
```

Tests are little promises, checked by the machine:

```bash
hotgrin test invoice.hot
```

Green `PASS` lines mean the maths holds. Now try this: break the program on
purpose — change `times` to `plus` inside `line total` — and run the tests
again. Watch them catch it instantly, in plain English. Fix it back. **That**
is why programmers write tests: not for today, but for the day someone
"improves" the maths.

## Where to go from here

- **Peek behind the curtain:** `hotgrin reveal invoice.hot` shows the real Go
  code your program becomes. Nothing is hidden.
- **Make it yours:** your company, your products, your rate. Then
  `hotgrin build --windows invoice.hot` produces an `invoice.exe` you can put
  on any Windows machine — nothing to install.
- **Keep learning:** the [cookbook](../../../docs/cookbook.md) has 21 more
  verified recipes, and the [gentle tutorial](../../../docs/tutorial.md)
  covers everything this program used, step by step.

If anything in this walkthrough confused you, that's a bug in the
walkthrough — [tell us](https://github.com/Hotgrin/hotgrin/issues) and we'll
fix it. Happy invoicing! 😁
