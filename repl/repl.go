package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/minond/calc/parser"
	"github.com/minond/calc/value"
	"github.com/minond/calc/value/evaluator"
)

type Repl struct {
	Input  io.Reader
	Output io.Writer

	parser *parser.Parser
	env    *value.Environment

	running   bool
	debugging bool
}

func NewRepl(input io.Reader, output io.Writer) *Repl {
	env := value.NewEnvironment()
	parse := parser.NewParser(env)
	return &Repl{
		Input:     input,
		Output:    output,
		parser:    parse,
		env:       env,
		running:   true,
		debugging: false,
	}
}

func (repl Repl) Running() bool {
	return repl.running
}

func (repl *Repl) Stop() {
	repl.running = false
}

func (repl Repl) Debug() bool {
	return repl.debugging
}

func (repl *Repl) SetDebug(debugging bool) {
	repl.debugging = debugging
}

func (repl Repl) Read() (string, error) {
	reader := bufio.NewReader(repl.Input)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (repl Repl) Write(s string, a ...interface{}) error {
	_, err := fmt.Fprintf(repl.Output, s, a...)
	return err
}

func (repl Repl) Eval(code string) {
	expr, err := repl.parser.Parse(code)
	if err != nil {
		repl.Write("syntax error: %v\n\n", err)
		return
	} else if repl.debugging {
		repl.Write("%s\n\n", expr.Stringify(0))
		return
	}

	val, err := evaluator.Eval(repl.env, expr)
	if err != nil {
		repl.Write("error: %v\n\n", err)
	} else {
		repl.env.SetVal("_", val)
		repl.Write("= %s\n\n", val.Stringify())
	}
}
