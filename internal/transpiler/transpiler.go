// Package transpiler walks a hotgrin AST and writes the equivalent Go
// source. The Go is then compiled with the normal Go toolchain into a real
// executable — which is how hotgrin turns plain English into a program.
//
// It carries a small, best-effort type inferer so the generated Go is honest
// and clean: numbers stay int/float64, text stays string, records become
// structs, and hotgrin's own rules (e.g. "plus" joins text, "divided by"
// always gives a decimal) are respected.
package transpiler

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"
	"unicode"

	"github.com/hotgrin/hotgrin/internal/ast"
)

// --- types --------------------------------------------------------------

type ssType struct {
	kind string  // int, float, string, bool, list, record, unknown
	elem *ssType // list element type
	name string  // Go type name for a record
}

var (
	tInt     = ssType{kind: "int"}
	tFloat   = ssType{kind: "float"}
	tString  = ssType{kind: "string"}
	tBool    = ssType{kind: "bool"}
	tUnknown = ssType{kind: "unknown"}
)

func (t ssType) goName() string {
	switch t.kind {
	case "int":
		return "int"
	case "float":
		return "float64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "list":
		if t.elem != nil {
			return "[]" + t.elem.goName()
		}
		return "[]any"
	case "record":
		return t.name
	}
	return "any"
}

type recordInfo struct {
	goType  string
	order   []string          // ss field names, in order
	goField map[string]string // ss field -> Go field
	ftype   map[string]ssType // ss field -> type
}

type funcSig struct {
	goName      string
	ssParams    []string
	paramTypes  []ssType
	ret         ssType
	hasGiveBack bool
	fallible    bool // contains "give back problem ..." -> returns (T, error) in Go
}

// --- transpiler ---------------------------------------------------------

type Transpiler struct {
	prog *ast.Program

	records         map[string]recordInfo // keyed by Go type name
	funcs           map[string]funcSig    // keyed by Go func name
	scope           map[string]ssType     // variable Go name -> type (current function)
	declared        map[string]bool       // variable Go name -> already declared in Go
	scopeReads      map[string]bool       // Go names read somewhere in the current scope
	recordDeclOrder []string              // record Go types, in first-seen order

	needFmt         bool
	needStrings     bool
	needItemOf      bool
	needContains    bool
	needSync        bool
	needBackground  bool
	needTestStrings bool
	needErrors      bool
	needFlag        bool
	inFallible      bool // currently emitting a fallible function body
	fallibleRet     ssType

	tmp    int
	errors []string
}

func New(prog *ast.Program) *Transpiler {
	return &Transpiler{
		prog:     prog,
		records:  map[string]recordInfo{},
		funcs:    map[string]funcSig{},
		scope:    map[string]ssType{},
		declared: map[string]bool{},
	}
}

// Transpile returns formatted Go source (and any problems encountered).
func (t *Transpiler) Transpile() (string, string, []string) {
	t.collectRecords()
	t.seedTopLevelVars()
	t.collectFuncs()

	var typeDecls, funcs, mainBody, tests strings.Builder

	// Inputs are hoisted to the top of main() as command-line flags.
	var inputs []*ast.InputStmt
	for _, s := range t.prog.Statements {
		if in, ok := s.(*ast.InputStmt); ok {
			inputs = append(inputs, in)
		}
	}
	inputPreamble := t.emitInputs(inputs)

	// Reads across the top-level (main) scope, for the unused-variable guard.
	t.scopeReads = t.collectReads(t.prog.Statements)

	for _, s := range t.prog.Statements {
		switch n := s.(type) {
		case *ast.ActionStmt:
			funcs.WriteString(t.emitFunc(n))
			funcs.WriteString("\n\n")
		case *ast.TestStmt:
			tests.WriteString(t.emitTest(n))
			tests.WriteString("\n\n")
		case *ast.InputStmt:
			// already handled in the preamble
		case *ast.UseStmt:
			// libraries are resolved by the loader before transpiling
		default:
			mainBody.WriteString(t.emitStmt(s))
			mainBody.WriteString("\n")
		}
	}

	// Record type declarations (collected during emission).
	for _, r := range t.recordDeclOrder {
		ri := t.records[r]
		typeDecls.WriteString("type " + ri.goType + " struct {\n")
		for _, f := range ri.order {
			typeDecls.WriteString("\t" + ri.goField[f] + " " + ri.ftype[f].goName() + "\n")
		}
		typeDecls.WriteString("}\n\n")
	}

	var src strings.Builder
	src.WriteString("package main\n\n")

	imports := t.imports()
	if len(imports) > 0 {
		src.WriteString("import (\n")
		for _, im := range imports {
			src.WriteString("\t\"" + im + "\"\n")
		}
		src.WriteString(")\n\n")
	}

	src.WriteString(typeDecls.String())
	if t.needBackground {
		src.WriteString("var backgroundTasks sync.WaitGroup\n\n")
	}
	src.WriteString(t.helpers())
	src.WriteString(funcs.String())
	src.WriteString("func main() {\n")
	if t.needBackground {
		src.WriteString("defer backgroundTasks.Wait()\n")
	}
	src.WriteString(inputPreamble)
	src.WriteString(mainBody.String())
	src.WriteString("}\n")

	mainSrc := t.format(src.String())

	// Tests go into a separate _test.go file (same package main).
	testSrc := ""
	if tests.Len() > 0 {
		var ts strings.Builder
		ts.WriteString("package main\n\nimport (\n\t\"testing\"\n")
		if t.needTestStrings {
			ts.WriteString("\t\"strings\"\n")
		}
		ts.WriteString(")\n\n")
		ts.WriteString(tests.String())
		testSrc = t.format(ts.String())
	}

	return mainSrc, testSrc, t.errors
}

func (t *Transpiler) format(src string) string {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		t.errors = append(t.errors, "go formatting failed: "+err.Error())
		return src
	}
	return string(formatted)
}

func (t *Transpiler) imports() []string {
	var out []string
	if t.needErrors {
		out = append(out, "errors")
	}
	if t.needFlag {
		out = append(out, "flag")
	}
	if t.needFmt {
		out = append(out, "fmt")
	}
	if t.needStrings {
		out = append(out, "strings")
	}
	if t.needSync {
		out = append(out, "sync")
	}
	return out
}

func (t *Transpiler) helpers() string {
	var b strings.Builder
	if t.needItemOf {
		b.WriteString("func itemOf[T any](s []T, i int) T {\n")
		b.WriteString("\tvar zero T\n")
		b.WriteString("\tif i < 0 || i >= len(s) {\n\t\treturn zero\n\t}\n")
		b.WriteString("\treturn s[i]\n}\n\n")
	}
	if t.needContains {
		b.WriteString("func contains[T comparable](s []T, x T) bool {\n")
		b.WriteString("\tfor _, v := range s {\n\t\tif v == x {\n\t\t\treturn true\n\t\t}\n\t}\n\treturn false\n}\n\n")
	}
	return b.String()
}

// --- pre-passes ---------------------------------------------------------

func (t *Transpiler) collectRecords() {
	var walk func(stmts []ast.Stmt)
	walk = func(stmts []ast.Stmt) {
		for _, s := range stmts {
			switch n := s.(type) {
			case *ast.DescribeStmt:
				t.registerRecord(n)
			case *ast.IfStmt:
				for _, c := range n.Clauses {
					walk(c.Body)
				}
				walk(n.Else)
			case *ast.RepeatTimesStmt:
				walk(n.Body)
			case *ast.RepeatWhileStmt:
				walk(n.Body)
			case *ast.ForEachStmt:
				walk(n.Body)
			case *ast.ActionStmt:
				walk(n.Body)
			}
		}
	}
	walk(t.prog.Statements)
}

func (t *Transpiler) registerRecord(n *ast.DescribeStmt) {
	goType := titleFirst(sanitize(n.Name)) + "T"
	ri := recordInfo{
		goType:  goType,
		goField: map[string]string{},
		ftype:   map[string]ssType{},
	}
	for _, f := range n.Fields {
		ri.order = append(ri.order, f.Name)
		ri.goField[f.Name] = titleFirst(sanitize(f.Name))
		ri.ftype[f.Name] = t.typeOf(f.Value)
	}
	t.records[goType] = ri
	t.recordDeclOrder = append(t.recordDeclOrder, goType)
	// the record's value variable has this record type
	t.scope[sanitize(n.Name)] = ssType{kind: "record", name: goType}
}

// seedTopLevelVars pre-computes the type of each top-level variable so that
// function parameter inference (which reads call-site argument types) can
// resolve arguments that are variables, not just literals.
func (t *Transpiler) seedTopLevelVars() {
	for _, s := range t.prog.Statements {
		switch n := s.(type) {
		case *ast.SetStmt:
			t.scope[sanitize(n.Name)] = t.typeOf(n.Value)
		case *ast.ForEachStmt:
			it := t.typeOf(n.Iterable)
			if it.kind == "list" && it.elem != nil {
				t.scope[sanitize(n.Var)] = *it.elem
			}
		case *ast.InputStmt:
			t.scope[sanitize(n.Name)] = inputType(n.Type)
		}
	}
}

func (t *Transpiler) collectFuncs() {
	// 1. register each action's name + param names
	for _, s := range t.prog.Statements {
		a, ok := s.(*ast.ActionStmt)
		if !ok {
			continue
		}
		sig := funcSig{
			goName:     sanitize(a.Name),
			ssParams:   a.Params,
			paramTypes: make([]ssType, len(a.Params)),
			ret:        tUnknown,
		}
		for i := range sig.paramTypes {
			sig.paramTypes[i] = tUnknown
		}
		t.funcs[sanitize(a.Name)] = sig
	}

	// 2. infer param types from the first call site of each function
	t.walkCalls(t.prog.Statements)

	// 3. infer return type from the body (with params bound)
	for _, s := range t.prog.Statements {
		a, ok := s.(*ast.ActionStmt)
		if !ok {
			continue
		}
		key := sanitize(a.Name)
		sig := t.funcs[key]
		saved := t.scope
		t.scope = map[string]ssType{}
		for k, v := range saved {
			t.scope[k] = v
		}
		for i, p := range a.Params {
			t.scope[sanitize(p)] = sig.paramTypes[i]
		}
		if ty, ok := t.firstGiveBackType(a.Body); ok {
			sig.ret = ty
			sig.hasGiveBack = true
		}
		if hasProblemGiveBack(a.Body) {
			sig.fallible = true
			sig.hasGiveBack = true
		}
		t.funcs[key] = sig
		t.scope = saved
	}
}

// hasProblemGiveBack reports whether a body can return an error.
func hasProblemGiveBack(stmts []ast.Stmt) bool {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.GiveBackStmt:
			if n.Problem {
				return true
			}
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				if hasProblemGiveBack(c.Body) {
					return true
				}
			}
			if hasProblemGiveBack(n.Else) {
				return true
			}
		case *ast.RepeatTimesStmt:
			if hasProblemGiveBack(n.Body) {
				return true
			}
		case *ast.RepeatWhileStmt:
			if hasProblemGiveBack(n.Body) {
				return true
			}
		case *ast.ForEachStmt:
			if hasProblemGiveBack(n.Body) {
				return true
			}
		}
	}
	return false
}

func (t *Transpiler) walkCalls(stmts []ast.Stmt) {
	var walkExpr func(e ast.Expr)
	walkExpr = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.CallExpr:
			if sig, ok := t.funcs[sanitize(x.Name)]; ok && len(x.Args) == len(sig.paramTypes) {
				allUnknown := true
				for _, pt := range sig.paramTypes {
					if pt.kind != "unknown" {
						allUnknown = false
					}
				}
				if allUnknown {
					for i, a := range x.Args {
						sig.paramTypes[i] = t.typeOf(a)
					}
					t.funcs[sanitize(x.Name)] = sig
				}
			}
			for _, a := range x.Args {
				walkExpr(a)
			}
		case *ast.BinaryExpr:
			walkExpr(x.Left)
			walkExpr(x.Right)
		case *ast.FieldExpr:
			walkExpr(x.Target)
		case *ast.IndexExpr:
			walkExpr(x.Index)
			walkExpr(x.Target)
		case *ast.ListLit:
			for _, el := range x.Elements {
				walkExpr(el)
			}
		}
	}
	var walk func(stmts []ast.Stmt)
	walk = func(stmts []ast.Stmt) {
		for _, s := range stmts {
			switch n := s.(type) {
			case *ast.SayStmt:
				walkExpr(n.Value)
			case *ast.SetStmt:
				walkExpr(n.Value)
			case *ast.PutStmt:
				walkExpr(n.Value)
				walkExpr(n.List)
			case *ast.GiveBackStmt:
				walkExpr(n.Value)
			case *ast.ExprStmt:
				walkExpr(n.Value)
			case *ast.IncreaseStmt:
				walkExpr(n.Amount)
			case *ast.IfStmt:
				for _, c := range n.Clauses {
					walkExpr(c.Cond)
					walk(c.Body)
				}
				walk(n.Else)
			case *ast.RepeatTimesStmt:
				walk(n.Body)
			case *ast.RepeatWhileStmt:
				walk(n.Body)
			case *ast.ForEachStmt:
				walk(n.Body)
			case *ast.ActionStmt:
				walk(n.Body)
			case *ast.AtSameTimeStmt:
				for _, task := range n.Tasks {
					walkExpr(task.Call)
					if task.Into != nil {
						walkExpr(task.Into)
					}
				}
			case *ast.StartStmt:
				walkExpr(n.Call)
			case *ast.TestStmt:
				walk(n.Body)
			case *ast.TryStmt:
				walk(n.Body)
				walk(n.Handler)
			case *ast.ExpectStmt:
				walkExpr(n.Actual)
				if n.Expected != nil {
					walkExpr(n.Expected)
				}
			}
		}
	}
	walk(stmts)
}

func (t *Transpiler) firstGiveBackType(stmts []ast.Stmt) (ssType, bool) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.GiveBackStmt:
			if n.Problem {
				continue // a problem path doesn't define the value type
			}
			return t.typeOf(n.Value), true
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				if ty, ok := t.firstGiveBackType(c.Body); ok {
					return ty, true
				}
			}
			if ty, ok := t.firstGiveBackType(n.Else); ok {
				return ty, true
			}
		case *ast.RepeatTimesStmt:
			if ty, ok := t.firstGiveBackType(n.Body); ok {
				return ty, true
			}
		case *ast.RepeatWhileStmt:
			if ty, ok := t.firstGiveBackType(n.Body); ok {
				return ty, true
			}
		case *ast.ForEachStmt:
			if ty, ok := t.firstGiveBackType(n.Body); ok {
				return ty, true
			}
		}
	}
	return tUnknown, false
}

// --- type inference -----------------------------------------------------

func (t *Transpiler) typeOf(e ast.Expr) ssType {
	switch x := e.(type) {
	case *ast.NumberLit:
		if strings.Contains(x.Value, ".") {
			return tFloat
		}
		return tInt
	case *ast.StringLit:
		return tString
	case *ast.BoolLit:
		return tBool
	case *ast.NothingLit:
		return tUnknown
	case *ast.Identifier:
		if ty, ok := t.scope[sanitize(x.Name)]; ok {
			return ty
		}
		return tUnknown
	case *ast.CallExpr:
		if sig, ok := t.funcs[sanitize(x.Name)]; ok {
			return sig.ret
		}
		return tUnknown
	case *ast.FieldExpr:
		if x.Member == "count" {
			return tInt
		}
		tt := t.typeOf(x.Target)
		if tt.kind == "record" {
			if ft, ok := t.records[tt.name].ftype[x.Member]; ok {
				return ft
			}
		}
		return tUnknown
	case *ast.IndexExpr:
		tt := t.typeOf(x.Target)
		if tt.kind == "list" && tt.elem != nil {
			return *tt.elem
		}
		return tUnknown
	case *ast.ListLit:
		if len(x.Elements) > 0 {
			el := t.typeOf(x.Elements[0])
			return ssType{kind: "list", elem: &el}
		}
		el := tUnknown
		return ssType{kind: "list", elem: &el}
	case *ast.BinaryExpr:
		return t.binaryType(x)
	}
	return tUnknown
}

func (t *Transpiler) binaryType(x *ast.BinaryExpr) ssType {
	switch x.Op {
	case "and", "or", "=", "!=", ">", "<", ">=", "<=", "contains":
		return tBool
	case "/":
		return tFloat
	case "+":
		lt, rt := t.typeOf(x.Left), t.typeOf(x.Right)
		if lt.kind == "string" || rt.kind == "string" {
			return tString
		}
		if lt.kind == "float" || rt.kind == "float" {
			return tFloat
		}
		return tInt
	case "-", "*":
		lt, rt := t.typeOf(x.Left), t.typeOf(x.Right)
		if lt.kind == "float" || rt.kind == "float" {
			return tFloat
		}
		return tInt
	}
	return tUnknown
}

// --- statement emission -------------------------------------------------

func (t *Transpiler) emitFunc(n *ast.ActionStmt) string {
	sig := t.funcs[sanitize(n.Name)]
	saved := t.scope
	savedDecl := t.declared
	t.scope = map[string]ssType{}
	for k, v := range saved {
		t.scope[k] = v
	}
	t.declared = map[string]bool{}
	savedReads := t.scopeReads
	t.scopeReads = t.collectReads(n.Body)

	var b strings.Builder
	b.WriteString("func " + sig.goName + "(")
	for i, p := range n.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		gp := sanitize(p)
		b.WriteString(gp + " " + sig.paramTypes[i].goName())
		t.scope[gp] = sig.paramTypes[i]
		t.declared[gp] = true
	}
	b.WriteString(")")
	if sig.fallible {
		t.needErrors = true
		if sig.ret.kind != "" && sig.ret.kind != "void" {
			b.WriteString(" (" + sig.ret.goName() + ", error)")
		} else {
			b.WriteString(" error")
		}
	} else if sig.hasGiveBack {
		b.WriteString(" " + sig.ret.goName())
	}
	b.WriteString(" {\n")
	savedFallible := t.inFallible
	t.inFallible = sig.fallible
	t.fallibleRet = sig.ret
	for _, s := range n.Body {
		b.WriteString(t.emitStmt(s))
		b.WriteString("\n")
	}
	t.inFallible = savedFallible
	b.WriteString("}")

	t.scope = saved
	t.declared = savedDecl
	t.scopeReads = savedReads
	return b.String()
}

// emitTry turns a try/if-it-fails block into an error-returning closure plus a
// handler that runs when something inside failed.
// collectReads returns the set of Go variable names read anywhere in stmts
// (within this scope — it does not descend into action or test bodies, which
// are separate scopes). It powers the unused-variable guard in SetStmt.
func (t *Transpiler) collectReads(stmts []ast.Stmt) map[string]bool {
	reads := map[string]bool{}
	var ex func(e ast.Expr)
	ex = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.Identifier:
			reads[sanitize(x.Name)] = true
		case *ast.BinaryExpr:
			ex(x.Left)
			ex(x.Right)
		case *ast.CallExpr:
			for _, a := range x.Args {
				ex(a)
			}
		case *ast.FieldExpr:
			ex(x.Target)
		case *ast.IndexExpr:
			ex(x.Index)
			ex(x.Target)
		case *ast.ListLit:
			for _, el := range x.Elements {
				ex(el)
			}
		}
	}
	var walk func(ss []ast.Stmt)
	walk = func(ss []ast.Stmt) {
		for _, s := range ss {
			switch n := s.(type) {
			case *ast.SayStmt:
				ex(n.Value)
			case *ast.SetStmt:
				ex(n.Value)
			case *ast.IncreaseStmt:
				ex(n.Target)
				ex(n.Amount)
			case *ast.PutStmt:
				ex(n.Value)
				ex(n.List)
			case *ast.GiveBackStmt:
				ex(n.Value)
			case *ast.ExprStmt:
				ex(n.Value)
			case *ast.IfStmt:
				for _, c := range n.Clauses {
					ex(c.Cond)
					walk(c.Body)
				}
				walk(n.Else)
			case *ast.RepeatTimesStmt:
				ex(n.Count)
				walk(n.Body)
			case *ast.RepeatWhileStmt:
				ex(n.Cond)
				walk(n.Body)
			case *ast.ForEachStmt:
				ex(n.Iterable)
				walk(n.Body)
			case *ast.AtSameTimeStmt:
				for _, task := range n.Tasks {
					ex(task.Call)
					if task.Into != nil {
						ex(task.Into)
					}
				}
			case *ast.StartStmt:
				ex(n.Call)
			case *ast.TryStmt:
				walk(n.Body)
				walk(n.Handler)
			case *ast.ExpectStmt:
				ex(n.Actual)
				if n.Expected != nil {
					ex(n.Expected)
				}
			}
		}
	}
	walk(stmts)
	return reads
}

// emitInputs builds the command-line flag preamble for main().
func (t *Transpiler) emitInputs(inputs []*ast.InputStmt) string {
	if len(inputs) == 0 {
		return ""
	}
	t.needFlag = true
	var b strings.Builder
	for _, in := range inputs {
		fn, zero, _ := inputSpec(in.Type)
		def := zero
		if in.Default != nil {
			def = t.emitExpr(in.Default)
		}
		fv := sanitize(in.Name) + "Flag"
		flagName := strings.ReplaceAll(in.Name, " ", "-")
		b.WriteString(fmt.Sprintf("%s := flag.%s(%q, %s, %q)\n", fv, fn, flagName, def, in.Name))
	}
	b.WriteString("flag.Parse()\n")
	for _, in := range inputs {
		name := sanitize(in.Name)
		t.scope[name] = inputType(in.Type)
		t.declared[name] = true
		fv := name + "Flag"
		b.WriteString(name + " := *" + fv + "\n_ = " + name + "\n")
	}
	return b.String()
}

// inputSpec maps an input type word to its Go flag constructor and zero default.
func inputSpec(typ string) (flagFunc, zero string, ty ssType) {
	switch typ {
	case "whole":
		return "Int", "0", tInt
	case "number", "decimal":
		return "Float64", "0", tFloat
	case "truth":
		return "Bool", "false", tBool
	default: // text and anything unrecognised
		return "String", `""`, tString
	}
}

func inputType(typ string) ssType {
	_, _, ty := inputSpec(typ)
	return ty
}

func (t *Transpiler) emitTry(n *ast.TryStmt) string {
	var b strings.Builder
	b.WriteString("err := func() error {\n")
	for _, s := range n.Body {
		b.WriteString(t.emitTryStmt(s))
		b.WriteString("\n")
	}
	b.WriteString("return nil\n}()\n")
	b.WriteString("if err != nil {\n")
	t.scope["theProblem"] = tString
	t.declared["theProblem"] = true
	b.WriteString("theProblem := err.Error()\n_ = theProblem\n")
	for _, s := range n.Handler {
		b.WriteString(t.emitStmt(s))
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

// emitTryStmt emits a statement inside a try body, special-casing fallible
// calls so their error is checked and propagated to the handler.
func (t *Transpiler) emitTryStmt(s ast.Stmt) string {
	switch n := s.(type) {
	case *ast.SetStmt:
		if call, ok := n.Value.(*ast.CallExpr); ok {
			if sig, ok := t.funcs[sanitize(call.Name)]; ok && sig.fallible {
				v := sanitize(n.Name)
				t.scope[v] = sig.ret
				t.declared[v] = true
				return v + ", err := " + t.emitCall(call) + "\nif err != nil {\nreturn err\n}"
			}
		}
	case *ast.ExprStmt:
		if call, ok := n.Value.(*ast.CallExpr); ok {
			if sig, ok := t.funcs[sanitize(call.Name)]; ok && sig.fallible {
				return "if _, err := " + t.emitCall(call) + "; err != nil {\nreturn err\n}"
			}
		}
	case *ast.SayStmt:
		// "say <fallible call>" must check the error and print only the value,
		// otherwise Go's Println would show the raw (value, <nil>) pair.
		if call, ok := n.Value.(*ast.CallExpr); ok {
			if sig, ok := t.funcs[sanitize(call.Name)]; ok && sig.fallible {
				t.needFmt = true
				v := fmt.Sprintf("sayVal%d", t.uid())
				return v + ", err := " + t.emitCall(call) + "\nif err != nil {\nreturn err\n}\nfmt.Println(" + v + ")"
			}
		}
	}
	return t.emitStmt(s)
}

// zeroValue is the Go zero literal for a type, used for the value slot when a
// fallible function returns an error.
func zeroValue(ty ssType) string {
	switch ty.kind {
	case "int", "float":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "record":
		return ty.name + "{}"
	case "list":
		return "nil"
	}
	return "nil"
}

func (t *Transpiler) emitStmt(s ast.Stmt) string {
	switch n := s.(type) {
	case *ast.SayStmt:
		t.needFmt = true
		return "fmt.Println(" + t.emitExpr(n.Value) + ")"

	case *ast.SetStmt:
		name := sanitize(n.Name)
		val := t.emitExpr(n.Value)
		t.scope[name] = t.typeOf(n.Value)
		if t.declared[name] {
			return name + " = " + val
		}
		t.declared[name] = true
		decl := name + " := " + val
		// Go rejects a declared-but-unused variable. If nothing in this scope
		// ever reads it, add a blank use so the generated Go still compiles —
		// the Watcher has already offered the friendly "set but never used" note.
		if t.scopeReads != nil && !t.scopeReads[name] {
			decl += "\n_ = " + name
		}
		return decl

	case *ast.IncreaseStmt:
		op := "+="
		if n.Down {
			op = "-="
		}
		return t.emitExpr(n.Target) + " " + op + " " + t.emitExpr(n.Amount)

	case *ast.PutStmt:
		list := t.emitExpr(n.List)
		return list + " = append(" + list + ", " + t.emitExpr(n.Value) + ")"

	case *ast.GiveBackStmt:
		if n.Problem {
			t.needErrors = true
			return "return " + zeroValue(t.fallibleRet) + ", errors.New(" + t.emitExpr(n.Value) + ")"
		}
		if t.inFallible {
			return "return " + t.emitExpr(n.Value) + ", nil"
		}
		return "return " + t.emitExpr(n.Value)

	case *ast.TryStmt:
		return t.emitTry(n)

	case *ast.ExprStmt:
		return t.emitCall(n.Value)

	case *ast.AtSameTimeStmt:
		return t.emitAtSameTime(n)

	case *ast.StartStmt:
		t.needSync = true
		t.needBackground = true
		return "backgroundTasks.Add(1)\n" +
			"go func() { defer backgroundTasks.Done(); " + t.emitCall(n.Call) + " }()"

	case *ast.ExpectStmt:
		return t.emitExpect(n)

	case *ast.DescribeStmt:
		return t.emitDescribe(n)

	case *ast.IfStmt:
		return t.emitIf(n)

	case *ast.RepeatTimesStmt:
		i := t.tempName()
		var b strings.Builder
		b.WriteString("for " + i + " := 0; " + i + " < " + t.emitExpr(n.Count) + "; " + i + "++ {\n")
		b.WriteString(t.emitBody(n.Body))
		b.WriteString("}")
		return b.String()

	case *ast.RepeatWhileStmt:
		var b strings.Builder
		b.WriteString("for " + t.emitExpr(n.Cond) + " {\n")
		b.WriteString(t.emitBody(n.Body))
		b.WriteString("}")
		return b.String()

	case *ast.ForEachStmt:
		v := sanitize(n.Var)
		it := t.typeOf(n.Iterable)
		if it.kind == "list" && it.elem != nil {
			t.scope[v] = *it.elem
		} else {
			t.scope[v] = tUnknown
		}
		t.declared[v] = true
		var b strings.Builder
		b.WriteString("for _, " + v + " := range " + t.emitExpr(n.Iterable) + " {\n")
		b.WriteString(t.emitBody(n.Body))
		b.WriteString("}")
		return b.String()
	}
	return "// (unhandled statement)"
}

func (t *Transpiler) emitBody(stmts []ast.Stmt) string {
	var b strings.Builder
	for _, s := range stmts {
		b.WriteString(t.emitStmt(s))
		b.WriteString("\n")
	}
	return b.String()
}

func (t *Transpiler) emitDescribe(n *ast.DescribeStmt) string {
	goType := titleFirst(sanitize(n.Name)) + "T"
	ri := t.records[goType]
	name := sanitize(n.Name)
	t.scope[name] = ssType{kind: "record", name: goType}
	t.declared[name] = true
	var b strings.Builder
	b.WriteString(name + " := " + goType + "{\n")
	for _, f := range n.Fields {
		b.WriteString(ri.goField[f.Name] + ": " + t.emitExpr(f.Value) + ",\n")
	}
	b.WriteString("}")
	return b.String()
}

func (t *Transpiler) emitIf(n *ast.IfStmt) string {
	var b strings.Builder
	for i, c := range n.Clauses {
		if i == 0 {
			b.WriteString("if " + t.emitExpr(c.Cond) + " {\n")
		} else {
			b.WriteString("} else if " + t.emitExpr(c.Cond) + " {\n")
		}
		b.WriteString(t.emitBody(c.Body))
	}
	if n.Else != nil {
		b.WriteString("} else {\n")
		b.WriteString(t.emitBody(n.Else))
	}
	b.WriteString("}")
	return b.String()
}

// emitCall emits a call. A bare identifier is treated as a zero-argument call.
func (t *Transpiler) emitCall(e ast.Expr) string {
	if id, ok := e.(*ast.Identifier); ok {
		return sanitize(id.Name) + "()"
	}
	return t.emitExpr(e)
}

func (t *Transpiler) emitAtSameTime(n *ast.AtSameTimeStmt) string {
	t.needSync = true
	id := t.uid()
	wg := fmt.Sprintf("wg%d", id)
	mu := fmt.Sprintf("mu%d", id)

	hasInto := false
	for _, task := range n.Tasks {
		if task.Into != nil {
			hasInto = true
		}
	}

	var b strings.Builder
	b.WriteString("var " + wg + " sync.WaitGroup\n")
	if hasInto {
		b.WriteString("var " + mu + " sync.Mutex\n")
	}
	b.WriteString(fmt.Sprintf("%s.Add(%d)\n", wg, len(n.Tasks)))
	for _, task := range n.Tasks {
		b.WriteString("go func() {\n")
		b.WriteString("defer " + wg + ".Done()\n")
		if task.Into != nil {
			list := t.emitExpr(task.Into)
			b.WriteString("r := " + t.emitCall(task.Call) + "\n")
			b.WriteString(mu + ".Lock()\n")
			b.WriteString(list + " = append(" + list + ", r)\n")
			b.WriteString(mu + ".Unlock()\n")
		} else {
			b.WriteString(t.emitCall(task.Call) + "\n")
		}
		b.WriteString("}()\n")
	}
	b.WriteString(wg + ".Wait()")
	return b.String()
}

// --- test emission ------------------------------------------------------

func (t *Transpiler) emitTest(n *ast.TestStmt) string {
	saved := t.scope
	savedDecl := t.declared
	t.scope = map[string]ssType{}
	for k, v := range saved {
		t.scope[k] = v
	}
	t.declared = map[string]bool{}
	savedReads := t.scopeReads
	t.scopeReads = t.collectReads(n.Body)

	var b strings.Builder
	b.WriteString("func Test" + testName(n.Name) + "(t *testing.T) {\n")
	for _, s := range n.Body {
		b.WriteString(t.emitStmt(s))
		b.WriteString("\n")
	}
	b.WriteString("}")

	t.scope = saved
	t.declared = savedDecl
	t.scopeReads = savedReads
	return b.String()
}

func (t *Transpiler) emitExpect(n *ast.ExpectStmt) string {
	gv := fmt.Sprintf("got%d", t.uid())
	var b strings.Builder
	b.WriteString(gv + " := " + t.emitExpr(n.Actual) + "\n")

	switch n.Op {
	case "=", ">=", "<=", "<", ">":
		wv := fmt.Sprintf("want%d", t.uid())
		b.WriteString(wv + " := " + t.emitExpr(n.Expected) + "\n")
		goOp := n.Op
		if goOp == "=" {
			goOp = "=="
		}
		// Reconcile int vs float so a decimal result compares to an int literal.
		left, right := gv, wv
		at, et := t.typeOf(n.Actual), t.typeOf(n.Expected)
		if at.kind == "float" && et.kind == "int" {
			right = "float64(" + wv + ")"
		} else if at.kind == "int" && et.kind == "float" {
			left = "float64(" + gv + ")"
		}
		b.WriteString("if !(" + left + " " + goOp + " " + right + ") {\n")
		b.WriteString(fmt.Sprintf("t.Errorf(\"expected %%v %s %%v\", %s, %s)\n}", opWord(n.Op), gv, wv))
	case "contains":
		wv := fmt.Sprintf("want%d", t.uid())
		b.WriteString(wv + " := " + t.emitExpr(n.Expected) + "\n")
		if t.typeOf(n.Actual).kind == "string" {
			t.needTestStrings = true
			b.WriteString("if !strings.Contains(" + gv + ", " + wv + ") {\n")
		} else {
			t.needContains = true
			b.WriteString("if !contains(" + gv + ", " + wv + ") {\n")
		}
		b.WriteString(fmt.Sprintf("t.Errorf(\"expected %%v to contain %%v\", %s, %s)\n}", gv, wv))
	case "is-nothing":
		b.WriteString("_ = " + gv + " // 'to be nothing' is limited in this version")
	}
	return b.String()
}

func testName(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r != '_' && !isAlnum(r)
	})
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(titleFirst(p))
	}
	if b.Len() == 0 {
		return "Unnamed"
	}
	return b.String()
}

func opWord(op string) string {
	switch op {
	case "=":
		return "to be"
	case ">=":
		return "to be at least"
	case "<=":
		return "to be at most"
	case "<":
		return "to be less than"
	case ">":
		return "to be greater than"
	}
	return op
}

// --- expression emission ------------------------------------------------

func (t *Transpiler) emitExpr(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.NumberLit:
		return x.Value
	case *ast.StringLit:
		return strconv.Quote(x.Value)
	case *ast.BoolLit:
		if x.Value {
			return "true"
		}
		return "false"
	case *ast.NothingLit:
		return "nil"
	case *ast.Identifier:
		return sanitize(x.Name)
	case *ast.CallExpr:
		var args []string
		for _, a := range x.Args {
			args = append(args, t.emitExpr(a))
		}
		return sanitize(x.Name) + "(" + strings.Join(args, ", ") + ")"
	case *ast.FieldExpr:
		if x.Member == "count" {
			return "len(" + t.emitExpr(x.Target) + ")"
		}
		tt := t.typeOf(x.Target)
		if tt.kind == "record" {
			if gf, ok := t.records[tt.name].goField[x.Member]; ok {
				return t.emitExpr(x.Target) + "." + gf
			}
		}
		return t.emitExpr(x.Target) + "." + titleFirst(sanitize(x.Member))
	case *ast.IndexExpr:
		t.needItemOf = true
		return "itemOf(" + t.emitExpr(x.Target) + ", " + t.emitExpr(x.Index) + ")"
	case *ast.ListLit:
		ty := t.typeOf(x)
		var els []string
		for _, el := range x.Elements {
			els = append(els, t.emitExpr(el))
		}
		return ty.goName() + "{" + strings.Join(els, ", ") + "}"
	case *ast.BinaryExpr:
		return t.emitBinary(x)
	}
	return "/* expr? */"
}

func (t *Transpiler) emitBinary(x *ast.BinaryExpr) string {
	switch x.Op {
	case "and":
		return "(" + t.emitExpr(x.Left) + " && " + t.emitExpr(x.Right) + ")"
	case "or":
		return "(" + t.emitExpr(x.Left) + " || " + t.emitExpr(x.Right) + ")"
	case "=":
		return "(" + t.emitExpr(x.Left) + " == " + t.emitExpr(x.Right) + ")"
	case "!=", ">", "<", ">=", "<=":
		return "(" + t.emitExpr(x.Left) + " " + x.Op + " " + t.emitExpr(x.Right) + ")"
	case "contains":
		if t.typeOf(x.Left).kind == "string" {
			t.needStrings = true
			return "strings.Contains(" + t.emitExpr(x.Left) + ", " + t.emitExpr(x.Right) + ")"
		}
		t.needContains = true
		return "contains(" + t.emitExpr(x.Left) + ", " + t.emitExpr(x.Right) + ")"
	case "/":
		return "(float64(" + t.emitExpr(x.Left) + ") / float64(" + t.emitExpr(x.Right) + "))"
	case "+":
		lt, rt := t.typeOf(x.Left), t.typeOf(x.Right)
		if lt.kind == "string" || rt.kind == "string" {
			return "(" + t.stringify(x.Left, lt) + " + " + t.stringify(x.Right, rt) + ")"
		}
		return t.numericBinary("+", x.Left, x.Right)
	case "-", "*":
		return t.numericBinary(x.Op, x.Left, x.Right)
	}
	return "/* op? */"
}

func (t *Transpiler) numericBinary(op string, left, right ast.Expr) string {
	lt, rt := t.typeOf(left), t.typeOf(right)
	l, r := t.emitExpr(left), t.emitExpr(right)
	// If one side is float and the other int, promote the int side.
	if lt.kind == "float" && rt.kind == "int" {
		r = "float64(" + r + ")"
	} else if rt.kind == "float" && lt.kind == "int" {
		l = "float64(" + l + ")"
	}
	return "(" + l + " " + op + " " + r + ")"
}

// stringify turns an operand into a Go string expression for "plus" joins.
func (t *Transpiler) stringify(e ast.Expr, ty ssType) string {
	if ty.kind == "string" {
		return t.emitExpr(e)
	}
	t.needFmt = true
	return "fmt.Sprint(" + t.emitExpr(e) + ")"
}

func (t *Transpiler) tempName() string {
	n := fmt.Sprintf("i%d", t.tmp)
	t.tmp++
	return n
}

func (t *Transpiler) uid() int {
	n := t.tmp
	t.tmp++
	return n
}

// --- identifier sanitising ----------------------------------------------

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// sanitize turns a hotgrin name (which may contain spaces) into a valid Go
// identifier in camelCase: "cart total" -> "cartTotal".
func sanitize(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "_"
	}
	var b strings.Builder
	for i, p := range parts {
		if i == 0 {
			b.WriteString(lowerFirst(p))
		} else {
			b.WriteString(titleFirst(p))
		}
	}
	s := b.String()
	if goKeywords[s] {
		s += "_"
	}
	return s
}

func titleFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func lowerFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
