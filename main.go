package main

import "os"

func main() {
	r := repl{
		input:  os.Stdin,
		output: os.Stdout,
	}

	for {
		r.write("? ")
		input, _ := r.read()
		expr, err := parse(input)
		if err != nil {
			r.write("syntax error: %v\n\n", err)
		} else {
			r.write("%s\n\n", expr)
		}
	}
}
