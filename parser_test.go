package main

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		label  string
		input  string
		output string
	}{
		{"number", "1", "(number 1)"},
		{"long number", "78934430289340", "(number 7.893443029e+13)"},
		{"identifier", "a", "(identifier a)"},
		{"long identifier", "jfkdlsa$%%@$@#", "(identifier jfkdlsa$%%@$@#)"},
		{"empty group", "()", "(group empty)"},
		{"nested empty group", "((()))", "(group\n  (group\n    (group empty)))"},
		{"prefix expression", "abs 1", "(prefix-app (word `abs`)\n  (number 1))"},
		{"infix expression", "1 + 2", "(infix-app (word `+`)\n  (number 1)\n  (number 2))"},
		{"multiple infix expressions", "1 + 2 + 3 + 4 + 5", "(infix-app (word `+`)\n  (number 1)\n  (infix-app (word `+`)\n    (number 2)\n    (infix-app (word `+`)\n      (number 3)\n      (infix-app (word `+`)\n        (number 4)\n        (number 5)))))"},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			ast, err := parse(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if ast.Stringify(0) != test.output {
				t.Errorf("invalid ast for `%s`:\nexpected: %s\nreturned: %s",
					test.input, test.output, ast.Stringify(0))
			}
		})
	}
}
