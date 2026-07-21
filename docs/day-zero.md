# Day Zero: you already think like a programmer

No computer needed for this page. Not even hotgrin installed. Just you,
maybe a cup of coffee, and five minutes.

If you've never written a line of code in your life — if you're not
entirely sure what a "file" is, or you've avoided this because school
was a long time ago, or you've quietly worried you might be too old or
not clever enough to start — this page is for you specifically. Not a
watered-down version of the real thing. The real thing, just without
the screen, so nothing about a keyboard or a red error message can get
in the way of what's actually true:

**You already think like a programmer. You've just never called it that.**

Here's the proof.

---

## 1. The recipe — an "algorithm"

When you make your family's chicken pie, you don't do the steps in a
random order. You don't add the pastry before the filling. You follow
the steps, in order, one at a time, and the pie comes out right.

That's it. That's an **algorithm**: a list of exact steps, done in
order, to get a result. Programmers didn't invent the idea — you were
doing it before you ever touched a computer.

In hotgrin, a recipe for making tea looks like this. Don't worry about
reading it perfectly — just notice it's a list of steps, top to
bottom, same as your pie:

```
say "Boil the kettle"
say "Warm the pot"
say "Add the tea leaves"
say "Pour in the hot water"
say "Wait four minutes"
say "Pour and serve"
```

**Now break it on purpose.** Swap two of those lines — say, "pour in
the hot water" and "add the tea leaves." Read it out loud in the new
order. It still *sounds* fine as a sentence... but try to actually
picture doing it that way. Pouring water into an empty pot, then
tossing tea leaves into a pot that's already full — it doesn't ruin
the tea completely, but it's not how you'd actually make it, and
anyone watching would know something was off.

That's exactly how a computer behaves. It will follow your steps
*precisely* as written — even the ones you didn't mean to write that
way. It never assumes, never guesses what you "obviously" meant, never
quietly fixes your order for you. That's not the computer being
difficult. That's the entire rulebook, and you already just followed
it — and broke it, on purpose, and understood exactly why — without
touching a keyboard.

**Try it now, unplugged:** write down the exact steps to make a cup of
tea, one per line, on paper. Show it to someone and ask them to follow
it *exactly* as written, no assuming. Then swap two steps yourself, the
way we just did above, and notice what changes. That question — "did I
actually say what I meant, in the order I meant it?" — is 90% of
programming.

---

## 2. The grocery list — a "loop"

You don't write a new plan for every item in your trolley. You do the
same thing, again and again, until the list is empty: pick an item,
check it off, next item, pick an item, check it off, next item... stop
when the list is done.

That repeating pattern is a **loop**.

```
set groceries to list of "bread", "milk", "eggs"
repeat for each thing in groceries
    say "Buying: " plus thing
end repeat
```

**Try it now, unplugged:** think of something else in your day that's
really a loop in disguise — hanging washing piece by piece, greeting
each guest at the door, watering each pot on the stoep. You've been
running loops your whole life.

---

## 3. The umbrella rule — a "decision"

"If it's raining, take an umbrella. Otherwise, leave it." You make
dozens of these calls a day without thinking twice: if the pot's
boiling, turn it down; if the light's red, stop; if it's Sunday,
sleep in.

A computer needs exactly the same kind of instruction — it just needs
you to spell it out, because it can't glance out the window itself.

```
set raining to true
if raining
    say "Take the umbrella"
else
    say "Leave the umbrella"
end if
```

**Try it now, unplugged:** write three "if this, then that" rules from
your actual morning routine. That's a program. You just wrote one.

---

## 4. The labelled jar — a "variable"

Picture a jam jar with a sticky label on it that says "sugar." Today
it's full. Tomorrow you use half of it. The label doesn't change — the
jar's still called "sugar" — but what's *inside* it changes.

That's a **variable**: a labelled space that holds a value, and the
value is allowed to change while the label stays the same.

```
set sugar to "half a jar"
say sugar
```

**Try it now, unplugged:** look around your kitchen for three labelled
containers. For each one, say out loud what's in it right now, and
imagine saying the same sentence again next week with a different
answer. That's exactly how a variable behaves.

---

## 5. The recipe you lend a friend — an "action"

You know your chicken pie recipe so well you don't write it down
anymore — you just tell people "add the filling, same as always."
Once you've taught someone a recipe properly, they can use it
whenever they like, on their own, without you standing there. You
just say the name of the recipe, and they know what to do.

That's an **action** (some languages call it a "function"): a
recipe you teach the computer once, then reuse by name, as many times
as you like.

```
action make tea with cups
    say "Boil the kettle"
    say "Pour " plus cups plus " cups"
end action

make tea with 2
make tea with 4
```

**Try it now, unplugged:** think of a routine you've explained to
someone else so well that they can now do it without you — showing a
grandchild how to make toast, teaching a colleague how you file
invoices. You already "wrote" that action. You just used words
instead of a keyboard.

---

## The part nobody tells you

A few honest answers to the worries that stop most people before they
start — because the worries are common, and none of them are true:

**"What if I break the computer?"**
You can't. Not with hotgrin, not from typing something wrong. The
worst that happens is a friendly message telling you what to fix —
in English or Afrikaans, your choice. Nothing you type here can damage
your computer.

**"What if I forget everything by next week?"**
You will forget some of it. Everyone does. That's not a sign you
can't do this — it's just how learning works at any age. Come back,
re-read this page, run the first lesson again. Repetition is the plan,
not a failure of it.

**"Aren't I too old to start this?"**
No. People take this up in their 50s, 60s, and beyond — often *because*
they want to keep their mind sharp, not despite it. The steps above
weren't simplified for you. They're the actual ideas every programmer
uses, every day, for their entire career. You just met them first,
here, in plain language.

**"I don't even know how to open a terminal."**
You don't need to yet. The [browser playground](https://hotgrin.github.io/hotgrin/playground/)
opens like any website — type on the left, see it run on the right.
Nothing to install, nothing to configure, nothing to get wrong on the
way in.

---

## What you actually just learned

Recipe → **algorithm**. Grocery list → **loop**. Umbrella rule →
**decision**. Labelled jar → **variable**. Recipe you taught a friend
→ **action**. Five ideas, no computer, and you already had all five
before you started reading.

Everything from here is just learning the exact words hotgrin wants
you to use for ideas you already own. That part is easy — you're not
learning to *think* differently, only to *spell out* what you already
do, every day, without noticing.

When you're ready: **[Lesson 01 — say hello](../examples/learn/01-say-hello.hot)**
is next. Open the [playground](https://hotgrin.github.io/hotgrin/playground/),
paste it in, and press run. That's the whole first step.

Baie geluk — you started. 🎉
