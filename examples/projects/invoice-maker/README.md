# Invoice Maker — the flagship worked example

A real invoice generator in under 200 heavily-commented lines: items, VAT,
a text file, and a web page — with tests proving the maths.

- **[invoice.hot](invoice.hot)** — the program, with a numbered section
  index `[1]`–`[8]` in its comments
- **[TUTORIAL.md](TUTORIAL.md)** — an extensive beginner walkthrough,
  section by section

```bash
hotgrin run invoice.hot                       # asks who the invoice is for
hotgrin run invoice.hot --company "My Shop"   # your details via flags
hotgrin test invoice.hot                      # prove the maths
hotgrin build --windows invoice.hot           # share it as invoice.exe
```
