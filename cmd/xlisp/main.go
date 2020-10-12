package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/chzyer/readline"
	"github.com/issadarkthing/xlisp"
	"github.com/spy16/sabre"
	"github.com/spy16/sabre/repl"
)

const help = `Xlisp %s [Commit: %s] [Compiled with %s]
Visit https://github.com/issadarkthing/xlisp for more.`

var (
	version = "N/A"
	commit  = "N/A"

	executeStr   = flag.String("e", "", "Execute string")
	unload       = flag.Bool("u", false, "Unload core library")
	printVersion = flag.Bool("v", false, "Prints slang version and exit")
)

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	xl := xlisp.New()
	xl.BindGo("*version*", version)

	var result sabre.Value
	var err error

	resolvedPath, err := filepath.Abs("/etc/xlisp/core.xlisp")
	if err != nil {
		fatalf("error: %v\n", err)
	}

	core, err := os.Open(resolvedPath)
	if err != nil {
		fatalf("error: %v\n", err)
	}
	defer core.Close()

	if !*unload {
		_, err = xl.ReadEval(core)
	}

	xl.SwitchNS(sabre.Symbol{Value: "user"})

	if len(os.Args) > 1 {

		var file int
		file = flag.NFlag() + 1

		fh, err := os.Open(os.Args[file])
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer fh.Close()

		xl.BindGo("*file*", fh.Name())
		_, err = xl.ReadEval(fh)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	if *executeStr != "" {
		result, err = xl.ReadEvalStr(*executeStr)
		fmt.Println(result)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		return
	}

	lr, errMapper := readlineInstance()

	repl := repl.New(xl,
		repl.WithBanner(fmt.Sprintf(help, version, commit, runtime.Version())),
		repl.WithInput(lr, errMapper),
		repl.WithOutput(lr.Stdout()),
		repl.WithPrompts("=>", "|"),
	)

	if err := repl.Loop(context.Background()); err != nil {
		fatalf("REPL exited with error: %v", err)
	}
	fmt.Println("Bye!")
}

func readlineInstance() (*readline.Instance, func(error) error) {
	lr, err := readline.New("")
	if err != nil {
		fatalf("readline: %v", err)
	}

	errMapper := func(e error) error {
		if errors.Is(e, readline.ErrInterrupt) {
			return nil
		}

		return e
	}

	return lr, errMapper
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
