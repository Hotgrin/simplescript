# The hotgrin cookbook

Copy-paste recipes for everyday tasks. Every recipe on this page runs exactly as
shown — paste it into a `.hot` file and `hotgrin run` it. Recipes marked with
`--flags` take inputs: `hotgrin run recipe.hot --amount 500`.

## 1. Greet someone

```
set name to "Adriaan"
say "Hello, " plus name plus "!"
```

## 2. Add VAT to a price (South African 15%)

```
set price to 500
set vat to price times 15 divided by 100
say "Price: R" plus price
say "VAT: R" plus vat
say "Total: R" plus (price plus vat)
```

## 3. Take the price as an input instead

```
input price as decimal default 500
set total to price plus (price times 15 divided by 100)
say "Total with VAT: R" plus total
```

Run it: `hotgrin run vat.hot --price 1250`

## 4. Find the highest mark

```
set marks to list of 67, 82, 45, 91, 78
set highest to 0

repeat for each m in marks
    if m is greater than highest
        set highest to m
    end if
end repeat

say "Highest mark: " plus highest
```

## 5. Average of a list

```
set marks to list of 67, 82, 45, 91, 78
set total to 0

repeat for each m in marks
    set total to total plus m
end repeat

say "Average: " plus (total divided by count of marks)
```

## 6. Countdown

```
set n to 5
repeat while n is greater than 0
    say n
    decrease n by 1
end repeat
say "Liftoff!"
```

## 7. A times table

```
input table as whole default 7
set n to 1
repeat 12 times
    say table plus " x " plus n plus " = " plus (table times n)
    increase n by 1
end repeat
```

## 8. Pass or fail, for a whole class

```
action grade with mark
    if mark is at least 50
        give back "passed"
    else
        give back "must retry"
    end if
end action

set marks to list of 82, 47, 65

repeat for each m in marks
    say m plus ": " plus grade with m
end repeat
```

## 9. A shopping cart with a record

```
describe cart
    item is "Wireless mouse"
    price is 299
    quantity is 3
end describe

set total to price of cart times quantity of cart
say item of cart plus " x" plus quantity of cart plus " = R" plus total
```

## 10. Apply a discount (as a reusable action)

```
action discount with amount, percent
    give back amount minus (amount times percent divided by 100)
end action

say "R" plus discount with 897, 10
```

## 11. Celsius to Fahrenheit

```
input celsius as decimal default 25
set fahrenheit to celsius times 9 divided by 5 plus 32
say celsius plus " C is " plus fahrenheit plus " F"
```

## 12. Guard against division by zero

```
action safe divide with a, b
    if b is 0
        give back problem "cannot divide by zero"
    end if
    give back a divided by b
end action

try
    say safe divide with 10, 2
    say safe divide with 10, 0
if it fails
    say "Caught: " plus the problem
end try
```

## 13. Is it in the list?

```
set colours to list of "red", "green", "blue"

if colours contains "green"
    say "Green is available"
end if
```

## 14. Grow a list as you go

```
set scores to list of 90, 85, 100
put 75 into scores
put 62 into scores

say count of scores plus " scores collected"

repeat for each s in scores
    say s
end repeat
```

## 15. Prove your code works — with a test

```
action add with a, b
    give back a plus b
end action

test "addition works"
    expect add with 2, 3 to be 5
    expect add with 10, 10 to be at least 15
end test
```

Run it: `hotgrin test recipe.hot`

---

Want a recipe that isn't here? [Open an issue](https://github.com/Hotgrin/hotgrin/issues)
and describe what you're trying to cook.
