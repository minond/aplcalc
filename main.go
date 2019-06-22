package main

import (
	"os"

	"github.com/minond/calc/parser"
)

func main() {
	var debug bool
	env := newEnvironment()

	r := repl{
		input:  os.Stdin,
		output: os.Stdout,
	}

loop:
	for {
		r.write("? ")
		input, _ := r.read()
		switch input {
		case "":
			continue
		case "exit":
			break loop
		case "debug":
			debug = !debug
		default:
			expr, err := parser.Parse(input)
			if err != nil {
				r.write("syntax error: %v\n\n", err)
				continue
			} else if debug {
				r.write("%s\n\n", expr.Stringify(0))
				continue
			}

			val, err := eval(env, expr)
			if err != nil {
				r.write("error: %v\n\n", err)
			} else {
				env.set("_", val)
				r.write("= %s\n\n", val.Stringify())
			}
		}
	}
}
