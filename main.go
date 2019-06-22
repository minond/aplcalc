package main

import (
	"os"

	p "github.com/minond/calc/parser"
	r "github.com/minond/calc/repl"
	v "github.com/minond/calc/value"
	"github.com/minond/calc/value/evaluator"
)

func main() {
	debugging := false
	running := true

	env := v.NewEnvironment()
	parse := p.NewParser(env)
	repl := r.NewRepl(os.Stdin, os.Stdout)

	for running {
		repl.Write("? ")
		input, _ := repl.Read()
		switch input {
		case "":
			continue
		case "exit":
			running = false
		case "debug":
			debugging = !debugging
		default:
			expr, err := parse.Parse(input)
			if err != nil {
				repl.Write("syntax error: %v\n\n", err)
				continue
			} else if debugging {
				repl.Write("%s\n\n", expr.Stringify(0))
				continue
			}

			val, err := evaluator.Eval(env, expr)
			if err != nil {
				repl.Write("error: %v\n\n", err)
			} else {
				env.SetVal("_", val)
				repl.Write("= %s\n\n", val.Stringify())
			}
		}
	}
}
