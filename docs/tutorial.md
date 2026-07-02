# Your first SimpleScript program — a gentle tutorial

Welcome! This tutorial takes you from nothing to a real, working program — and
then to a program you can *share* as a Windows `.exe`. No programming experience
needed. Every example here really runs.

## 0. Setting up (one time only)

1. Install **Go** from [go.dev/dl](https://go.dev/dl/) (SimpleScript uses it
   behind the scenes — you will never have to write any Go).
2. Download SimpleScript and build the command:

```bash
git clone https://github.com/Hotgrin/simplescript
cd simplescript
go build -o simplescript ./cmd/simplescript
```

That's it. You now have a `simplescript` command (on Windows: `simplescript.exe`).

## 1. Say hello

Make a file called `mine.ss` with one line:

```
say "Hello, world"
```

Run it:

```bash
./simplescript run mine.ss
```

`say` shows something on the screen. Anything in quotes is **text**.

## 2. Remember things with `set`

```
set name to "Adriaan"
set year to 2026
say "Hello, " plus name
say "Next year is " plus (year plus 1)
```

- `set ... to ...` stores a value under a name. Names can even have spaces:
  `set cart total to 0` is one name, `cart total`.
- `plus` adds numbers, and joins text. If either side is text, the other side is
  converted for you — no ceremony.
- Parentheses group, exactly like school math.

Try changing the name and run it again.

## 3. Make decisions with `if`

```
set score to 82

if score is at least 50
    say "You passed!"
else
    say "Not yet - try again"
end if
```

The comparisons read like English: `is`, `is not`, `is greater than`,
`is less than`, `is at least`, `is at most`. Every block closes with an `end`
line (`end if`), so you always know where things finish.

## 4. Repeat things

```
repeat 3 times
    say "hip hip hooray"
end repeat

set scores to list of 90, 85, 100
repeat for each s in scores
    say s
end repeat
```

A `list of` holds several values. `put 75 into scores` adds one more;
`count of scores` tells you how many; `item 0 of scores` is the first one
(computers count from 0).

## 5. Teach the computer a new trick: actions

An **action** is a recipe you name once and use many times:

```
action greet with who
    give back "Hello, " plus who
end action

say greet with "AJ"
say greet with "the whole world"
```

- `with who` means the action takes one input, called `who`.
- `give back` hands a result to whoever asked.
- You call it by name: `greet with "AJ"`.

## 6. When things can go wrong

Real programs face problems: a file is missing, a number is zero. SimpleScript
makes handling that honest and gentle:

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
    say "Could not do that: " plus the problem
end try
```

`give back problem "..."` says "this went wrong". A `try ... if it fails` block
catches it — `the problem` holds the message. And here's the kind part: if you
*forget* the `try`, SimpleScript stops you **before running** with a friendly
note, so a failure can never slip past silently.

## 7. Check your own work: tests

Tests are part of the language, not an add-on:

```
test "greeting works"
    expect greet with "AJ" to be "Hello, AJ"
end test
```

```bash
./simplescript test mine.ss
```

You'll see a green PASS — or a plain-English explanation of what didn't match.

## 8. The Watcher — a friend looking over your shoulder

Make a deliberate mistake:

```
set total to 100
say totall
```

```bash
./simplescript check mine.ss
```

```
error   line 2: there is no value called 'totall' here — is it a typo, or did you forget to set it?
```

The Watcher only speaks when it is *sure* something is wrong — it never nags
about things that are fine. Prefer Afrikaans? `./simplescript check --af mine.ss`.

## 9. Take inputs, share your program

```
input name as text default "world"
say "Hello, " plus name
```

```bash
./simplescript run mine.ss --name AJ
```

And when you're proud of it, make a real program file you can give to anyone:

```bash
./simplescript build mine.ss              # a program for this computer
./simplescript build --windows mine.ss    # a Windows .exe
```

The person you give it to needs nothing installed — it just runs.

## 10. Where to next?

- Split reusable actions into their own file and `use "myhelpers"` to share them.
- Peek behind the curtain any time: `./simplescript reveal mine.ss` shows the
  real Go code your program becomes. Nothing is hidden.
- Read the [language reference](language-reference.md) for every construct.

Happy building! And if something confuses you, that's a documentation bug —
please [open an issue](https://github.com/Hotgrin/simplescript/issues) and tell
us.
