# hotgrin for AI assistants (the prompt pack)

**Instructions for the AI reading this:** you are being taught **hotgrin**, a
programming language released in 2026 — it is NOT in your training data. When
asked to write hotgrin, use ONLY the syntax in this document. Never invent
constructs that are not listed here. Follow the GOTCHAS section strictly —
each one is a mistake AIs are likely to make. hotgrin v0.5.

## What hotgrin is

A language that reads like plain English. Files end in `.hot`. Programs are
line-oriented (one statement per line, `#` starts a comment) and every block
closes with an explicit `end` line. hotgrin transpiles to Go and compiles to a
native executable. Run with `hotgrin run file.hot`.

## Names (critical — different from every other language)

Names may contain **spaces**: `set cart total to 0` declares one variable
named `cart total`. Names are bounded by reserved words (`to`, `of`, `with`,
`is`, `plus`, `in`, ...). Consequences:
- Never use a reserved word inside a name.
- camelCase and snake_case are legal but unidiomatic — prefer spaced names.
- Unicode names are fine.

## Types (inferred — never annotated)

whole numbers (`42`), decimals (`3.14`), text (`"hi"` or `'hi'`), truth
(`true`/`yes`, `false`/`no`), `nothing`, lists, records, measurements.
There is NO type annotation syntax. The first `set` declares a variable;
later `set`s reassign. A list's elements must share one type.

## Statements

```
say <expr>                          # print (with newline)
set <name> to <expr>                # declare or assign
increase <name> by <expr>           # += (also works on record fields)
decrease <name> by <expr>           # -=

if <cond>
    ...
else if <cond>
    ...
else
    ...
end if

repeat <expr> times ... end repeat
repeat while <cond> ... end repeat
repeat for each <name> in <list> ... end repeat

set xs to list of 1, 2, 3
put 4 into xs                       # append
say count of xs                     # length
say item 0 of xs                    # 0-based index; variables ok: item i of xs
if xs contains 3 ...                # membership

describe order                      # a record with fields
    item is "Mouse"
    price is 299
end describe
say price of order                  # field read
set price of order to 249           # field write

action <name> with <p1>, <p2>       # define a function ("action")
    give back <expr>                # return
end action
set r to <name> with <arg1>, <arg2> # call; zero-arg call is just the bare name

give back problem "message"         # inside an action: this action can FAIL
try
    set r to <fallible call>
if it fails
    say "Reason: " plus the problem # 'the problem' = the failure message
end try

test "name"                         # tests live next to code; run: hotgrin test f.hot
    expect <expr> to be <expr>      # also: to be at least / at most /
end test                            #   less than / greater than / contains

at the same time                    # run calls concurrently, wait for all
    do <call>
    do <call> into <list>           # collect results safely (order not guaranteed)
end at the same time
start <call>                        # fire in background; program waits at exit

input name as text default "world"  # command-line flag --name (free --help)
input n as whole default 3          # types: text, whole, decimal, truth
ask "What is your name?" into name  # interactive prompt; answer is ALWAYS text
stop with error "message"           # print to stderr, exit code 1

use "lib/helpers"                   # local library (path relative to this file)
use "std/text"                      # standard library (see list below)
use tools from "github.com/user/repo"       # remote library, fetched with git
use tools from "github.com/user/repo@v1.0"  # pinned to a tag

use go                              # escape hatch: verbatim Go
import "strings"
func shoutCase(s string) string { return strings.ToUpper(s) + "!" }
end go
say shout case with "howzit"        # camelCase is called as spaced words
```

`use go` rules: one `import "pkg"` per line; parameters/returns limited to
`string`, `int`, `float64`, `bool`; a function returning `(T, error)` becomes
a FALLIBLE action (callers must use `try`). Blocks are top-level only.

## Expressions and precedence (loosest → tightest)

1. `or`
2. `and`
3. comparisons: `is`, `is not`, `is greater than`, `is less than`,
   `is at least`, `is at most`, `contains`
4. `rounded to` (round to N decimal places), `in` (unit conversion)
5. `plus`, `minus`
6. `times`, `divided by`
7. `of`

- `plus` adds numbers OR joins text (if either side is text, the other is
  converted automatically).
- `divided by` ALWAYS yields a decimal: `7 divided by 2` is `3.5`.
- Parentheses group as in maths.

## Units of measure

```
set weight to 129 kg
say weight                       # prints: 129 kg  (units print themselves)
say weight in g                  # prints: 129000 g
set walk to 2 km plus 500 m      # 2.5 km (right side converts into LEFT unit)
if 90 min is greater than 1 h ...
set double to weight times 2     # scaling by plain numbers is fine
```

Units — mass: `mg g kg t` · length: `mm cm m km` · time: `ms s seconds min
minutes h hours` · volume: `ml l`. Rules: same-dimension values combine and
compare; `unit times/divided by plain-number` is fine; `unit plus/minus
plain-number` is an ERROR (write the unit on both sides); mixing dimensions
(kg with m) is an ERROR; `unit times unit` is an ERROR.

## Standard libraries (exact action names)

- `use "std/text"` → `upper case with s` · `lower case with s` ·
  `trim spaces with s` · `replace all with s, old, new` · `text length with s`
  · `starts with with s, prefix` · `ends with with s, suffix`
- `use "std/data"` → `read file with path` · `write file with path, content`
  — **both FALLIBLE: must be called inside try**
- `use "std/random"` → `random up to with n` (0..n-1) ·
  `random between with low, high` (inclusive)
- `use "std/web"` → `fetch text with url` (HTTP GET) ·
  `json value with doc, "dotted.path"` (extract a JSON field as text)
  — **both FALLIBLE: must be called inside try**

## GOTCHAS — follow these strictly

1. **Fallible calls MUST be inside `try ... if it fails ... end try`.** Any
   action containing `give back problem`, any `(T, error)` go-func, and all of
   std/data and std/web. The checker refuses the program otherwise.
2. **`rounded to` binds looser than arithmetic but attaches to the last
   argument of a call.** `say pay with a, b rounded to 2` rounds only `b`.
   Correct: `set p to pay with a, b` then `say p rounded to 2`, or
   `say (pay with a, b) rounded to 2`... prefer the two-line form.
3. **`ask` always yields TEXT.** There is no text→number conversion. For
   numeric input, use `input n as whole default 5` (a --flag) instead of ask.
4. **Indexes are 0-based.** `item 0 of xs` is the first element.
5. **Do not mix list element types.** `list of 1, 2, "three"` is wrong.
6. **A void action (no `give back`) cannot be used as a value.**
   `set x to log it with "hi"` is an error if `log it` gives nothing back.
7. **Every block needs its `end` line:** `end if`, `end repeat`, `end action`,
   `end describe`, `end try`, `end test`, `end at the same time`, `end go`.
8. **Don't invent stdlib actions.** Only the ones listed above exist. For
   anything else, use the `use go` escape hatch.
9. **Measurements:** never add a bare number to a unit value; never mix
   dimensions; remember `say` prints the unit for you.
10. **There is no `while`, `for`, `func`, `return`, `elif`, `print`, `import`
    keyword.** The words are `repeat while`, `repeat for each`, `action`,
    `give back`, `else if`, `say`, `use`.

## Verified example (uses much of the language)

```
use "std/text"

input rate as decimal default 11.5

action monthly payment with loan, yearly rate, term years
    if yearly rate is 0
        give back loan divided by (term years times 12)
    end if
    set monthly rate to yearly rate divided by 100 divided by 12
    set months to term years times 12
    set factor to 1.0
    repeat months times
        set factor to factor times (1 plus monthly rate)
    end repeat
    give back loan times monthly rate times factor divided by (factor minus 1)
end action

set payment to monthly payment with 250000.0, rate, 20
say "Monthly: R" plus (payment rounded to 2)
say upper case with "done"

test "zero interest is simple division"
    expect monthly payment with 12000.0, 0.0, 1 to be 1000
end test
```

## CLI

```
hotgrin run file.hot [--yourflags]   # check, compile, execute
hotgrin test file.hot                # run the test blocks
hotgrin check [--af] file.hot        # problems only (--af = Afrikaans)
hotgrin build [--windows] file.hot   # native binary / Windows .exe
hotgrin reveal file.hot              # show the generated Go
```

The Watcher checks every program before it runs and explains problems kindly
(English or Afrikaans). If it is silent, the program compiles.

Docs: https://github.com/Hotgrin/hotgrin — playground:
https://hotgrin.github.io/hotgrin/playground/
