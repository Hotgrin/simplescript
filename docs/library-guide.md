# Writing hotgrin libraries

A library is just a `.hot` file whose **actions** other programs can use. This
guide covers local libraries, publishing to GitHub, and the `use go` escape
hatch.

## Local libraries

Put actions in a file; import it by path (relative to the importing file):

```
# lib/greetings.hot
action greet with who
    give back "Hello, " plus who
end action
```

```
use "lib/greetings"
say greet with "AJ"
```

Only actions (and `use go` blocks) are taken from a library — top-level
statements in it are ignored. Imports are transitive, each file loads once,
and the Watcher checks imported code too.

## Publishing on GitHub

1. Make a public repo with a **`lib.hot`** at its root containing your actions.
2. Anyone can then write:

```
use tools from "github.com/you/your-repo"
```

hotgrin fetches it with git (so git must be installed), caches it under
`~/.hotgrin/cache/`, and compiles it into the program. Details:

- **Subpaths work:** `use "github.com/you/repo/helpers/text"` loads
  `helpers/text.hot` (or `helpers/text/lib.hot` if that's a folder).
- **Versions:** pin a tag with `@`: `use "github.com/you/repo@v1.2.0"`.
- **Cache:** fetched once, reused offline. Delete the folder under
  `~/.hotgrin/cache/` to force a refresh.

## The `use go` escape hatch

When hotgrin doesn't have something, wrap Go directly:

```
use go
import "strings"
func shoutCase(s string) string { return strings.ToUpper(s) + "!" }
end go

say shout case with "howzit"
```

Rules of the hatch (v0.4):

- **camelCase becomes spoken form:** `shoutCase` is callable as `shout case`.
- **Imports:** one per line, `import "pkg"` form; hotgrin merges and
  de-duplicates them with its own.
- **Failure-capable functions:** return `(T, error)` and hotgrin treats the
  action as fallible — callers must use `try / if it fails`, exactly like
  `give back problem`. This is how `std/data`'s `read file` works.
- **Supported signatures:** plain parameters and returns of `string`, `int`,
  `float64`, `bool` (or `(T, error)`). Keep exotic Go (channels, generics,
  methods) inside the block — expose simple functions.
- Blocks are top-level only, and everything between `use go` and `end go` is
  passed to the Go compiler verbatim — Go's rules apply inside.

## The standard libraries

Shipped inside hotgrin itself (no download):

- `use "std/text"` — `upper case`, `lower case`, `trim spaces`, `replace all`,
  `text length`, `starts with`, `ends with`
- `use "std/data"` — `read file`, `write file` (fallible — use `try`)
- `use "std/random"` — `random up to`, `random between`
- `use "std/web"` — `fetch text`, `json value` (fallible)

They're written in hotgrin using `use go` — read them in
[`internal/loader/std/`](../internal/loader/std/) as worked examples of this
guide.

## Conventions for a friendly library

- Name actions the way they'll be *read*: `send invoice with ...`, not `sndInv`.
- One purpose per library; small beats sprawling.
- Include a `test "..."` block or two — users can run `hotgrin test lib.hot`.
- Say in a comment what can fail and why.
