// Package parser turns a stream of tokens into an abstract syntax tree (AST).
//
// Statements are parsed by straightforward recursive descent (one method per
// kind of statement). Expressions are parsed by a small Pratt / precedence-
// climbing engine so that "times" binds tighter than "plus", comparisons sit
// below arithmetic, and "and"/"or" sit below that.
//
// This version covers the core spine (Mapping Part I). Constructs from Mapping
// Part II — concurrency, try/error handling, tests, CLI inputs, libraries, the
// Go escape hatch — are reported as "not yet supported" and come next.
package parser

import (
	"fmt"
	"strings"

	"github.com/hotgrin/hotgrin/internal/ast"
	"github.com/hotgrin/hotgrin/internal/lexer"
	"github.com/hotgrin/hotgrin/internal/units"
)

// Error is a parse problem, with the line so messages can point at the source.
type Error struct {
	Line    int
	Message string
}

func (e Error) String() string { return fmt.Sprintf("line %d: %s", e.Line, e.Message) }

// Parser holds the token stream and the current position.
type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []Error
}

// New builds a Parser from tokens (as produced by the lexer).
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse runs the parser and returns the program plus any errors.
func (p *Parser) Parse() (*ast.Program, []Error) {
	prog := &ast.Program{}
	p.skipNewlines()
	for !p.is(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.skipNewlines()
	}
	return prog, p.errors
}

// --- token helpers ------------------------------------------------------

func (p *Parser) cur() lexer.Token { return p.tokens[p.pos] }

func (p *Parser) peekAt(n int) lexer.Token {
	i := p.pos + n
	if i >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // EOF
	}
	return p.tokens[i]
}

func (p *Parser) is(t lexer.TokenType) bool { return p.cur().Type == t }

func (p *Parser) advance() lexer.Token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return t
}

func (p *Parser) accept(t lexer.TokenType) bool {
	if p.is(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(t lexer.TokenType) lexer.Token {
	if p.is(t) {
		return p.advance()
	}
	p.errorf("expected %s but found %s", t, p.cur().Type)
	return p.cur()
}

func (p *Parser) skipNewlines() {
	for p.is(lexer.NEWLINE) {
		p.advance()
	}
}

func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, Error{Line: p.cur().Line, Message: fmt.Sprintf(format, args...)})
}

// recover skips to the next newline so one bad line doesn't derail the rest.
func (p *Parser) recover() {
	for !p.is(lexer.NEWLINE) && !p.is(lexer.EOF) {
		p.advance()
	}
}

// --- statements ---------------------------------------------------------

func (p *Parser) parseStatement() ast.Stmt {
	line := p.cur().Line
	s := p.parseStatementInner()
	if s != nil {
		if sb, ok := s.(interface{ SetLine(int) }); ok {
			sb.SetLine(line)
		}
	}
	return s
}

func (p *Parser) parseStatementInner() ast.Stmt {
	switch p.cur().Type {
	case lexer.SAY:
		return p.parseSay()
	case lexer.SET:
		return p.parseSet()
	case lexer.INCREASE, lexer.DECREASE:
		return p.parseIncrease()
	case lexer.PUT:
		return p.parsePut()
	case lexer.GIVE_BACK:
		return p.parseGiveBack()
	case lexer.IF:
		return p.parseIf()
	case lexer.REPEAT:
		return p.parseRepeat()
	case lexer.ACTION:
		return p.parseAction()
	case lexer.DESCRIBE:
		return p.parseDescribe()
	case lexer.AT_THE_SAME_TIME:
		return p.parseAtSameTime()
	case lexer.START:
		return p.parseStart()
	case lexer.TEST:
		return p.parseTest()
	case lexer.EXPECT:
		return p.parseExpect()
	case lexer.TRY:
		return p.parseTry()
	case lexer.INPUT:
		return p.parseInput()
	case lexer.USE:
		return p.parseUse()
	case lexer.GO_BLOCK:
		return &ast.GoBlockStmt{Code: p.advance().Literal}
	case lexer.ASK:
		return p.parseAsk()
	case lexer.STOP_WITH_ERROR:
		return p.parseStop()
	default:
		// A bare expression statement, e.g. a call: greet with "AJ"
		expr := p.parseExpr(0)
		if expr == nil {
			p.recover()
			return nil
		}
		return &ast.ExprStmt{Value: expr}
	}
}

func (p *Parser) parseSay() ast.Stmt {
	p.expect(lexer.SAY)
	return &ast.SayStmt{Value: p.parseExpr(0)}
}

func (p *Parser) parseSet() ast.Stmt {
	p.expect(lexer.SET)
	name := p.expect(lexer.IDENT).Literal
	p.expect(lexer.TO)
	return &ast.SetStmt{Name: name, Value: p.parseExpr(0)}
}

func (p *Parser) parseIncrease() ast.Stmt {
	down := p.is(lexer.DECREASE)
	p.advance()              // increase / decrease
	target := p.parseExpr(0) // stops at BY (reserved, not an operator)
	p.expect(lexer.BY)
	amount := p.parseExpr(0)
	return &ast.IncreaseStmt{Target: target, Amount: amount, Down: down}
}

func (p *Parser) parsePut() ast.Stmt {
	p.expect(lexer.PUT)
	value := p.parseExpr(0) // stops at INTO
	p.expect(lexer.INTO)
	list := p.parseExpr(0)
	return &ast.PutStmt{Value: value, List: list}
}

func (p *Parser) parseGiveBack() ast.Stmt {
	p.expect(lexer.GIVE_BACK)
	// "give back problem <message>" returns an error from a fallible action.
	if p.is(lexer.IDENT) && p.cur().Literal == "problem" {
		p.advance()
		return &ast.GiveBackStmt{Problem: true, Value: p.parseExpr(0)}
	}
	return &ast.GiveBackStmt{Value: p.parseExpr(0)}
}

func (p *Parser) parseInput() ast.Stmt {
	p.expect(lexer.INPUT)
	name := p.expect(lexer.IDENT).Literal
	p.expect(lexer.AS)
	typ := p.expect(lexer.IDENT).Literal
	stmt := &ast.InputStmt{Name: name, Type: typ}
	if p.is(lexer.DEFAULT) {
		p.advance()
		stmt.Default = p.parseExpr(0)
	}
	return stmt
}

func (p *Parser) parseUse() ast.Stmt {
	p.expect(lexer.USE)
	stmt := &ast.UseStmt{}
	if p.is(lexer.STRING) {
		stmt.Path = p.advance().Literal
		return stmt
	}
	if p.is(lexer.IDENT) {
		stmt.Name = p.advance().Literal
		p.expect(lexer.FROM)
		stmt.Path = p.expect(lexer.STRING).Literal
		return stmt
	}
	p.errorf("expected a library path in quotes after 'use'")
	return stmt
}

func (p *Parser) parseAsk() ast.Stmt {
	p.expect(lexer.ASK)
	prompt := p.parseExpr(0)
	p.expect(lexer.INTO)
	name := p.expect(lexer.IDENT).Literal
	return &ast.AskStmt{Prompt: prompt, Var: name}
}

func (p *Parser) parseStop() ast.Stmt {
	p.expect(lexer.STOP_WITH_ERROR)
	return &ast.StopStmt{Message: p.parseExpr(0)}
}

func (p *Parser) parseTry() ast.Stmt {
	p.expect(lexer.TRY)
	p.expect(lexer.NEWLINE)
	stmt := &ast.TryStmt{Body: p.parseBlock()}
	if p.is(lexer.IF_IT_FAILS) {
		p.advance()
		p.skipNewlines()
		stmt.Handler = p.parseBlock()
	}
	p.expect(lexer.END)
	p.expect(lexer.TRY)
	return stmt
}

func (p *Parser) parseIf() ast.Stmt {
	stmt := &ast.IfStmt{}
	p.expect(lexer.IF)
	cond := p.parseExpr(0)
	p.expect(lexer.NEWLINE)
	body := p.parseBlock()
	stmt.Clauses = append(stmt.Clauses, ast.IfClause{Cond: cond, Body: body})

	for p.is(lexer.ELSE) {
		p.advance() // else
		if p.is(lexer.IF) {
			p.advance() // if  (this is "else if")
			c := p.parseExpr(0)
			p.expect(lexer.NEWLINE)
			b := p.parseBlock()
			stmt.Clauses = append(stmt.Clauses, ast.IfClause{Cond: c, Body: b})
			continue
		}
		// plain else
		p.expect(lexer.NEWLINE)
		stmt.Else = p.parseBlock()
		break
	}

	p.expect(lexer.END)
	p.expect(lexer.IF)
	return stmt
}

func (p *Parser) parseRepeat() ast.Stmt {
	p.expect(lexer.REPEAT)

	switch {
	case p.is(lexer.WHILE):
		p.advance()
		cond := p.parseExpr(0)
		p.expect(lexer.NEWLINE)
		body := p.parseBlock()
		p.expect(lexer.END)
		p.expect(lexer.REPEAT)
		return &ast.RepeatWhileStmt{Cond: cond, Body: body}

	case p.is(lexer.FOR_EACH):
		p.advance()
		varName := p.expect(lexer.IDENT).Literal
		p.expect(lexer.IN)
		iter := p.parseExpr(0)
		p.expect(lexer.NEWLINE)
		body := p.parseBlock()
		p.expect(lexer.END)
		p.expect(lexer.REPEAT)
		return &ast.ForEachStmt{Var: varName, Iterable: iter, Body: body}

	default: // repeat N times
		count := p.parseExpr(0) // the loop "times" is left for us (see Pratt rule)
		p.expect(lexer.TIMES)
		p.expect(lexer.NEWLINE)
		body := p.parseBlock()
		p.expect(lexer.END)
		p.expect(lexer.REPEAT)
		return &ast.RepeatTimesStmt{Count: count, Body: body}
	}
}

func (p *Parser) parseAction() ast.Stmt {
	p.expect(lexer.ACTION)
	name := p.expect(lexer.IDENT).Literal
	var params []string
	if p.accept(lexer.WITH) {
		params = append(params, p.expect(lexer.IDENT).Literal)
		for p.accept(lexer.COMMA) {
			params = append(params, p.expect(lexer.IDENT).Literal)
		}
	}
	p.expect(lexer.NEWLINE)
	body := p.parseBlock()
	p.expect(lexer.END)
	p.expect(lexer.ACTION)
	return &ast.ActionStmt{Name: name, Params: params, Body: body}
}

func (p *Parser) parseDescribe() ast.Stmt {
	p.expect(lexer.DESCRIBE)
	name := p.expect(lexer.IDENT).Literal
	p.expect(lexer.NEWLINE)
	stmt := &ast.DescribeStmt{Name: name}
	for {
		p.skipNewlines()
		if p.is(lexer.END) || p.is(lexer.EOF) {
			break
		}
		fieldName := p.expect(lexer.IDENT).Literal
		p.expect(lexer.IS)
		value := p.parseExpr(0)
		stmt.Fields = append(stmt.Fields, ast.DescribeField{Name: fieldName, Value: value})
		p.expect(lexer.NEWLINE)
	}
	p.expect(lexer.END)
	p.expect(lexer.DESCRIBE)
	return stmt
}

func (p *Parser) parseAtSameTime() ast.Stmt {
	p.expect(lexer.AT_THE_SAME_TIME)
	p.expect(lexer.NEWLINE)
	stmt := &ast.AtSameTimeStmt{}
	for {
		p.skipNewlines()
		if p.is(lexer.END) || p.is(lexer.EOF) {
			break
		}
		p.expect(lexer.DO)
		call := p.parseExpr(0) // stops at INTO (reserved, not an operator)
		var into ast.Expr
		if p.accept(lexer.INTO) {
			into = p.parseExpr(0)
		}
		stmt.Tasks = append(stmt.Tasks, ast.ConcurrentTask{Call: call, Into: into})
		p.expect(lexer.NEWLINE)
	}
	p.expect(lexer.END)
	p.expect(lexer.AT_THE_SAME_TIME)
	return stmt
}

func (p *Parser) parseStart() ast.Stmt {
	p.expect(lexer.START)
	return &ast.StartStmt{Call: p.parseExpr(0)}
}

func (p *Parser) parseTest() ast.Stmt {
	p.expect(lexer.TEST)
	name := p.expect(lexer.STRING).Literal
	p.expect(lexer.NEWLINE)
	body := p.parseBlock()
	p.expect(lexer.END)
	p.expect(lexer.TEST)
	return &ast.TestStmt{Name: name, Body: body}
}

func (p *Parser) parseExpect() ast.Stmt {
	p.expect(lexer.EXPECT)
	// Parse the actual value above comparison precedence, so the comparison
	// keyword (to be / contains / ...) is left for us to read.
	actual := p.parseExpr(4)
	stmt := &ast.ExpectStmt{Actual: actual}
	switch p.cur().Type {
	case lexer.TO_BE:
		p.advance()
		if p.is(lexer.NOTHING) {
			p.advance()
			stmt.Op = "is-nothing"
		} else {
			stmt.Op = "="
			stmt.Expected = p.parseExpr(0)
		}
	case lexer.TO_BE_AT_LEAST:
		p.advance()
		stmt.Op = ">="
		stmt.Expected = p.parseExpr(0)
	case lexer.TO_BE_AT_MOST:
		p.advance()
		stmt.Op = "<="
		stmt.Expected = p.parseExpr(0)
	case lexer.TO_BE_LESS_THAN:
		p.advance()
		stmt.Op = "<"
		stmt.Expected = p.parseExpr(0)
	case lexer.TO_BE_GREATER_THAN:
		p.advance()
		stmt.Op = ">"
		stmt.Expected = p.parseExpr(0)
	case lexer.CONTAINS:
		p.advance()
		stmt.Op = "contains"
		stmt.Expected = p.parseExpr(0)
	default:
		p.errorf("expected 'to be ...' or 'contains' in this check")
	}
	return stmt
}

// parseBlock parses statements until a block terminator (end / else / EOF).
func (p *Parser) parseBlock() []ast.Stmt {
	var stmts []ast.Stmt
	for {
		p.skipNewlines()
		if p.is(lexer.END) || p.is(lexer.ELSE) || p.is(lexer.IF_IT_FAILS) || p.is(lexer.EOF) {
			break
		}
		s := p.parseStatement()
		if s != nil {
			stmts = append(stmts, s)
		} else {
			// avoid an infinite loop on an unrecognised line
			if !p.is(lexer.NEWLINE) {
				p.recover()
			}
		}
	}
	return stmts
}

// --- expressions (Pratt) ------------------------------------------------

const ofPrec = 7 // "of" binds tighter than arithmetic

func infixPrec(t lexer.TokenType) int {
	switch t {
	case lexer.OR:
		return 1
	case lexer.AND:
		return 2
	case lexer.IS, lexer.IS_NOT, lexer.IS_GREATER_THAN, lexer.IS_LESS_THAN,
		lexer.IS_AT_LEAST, lexer.IS_AT_MOST, lexer.CONTAINS:
		return 3
	case lexer.ROUNDED_TO, lexer.IN:
		return 4 // looser than arithmetic: "a plus b rounded to 2" rounds the sum
	case lexer.PLUS, lexer.MINUS:
		return 5
	case lexer.TIMES, lexer.DIVIDED_BY:
		return 6
	case lexer.OF:
		return ofPrec
	}
	return 0
}

func opSymbol(t lexer.TokenType) string {
	switch t {
	case lexer.PLUS:
		return "+"
	case lexer.MINUS:
		return "-"
	case lexer.TIMES:
		return "*"
	case lexer.DIVIDED_BY:
		return "/"
	case lexer.IS:
		return "="
	case lexer.IS_NOT:
		return "!="
	case lexer.IS_GREATER_THAN:
		return ">"
	case lexer.IS_LESS_THAN:
		return "<"
	case lexer.IS_AT_LEAST:
		return ">="
	case lexer.IS_AT_MOST:
		return "<="
	case lexer.CONTAINS:
		return "contains"
	case lexer.ROUNDED_TO:
		return "rounded"
	case lexer.AND:
		return "and"
	case lexer.OR:
		return "or"
	}
	return "?"
}

func canStartExpr(t lexer.TokenType) bool {
	switch t {
	case lexer.NUMBER, lexer.STRING, lexer.TRUE, lexer.FALSE, lexer.NOTHING,
		lexer.IDENT, lexer.LPAREN, lexer.MINUS:
		return true
	}
	return false
}

func (p *Parser) parseExpr(minPrec int) ast.Expr {
	left := p.parsePrimary()
	if left == nil {
		return nil
	}
	for {
		op := p.cur().Type
		prec := infixPrec(op)
		if prec == 0 || prec < minPrec {
			break
		}
		// "times" before a newline/block is the repeat keyword, not multiply.
		if op == lexer.TIMES && !canStartExpr(p.peekAt(1).Type) {
			break
		}

		if op == lexer.OF {
			p.advance()
			right := p.parsePrimary()
			id, ok := left.(*ast.Identifier)
			if !ok {
				p.errorf("'of' needs a field name on its left")
				return left
			}
			left = &ast.FieldExpr{Member: id.Name, Target: right}
			continue
		}

		opLine := p.cur().Line
		if op == lexer.IN {
			p.advance()
			tok := p.expect(lexer.IDENT)
			u, ok := units.Lookup(tok.Literal)
			if !ok {
				p.errorf("'in' converts measurements, so %q must be a unit (kg, m, s, ...)", tok.Literal)
				return left
			}
			left = &ast.ConvExpr{X: left, Unit: u.Name, Line: opLine}
			continue
		}
		p.advance()
		right := p.parseExpr(prec + 1) // left-associative
		left = &ast.BinaryExpr{Op: opSymbol(op), Left: left, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parsePrimary() ast.Expr {
	line := p.cur().Line
	switch p.cur().Type {
	case lexer.NUMBER:
		num := p.advance().Literal
		// "129 kg" — a number directly followed by a unit word is a measurement
		if p.is(lexer.IDENT) {
			if u, ok := units.Lookup(p.cur().Literal); ok {
				p.advance()
				return &ast.UnitLit{Value: num, Unit: u.Name}
			}
		}
		return &ast.NumberLit{Value: num}
	case lexer.STRING:
		return &ast.StringLit{Value: p.advance().Literal}
	case lexer.TRUE:
		p.advance()
		return &ast.BoolLit{Value: true, Line: line}
	case lexer.FALSE:
		p.advance()
		return &ast.BoolLit{Value: false, Line: line}
	case lexer.NOTHING:
		p.advance()
		return &ast.NothingLit{}
	case lexer.LPAREN:
		p.advance()
		e := p.parseExpr(0)
		p.expect(lexer.RPAREN)
		return e
	case lexer.IDENT:
		name := p.advance().Literal

		// A multi-word name beginning "item ..." followed by 'of' is a
		// variable index: "item i of scores" (the lexer glued "item i").
		if strings.HasPrefix(name, "item ") && p.is(lexer.OF) {
			p.advance() // of
			target := p.parsePrimary()
			idx := &ast.Identifier{Name: strings.TrimPrefix(name, "item "), Line: line}
			return &ast.IndexExpr{Index: idx, Target: target}
		}
		// "item N of target" — but only when an index follows. If 'item' is
		// directly followed by 'of', it's a field name ("item of order").
		if name == "item" && !p.is(lexer.OF) {
			idx := p.parsePrimary()
			p.expect(lexer.OF)
			target := p.parsePrimary()
			return &ast.IndexExpr{Index: idx, Target: target}
		}
		// "list of a, b, c"
		if name == "list" && p.is(lexer.OF) {
			p.advance() // of
			return &ast.ListLit{Elements: p.parseCommaExprs()}
		}
		// "name with arg, arg"  (an action call)
		if p.is(lexer.WITH) {
			p.advance()
			return &ast.CallExpr{Name: name, Args: p.parseCommaExprs(), Line: line}
		}
		return &ast.Identifier{Name: name, Line: line}
	default:
		p.errorf("unexpected %s", p.cur().Type)
		return nil
	}
}

// parseCommaExprs parses one or more expressions separated by commas. Used for
// call arguments and list elements. Each element is parsed above comma level.
func (p *Parser) parseCommaExprs() []ast.Expr {
	var out []ast.Expr
	out = append(out, p.parseExpr(1))
	for p.accept(lexer.COMMA) {
		out = append(out, p.parseExpr(1))
	}
	return out
}
