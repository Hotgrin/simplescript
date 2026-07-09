// Package ast defines the tree shapes that a hotgrin program parses into.
//
// Every node can print itself as a small S-expression, which makes the tree
// easy to read in the demo tool and easy to assert on in tests. Statements and
// the key expressions also carry the source Line they came from, so the checker
// and later stages can point messages back at the user's hotgrin line.
package ast

import (
	"fmt"
	"strings"
)

// Node is anything in the tree.
type Node interface{ String() string }

// Stmt is a statement (an action the program takes).
type Stmt interface {
	Node
	stmtNode()
}

// Expr is an expression (something that produces a value).
type Expr interface {
	Node
	exprNode()
}

// stmtBase carries the source line for any statement. The parser sets it at a
// single choke point, so every statement gets a line for free.
type stmtBase struct{ Line int }

func (b *stmtBase) SetLine(n int) { b.Line = n }
func (b *stmtBase) GetLine() int  { return b.Line }

// ind indents every line of s by two spaces (for nesting children).
func ind(s string) string { return strings.ReplaceAll(s, "\n", "\n  ") }

// writeBlock appends a block of statements, each on its own indented line.
func writeBlock(b *strings.Builder, stmts []Stmt) {
	for _, s := range stmts {
		b.WriteString("\n  ")
		b.WriteString(ind(s.String()))
	}
}

// --- Program -----------------------------------------------------------

type Program struct{ Statements []Stmt }

func (p *Program) String() string {
	var b strings.Builder
	b.WriteString("(program")
	writeBlock(&b, p.Statements)
	b.WriteString(")")
	return b.String()
}

// --- Expressions -------------------------------------------------------

type NumberLit struct{ Value string }

func (n *NumberLit) exprNode()      {}
func (n *NumberLit) String() string { return "(num " + n.Value + ")" }

type StringLit struct{ Value string }

func (s *StringLit) exprNode()      {}
func (s *StringLit) String() string { return fmt.Sprintf("(str %q)", s.Value) }

type BoolLit struct {
	Line  int
	Value bool
}

func (bl *BoolLit) exprNode()      {}
func (bl *BoolLit) String() string { return fmt.Sprintf("(bool %v)", bl.Value) }

type NothingLit struct{}

func (n *NothingLit) exprNode()      {}
func (n *NothingLit) String() string { return "nothing" }

type Identifier struct {
	Line int
	Name string
}

func (i *Identifier) exprNode()      {}
func (i *Identifier) String() string { return "(id " + i.Name + ")" }

// BinaryExpr covers math, comparison, and logic: (+ a b), (>= a b), (and a b).
type BinaryExpr struct {
	Line        int
	Op          string // a printable symbol: + - * / = != > < >= <= contains and or
	Left, Right Expr
}

func (be *BinaryExpr) exprNode() {}
func (be *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Op, be.Left, be.Right)
}

// CallExpr is calling an action: grade with "Adriaan", 82.
type CallExpr struct {
	Line int
	Name string
	Args []Expr
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) String() string {
	var b strings.Builder
	b.WriteString("(call " + c.Name)
	for _, a := range c.Args {
		b.WriteString(" " + a.String())
	}
	b.WriteString(")")
	return b.String()
}

// FieldExpr is "member of target": location of Adriaan, count of scores.
type FieldExpr struct {
	Member string
	Target Expr
}

func (f *FieldExpr) exprNode()      {}
func (f *FieldExpr) String() string { return "(of " + f.Member + " " + f.Target.String() + ")" }

// IndexExpr is "item N of target": item 0 of scores.
type IndexExpr struct {
	Index  Expr
	Target Expr
}

func (ix *IndexExpr) exprNode() {}
func (ix *IndexExpr) String() string {
	return "(item " + ix.Index.String() + " " + ix.Target.String() + ")"
}

// UnitLit is a measured value: 129 kg, 2.5 km.
type UnitLit struct {
	Value string
	Unit  string // canonical unit name
}

func (u *UnitLit) exprNode()      {}
func (u *UnitLit) String() string { return "(unit " + u.Value + " " + u.Unit + ")" }

// ConvExpr converts a measured value to another unit: weight in g.
type ConvExpr struct {
	Line int
	X    Expr
	Unit string
}

func (c *ConvExpr) exprNode()      {}
func (c *ConvExpr) String() string { return "(in " + c.X.String() + " " + c.Unit + ")" }

// ListLit is "list of a, b, c".
type ListLit struct{ Elements []Expr }

func (l *ListLit) exprNode() {}
func (l *ListLit) String() string {
	var b strings.Builder
	b.WriteString("(list")
	for _, e := range l.Elements {
		b.WriteString(" " + e.String())
	}
	b.WriteString(")")
	return b.String()
}

// --- Statements --------------------------------------------------------

type SayStmt struct {
	stmtBase
	Value Expr
}

func (s *SayStmt) stmtNode()      {}
func (s *SayStmt) String() string { return "(say " + s.Value.String() + ")" }

type SetStmt struct {
	stmtBase
	Name  string
	Value Expr
}

func (s *SetStmt) stmtNode()      {}
func (s *SetStmt) String() string { return fmt.Sprintf("(set %q %s)", s.Name, s.Value) }

type IncreaseStmt struct {
	stmtBase
	Target Expr
	Amount Expr
	Down   bool // true for decrease
}

func (s *IncreaseStmt) stmtNode() {}
func (s *IncreaseStmt) String() string {
	verb := "increase"
	if s.Down {
		verb = "decrease"
	}
	return fmt.Sprintf("(%s %s %s)", verb, s.Target, s.Amount)
}

// PutStmt is "put value into list".
type PutStmt struct {
	stmtBase
	Value Expr
	List  Expr
}

func (s *PutStmt) stmtNode() {}
func (s *PutStmt) String() string {
	return "(put " + s.Value.String() + " into " + s.List.String() + ")"
}

type GiveBackStmt struct {
	stmtBase
	Value   Expr
	Problem bool // true for "give back problem <message>"
}

func (s *GiveBackStmt) stmtNode() {}
func (s *GiveBackStmt) String() string {
	if s.Problem {
		return "(give-back-problem " + s.Value.String() + ")"
	}
	return "(give-back " + s.Value.String() + ")"
}

// UseStmt imports a hotgrin library file. Name is an optional label.
type UseStmt struct {
	stmtBase
	Name string
	Path string
}

func (s *UseStmt) stmtNode() {}
func (s *UseStmt) String() string {
	if s.Name != "" {
		return fmt.Sprintf("(use %q from %q)", s.Name, s.Path)
	}
	return fmt.Sprintf("(use %q)", s.Path)
}

// InputStmt declares a command-line input: input <name> as <type> [default <v>].
type InputStmt struct {
	stmtBase
	Name    string
	Type    string // text, number, whole, decimal, truth
	Default Expr   // nil if no default given
}

func (s *InputStmt) stmtNode() {}
func (s *InputStmt) String() string {
	if s.Default != nil {
		return fmt.Sprintf("(input %q %s default %s)", s.Name, s.Type, s.Default)
	}
	return fmt.Sprintf("(input %q %s)", s.Name, s.Type)
}

// GoBlockStmt is a verbatim block of Go from 'use go ... end go'. Its funcs
// become callable actions; single-line imports are merged with generated ones.
type GoBlockStmt struct {
	stmtBase
	Code string
}

func (s *GoBlockStmt) stmtNode()      {}
func (s *GoBlockStmt) String() string { return "(use-go ...)" }

// SetFieldStmt writes a record field: set price of order to 249.
type SetFieldStmt struct {
	stmtBase
	Member string
	Target Expr
	Value  Expr
}

func (s *SetFieldStmt) stmtNode() {}
func (s *SetFieldStmt) String() string {
	return "(set-field " + s.Member + " of " + s.Target.String() + " " + s.Value.String() + ")"
}

// AskStmt prompts the person running the program and stores their answer
// (as text) under Var: ask "What is your name?" into name.
type AskStmt struct {
	stmtBase
	Prompt Expr
	Var    string
}

func (s *AskStmt) stmtNode()      {}
func (s *AskStmt) String() string { return "(ask " + s.Prompt.String() + " into " + s.Var + ")" }

// StopStmt ends the program immediately with an error message and a non-zero
// exit code: stop with error "no input file given".
type StopStmt struct {
	stmtBase
	Message Expr
}

func (s *StopStmt) stmtNode()      {}
func (s *StopStmt) String() string { return "(stop-with-error " + s.Message.String() + ")" }

// TryStmt runs Body; if a fallible step fails, Handler runs with the error
// available as the value "the problem".
type TryStmt struct {
	stmtBase
	Body    []Stmt
	Handler []Stmt
}

func (s *TryStmt) stmtNode() {}
func (s *TryStmt) String() string {
	var b strings.Builder
	b.WriteString("(try")
	writeBlock(&b, s.Body)
	b.WriteString("\n  (if-it-fails")
	var hb strings.Builder
	writeBlock(&hb, s.Handler)
	b.WriteString(ind(hb.String()))
	b.WriteString("))")
	return b.String()
}

// ExprStmt is a bare expression used as a statement, e.g. a call: greet with "AJ".
type ExprStmt struct {
	stmtBase
	Value Expr
}

func (s *ExprStmt) stmtNode()      {}
func (s *ExprStmt) String() string { return s.Value.String() }

// IfClause is one condition + body within an if statement.
type IfClause struct {
	Cond Expr
	Body []Stmt
}

type IfStmt struct {
	stmtBase
	Clauses []IfClause // the if and any else-if clauses, in order
	Else    []Stmt     // may be nil
}

func (s *IfStmt) stmtNode() {}
func (s *IfStmt) String() string {
	var b strings.Builder
	b.WriteString("(if")
	for _, c := range s.Clauses {
		b.WriteString("\n  (clause " + ind(c.Cond.String()))
		var bb strings.Builder
		writeBlock(&bb, c.Body)
		b.WriteString(ind(bb.String()))
		b.WriteString(")")
	}
	if s.Else != nil {
		b.WriteString("\n  (else")
		var bb strings.Builder
		writeBlock(&bb, s.Else)
		b.WriteString(ind(bb.String()))
		b.WriteString(")")
	}
	b.WriteString(")")
	return b.String()
}

type RepeatTimesStmt struct {
	stmtBase
	Count Expr
	Body  []Stmt
}

func (s *RepeatTimesStmt) stmtNode() {}
func (s *RepeatTimesStmt) String() string {
	var b strings.Builder
	b.WriteString("(repeat-times " + s.Count.String())
	writeBlock(&b, s.Body)
	b.WriteString(")")
	return b.String()
}

type RepeatWhileStmt struct {
	stmtBase
	Cond Expr
	Body []Stmt
}

func (s *RepeatWhileStmt) stmtNode() {}
func (s *RepeatWhileStmt) String() string {
	var b strings.Builder
	b.WriteString("(repeat-while " + s.Cond.String())
	writeBlock(&b, s.Body)
	b.WriteString(")")
	return b.String()
}

type ForEachStmt struct {
	stmtBase
	Var      string
	Iterable Expr
	Body     []Stmt
}

func (s *ForEachStmt) stmtNode() {}
func (s *ForEachStmt) String() string {
	var b strings.Builder
	b.WriteString("(for-each " + s.Var + " " + s.Iterable.String())
	writeBlock(&b, s.Body)
	b.WriteString(")")
	return b.String()
}

type ActionStmt struct {
	stmtBase
	Name   string
	Params []string
	Body   []Stmt
}

func (s *ActionStmt) stmtNode() {}
func (s *ActionStmt) String() string {
	var b strings.Builder
	b.WriteString("(action " + s.Name + " (params")
	for _, p := range s.Params {
		b.WriteString(" " + p)
	}
	b.WriteString(")")
	writeBlock(&b, s.Body)
	b.WriteString(")")
	return b.String()
}

type DescribeField struct {
	Name  string
	Value Expr
}

type DescribeStmt struct {
	stmtBase
	Name   string
	Fields []DescribeField
}

func (s *DescribeStmt) stmtNode() {}
func (s *DescribeStmt) String() string {
	var b strings.Builder
	b.WriteString("(describe " + s.Name)
	for _, f := range s.Fields {
		b.WriteString(fmt.Sprintf("\n  (field %q %s)", f.Name, f.Value))
	}
	b.WriteString(")")
	return b.String()
}

// ConcurrentTask is one "do <call>" line, optionally "into <list>".
type ConcurrentTask struct {
	Call Expr
	Into Expr // nil unless the result is collected into a list
}

// AtSameTimeStmt runs its tasks concurrently and waits for all of them.
type AtSameTimeStmt struct {
	stmtBase
	Tasks []ConcurrentTask
}

func (s *AtSameTimeStmt) stmtNode() {}
func (s *AtSameTimeStmt) String() string {
	var b strings.Builder
	b.WriteString("(at-same-time")
	for _, task := range s.Tasks {
		if task.Into != nil {
			b.WriteString("\n  (do " + task.Call.String() + " into " + task.Into.String() + ")")
		} else {
			b.WriteString("\n  (do " + task.Call.String() + ")")
		}
	}
	b.WriteString(")")
	return b.String()
}

// StartStmt launches a call as a background task (fire-and-forget).
type StartStmt struct {
	stmtBase
	Call Expr
}

func (s *StartStmt) stmtNode()      {}
func (s *StartStmt) String() string { return "(start " + s.Call.String() + ")" }

// ExpectStmt is one assertion inside a test: expect X to be Y.
// Op is one of: = >= <= < > contains is-nothing. Expected is nil for is-nothing.
type ExpectStmt struct {
	stmtBase
	Actual   Expr
	Op       string
	Expected Expr
}

func (s *ExpectStmt) stmtNode() {}
func (s *ExpectStmt) String() string {
	if s.Expected == nil {
		return "(expect " + s.Actual.String() + " " + s.Op + ")"
	}
	return "(expect " + s.Actual.String() + " " + s.Op + " " + s.Expected.String() + ")"
}

// TestStmt is a named, first-class test.
type TestStmt struct {
	stmtBase
	Name string
	Body []Stmt
}

func (s *TestStmt) stmtNode() {}
func (s *TestStmt) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("(test %q", s.Name))
	writeBlock(&b, s.Body)
	b.WriteString(")")
	return b.String()
}
