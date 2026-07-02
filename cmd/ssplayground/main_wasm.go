//go:build js && wasm

// Command ssplayground compiles the SimpleScript pipeline to WebAssembly for
// the browser playground. It exposes one JavaScript function:
//
//	transpileSS(source) -> {
//	    goCode:      string,          // generated main.go ("" on parse errors)
//	    testCode:    string,          // generated main_test.go, if any tests
//	    parseErrors: [{line, message}],
//	    findings:    [{severity, line, en, af}],
//	}
//
// Everything runs client-side: no server, nothing leaves the page.
package main

import (
	"syscall/js"

	"github.com/hotgrin/simplescript/internal/lexer"
	"github.com/hotgrin/simplescript/internal/parser"
	"github.com/hotgrin/simplescript/internal/transpiler"
	"github.com/hotgrin/simplescript/internal/watcher"
)

func transpileSS(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf(map[string]any{"error": "transpileSS needs one argument"})
	}
	source := args[0].String()

	tokens := lexer.New(source).Tokenize()
	prog, perrs := parser.New(tokens).Parse()

	parseErrors := make([]any, 0, len(perrs))
	for _, e := range perrs {
		parseErrors = append(parseErrors, map[string]any{
			"line":    e.Line,
			"message": e.Message,
		})
	}

	findings := make([]any, 0)
	goCode, testCode := "", ""

	// Only check and transpile what parsed cleanly; on parse errors the tree
	// may be partial, and the parse messages are the ones that matter.
	if len(perrs) == 0 {
		for _, f := range watcher.New(prog).Check() {
			findings = append(findings, map[string]any{
				"severity": f.Severity.String(),
				"line":     f.Line,
				"en":       f.Message("en"),
				"af":       f.Message("af"),
			})
		}
		main, test, _ := transpiler.New(prog).Transpile()
		goCode, testCode = main, test
	}

	return js.ValueOf(map[string]any{
		"goCode":      goCode,
		"testCode":    testCode,
		"parseErrors": parseErrors,
		"findings":    findings,
	})
}

func main() {
	js.Global().Set("transpileSS", js.FuncOf(transpileSS))
	js.Global().Set("simplescriptReady", js.ValueOf(true))
	select {} // keep the Go runtime alive for future calls
}
