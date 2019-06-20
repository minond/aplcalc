package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type repl struct {
	input  io.Reader
	output io.Writer
}

func (repl repl) read() (string, error) {
	reader := bufio.NewReader(repl.input)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (repl repl) write(s string, a ...interface{}) error {
	_, err := fmt.Fprintf(repl.output, s, a...)
	return err
}
