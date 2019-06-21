package main

import "os"

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
			expr, err := parse(input)
			if err != nil {
				r.write("tokens: %s\n", tokenize(input))
				r.write("syntax error: %v\n\n", err)
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
