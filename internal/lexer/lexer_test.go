package lexer

import "testing"

// check lexes src and asserts the token TYPES match want (ignoring the final EOF
// unless you include it). It's a quick way to pin down tokenisation shape.
func typesOf(src string) []TokenType {
	toks := New(src).Tokenize()
	out := make([]TokenType, len(toks))
	for i, t := range toks {
		out[i] = t.Type
	}
	return out
}

func equal(a, b []TokenType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTokenTypes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []TokenType
	}{
		{
			name: "hello world",
			src:  `say "Hello, world"`,
			want: []TokenType{SAY, STRING, NEWLINE, EOF},
		},
		{
			name: "multi-word name bounded by 'to'",
			src:  `set cart total to 0`,
			want: []TokenType{SET, IDENT, TO, NUMBER, NEWLINE, EOF},
		},
		{
			name: "longest-match comparison phrase",
			src:  `if score is greater than 40`,
			want: []TokenType{IF, IDENT, IS_GREATER_THAN, NUMBER, NEWLINE, EOF},
		},
		{
			name: "text join with plus",
			src:  `say learner plus " scored " plus score`,
			want: []TokenType{SAY, IDENT, PLUS, STRING, PLUS, IDENT, NEWLINE, EOF},
		},
		{
			name: "comma-separated arguments",
			src:  `set total to add with 3, 4`,
			want: []TokenType{SET, IDENT, TO, IDENT, WITH, NUMBER, COMMA, NUMBER, NEWLINE, EOF},
		},
		{
			name: "give back is one token",
			src:  `give back a plus b`,
			want: []TokenType{GIVE_BACK, IDENT, PLUS, IDENT, NEWLINE, EOF},
		},
		{
			name: "if it fails beats if",
			src:  `if it fails`,
			want: []TokenType{IF_IT_FAILS, NEWLINE, EOF},
		},
		{
			name: "describe block lines",
			src: "describe Adriaan\n" +
				"age is 56\n" +
				"location is \"Johannesburg\"\n" +
				"end describe",
			want: []TokenType{
				DESCRIBE, IDENT, NEWLINE,
				IDENT, IS, NUMBER, NEWLINE,
				IDENT, IS, STRING, NEWLINE,
				END, DESCRIBE, NEWLINE,
				EOF,
			},
		},
		{
			name: "to be assertion",
			src:  `expect total to be 5`,
			want: []TokenType{EXPECT, IDENT, TO_BE, NUMBER, NEWLINE, EOF},
		},
		{
			name: "hash inside a string is not a comment",
			src:  `say "C# is still cool"`,
			want: []TokenType{SAY, STRING, NEWLINE, EOF},
		},
		{
			name: "comment-only line yields nothing but EOF",
			src:  `# this whole line is a comment`,
			want: []TokenType{EOF},
		},
		{
			name: "at the same time is one token",
			src:  `at the same time`,
			want: []TokenType{AT_THE_SAME_TIME, NEWLINE, EOF},
		},
		{
			name: "unicode identifier",
			src:  `set prénom to "José"`,
			want: []TokenType{SET, IDENT, TO, STRING, NEWLINE, EOF},
		},
		{
			name: "record read with of",
			src:  `say location of Adriaan`,
			want: []TokenType{SAY, IDENT, OF, IDENT, NEWLINE, EOF},
		},
		{
			name: "increase keeps the name separate (the demo bug)",
			src:  `increase cart total by 199`,
			want: []TokenType{INCREASE, IDENT, BY, NUMBER, NEWLINE, EOF},
		},
		{
			// 'add' and 'list' are NOT reserved (so 'add' stays usable as a
			// function name). The parser resolves 'list of ...' and the
			// list-append statement by grammatical context.
			name: "list and add lex as identifiers for the parser to resolve",
			src: "set scores to list of 90, 85, 100\n" +
				"add 75 to scores",
			want: []TokenType{
				SET, IDENT, TO, IDENT, OF, NUMBER, COMMA, NUMBER, COMMA, NUMBER, NEWLINE,
				IDENT, NUMBER, TO, IDENT, NEWLINE,
				EOF,
			},
		},
		{
			// 'put X into L' is the list-append form. 'put' is reserved so the
			// name after it (even a multi-word one) stays separate.
			name: "put into list-append, multi-word value",
			src:  `put chosen product into cart`,
			want: []TokenType{PUT, IDENT, INTO, IDENT, NEWLINE, EOF},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := typesOf(c.src)
			if !equal(got, c.want) {
				t.Errorf("\nsrc:  %q\n got: %v\nwant: %v", c.src, got, c.want)
			}
		})
	}
}

// Spot-check that literals carry the right text (and that names keep casing).
func TestLiterals(t *testing.T) {
	toks := New(`set Cart Total to "R49"`).Tokenize()
	// SET, IDENT("Cart Total"), TO, STRING("R49"), NEWLINE, EOF
	if toks[1].Type != IDENT || toks[1].Literal != "Cart Total" {
		t.Errorf("expected IDENT 'Cart Total', got %v", toks[1])
	}
	if toks[3].Type != STRING || toks[3].Literal != "R49" {
		t.Errorf("expected STRING 'R49', got %v", toks[3])
	}
}

// 'yes' and 'no' must lex as TRUE and FALSE.
func TestYesNoAreTruth(t *testing.T) {
	toks := New("set ready to yes").Tokenize()
	if toks[3].Type != TRUE {
		t.Errorf("expected 'yes' to lex as TRUE, got %v", toks[3])
	}
}

// Line numbers must be tracked so later stages can point at the right line.
func TestLineNumbers(t *testing.T) {
	src := "say \"one\"\n\nsay \"three\""
	toks := New(src).Tokenize()
	last := toks[len(toks)-2] // the STRING "three"
	if last.Line != 3 {
		t.Errorf("expected 'three' on line 3, got line %d", last.Line)
	}
}

func TestStringEscapes(t *testing.T) {
	toks := New(`say "line one\nline two\ttabbed \"quoted\""`).Tokenize()
	got := ""
	for _, tk := range toks {
		if tk.Type == STRING {
			got = tk.Literal
		}
	}
	want := "line one\nline two\ttabbed \"quoted\""
	if got != want {
		t.Errorf("escapes wrong: got %q want %q", got, want)
	}
}
