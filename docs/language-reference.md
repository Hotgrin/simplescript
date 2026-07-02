# SimpleScript language reference (v0.1)

A precise reference for the current implementation. For a gentle introduction,
see the [tutorial](tutorial.md). SimpleScript transpiles to Go; each section
notes the mapping, and `simplescript reveal file.ss` shows the exact output for
your program.

## Lexical structure

- **Line-oriented.** One statement per line; blocks close with `end <keyword>`.
- **Comments** start with `#` and run to end of line.
- **Names may contain spaces** (`cart total` is one identifier). This works
  because a small fixed set of reserved *connector words* (`to`, `of`, `with`,
  `is`, ...) bound them; a name is every word up to the next reserved word.
  Longest-match wins: `is greater than` lexes as one token, beating `is`.
- **Unicode identifiers** are supported (`prénom`, `имя`, `名前`); names keep
  their original casing.
- **Strings** use double or single quotes. **Numbers** are integer or decimal
  literals; a leading `-` is part of the literal.
- Reserved words (e.g. `times`, `list`-with-`of`, connectors) cannot be used as
  plain variable names.

## Types

Types are inferred and hidden (no annotations in v0.1). The ladder:

| SimpleScript value | Go type |
|---|---|
| `"text"` | `string` |
| `42` | `int` |
| `3.14` | `float64` |
| `true` / `yes`, `false` / `no` | `bool` |
| `nothing` | zero value (context-dependent) |
| `list of ...` | `[]T` (element type inferred; mixed types are an error) |
| `describe` entity | inferred `struct` |

Numeric promotion: mixing `int` and `float64` in arithmetic promotes to
`float64`. `divided by` **always** yields `float64` (`7 divided by 2` is `3.5`).

## Statements

### Output and variables
```
say <expr>                      # fmt.Println
set <name> to <expr>            # first set declares (:=), later sets assign (=)
increase <name> by <expr>       # +=
decrease <name> by <expr>       # -=
```
A variable that is set but never read is guarded with `_ = name` in the
generated Go (so it always compiles) and reported by the Watcher as a
suggestion.

### Conditionals
```
if <cond>
    ...
else if <cond>
    ...
else
    ...
end if
```

### Loops
```
repeat <expr> times ... end repeat        # for i := 0; i < n; i++
repeat while <cond> ... end repeat        # for cond
repeat for each <name> in <list> ... end repeat   # for _, v := range
```

### Lists
```
set xs to list of 1, 2, 3
put 4 into xs                   # append
say count of xs                 # len
say item 0 of xs                # safe indexed access (0-based)
xs contains 3                   # membership (generated helper)
```

### Records
```
describe order
    item is "Wireless mouse"
    price is 299
end describe

say item of order               # field read
set price of order to 249      # field write
increase price of order by 10
```
A `describe` block generates a struct type (`orderT`) and a value (`order`);
fields are TitleCased in Go. Inside the block, `is` sets a field. `X of Y`
resolves by the type of `Y`: list operation for lists, field access for records.

### Actions
```
action <name> with <p1>, <p2>
    ...
    give back <expr>
end action

set r to <name> with <arg1>, <arg2>
```
Parameter types are inferred from call sites; the return type from `give back`.
Top-level statements form `main()`; actions become top-level funcs.

### Error handling
```
action risky with x
    if x is 0
        give back problem "message"     # return zero, errors.New(...)
    end if
    give back x                          # return x, nil
end action

try
    set r to risky with 5
if it fails
    say the problem                      # the error's message
end try
```
An action containing `give back problem` is *fallible* and compiles to
`(T, error)`. The `try` body becomes a closure returning `error`; the first
failure short-circuits to the handler. **Calling a fallible action outside
`try` is a Watcher error** — failure cannot be silently ignored.

### Tests
```
test "name"
    expect <expr> to be <expr>
end test
```
Assertions: `to be`, `to be at least`, `to be at most`, `to be less than`,
`to be greater than`, `contains`. Tests emit to a generated `_test.go`;
`simplescript test file.ss` runs them via `go test`. Int/float mismatches in
assertions are reconciled automatically.

### Concurrency
```
at the same time
    do <call>
    do <call> into <list>       # mutex-guarded collect (safe by design)
end at the same time            # waits for all (WaitGroup)

start <call>                    # background; program waits at exit
```
Collected order is non-deterministic. Writes to shared variables inside the
block other than `into` are unsupported.

### Command-line inputs
```
input name as text default "world"
input count as whole default 3
input rate as decimal
input loud as truth
```
Types: `text`, `whole` (int), `number`/`decimal` (float64), `truth` (bool).
Inputs become Go `flag` declarations at the top of `main()`; the program gets
`--name`-style options and a generated `--help`. Pass values through the CLI:
`simplescript run app.ss --name AJ`.

### Libraries
```
use "lib/textutils"             # path relative to the importing file (.ss optional)
use textutils from "lib/textutils"
```
A library is a plain `.ss` file; its **actions** are merged into the consuming
program (whole-program transpile). Transitive `use` works; each file loads once
(cycles and diamonds are safe); the Watcher checks imported code too. Non-action
top-level statements in a library are ignored. Remote paths (GitHub/URLs) are
not fetched yet and produce a friendly error.

## Expressions

Operator precedence, loosest to tightest:

1. `or`
2. `and`
3. comparisons: `is`, `is not`, `is greater than`, `is less than`,
   `is at least`, `is at most`, `contains`
4. `plus`, `minus`
5. `times`, `divided by`
6. `of` (field/list access)

- `plus` is numeric addition, or concatenation when either operand is text
  (the other side is auto-converted).
- Parentheses group. Calls use `name with arg, arg`.

## The Watcher

The always-on checker. Iron rule: **a finding is always a real problem** (no
false alarms). Current rules: unknown variable (typo), unknown action, wrong
argument count, divide-by-zero literal, duplicate action, duplicate parameter,
unreachable code after `give back`, constant conditions, set-but-never-used
(suggestion), unknown record field, and unhandled fallible calls. Findings are
bilingual (English/Afrikaans, `--af`).

Severities: `error` blocks run/build; `warning` and suggestions do not.

## The CLI

```
simplescript run     <file.ss> [--flags for your program]
simplescript test    <file.ss>
simplescript build   [--windows] <file.ss>
simplescript check   [--af] <file.ss>
simplescript reveal  <file.ss>
simplescript version | help
```

## Not in v0.1 (roadmap)

Interactive `ask` prompts · `stop with error` · remote libraries · type
annotations · units of measure · a browser playground · the Assessor and
AI-mentor checking layers.
