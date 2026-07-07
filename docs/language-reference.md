# hotgrin language reference (v0.5)

A precise reference for the current implementation. For a gentle introduction,
see the [tutorial](tutorial.md). hotgrin transpiles to Go; each section
notes the mapping, and `hotgrin reveal file.hot` shows the exact output for
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

| hotgrin value | Go type |
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
say item i of xs                # variable indexes work too
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
`hotgrin test file.hot` runs them via `go test`. Int/float mismatches in
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

### Interactive prompts and stopping
```
ask "What is your name?" into name    # prints the prompt, reads a line (text)
stop with error "no file given"       # message to stderr, exit code 1
```
`ask` always yields text; convert as needed. The Watcher treats code after
`stop with error` as unreachable.

### Units of measure
```
set weight to 129 kg
say weight                       # 129 kg  (units print themselves)
say weight in g                  # 129000 g
set walk to 2 km plus 500 m      # 2.5 km  (converted into the left unit)
if 90 min is greater than 1 h ...
```
Known units — mass: `mg g kg t` · length: `mm cm m km` · time:
`ms s/seconds min/minutes h/hours` · volume: `ml l`. Same-dimension values
combine and compare (the right side converts into the left side's unit);
scaling by a plain number is fine (`weight times 2`); mixing dimensions, or
adding a bare number to a measurement, is an error caught before the program
runs. `x in <unit>` converts explicitly.

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
`hotgrin run app.hot --name AJ`.

### Libraries
```
use "lib/textutils"             # local: path relative to the importing file
use "std/text"                  # standard library, shipped inside hotgrin
use tools from "github.com/you/repo"        # remote, fetched with git + cached
use tools from "github.com/you/repo@v1.0"   # pinned to a tag
```
A library is a plain `.hot` file; its **actions** are merged into the consuming
program (whole-program transpile). Transitive `use` works; each file loads once
(cycles and diamonds are safe); the Watcher checks imported code too. Non-action
top-level statements in a library are ignored. Remote github.com libraries are
fetched with git and cached under `~/.hotgrin/cache/`. See the
[library guide](library-guide.md).

### The `use go` escape hatch
```
use go
import "strings"
func shoutCase(s string) string { return strings.ToUpper(s) + "!" }
end go

say shout case with "howzit"
```
Everything between `use go` and `end go` is Go, compiled verbatim. Declared
functions become callable actions (camelCase reads as spaced words); functions
returning `(T, error)` are fallible and must be called inside `try`. Simple
parameter/return types only (`string`, `int`, `float64`, `bool`).

## Expressions

Operator precedence, loosest to tightest:

1. `or`
2. `and`
3. comparisons: `is`, `is not`, `is greater than`, `is less than`,
   `is at least`, `is at most`, `contains`
4. `rounded to`
5. `plus`, `minus`
6. `times`, `divided by`
7. `of` (field/list access)

- `plus` is numeric addition, or concatenation when either operand is text
  (the other side is auto-converted).
- `<expr> rounded to <n>` rounds to n decimal places (always yields a decimal).
  It binds looser than arithmetic — `a plus b rounded to 2` rounds the sum —
  but tighter than comparisons. To round a call's result, parenthesise the
  call: `(monthly payment with a, r, y) rounded to 2`.
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
hotgrin run     <file.hot> [--flags for your program]
hotgrin test    <file.hot>
hotgrin build   [--windows] <file.hot>
hotgrin check   [--af] <file.hot>
hotgrin reveal  <file.hot>
hotgrin version | help
```

## Not in v0.1 (roadmap)

Type annotations · units of measure · an interpreter mode
(no Go install) · the Assessor and AI-mentor checking layers.
