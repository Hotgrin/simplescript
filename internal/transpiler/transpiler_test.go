package transpiler

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hotgrin/hotgrin/internal/lexer"
	"github.com/hotgrin/hotgrin/internal/parser"
)

// gen lexes, parses, and transpiles src to Go, failing on parse errors.
func gen(t *testing.T, src string) string {
	t.Helper()
	toks := lexer.New(src).Tokenize()
	prog, perrs := parser.New(toks).Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	out, _, _ := New(prog).Transpile()
	return out
}

// wants checks that every expected fragment appears in the generated Go.
func wants(t *testing.T, src string, fragments ...string) {
	t.Helper()
	out := gen(t, src)
	for _, f := range fragments {
		if !strings.Contains(out, f) {
			t.Errorf("\nfor src: %s\nexpected Go to contain: %q\ngot:\n%s", src, f, out)
		}
	}
}

func TestSayAndSet(t *testing.T) {
	wants(t, `say "hi"`, `fmt.Println("hi")`)
	wants(t, "set x to 5", "x := 5")
	wants(t, "set x to 5\nset x to 6", "x := 5", "x = 6") // declare then assign
}

func TestPlusRule(t *testing.T) {
	// number + number is arithmetic
	wants(t, "set x to 2 plus 3", "(2 + 3)")
	// text + number concatenates, converting the number
	wants(t, `say "score: " plus 5`, `("score: " + fmt.Sprint(5))`)
	// text + text is plain concat
	wants(t, `say "a" plus "b"`, `("a" + "b")`)
}

func TestDivisionIsDecimal(t *testing.T) {
	wants(t, "set x to 7 divided by 2", "(float64(7) / float64(2))")
}

func TestTimesBindsTighter(t *testing.T) {
	wants(t, "set x to 2 plus 3 times 4", "(2 + (3 * 4))")
}

func TestRecord(t *testing.T) {
	src := "describe Adriaan\n" +
		"age is 56\n" +
		"location is \"Johannesburg\"\n" +
		"end describe\n" +
		"say location of Adriaan"
	wants(t, src,
		"type AdriaanT struct {",
		"Age      int",
		"Location string",
		"adriaan := AdriaanT{",
		"adriaan.Location",
	)
}

func TestActionSignatureInference(t *testing.T) {
	src := "action grade with name, mark\n" +
		"if mark is at least 50\n" +
		"give back name plus \" passed\"\n" +
		"else\n" +
		"give back name plus \" no\"\n" +
		"end if\n" +
		"end action\n" +
		"say grade with \"Adriaan\", 82"
	wants(t, src,
		"func grade(name string, mark int) string {",
		"if mark >= 50 {",
		"grade(\"Adriaan\", 82)",
	)
}

func TestCollections(t *testing.T) {
	src := "set scores to list of 90, 85, 100\n" +
		"put 75 into scores\n" +
		"say count of scores"
	wants(t, src,
		"scores := []int{90, 85, 100}",
		"scores = append(scores, 75)",
		"len(scores)",
	)
}

func TestLoops(t *testing.T) {
	wants(t, "repeat 3 times\nsay \"x\"\nend repeat", "for i0 := 0; i0 < 3; i0++ {")
	wants(t, "set xs to list of 1, 2\nrepeat for each n in xs\nsay n\nend repeat",
		"for _, n := range xs {")
}

func TestMultiWordNamesBecomeCamelCase(t *testing.T) {
	wants(t, "set cart total to 0\nincrease cart total by 5",
		"cartTotal := 0", "cartTotal += 5")
}

func TestConcurrencyWaitGroup(t *testing.T) {
	src := "action ping with x\nsay x\nend action\n" +
		"at the same time\n" +
		"do ping with \"a\"\n" +
		"do ping with \"b\"\n" +
		"end at the same time"
	wants(t, src,
		"var wg0 sync.WaitGroup",
		"wg0.Add(2)",
		"go func() {",
		"defer wg0.Done()",
		"ping(\"a\")",
		"wg0.Wait()",
	)
}

func TestConcurrentCollectIsMutexGuarded(t *testing.T) {
	src := "action sq with n\ngive back n times n\nend action\n" +
		"set answers to list of 0\n" +
		"at the same time\n" +
		"do sq with 3 into answers\n" +
		"end at the same time"
	wants(t, src,
		"var mu0 sync.Mutex",
		"r := sq(3)",
		"mu0.Lock()",
		"answers = append(answers, r)",
		"mu0.Unlock()",
	)
}

func TestUnusedVariableGuarded(t *testing.T) {
	// A set-but-never-used variable must still produce compilable Go
	// (Go rejects unused declarations), so the transpiler adds "_ = name".
	wants(t, "set x to 5\nsay \"hi\"", "x := 5", "_ = x")

	// A variable that IS used must NOT get a blank-assignment guard.
	out := gen(t, "set y to 5\nsay y")
	if strings.Contains(out, "_ = y") {
		t.Errorf("used variable should not be guarded:\n%s", out)
	}
}

func TestInputFlags(t *testing.T) {
	src := "input name as text default \"world\"\ninput count as whole default 3\nsay name"
	wants(t, src,
		"\"flag\"",
		"nameFlag := flag.String(\"name\", \"world\", \"name\")",
		"countFlag := flag.Int(\"count\", 3, \"count\")",
		"flag.Parse()",
		"name := *nameFlag",
	)
}

func TestSayFallibleInsideTry(t *testing.T) {
	// "say <fallible call>" inside try must never print Go's (value, <nil>) pair.
	src := "action half with n\nif n is 0\ngive back problem \"no\"\nend if\ngive back n divided by 2\nend action\n" +
		"try\nsay half with 10\nif it fails\nsay the problem\nend try"
	out := gen(t, src)
	if strings.Contains(out, "fmt.Println(half(10))") {
		t.Errorf("say of fallible call leaks the error tuple:\n%s", out)
	}
	for _, want := range []string{", err := half(10)", "fmt.Println(sayVal"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestSetField(t *testing.T) {
	src := "describe order\nprice is 100\nend describe\nset price of order to 249\nsay price of order"
	wants(t, src, "order.Price = 249")
}

func TestUnits(t *testing.T) {
	wants(t, "set w to 129 kg\nsay w", "w := 129.0", "fmt.Println(w, \"kg\")")
	wants(t, "set w to 129 kg\nsay w in g", "(w * 1000)", "\"g\")")
	// cross-unit arithmetic converts into the left unit
	wants(t, "set d to 2 km plus 500 m\nsay d", "(2.0 + (500.0 * 0.001))")
	// text joining shows the unit
	wants(t, "set w to 129 kg\nsay \"W: \" plus w", "fmt.Sprintf(\"%v \"+\"kg\", w)")
}

func TestUnitRateDivision(t *testing.T) {
	// Different dimensions divide into a plain base-unit rate: 5 km / 25 min
	// = 5000 m / 1500 s. Same dimension stays a plain ratio.
	wants(t, "set d to 5 km\nset p to 25 min\nsay d divided by p",
		"(d * 1000)", "(p * 60)")
	wants(t, "set a to 2 km\nset b to 500 m\nsay a divided by b",
		"(b * 0.001)")
}

func TestUnitDimensionErrors(t *testing.T) {
	toks := lexer.New("set w to 1 kg\nset h to 1 m\nsay w plus h").Tokenize()
	prog, _ := parser.New(toks).Parse()
	_, _, errs := New(prog).Transpile()
	if len(errs) == 0 || !strings.Contains(strings.Join(errs, " "), "mass") {
		t.Errorf("expected a dimension error, got %v", errs)
	}
	toks2 := lexer.New("set w to 1 kg\nsay w plus 5").Tokenize()
	prog2, _ := parser.New(toks2).Parse()
	_, _, errs2 := New(prog2).Transpile()
	if len(errs2) == 0 {
		t.Error("expected an error mixing a unit and a bare number")
	}
}

func TestUseGoBlock(t *testing.T) {
	src := "use go\nimport \"math/rand\"\nfunc luckyNumber() int { return rand.Intn(100) }\nend go\nsay lucky number"
	wants(t, src,
		"\"math/rand\"",              // import merged
		"func luckyNumber() int",     // body verbatim
		"fmt.Println(luckyNumber())", // bare name = zero-arg call
	)
	out := gen(t, src)
	if strings.Contains(out, "import \"math/rand\"\nfunc") {
		t.Errorf("import line should be lifted out of the body:\n%s", out)
	}
}

func TestRepeatedFallibleSetSameVar(t *testing.T) {
	// Two fallible sets of the SAME variable in one try body must not emit a
	// duplicate := (found by the AI prompt-pack QA). The second becomes "=".
	fall := "action risky with p\nif p is 0\ngive back problem \"no\"\nend if\ngive back p\nend action\n"
	src := fall +
		"try\nset saved to risky with 1\nset saved to risky with 2\nsay saved\n" +
		"if it fails\nsay the problem\nend try"
	out := gen(t, src)
	if len(regexp.MustCompile(`saved, err\d+ :=`).FindAllString(out, -1)) != 1 ||
		len(regexp.MustCompile(`saved, err\d+ =[^=]`).FindAllString(out, -1)) != 1 {
		t.Errorf("want exactly one := and one = for repeated set:\n%s", out)
	}
	// Sibling try scopes must not leak declarations into each other.
	src2 := fall +
		"try\nset saved to risky with 1\nsay saved\n" +
		"if it fails\ntry\nset saved to risky with 2\nsay saved\n" +
		"if it fails\nsay the problem\nend try\nend try"
	out2 := gen(t, src2)
	if len(regexp.MustCompile(`saved, err\d+ :=`).FindAllString(out2, -1)) != 2 {
		t.Errorf("sibling closures must each declare independently:\n%s", out2)
	}
	// Two SEQUENTIAL tries in one scope must use distinct error names.
	src4 := fall + "try\nset a to risky with 1\nsay a\nif it fails\nsay the problem\nend try\n" +
		"try\nset b to risky with 2\nsay b\nif it fails\nsay the problem\nend try"
	out4 := gen(t, src4)
	if len(regexp.MustCompile(`tryErr\d+ := func`).FindAllString(out4, -1)) != 2 ||
		strings.Count(out4, "tryErr0 :=") > 1 {
		t.Errorf("sequential tries must not share an error name:\n%s", out4)
	}
	// An unused fallible-set variable still needs its _ = guard.
	src3 := fall + "try\nset ignored to risky with 1\nif it fails\nsay the problem\nend try"
	if out3 := gen(t, src3); !strings.Contains(out3, "_ = ignored") {
		t.Errorf("unused fallible set missing guard:\n%s", out3)
	}
}

func TestGoFallibleBridge(t *testing.T) {
	src := "use go\nimport \"os\"\nfunc readFile(p string) (string, error) { b, e := os.ReadFile(p); return string(b), e }\nend go\n" +
		"try\nset c to read file with \"x.txt\"\nsay c\nif it fails\nsay the problem\nend try"
	wants(t, src, "c, err", ":= readFile(\"x.txt\")")
}

func TestAskStopRounded(t *testing.T) {
	wants(t, "ask \"Name?\" into name\nsay name",
		"\"bufio\"", "\"strings\"", "stdinReader = bufio.NewReader(os.Stdin)",
		"name, _ := stdinReader.ReadString('\\n')", "strings.TrimSpace(name)")
	wants(t, "stop with error \"bad\"",
		"fmt.Fprintln(os.Stderr, \"bad\")", "os.Exit(1)")
	wants(t, "say 3.14159 rounded to 2",
		"\"math\"", "func roundTo(", "roundTo(3.14159, 2)")
	// int operand gets promoted
	wants(t, "set x to 10\nsay x rounded to 1", "roundTo(float64(x), 1)")
}

func TestInferenceThroughActionLocals(t *testing.T) {
	// An action's locals must drive parameter AND return inference of callees.
	src := "input years as whole default 2\n" +
		"action grow with r, n\nset f to 1.0\nrepeat n times\nset f to f times (1 plus r)\nend repeat\ngive back f\nend action\n" +
		"action pay with term\nset months to term times 12\nset rate to 0.01\ngive back grow with rate, months\nend action\n" +
		"say pay with years"
	wants(t, src,
		"func grow(r float64, n int) float64",
		"func pay(term int) float64",
	)
}

func TestErrorHandling(t *testing.T) {
	src := "action risky with x\n" +
		"if x is 0\ngive back problem \"no zero\"\nend if\n" +
		"give back x\nend action\n" +
		"try\nset r to risky with 5\nsay r\n" +
		"if it fails\nsay the problem\nend try"
	wants(t, src,
		"func risky(x int) (int, error)", // fallible signature
		"errors.New(\"no zero\")",        // problem path
		"return x, nil",                  // value path returns nil error
		"\"errors\"",                     // import
		"r, err",                         // checked call inside try
		":= risky(5)",
		".Error()",
	)
}

func TestTestEmission(t *testing.T) {
	src := "action add with a, b\ngive back a plus b\nend action\n" +
		"test \"adds\"\nexpect add with 2, 3 to be 5\nend test"
	toks := lexer.New(src).Tokenize()
	prog, perrs := parser.New(toks).Parse()
	if len(perrs) > 0 {
		t.Fatalf("parse errors: %v", perrs)
	}
	mainSrc, testSrc, _ := New(prog).Transpile()

	if strings.Contains(mainSrc, "func TestAdds") {
		t.Errorf("test leaked into main.go:\n%s", mainSrc)
	}
	for _, want := range []string{
		"package main", "\"testing\"", "func TestAdds(t *testing.T)",
		"add(2, 3)", "t.Errorf(\"expected %v to be %v\"",
	} {
		if !strings.Contains(testSrc, want) {
			t.Errorf("test src missing %q:\n%s", want, testSrc)
		}
	}
}

func TestExpectFloatIntComparison(t *testing.T) {
	src := "action half with n\ngive back n divided by 2\nend action\n" +
		"test \"t\"\nexpect half with 10 to be 5\nend test"
	toks := lexer.New(src).Tokenize()
	prog, _ := parser.New(toks).Parse()
	_, testSrc, _ := New(prog).Transpile()
	if !strings.Contains(testSrc, "float64(") {
		t.Errorf("expected float promotion in test src:\n%s", testSrc)
	}
}

func TestParamInferenceFromVariableArg(t *testing.T) {
	// 'discount' is called with a variable; its parameter types must still be
	// inferred (not fall back to 'any').
	src := "set total to 897\n" +
		"action discount with amount, percent\n" +
		"give back amount minus percent\n" +
		"end action\n" +
		"say discount with total, 10"
	wants(t, src, "func discount(amount int, percent int) int {")
}
