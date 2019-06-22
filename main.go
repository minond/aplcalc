package main

import (
	"os"

	"github.com/minond/calc/parser"
	"github.com/minond/calc/repl"
	"github.com/minond/calc/value"
	"github.com/minond/calc/value/evaluator"
)

func main() {
	debug := false
	running := true

	e := value.NewEnvironment()
	p := parser.NewParser(e)
	r := repl.Repl{
		Input:  os.Stdin,
		Output: os.Stdout,
	}

	for running {
		r.Write("? ")
		input, _ := r.Read()
		switch input {
		case "":
			continue
		case "exit":
			running = false
		case "debug":
			debug = !debug
		default:
			expr, err := p.Parse(input)
			if err != nil {
				r.Write("syntax error: %v\n\n", err)
				continue
			} else if debug {
				r.Write("%s\n\n", expr.Stringify(0))
				continue
			}

			val, err := evaluator.Eval(e, expr)
			if err != nil {
				r.Write("error: %v\n\n", err)
			} else {
				e.SetVal("_", val)
				r.Write("= %s\n\n", val.Stringify())
			}
		}
	}
}
