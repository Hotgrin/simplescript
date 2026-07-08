# hotgrin examples, by category

Every program here is machine-verified: it runs exactly as written. Categories
match the "What can you build?" grid on [hotgrin.com](https://hotgrin.com).

- **[seo/](seo/)** — check pages answer, title-tag length, robots.txt
- **[api/](api/)** — talk to real REST APIs, tiny uptime checker
- **[math/](math/)** — percentages, compound interest, list statistics
- **[finance/](finance/)** — savings goal timeline, ROI (see also the
  [loan calculator project](projects/loan-calculator/))
- **[science/](science/)** — speed with real units, conversions, recipe scaling
- **[text-files/](text-files/)** — word counts, find-and-replace, a persistent journal
- **[html/](html/)** — generate a real web page from data
- **[email/](email/)** — send mail via SMTP (through the `use go` escape hatch)
- **[games/](games/)** — a quiz with scoring, a dice battle

Plus the originals in this folder (`hello.hot`, `shop.hot`, ...) and two
complete worked projects: the
[invoice maker](projects/invoice-maker/) (flagship, with a full beginner
walkthrough) and the [loan calculator](projects/loan-calculator/).

Run any of them: `hotgrin run examples/<category>/<name>.hot` — most take
`--flags` (try `--help`).
