# Loan calculator — a complete hotgrin project

A real, useful program in one `.hot` file: give it a loan amount, an interest
rate, and a term, and it tells you the monthly payment, the total you'll pay,
and the total interest.

```bash
hotgrin run loan.hot
hotgrin run loan.hot --amount 100000 --rate 10 --years 5
hotgrin run loan.hot --help
hotgrin test loan.hot
hotgrin build --windows loan.hot     # share it as loan.exe
```

## What it shows off

- **Command-line inputs** with sensible defaults and a free `--help`
- **An action** doing real financial maths (the amortization formula)
- **A loop standing in for a power operator** — `(1+r)^n` built by multiplying
  in a `repeat`, with a comment explaining exactly that
- **Multi-word names** (`monthly rate`, `total paid`, `term years`) reading like
  the sentence you'd say aloud
- **Tests living next to the code they prove**, including a range assertion
  (`to be at least` / `to be at most`) for a floating-point answer

## Honest notes

The payment prints with full decimal places (R2666.0740…) because hotgrin has
no number-formatting helpers yet — that's on the [roadmap](../../../ROADMAP.md),
and this project is exactly the kind that motivates it.
