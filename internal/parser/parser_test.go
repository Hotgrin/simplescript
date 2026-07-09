package parser

import (
	"strings"
	"testing"

	"github.com/hotgrin/hotgrin/internal/lexer"
)

// parse lexes and parses src, failing the test on any parse error.
func parse(t *testing.T, src string) string {
	t.Helper()
	toks := lexer.New(src).Tokenize()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			b.WriteString("\n  " + e.String())
		}
		t.Fatalf("unexpected parse errors for %q:%s", src, b.String())
	}
	return prog.String()
}

func TestParseInput(t *testing.T) {
	got := parse(t, "input name as text default \"world\"\ninput count as whole")
	for _, want := range []string{`(input "name" text default (str "world"))`, `(input "count" whole)`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestParseUse(t *testing.T) {
	if got := parse(t, `use "lib/textutils"`); !strings.Contains(got, `(use "lib/textutils")`) {
		t.Errorf("plain use not parsed: %s", got)
	}
	if got := parse(t, `use math from "lib/math"`); !strings.Contains(got, `(use "math" from "lib/math")`) {
		t.Errorf("named use not parsed: %s", got)
	}
}

func TestParseSetField(t *testing.T) {
	got := parse(t, "set price of order to 249")
	if !strings.Contains(got, "(set-field price of (id order) (num 249))") {
		t.Errorf("field write not parsed: %s", got)
	}
	// plain set still plain
	if got := parse(t, "set price to 249"); !strings.Contains(got, `(set "price" (num 249))`) {
		t.Errorf("plain set broken: %s", got)
	}
}

func TestParseUnits(t *testing.T) {
	if got := parse(t, "set w to 129 kg"); !strings.Contains(got, "(unit 129 kg)") {
		t.Errorf("unit literal not parsed: %s", got)
	}
	if got := parse(t, "say w in g"); !strings.Contains(got, "(in (id w) g)") {
		t.Errorf("conversion not parsed: %s", got)
	}
	// aliases normalise; for-each 'in' unaffected
	if got := parse(t, "set p to 90 minutes"); !strings.Contains(got, "(unit 90 min)") {
		t.Errorf("alias not normalised: %s", got)
	}
	if got := parse(t, "repeat for each s in xs\nsay s\nend repeat"); !strings.Contains(got, "(for-each s (id xs)") {
		t.Errorf("for-each in broken: %s", got)
	}
	// a non-unit word after a number stays a name usage
	if got := parse(t, "set n to 5\nsay n"); !strings.Contains(got, "(num 5)") {
		t.Errorf("plain number broken: %s", got)
	}
}

func TestParseV03Features(t *testing.T) {
	got := parse(t, "ask \"Name?\" into name\nstop with error \"bye\"")
	for _, want := range []string{`(ask (str "Name?") into name)`, `(stop-with-error (str "bye"))`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// variable list index: "item i of xs"
	if got := parse(t, "say item i of xs"); !strings.Contains(got, "(item (id i) (id xs))") {
		t.Errorf("variable index not parsed: %s", got)
	}
	// multi-word variable index
	if got := parse(t, "say item current pos of xs"); !strings.Contains(got, "(item (id current pos) (id xs))") {
		t.Errorf("multi-word variable index not parsed: %s", got)
	}
	// rounded to binds looser than arithmetic
	if got := parse(t, "say a plus b rounded to 2"); !strings.Contains(got, "(rounded (+ (id a) (id b)) (num 2))") {
		t.Errorf("rounded precedence wrong: %s", got)
	}
}

func TestParseTryAndProblem(t *testing.T) {
	got := parse(t, "try\nset x to f with 1\nif it fails\nsay the problem\nend try")
	for _, want := range []string{"(try", "(if-it-fails", "(id the problem)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	gb := parse(t, "action f\ngive back problem \"boom\"\nend action")
	if !strings.Contains(gb, "(give-back-problem (str \"boom\"))") {
		t.Errorf("give back problem not parsed:\n%s", gb)
	}
}

func TestParseTestAndExpect(t *testing.T) {
	got := parse(t, "test \"adds\"\nexpect add with 2, 3 to be 5\nend test")
	want := `(program
  (test "adds"
    (expect (call add (num 2) (num 3)) = (num 5))))`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}

	for _, c := range []struct{ src, op string }{
		{"test \"t\"\nexpect x to be at least 1\nend test", ">="},
		{"test \"t\"\nexpect x to be less than 1\nend test", "<"},
		{"test \"t\"\nexpect xs contains 3\nend test", "contains"},
		{"test \"t\"\nexpect x to be nothing\nend test", "is-nothing"},
	} {
		out := parse(t, c.src)
		if !strings.Contains(out, " "+c.op) {
			t.Errorf("for %q expected op %q in:\n%s", c.src, c.op, out)
		}
	}
}

func TestExpressions(t *testing.T) {
	cases := []struct{ src, want string }{
		{`say "hi"`, `(program
  (say (str "hi")))`},
		{`set total to 40 plus 2`, `(program
  (set "total" (+ (num 40) (num 2))))`},
		// times binds tighter than plus
		{`set x to 2 plus 3 times 4`, `(program
  (set "x" (+ (num 2) (* (num 3) (num 4)))))`},
		// parentheses override precedence
		{`set x to (2 plus 3) times 4`, `(program
  (set "x" (* (+ (num 2) (num 3)) (num 4))))`},
		// comparison + logic
		{`set ok to score is at least 50 and score is at most 100`, `(program
  (set "ok" (and (>= (id score) (num 50)) (<= (id score) (num 100)))))`},
		// text join
		{`say learner plus " scored " plus score`, `(program
  (say (+ (+ (id learner) (str " scored ")) (id score))))`},
		// call with comma args
		{`say grade with "Adriaan", 82`, `(program
  (say (call grade (str "Adriaan") (num 82))))`},
		// field access via of
		{`say location of Adriaan`, `(program
  (say (of location (id Adriaan))))`},
		// list literal
		{`set scores to list of 90, 85, 100`, `(program
  (set "scores" (list (num 90) (num 85) (num 100))))`},
		// item index
		{`say item 0 of scores`, `(program
  (say (item (num 0) (id scores))))`},
		// 'item' as a record field name (not the index operator) — regression
		{`say item of order`, `(program
  (say (of item (id order))))`},
	}
	for _, c := range cases {
		if got := parse(t, c.src); got != c.want {
			t.Errorf("\nsrc:  %s\ngot:  %s\nwant: %s", c.src, got, c.want)
		}
	}
}

func TestStatements(t *testing.T) {
	cases := []struct{ src, want string }{
		{`increase cart total by 199`, `(program
  (increase (id cart total) (num 199)))`},
		{`put chosen product into cart`, `(program
  (put (id chosen product) into (id cart)))`},
		{"describe Adriaan\nage is 56\nlocation is \"Johannesburg\"\nend describe",
			`(program
  (describe Adriaan
    (field "age" (num 56))
    (field "location" (str "Johannesburg"))))`},
	}
	for _, c := range cases {
		if got := parse(t, c.src); got != c.want {
			t.Errorf("\nsrc:  %s\ngot:  %s\nwant: %s", c.src, got, c.want)
		}
	}
}

func TestRepeatTimesNotMultiply(t *testing.T) {
	got := parse(t, "repeat 3 times\nsay \"hi\"\nend repeat")
	want := `(program
  (repeat-times (num 3)
    (say (str "hi"))))`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestIfElse(t *testing.T) {
	src := "if mark is at least 50\n" +
		"give back name plus \" passed\"\n" +
		"else\n" +
		"give back name plus \" must retry\"\n" +
		"end if"
	got := parse(t, src)
	want := `(program
  (if
    (clause (>= (id mark) (num 50))
      (give-back (+ (id name) (str " passed"))))
    (else
      (give-back (+ (id name) (str " must retry"))))))`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

func TestActionAndForEach(t *testing.T) {
	src := "action grade with name, mark\n" +
		"repeat for each person in learners\n" +
		"say person\n" +
		"end repeat\n" +
		"end action"
	got := parse(t, src)
	want := `(program
  (action grade (params name mark)
    (for-each person (id learners)
      (say (id person)))))`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

// A deferred Part II construct should be reported, not silently mis-parsed.
func TestDeferredConstructReportsError(t *testing.T) {
	// 'use' (libraries) is still on the roadmap, so it should report cleanly.
	toks := lexer.New("ask \"name?\"").Tokenize()
	_, errs := New(toks).Parse()
	if len(errs) == 0 {
		t.Fatal("expected a 'not supported yet' error for 'ask'")
	}
}

func TestConcurrency(t *testing.T) {
	src := "at the same time\n" +
		"do announce with \"a\"\n" +
		"do square with 3 into answers\n" +
		"end at the same time\n" +
		"start cleanup"
	got := parse(t, src)
	want := `(program
  (at-same-time
    (do (call announce (str "a")))
    (do (call square (num 3)) into (id answers)))
  (start (id cleanup)))`
	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}
