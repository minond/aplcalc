package parser

import (
	"testing"

	"github.com/minond/calc/value"
)

func TestParse(t *testing.T) {
	tests := []struct {
		label  string
		input  string
		output string
	}{
		{"number", "1", "(num 1)"},
		{"long number", "78934430289340", "(num 7.893443029e+13)"},
		{"identifier", "a", "(id a)"},
		{"long identifier", "jfkdlsa$%%@$@#", "(id jfkdlsa$%%@$@#)"},
		{"empty group", "()", "(group empty)"},
		{"nested empty group", "((()))", "(group\n  (group\n    (group empty)))"},
		{"prefix expression for number", "abs 1", "(app abs\n  (num 1))"},
		{"prefix expression for identifier", "abs abc", "(app abs\n  (id abc))"},
		{"infix expression", "1 + 2", "(op +\n  (num 1)\n  (num 2))"},
		{"multiple infix expressions", "1 + 2 + 3 + 4 + 5", "(op +\n  (num 1)\n  (op +\n    (num 2)\n    (op +\n      (num 3)\n      (op +\n        (num 4)\n        (num 5)))))"},
		{"infix with an identifier and a number", "a + 1", "(op +\n  (id a)\n  (num 1))"},
		{"infix with a number and an identifier", "1 + a", "(op +\n  (num 1)\n  (id a))"},
		{"infix with two identifiers", "a + b", "(op +\n  (id a)\n  (id b))"},
	}

	e := value.NewEnvironment()
	p := NewParser(e)

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			ast, err := p.Parse(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if ast.Stringify(0) != test.output {
				t.Errorf("invalid ast for `%s`:\nexpected: %s\nreturned: %s",
					test.input, test.output, ast.Stringify(0))
			}
		})
	}
}
