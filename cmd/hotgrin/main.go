// Command hotgrin is the friendly front door to the whole toolchain.
//
//	hotgrin run     hello.hot      run a program
//	hotgrin build   hello.hot      build a standalone program
//	hotgrin build --windows x.hot  build a Windows .exe
//	hotgrin check   hello.hot      check a program for problems
//	hotgrin reveal  hello.hot      show the Go a program turns into
//	hotgrin help                  show help
//
// It runs the Watcher before running or building, so a beginner sees friendly
// hotgrin messages — never raw Go errors.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hotgrin/hotgrin/internal/ast"
	"github.com/hotgrin/hotgrin/internal/loader"
	"github.com/hotgrin/hotgrin/internal/transpiler"
	"github.com/hotgrin/hotgrin/internal/watcher"
)

const version = "hotgrin 0.5.9"

func main() {
	af := false
	windows := false
	var rest []string
	for _, a := range os.Args[1:] {
		switch a {
		case "--af", "-af", "--afrikaans":
			af = true
		case "--windows", "-windows", "--exe":
			windows = true
		default:
			rest = append(rest, a)
		}
	}

	lang := "en"
	if af {
		lang = "af"
	}

	if len(rest) == 0 {
		printHelp()
		return
	}

	cmd := rest[0]
	file := ""
	if len(rest) > 1 {
		file = rest[1]
	}
	var progArgs []string
	if len(rest) > 2 {
		progArgs = rest[2:]
	}

	switch cmd {
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		fmt.Println(version)
	case "check":
		cmdCheck(file, lang)
	case "run":
		cmdRun(file, lang, progArgs)
	case "test":
		cmdTest(file, lang)
	case "build":
		cmdBuild(file, lang, windows)
	case "reveal":
		cmdReveal(file, lang)
	default:
		fmt.Fprintf(os.Stderr, "I don't know the command %q. Try: hotgrin help\n", cmd)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`hotgrin - a programming language that reads like plain English.

Usage:
  hotgrin run     <file.hot>     Run a program (extra --flags pass to it)
  hotgrin test    <file.hot>     Run the tests in a program
  hotgrin build   <file.hot>     Build a standalone program you can share
  hotgrin check   <file.hot>     Check a program for problems
  hotgrin reveal  <file.hot>     Show the Go code a program becomes
  hotgrin help                  Show this help
  hotgrin version               Show the version

Options:
  --windows    With 'build', make a Windows .exe
  --af         Show messages in Afrikaans

Examples:
  hotgrin run hello.hot
  hotgrin build --windows hello.hot
  hotgrin check --af hello.hot
`)
}

// load reads, lexes, and parses a file. It exits with a friendly message on a
// missing file or parse problems.
func load(file string) *ast.Program {
	if file == "" {
		fmt.Fprintln(os.Stderr, "Please tell me which file, e.g.  hotgrin run hello.hot")
		os.Exit(1)
	}
	if _, err := os.Stat(file); err != nil {
		fmt.Fprintf(os.Stderr, "I couldn't open %q. Is the name right?\n", file)
		os.Exit(1)
	}
	prog, errs := loader.LoadFile(file)
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "I couldn't understand part of your program:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e)
		}
		os.Exit(1)
	}
	return prog
}

// report runs the Watcher and prints findings. It returns true if there were
// any errors (problems that must be fixed before running or building).
func report(prog *ast.Program, lang string) bool {
	findings := watcher.New(prog).Check()
	hasError := false
	for _, f := range findings {
		label := "idea   "
		switch f.Severity {
		case watcher.Error:
			label = "error  "
			hasError = true
		case watcher.Warning:
			label = "warning"
		}
		fmt.Fprintf(os.Stderr, "  %s line %d: %s\n", label, f.Line, f.Message(lang))
	}
	return hasError
}

func cmdCheck(file, lang string) {
	prog := load(file)
	if report(prog, lang) {
		os.Exit(1)
	}
	if len(watcher.New(prog).Check()) == 0 {
		fmt.Println("All good - I found no problems.")
	}
}

func cmdReveal(file, lang string) {
	prog := load(file)
	goSrc, _, terrs := transpiler.New(prog).Transpile()
	reportTranspile(terrs)
	fmt.Print(goSrc)
}

// reportTranspile prints problems the transpiler itself can prove (like
// mixing kg and km) in the same friendly voice, and stops before Go ever
// sees the program.
func reportTranspile(errs []string) {
	if len(errs) == 0 {
		return
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "  error   "+e)
	}
	fmt.Fprintln(os.Stderr, "\nI found problems above, so I stopped. Fix those and try again.")
	os.Exit(1)
}

func cmdRun(file, lang string, progArgs []string) {
	prog := load(file)
	if report(prog, lang) {
		fmt.Fprintln(os.Stderr, "\nI found problems above, so I didn't run it. Fix those and try again.")
		os.Exit(1)
	}
	goSrc, _, terrs := transpiler.New(prog).Transpile()
	reportTranspile(terrs)
	dir := tempModule(goSrc)
	defer os.RemoveAll(dir)

	if !haveGo() {
		os.Exit(1)
	}
	// Build to a temp binary and run it directly: 'go run' would append its
	// own "exit status 1" noise to stderr, which breaks our no-raw-Go promise.
	bin := filepath.Join(dir, "program")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = dir
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		os.Exit(1)
	}
	c := exec.Command(bin, progArgs...)
	c.Stdin = os.Stdin // so 'ask' can read the person's answers
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(1)
	}
}

func cmdTest(file, lang string) {
	prog := load(file)
	if report(prog, lang) {
		fmt.Fprintln(os.Stderr, "\nI found problems above, so I didn't run the tests.")
		os.Exit(1)
	}
	mainSrc, testSrc, terrs := transpiler.New(prog).Transpile()
	reportTranspile(terrs)
	if testSrc == "" {
		fmt.Println("No tests found. Add a 'test \"...\" ... end test' block.")
		return
	}
	if !haveGo() {
		os.Exit(1)
	}
	dir, _ := os.MkdirTemp("", "sstest")
	defer os.RemoveAll(dir)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainSrc), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testSrc), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module hotprogram\n\ngo 1.22\n"), 0o644)

	c := exec.Command("go", "test", "-v", ".")
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(1)
	}
}

func cmdBuild(file, lang string, windows bool) {
	prog := load(file)
	if report(prog, lang) {
		fmt.Fprintln(os.Stderr, "\nI found problems above, so I didn't build. Fix those and try again.")
		os.Exit(1)
	}
	goSrc, _, terrs := transpiler.New(prog).Transpile()
	reportTranspile(terrs)
	dir := tempModule(goSrc)
	defer os.RemoveAll(dir)

	if !haveGo() {
		os.Exit(1)
	}

	base := strings.TrimSuffix(filepath.Base(file), ".hot")
	out := base
	if windows {
		out += ".exe"
	}
	outAbs, _ := filepath.Abs(out)

	c := exec.Command("go", "build", "-o", outAbs, ".")
	c.Dir = dir
	if windows {
		c.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
	}
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "The build failed unexpectedly.")
		os.Exit(1)
	}
	fmt.Printf("Done. Built: %s\n", out)
	if !windows {
		fmt.Printf("Run it with:  ./%s\n", out)
	}
}

func tempModule(goSrc string) string {
	dir, err := os.MkdirTemp("", "hotgrin")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(goSrc), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module hotprogram\n\ngo 1.22\n"), 0o644)
	return dir
}

func haveGo() bool {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintln(os.Stderr, "To run or build programs, hotgrin needs Go installed for now.")
		fmt.Fprintln(os.Stderr, "Get it from https://go.dev/dl/ , then try again.")
		return false
	}
	return true
}
