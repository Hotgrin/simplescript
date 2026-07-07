//go:build js && wasm

// Command hotplayground compiles the hotgrin pipeline to WebAssembly for
// the browser playground. It exposes one JavaScript function:
//
//	transpileHot(source) -> {
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

	"github.com/hotgrin/hotgrin/internal/lexer"
	"github.com/hotgrin/hotgrin/internal/parser"
	"github.com/hotgrin/hotgrin/internal/transpiler"
	"github.com/hotgrin/hotgrin/internal/watcher"
)

func transpileHot(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return js.ValueOf(map[string]any{"error": "transpileHot needs one argument"})
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
		main, test, terrs := transpiler.New(prog).Transpile()
		goCode, testCode = main, test
		for _, e := range terrs {
			findings = append(findings, map[string]any{
				"severity": "error", "line": 0, "en": e, "af": e,
			})
		}
		if len(terrs) > 0 {
			goCode, testCode = "", ""
		}
	}

	return js.ValueOf(map[string]any{
		"goCode":      goCode,
		"testCode":    testCode,
		"parseErrors": parseErrors,
		"findings":    findings,
	})
}

func main() {
	js.Global().Set("transpileHot", js.FuncOf(transpileHot))
	js.Global().Set("hotgrinReady", js.ValueOf(true))
	select {} // keep the Go runtime alive for future calls
}
