package main

import "os"

func main() {
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
		default:
			expr, err := parse(input)
			if err != nil {
				r.write("tokens: %s\n", tokenize(input))
				r.write("syntax error: %v\n\n", err)
			} else {
				r.write("%s\n\n", expr.Stringify(0))
			}
		}
	}
}
