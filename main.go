package main

import (
	"os"

	r "github.com/minond/calc/repl"
)

func main() {
	repl := r.NewRepl(os.Stdin, os.Stdout)

	for repl.Running() {
		repl.Write("? ")

		input, _ := repl.Read()
		switch input {
		case "":
			continue
		case "exit":
			repl.Stop()
		case "debug":
			repl.SetDebug(!repl.Debug())
		default:
			repl.Eval(input)
		}
	}
}
