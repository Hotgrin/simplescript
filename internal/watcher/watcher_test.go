package watcher

import (
	"strings"
	"testing"

	"github.com/hotgrin/hotgrin/internal/lexer"
	"github.com/hotgrin/hotgrin/internal/parser"
)

func check(t *testing.T, src string) []Finding {
	t.Helper()
	toks := lexer.New(src).Tokenize()
	prog, _ := parser.New(toks).Parse()
	return New(prog).Check()
}

// hasFinding reports whether any finding's English text contains substr.
func hasFinding(fs []Finding, sev Severity, substr string) bool {
	for _, f := range fs {
		if f.Severity == sev && strings.Contains(f.En, substr) {
			return true
		}
	}
	return false
}

func TestUnknownAction(t *testing.T) {
	fs := check(t, `say frobnicate with 3`)
	if !hasFinding(fs, Error, "no action called 'frobnicate'") {
		t.Errorf("expected unknown-action error, got %v", fs)
	}
}

func TestWrongArgCount(t *testing.T) {
	fs := check(t, "action greet with name\nsay name\nend action\nsay greet with \"a\", \"b\"")
	if !hasFinding(fs, Error, "needs 1 input") {
		t.Errorf("expected wrong-arg-count error, got %v", fs)
	}
}

func TestUnknownVariable(t *testing.T) {
	fs := check(t, "set score to 10\nsay scoer")
	if !hasFinding(fs, Error, "no value called 'scoer'") {
		t.Errorf("expected unknown-variable error, got %v", fs)
	}
}

func TestDivideByZero(t *testing.T) {
	fs := check(t, "set x to 10 divided by 0")
	if !hasFinding(fs, Error, "divides by zero") {
		t.Errorf("expected divide-by-zero error, got %v", fs)
	}
}

func TestDuplicateAction(t *testing.T) {
	fs := check(t, "action go1\nsay \"a\"\nend action\naction go1\nsay \"b\"\nend action")
	if !hasFinding(fs, Error, "already an action called 'go1'") {
		t.Errorf("expected duplicate-action error, got %v", fs)
	}
}

func TestDuplicateParam(t *testing.T) {
	fs := check(t, "action add with a, a\ngive back a\nend action")
	if !hasFinding(fs, Error, "two inputs called 'a'") {
		t.Errorf("expected duplicate-param error, got %v", fs)
	}
}

func TestUnreachableAfterGiveBack(t *testing.T) {
	fs := check(t, "action f\ngive back 1\nsay \"dead\"\nend action")
	if !hasFinding(fs, Warning, "can never run") {
		t.Errorf("expected unreachable warning, got %v", fs)
	}
}

func TestConstantCondition(t *testing.T) {
	fs := check(t, "if 5 is greater than 3\nsay \"x\"\nend if")
	if !hasFinding(fs, Warning, "always the same") {
		t.Errorf("expected constant-condition warning, got %v", fs)
	}
}

func TestUnusedVariable(t *testing.T) {
	fs := check(t, "set used to 1\nset leftover to 2\nsay used")
	if !hasFinding(fs, Suggestion, "'leftover' is set but never used") {
		t.Errorf("expected unused suggestion, got %v", fs)
	}
	// 'used' must NOT be flagged
	if hasFinding(fs, Suggestion, "'used' is set") {
		t.Errorf("'used' should not be flagged as unused: %v", fs)
	}
}

func TestUnknownRecordField(t *testing.T) {
	src := "describe car\ncolour is \"red\"\nend describe\nsay speed of car"
	fs := check(t, src)
	if !hasFinding(fs, Error, "has no field called 'speed'") {
		t.Errorf("expected unknown-field error, got %v", fs)
	}
}

// The iron rule: good programs produce ZERO findings.
func TestNoFalsePositives(t *testing.T) {
	good := []string{
		`say "hello"`,
		"set x to 5\nincrease x by 1\nsay x",
		"action grade with name, mark\nif mark is at least 50\ngive back name plus \" ok\"\nelse\ngive back name plus \" no\"\nend if\nend action\nsay grade with \"AJ\", 80",
		"describe car\ncolour is \"red\"\nend describe\nsay colour of car",
		"set xs to list of 1, 2, 3\nput 4 into xs\nrepeat for each n in xs\nsay n\nend repeat",
		"action ping with x\nsay x\nend action\nat the same time\ndo ping with \"a\"\ndo ping with \"b\"\nend at the same time",
	}
	for _, src := range good {
		fs := check(t, src)
		if len(fs) != 0 {
			t.Errorf("FALSE POSITIVE on good program:\n%s\n-> %v", src, fs)
		}
	}
}

func TestInputsAreKnownNames(t *testing.T) {
	src := "input name as text default \"world\"\nsay \"Hello \" plus name"
	if fs := check(t, src); len(fs) != 0 {
		t.Errorf("false positive using a declared input: %v", fs)
	}
}

func TestVoidActionAsValue(t *testing.T) {
	src := "action log it with msg\nsay msg\nend action\nset x to log it with \"hi\"\nsay x"
	if !hasFinding(check(t, src), Error, "does not give anything back") {
		t.Errorf("expected void-use error")
	}
	// but an action that gives back is fine
	src2 := "action two\ngive back 2\nend action\nset x to two\nsay x"
	for _, f := range check(t, src2) {
		if f.Severity == Error {
			t.Errorf("false positive on value action: %v", f)
		}
	}
}

func TestGoFuncsAreKnownActions(t *testing.T) {
	src := "use go\nfunc luckyNumber() int { return 7 }\nend go\nsay lucky number"
	if fs := check(t, src); len(fs) != 0 {
		t.Errorf("false positive calling a go-block func: %v", fs)
	}
	// fallible go funcs must be guarded by try
	src2 := "use go\nimport \"os\"\nfunc readFile(p string) (string, error) { b, e := os.ReadFile(p); return string(b), e }\nend go\nset c to read file with \"x\"\nsay c"
	if !hasFinding(check(t, src2), Error, "can fail") {
		t.Errorf("expected fallible-outside-try error for go-bridged func")
	}
}

func TestAskDeclaresName(t *testing.T) {
	if fs := check(t, "ask \"Name?\" into name\nsay \"Hi \" plus name"); len(fs) != 0 {
		t.Errorf("false positive on ask-declared name: %v", fs)
	}
}

func TestUnreachableAfterStop(t *testing.T) {
	fs := check(t, "action f\nstop with error \"bye\"\nsay \"dead\"\nend action")
	if !hasFinding(fs, Warning, "can never run") {
		t.Errorf("expected unreachable-after-stop warning, got %v", fs)
	}
}

func TestFallibleCallOutsideTry(t *testing.T) {
	src := "action risky with x\nif x is 0\ngive back problem \"no\"\nend if\ngive back x\nend action\n" +
		"set r to risky with 5\nsay r"
	if !hasFinding(check(t, src), Error, "'risky' can fail") {
		t.Errorf("expected fallible-outside-try error")
	}
}

func TestFallibleCallInsideTryIsFine(t *testing.T) {
	src := "action risky with x\nif x is 0\ngive back problem \"no\"\nend if\ngive back x\nend action\n" +
		"try\nset r to risky with 5\nsay r\nif it fails\nsay the problem\nend try"
	if fs := check(t, src); len(fs) != 0 {
		t.Errorf("false positive on a correctly-handled fallible call: %v", fs)
	}
}

func TestChecksInsideTestBodies(t *testing.T) {
	fs := check(t, "test \"t\"\nexpect mistery with 3 to be 9\nend test")
	if !hasFinding(fs, Error, "no action called 'mistery'") {
		t.Errorf("expected the Watcher to check test bodies, got %v", fs)
	}
}

func TestGoodTestNoFalsePositive(t *testing.T) {
	src := "action add with a, b\ngive back a plus b\nend action\n" +
		"test \"adds\"\nexpect add with 2, 3 to be 5\nend test"
	if fs := check(t, src); len(fs) != 0 {
		t.Errorf("false positive on a good test program: %v", fs)
	}
}

func TestAfrikaansMessages(t *testing.T) {
	fs := check(t, "set x to 10 divided by 0")
	if len(fs) == 0 {
		t.Fatal("expected a finding")
	}
	af := fs[0].Message("af")
	if !strings.Contains(af, "deel deur nul") {
		t.Errorf("expected Afrikaans message, got %q", af)
	}
}
