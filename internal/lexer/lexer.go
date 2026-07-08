package lexer

import (
	"strings"
	"unicode"
)

// Lexer reads hotgrin source and produces a flat slice of tokens.
//
// hotgrin is line-oriented, so the lexer works one line at a time. Within
// a line it tokenises in two clear stages:
//
//	Stage 1 (lineToAtoms): split the line's characters into raw "atoms" —
//	        words, numbers, strings, and the few punctuation marks. Comments and
//	        whitespace fall away here.
//	Stage 2 (emitLine):    walk the atoms, grouping runs of words into either a
//	        reserved keyword/phrase (longest match wins) or, when no keyword
//	        matches, a multi-word identifier that runs until the next keyword.
//
// This two-stage shape is what lets "set cart total to 0" become
// SET, IDENT("cart total"), TO, NUMBER("0") — the name is simply every word up
// to the next rail.
type Lexer struct {
	source string
	tokens []Token
}

// New makes a Lexer for the given source text.
func New(source string) *Lexer {
	return &Lexer{source: source}
}

// Tokenize runs the lexer and returns every token, ending with EOF.
func (l *Lexer) Tokenize() []Token {
	normalised := strings.ReplaceAll(l.source, "\r\n", "\n")
	lines := strings.Split(normalised, "\n")

	inGo := false
	goStart := 0
	var goLines []string
	for idx, raw := range lines {
		lineNo := idx + 1
		trimmed := strings.TrimSpace(raw)
		if inGo {
			if strings.EqualFold(trimmed, "end go") {
				l.add(GO_BLOCK, strings.Join(goLines, "\n"), goStart)
				l.add(NEWLINE, "", lineNo)
				inGo, goLines = false, nil
				continue
			}
			goLines = append(goLines, raw)
			continue
		}
		if strings.EqualFold(trimmed, "use go") {
			inGo, goStart = true, lineNo
			continue
		}
		atoms := lineToAtoms(raw)
		if len(atoms) == 0 {
			continue // blank line or a line that was only a comment
		}
		before := len(l.tokens)
		l.emitLine(atoms, lineNo)
		if len(l.tokens) > before {
			l.add(NEWLINE, "", lineNo)
		}
	}

	l.add(EOF, "", len(lines))
	return l.tokens
}

func (l *Lexer) add(t TokenType, literal string, line int) {
	l.tokens = append(l.tokens, Token{Type: t, Literal: literal, Line: line})
}

// --- Stage 1: characters -> atoms ---------------------------------------

type atomKind int

const (
	atomWord atomKind = iota
	atomNumber
	atomString
	atomLParen
	atomRParen
	atomComma
	atomIllegal
)

type atom struct {
	kind atomKind
	text string
}

// lineToAtoms breaks one line into atoms. A '#' that is not inside a string
// begins a comment, so the rest of the line is ignored.
func lineToAtoms(line string) []atom {
	r := []rune(line)
	var atoms []atom
	i := 0

	for i < len(r) {
		c := r[i]
		switch {
		case c == ' ' || c == '\t':
			i++

		case c == '#':
			return atoms // comment to end of line

		case c == '(':
			atoms = append(atoms, atom{atomLParen, "("})
			i++
		case c == ')':
			atoms = append(atoms, atom{atomRParen, ")"})
			i++
		case c == ',':
			atoms = append(atoms, atom{atomComma, ","})
			i++

		case c == '"' || c == '\'':
			quote := c
			i++ // skip opening quote
			var sb []rune
			for i < len(r) && r[i] != quote {
				if r[i] == '\\' && i+1 < len(r) {
					switch r[i+1] {
					case 'n':
						sb = append(sb, '\n')
					case 't':
						sb = append(sb, '\t')
					case '\\', '"', '\'':
						sb = append(sb, r[i+1])
					default: // unknown escape: keep both characters
						sb = append(sb, r[i], r[i+1])
					}
					i += 2
					continue
				}
				sb = append(sb, r[i])
				i++
			}
			atoms = append(atoms, atom{atomString, string(sb)})
			if i < len(r) {
				i++ // skip closing quote
			}

		case unicode.IsDigit(c) || (c == '-' && i+1 < len(r) && unicode.IsDigit(r[i+1])):
			start := i
			i++ // sign or first digit
			for i < len(r) && (unicode.IsDigit(r[i]) || r[i] == '.') {
				i++
			}
			atoms = append(atoms, atom{atomNumber, string(r[start:i])})

		case unicode.IsLetter(c):
			// A "word" is letters/digits/underscore. unicode.IsLetter means
			// names in any language work: prénom, имя, 名前.
			start := i
			for i < len(r) && (unicode.IsLetter(r[i]) || unicode.IsDigit(r[i]) || r[i] == '_') {
				i++
			}
			atoms = append(atoms, atom{atomWord, string(r[start:i])})

		default:
			atoms = append(atoms, atom{atomIllegal, string(c)})
			i++
		}
	}

	return atoms
}

// --- Stage 2: atoms -> tokens -------------------------------------------

// emitLine turns a line's atoms into tokens.
func (l *Lexer) emitLine(atoms []atom, line int) {
	i := 0
	for i < len(atoms) {
		a := atoms[i]
		switch a.kind {
		case atomWord:
			// Longest reserved phrase starting here?
			if tt, n, ok := matchReserved(atoms, i); ok {
				l.add(tt, joinWords(atoms[i:i+n]), line)
				i += n
				continue
			}
			// Otherwise this is an identifier: take words until a reserved
			// phrase would begin (or a non-word atom interrupts).
			start := i
			i++ // include the first word
			for i < len(atoms) && atoms[i].kind == atomWord {
				if _, _, ok := matchReserved(atoms, i); ok {
					break
				}
				i++
			}
			l.add(IDENT, joinWords(atoms[start:i]), line)

		case atomNumber:
			l.add(NUMBER, a.text, line)
			i++
		case atomString:
			l.add(STRING, a.text, line)
			i++
		case atomLParen:
			l.add(LPAREN, "(", line)
			i++
		case atomRParen:
			l.add(RPAREN, ")", line)
			i++
		case atomComma:
			l.add(COMMA, ",", line)
			i++
		default:
			l.add(ILLEGAL, a.text, line)
			i++
		}
	}
}

// matchReserved tries to match the longest reserved phrase that begins at
// atoms[i]. It only considers consecutive word atoms. It returns the token
// type, how many words it consumed, and whether anything matched.
func matchReserved(atoms []atom, i int) (TokenType, int, bool) {
	var words []string
	for j := i; j < len(atoms) && j < i+maxPhraseWords && atoms[j].kind == atomWord; j++ {
		words = append(words, strings.ToLower(atoms[j].text))
	}
	for n := len(words); n >= 1; n-- {
		phrase := strings.Join(words[:n], " ")
		if tt, ok := reserved[phrase]; ok {
			return tt, n, true
		}
	}
	return ILLEGAL, 0, false
}

// joinWords rebuilds the original text of a run of word atoms, preserving the
// user's casing and script (so "Cart Total" stays "Cart Total").
func joinWords(atoms []atom) string {
	parts := make([]string, len(atoms))
	for i, a := range atoms {
		parts[i] = a.text
	}
	return strings.Join(parts, " ")
}
