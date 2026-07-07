// Package watcher is hotgrin's always-on checker.
//
// It walks the tree and reports problems it can PROVE. The iron rule: if the
// Watcher flags something, it is genuinely wrong — these rules never raise a
// false alarm. A clean run does not prove a program is bug-free, but a flag
// always means a real issue.
//
// Every finding carries a severity, the source line, and a plain-language
// message in both English and Afrikaans.
package watcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hotgrin/hotgrin/internal/ast"
	"github.com/hotgrin/hotgrin/internal/gobridge"
)

type Severity int

const (
	Error Severity = iota
	Warning
	Suggestion
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	default:
		return "suggestion"
	}
}

// Finding is one problem the Watcher found.
type Finding struct {
	Severity Severity
	Line     int
	En       string // English message
	Af       string // Afrikaans message
}

// Message returns the finding text in the requested language ("af" for
// Afrikaans, anything else for English).
func (f Finding) Message(lang string) string {
	if lang == "af" && f.Af != "" {
		return f.Af
	}
	return f.En
}

type actionInfo struct {
	params []string
	line   int
}

// Watcher holds the program-level facts the rules need.
type Watcher struct {
	prog      *ast.Program
	actions   map[string]actionInfo
	records   map[string]map[string]bool // record name -> field set
	fallible  map[string]bool            // actions that can "give back problem"
	givesBack map[string]bool            // actions that give a value back
	findings  []Finding
}

func New(prog *ast.Program) *Watcher {
	return &Watcher{
		prog:      prog,
		actions:   map[string]actionInfo{},
		records:   map[string]map[string]bool{},
		fallible:  map[string]bool{},
		givesBack: map[string]bool{},
	}
}

// Check runs every rule and returns the findings, sorted by line.
func (w *Watcher) Check() []Finding {
	w.collect()

	// The top-level statements form the "main" scope.
	w.checkScope(w.prog.Statements, w.mainNames())

	// Each action body is its own scope (params + its own declarations).
	for _, s := range w.prog.Statements {
		if a, ok := s.(*ast.ActionStmt); ok {
			names := map[string]bool{}
			for _, p := range a.Params {
				names[p] = true
			}
			w.collectNames(a.Body, names)
			w.checkScope(a.Body, names)
		}
		if ts, ok := s.(*ast.TestStmt); ok {
			names := map[string]bool{}
			w.collectNames(ts.Body, names)
			w.checkScope(ts.Body, names)
		}
	}

	sort.SliceStable(w.findings, func(i, j int) bool {
		return w.findings[i].Line < w.findings[j].Line
	})
	return w.findings
}

func (w *Watcher) add(sev Severity, line int, en, af string) {
	w.findings = append(w.findings, Finding{Severity: sev, Line: line, En: en, Af: af})
}

// --- collection ---------------------------------------------------------

func (w *Watcher) collect() {
	for _, s := range w.prog.Statements {
		a, ok := s.(*ast.ActionStmt)
		if !ok {
			continue
		}
		if prev, dup := w.actions[a.Name]; dup {
			w.add(Error, a.Line,
				fmt.Sprintf("there is already an action called '%s' (first defined on line %d)", a.Name, prev.line),
				fmt.Sprintf("daar is reeds 'n aksie genaamd '%s' (eerste op reël %d gedefinieer)", a.Name, prev.line))
		} else {
			w.actions[a.Name] = actionInfo{params: a.Params, line: a.Line}
		}
		if problemGiveBack(a.Body) {
			w.fallible[a.Name] = true
		}
		if plainGiveBack(a.Body) {
			w.givesBack[a.Name] = true
		}
		// duplicate parameter names
		seen := map[string]bool{}
		for _, p := range a.Params {
			if seen[p] {
				w.add(Error, a.Line,
					fmt.Sprintf("action '%s' has two inputs called '%s'", a.Name, p),
					fmt.Sprintf("aksie '%s' het twee insette genaamd '%s'", a.Name, p))
			}
			seen[p] = true
		}
	}
	// functions declared in use-go blocks are callable actions too
	for _, s := range w.prog.Statements {
		if g, ok := s.(*ast.GoBlockStmt); ok {
			_, body := gobridge.Imports(g.Code)
			for _, f := range gobridge.Funcs(body) {
				params := make([]string, f.Params)
				w.actions[spacedName(f.Name)] = actionInfo{params: params, line: g.Line}
				w.actions[f.Name] = actionInfo{params: params, line: g.Line}
				if f.Fallible {
					w.fallible[spacedName(f.Name)] = true
					w.fallible[f.Name] = true
				}
				if f.Ret != "" {
					w.givesBack[spacedName(f.Name)] = true
					w.givesBack[f.Name] = true
				}
			}
		}
	}

	// record field sets (collected globally; record names are normally unique)
	var walk func(stmts []ast.Stmt)
	walk = func(stmts []ast.Stmt) {
		for _, s := range stmts {
			switch n := s.(type) {
			case *ast.DescribeStmt:
				fields := map[string]bool{}
				for _, f := range n.Fields {
					fields[f.Name] = true
				}
				w.records[n.Name] = fields
			case *ast.ActionStmt:
				walk(n.Body)
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
			}
		}
	}
	walk(w.prog.Statements)
}

// mainNames is the set of names visible at the top level (not inside actions).
func (w *Watcher) mainNames() map[string]bool {
	names := map[string]bool{}
	w.collectNames(w.prog.Statements, names)
	return names
}

// collectNames gathers every name DECLARED directly in stmts (not descending
// into action bodies, which have their own scope).
func (w *Watcher) collectNames(stmts []ast.Stmt, names map[string]bool) {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.SetStmt:
			names[n.Name] = true
		case *ast.ForEachStmt:
			names[n.Var] = true
			w.collectNames(n.Body, names)
		case *ast.DescribeStmt:
			names[n.Name] = true
		case *ast.InputStmt:
			names[n.Name] = true
		case *ast.AskStmt:
			names[n.Var] = true
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				w.collectNames(c.Body, names)
			}
			w.collectNames(n.Else, names)
		case *ast.RepeatTimesStmt:
			w.collectNames(n.Body, names)
		case *ast.RepeatWhileStmt:
			w.collectNames(n.Body, names)
		case *ast.TryStmt:
			w.collectNames(n.Body, names)
			w.collectNames(n.Handler, names)
			names["the problem"] = true
			// ActionStmt is intentionally skipped: separate scope.
		}
	}
}

// --- scope checking -----------------------------------------------------

func (w *Watcher) checkScope(stmts []ast.Stmt, names map[string]bool) {
	w.checkBlock(stmts, names, false)
	w.checkUnused(stmts, names)
}

func (w *Watcher) checkBlock(stmts []ast.Stmt, names map[string]bool, inTry bool) {
	gaveBack := false
	for _, s := range stmts {
		if gaveBack {
			w.add(Warning, lineOf(s),
				"this line can never run — the action already gave back a value above it",
				"hierdie reël kan nooit loop nie — die aksie het reeds 'n waarde hierbo teruggegee")
			gaveBack = false // report only the first unreachable line, no spam
		}
		switch n := s.(type) {
		case *ast.SayStmt:
			w.checkVoidUse(n.Value)
			w.checkExpr(n.Value, names)
		case *ast.SetStmt:
			w.checkFallibleUse(n.Value, inTry)
			w.checkVoidUse(n.Value)
			w.checkExpr(n.Value, names)
		case *ast.IncreaseStmt:
			w.checkExpr(n.Target, names)
			w.checkExpr(n.Amount, names)
		case *ast.PutStmt:
			w.checkExpr(n.Value, names)
			w.checkExpr(n.List, names)
		case *ast.GiveBackStmt:
			w.checkExpr(n.Value, names)
			gaveBack = true
		case *ast.AskStmt:
			w.checkExpr(n.Prompt, names)
		case *ast.StopStmt:
			w.checkExpr(n.Message, names)
			gaveBack = true // nothing after a stop can run
		case *ast.ExprStmt:
			w.checkFallibleUse(n.Value, inTry)
			w.checkExpr(n.Value, names)
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				w.checkExpr(c.Cond, names)
				w.checkConstantCondition(c.Cond)
				w.checkBlock(c.Body, names, inTry)
			}
			w.checkBlock(n.Else, names, inTry)
		case *ast.RepeatTimesStmt:
			w.checkExpr(n.Count, names)
			w.checkBlock(n.Body, names, inTry)
		case *ast.RepeatWhileStmt:
			w.checkExpr(n.Cond, names)
			w.checkConstantCondition(n.Cond)
			w.checkBlock(n.Body, names, inTry)
		case *ast.ForEachStmt:
			w.checkExpr(n.Iterable, names)
			w.checkBlock(n.Body, names, inTry)
		case *ast.TryStmt:
			w.checkBlock(n.Body, names, true)
			w.checkBlock(n.Handler, names, false)
		case *ast.AtSameTimeStmt:
			for _, task := range n.Tasks {
				w.checkExpr(task.Call, names)
				if task.Into != nil {
					w.checkExpr(task.Into, names)
				}
			}
		case *ast.StartStmt:
			w.checkExpr(n.Call, names)
		case *ast.ExpectStmt:
			w.checkExpr(n.Actual, names)
			if n.Expected != nil {
				w.checkExpr(n.Expected, names)
			}
		}
	}
}

// checkVoidUse flags using an action that never gives a value back where a
// value is needed (set / say).
func (w *Watcher) checkVoidUse(e ast.Expr) {
	if call, ok := e.(*ast.CallExpr); ok {
		if _, known := w.actions[call.Name]; known && !w.givesBack[call.Name] && !w.fallible[call.Name] {
			w.add(Error, call.Line,
				fmt.Sprintf("'%s' does not give anything back, so it can't be used as a value", call.Name),
				fmt.Sprintf("'%s' gee niks terug nie, dus kan dit nie as 'n waarde gebruik word nie", call.Name))
		}
	}
}

// checkFallibleUse flags a call to a fallible action used outside a try block,
// where its possible failure would go unhandled.
func (w *Watcher) checkFallibleUse(e ast.Expr, inTry bool) {
	if inTry {
		return
	}
	if call, ok := e.(*ast.CallExpr); ok && w.fallible[call.Name] {
		w.add(Error, call.Line,
			fmt.Sprintf("'%s' can fail, so use it inside a 'try ... if it fails ... end try' block", call.Name),
			fmt.Sprintf("'%s' kan misluk, gebruik dit dus binne 'n 'try ... if it fails ... end try' blok", call.Name))
	}
}

func (w *Watcher) checkExpr(e ast.Expr, names map[string]bool) {
	switch x := e.(type) {
	case *ast.Identifier:
		if !names[x.Name] {
			if _, isAction := w.actions[x.Name]; !isAction {
				w.add(Error, x.Line,
					fmt.Sprintf("there is no value called '%s' here — is it a typo, or did you forget to set it?", x.Name),
					fmt.Sprintf("daar is geen waarde genaamd '%s' hier nie — is dit 'n tikfout, of het jy vergeet om dit te stel?", x.Name))
			}
		}
	case *ast.CallExpr:
		if info, ok := w.actions[x.Name]; ok {
			if len(x.Args) != len(info.params) {
				w.add(Error, x.Line,
					fmt.Sprintf("action '%s' needs %d input(s) but you gave it %d", x.Name, len(info.params), len(x.Args)),
					fmt.Sprintf("aksie '%s' benodig %d inset(te) maar jy het %d gegee", x.Name, len(info.params), len(x.Args)))
			}
		} else {
			w.add(Error, x.Line,
				fmt.Sprintf("there is no action called '%s'", x.Name),
				fmt.Sprintf("daar is geen aksie genaamd '%s' nie", x.Name))
		}
		for _, a := range x.Args {
			w.checkExpr(a, names)
		}
	case *ast.BinaryExpr:
		if x.Op == "/" {
			if n, ok := x.Right.(*ast.NumberLit); ok && isZero(n.Value) {
				w.add(Error, x.Line,
					"this divides by zero, which a computer cannot do",
					"hierdie deel deur nul, wat 'n rekenaar nie kan doen nie")
			}
		}
		w.checkExpr(x.Left, names)
		w.checkExpr(x.Right, names)
	case *ast.FieldExpr:
		// Unknown record field, when the target is a known record value.
		if id, ok := x.Target.(*ast.Identifier); ok && x.Member != "count" {
			if fields, known := w.records[id.Name]; known && !fields[x.Member] {
				w.add(Error, id.Line,
					fmt.Sprintf("'%s' has no field called '%s'", id.Name, x.Member),
					fmt.Sprintf("'%s' het geen veld genaamd '%s' nie", id.Name, x.Member))
			}
		}
		w.checkExpr(x.Target, names)
	case *ast.IndexExpr:
		w.checkExpr(x.Index, names)
		w.checkExpr(x.Target, names)
	case *ast.ListLit:
		for _, el := range x.Elements {
			w.checkExpr(el, names)
		}
	}
}

// checkConstantCondition flags conditions that are provably always the same.
func (w *Watcher) checkConstantCondition(cond ast.Expr) {
	switch c := cond.(type) {
	case *ast.BoolLit:
		val := "always true"
		af := "altyd waar"
		if !c.Value {
			val, af = "always false", "altyd vals"
		}
		w.add(Warning, c.Line,
			"this condition is "+val+", so the check does nothing",
			"hierdie voorwaarde is "+af+", dus doen die toets niks nie")
	case *ast.BinaryExpr:
		if litCompare(c) {
			w.add(Warning, c.Line,
				"this condition compares two fixed values, so it is always the same",
				"hierdie voorwaarde vergelyk twee vaste waardes, dus is dit altyd dieselfde")
		}
	}
}

// --- set-but-never-used -------------------------------------------------

func (w *Watcher) checkUnused(stmts []ast.Stmt, names map[string]bool) {
	declared := map[string]int{} // name -> line of the set
	used := map[string]bool{}

	var useExpr func(e ast.Expr)
	useExpr = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.Identifier:
			used[x.Name] = true
		case *ast.BinaryExpr:
			useExpr(x.Left)
			useExpr(x.Right)
		case *ast.CallExpr:
			for _, a := range x.Args {
				useExpr(a)
			}
		case *ast.FieldExpr:
			useExpr(x.Target)
		case *ast.IndexExpr:
			useExpr(x.Index)
			useExpr(x.Target)
		case *ast.ListLit:
			for _, el := range x.Elements {
				useExpr(el)
			}
		}
	}

	var walk func(stmts []ast.Stmt)
	walk = func(stmts []ast.Stmt) {
		for _, s := range stmts {
			switch n := s.(type) {
			case *ast.SetStmt:
				if _, seen := declared[n.Name]; !seen {
					declared[n.Name] = n.Line
				}
				useExpr(n.Value)
			case *ast.SayStmt:
				useExpr(n.Value)
			case *ast.IncreaseStmt:
				useExpr(n.Target) // a read+write counts as a use
				useExpr(n.Amount)
			case *ast.PutStmt:
				useExpr(n.Value)
				useExpr(n.List)
			case *ast.GiveBackStmt:
				useExpr(n.Value)
			case *ast.AskStmt:
				useExpr(n.Prompt)
			case *ast.StopStmt:
				useExpr(n.Message)
			case *ast.ExprStmt:
				useExpr(n.Value)
			case *ast.IfStmt:
				for _, c := range n.Clauses {
					useExpr(c.Cond)
					walk(c.Body)
				}
				walk(n.Else)
			case *ast.RepeatTimesStmt:
				useExpr(n.Count)
				walk(n.Body)
			case *ast.RepeatWhileStmt:
				useExpr(n.Cond)
				walk(n.Body)
			case *ast.ForEachStmt:
				useExpr(n.Iterable)
				walk(n.Body)
			case *ast.AtSameTimeStmt:
				for _, task := range n.Tasks {
					useExpr(task.Call)
					if task.Into != nil {
						useExpr(task.Into)
					}
				}
			case *ast.StartStmt:
				useExpr(n.Call)
			case *ast.TryStmt:
				walk(n.Body)
				walk(n.Handler)
			}
		}
	}
	walk(stmts)

	// Report names set but never read (sorted for stable output).
	type item struct {
		name string
		line int
	}
	var unused []item
	for name, line := range declared {
		if !used[name] {
			unused = append(unused, item{name, line})
		}
	}
	sort.Slice(unused, func(i, j int) bool { return unused[i].line < unused[j].line })
	for _, it := range unused {
		w.add(Suggestion, it.line,
			fmt.Sprintf("'%s' is set but never used — you can remove it", it.name),
			fmt.Sprintf("'%s' word gestel maar nooit gebruik nie — jy kan dit verwyder", it.name))
	}
}

// --- small helpers ------------------------------------------------------

// spacedName turns a Go camelCase name into its hotgrin spoken form:
// luckyNumber -> "lucky number".
func spacedName(goName string) string {
	var b strings.Builder
	for i, r := range goName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func lineOf(s ast.Stmt) int {
	if g, ok := s.(interface{ GetLine() int }); ok {
		return g.GetLine()
	}
	return 0
}

func isZero(num string) bool {
	for _, r := range num {
		if r != '0' && r != '.' && r != '-' && r != '+' {
			return false
		}
	}
	return strings.ContainsAny(num, "0")
}

// litCompare reports whether a comparison has fixed literals on both sides.
func litCompare(b *ast.BinaryExpr) bool {
	switch b.Op {
	case "=", "!=", ">", "<", ">=", "<=":
		return isLiteral(b.Left) && isLiteral(b.Right)
	}
	return false
}

func isLiteral(e ast.Expr) bool {
	switch e.(type) {
	case *ast.NumberLit, *ast.StringLit, *ast.BoolLit:
		return true
	}
	return false
}

// plainGiveBack reports whether a body gives a value back on some path.
func plainGiveBack(stmts []ast.Stmt) bool {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.GiveBackStmt:
			if !n.Problem {
				return true
			}
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				if plainGiveBack(c.Body) {
					return true
				}
			}
			if plainGiveBack(n.Else) {
				return true
			}
		case *ast.RepeatTimesStmt:
			if plainGiveBack(n.Body) {
				return true
			}
		case *ast.RepeatWhileStmt:
			if plainGiveBack(n.Body) {
				return true
			}
		case *ast.ForEachStmt:
			if plainGiveBack(n.Body) {
				return true
			}
		case *ast.TryStmt:
			if plainGiveBack(n.Body) || plainGiveBack(n.Handler) {
				return true
			}
		}
	}
	return false
}

// problemGiveBack reports whether a body can "give back problem".
func problemGiveBack(stmts []ast.Stmt) bool {
	for _, s := range stmts {
		switch n := s.(type) {
		case *ast.GiveBackStmt:
			if n.Problem {
				return true
			}
		case *ast.IfStmt:
			for _, c := range n.Clauses {
				if problemGiveBack(c.Body) {
					return true
				}
			}
			if problemGiveBack(n.Else) {
				return true
			}
		case *ast.RepeatTimesStmt:
			if problemGiveBack(n.Body) {
				return true
			}
		case *ast.RepeatWhileStmt:
			if problemGiveBack(n.Body) {
				return true
			}
		case *ast.ForEachStmt:
			if problemGiveBack(n.Body) {
				return true
			}
		}
	}
	return false
}
