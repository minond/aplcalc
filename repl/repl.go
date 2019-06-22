package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Repl struct {
	Input  io.Reader
	Output io.Writer
}

func NewRepl(input io.Reader, output io.Writer) Repl {
	return Repl{Input: input, Output: output}
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
