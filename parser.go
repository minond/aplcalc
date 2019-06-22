// Parses the following grammar:
//
// expr = app
//      | op
//      | unit
//      ;
//
// op = expr id expr
//    ;
//
// app = id expr ( expr ) *
//     ;
//
// unit = group
//      | num
//      | id
//      ;
//
// group = "(" expr ")"
//       ;
//
// id = ?? valid identifier characters ??
//
// num = ?? valid number characters ??

package main

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

type tok uint8

const (
	tokEOF tok = iota
	tokNum
	tokWord
)

type token struct {
	tok    tok
	lexeme string
}

func (t token) is(a tok) bool {
	return t.tok == a
}

func (t token) eqv(other token) bool {
	return t.tok == other.tok && t.lexeme == other.lexeme
}

func (t token) String() string {
	switch t.tok {
	case tokEOF:
		return "(token-eof)"
	case tokNum:
		return fmt.Sprintf("(token-num `%s`)", t.lexeme)
	case tokWord:
		return fmt.Sprintf("(token-word `%s`)", t.lexeme)
	default:
		return fmt.Sprintf("(token-unknown `%s`)", t.lexeme)
	}
}

var (
	tokenEOF        = token{tok: tokEOF}
	tokenCloseParen = token{tok: tokWord, lexeme: ")"}
	tokenOpenParen  = token{tok: tokWord, lexeme: "("}
)

func tokenize(input string) []token {
	runes := []rune(input)
	max := len(runes)

	var curr rune
	var tokens []token

	validchar := and(not(unicode.IsSpace), not(is(')')))

	for pos := 0; pos < max; {
		curr = runes[pos]
		switch {
		case unicode.IsSpace(curr):
			pos++
		case curr == '(':
			tokens = append(tokens, tokenOpenParen)
			pos++
		case curr == ')':
			tokens = append(tokens, tokenCloseParen)
			pos++
		case unicode.IsNumber(curr):
			num, size := eat(runes, pos, max, validchar)
			pos += size
			tokens = append(tokens, token{tok: tokNum, lexeme: string(num)})
		default:
			word, size := eat(runes, pos, max, validchar)
			pos += size
			tokens = append(tokens, token{tok: tokWord, lexeme: string(word)})
		}
	}

	return tokens
}

type runePred func(rune) bool

func is(r1 rune) runePred {
	return func(r2 rune) bool {
		return r1 == r2
	}
}

func not(fn runePred) runePred {
	return func(r rune) bool {
		return !fn(r)
	}
}

func and(fns ...runePred) runePred {
	return func(r rune) bool {
		for _, fn := range fns {
			if !fn(r) {
				return false
			}
		}
		return true
	}
}

func eat(runes []rune, pos, max int, pred runePred) ([]rune, int) {
	buff := []rune{}
	for ; pos < max; pos++ {
		if !pred(runes[pos]) {
			break
		}
		buff = append(buff, runes[pos])
	}
	return buff, len(buff)
}

type expression interface {
	Stringify(indent int) string
}

type groupExpr struct {
	sub expression
}

func (g groupExpr) Stringify(indent int) string {
	if g.sub == nil {
		return "(group empty)"
	}
	return fmt.Sprintf("(group\n%s%s)",
		strings.Repeat(" ", indent+2),
		g.sub.Stringify(indent+2))
}

type opExpr struct {
	op  string
	lhs expression
	rhs expression
}

func (b opExpr) Stringify(indent int) string {
	return fmt.Sprintf("(op %s\n%s%s\n%s%s)",
		b.op,
		strings.Repeat(" ", indent+2),
		b.lhs.Stringify(indent+2),
		strings.Repeat(" ", indent+2),
		b.rhs.Stringify(indent+2))
}

type appExpr struct {
	op   string
	args []expression
}

func (p appExpr) Stringify(indent int) string {
	var args []string
	for _, arg := range p.args {
		args = append(args, arg.Stringify(indent+2))
	}

	pad := "\n" + strings.Repeat(" ", indent+2)
	left := strings.Join(args, pad)
	return fmt.Sprintf("(app %s%s%s)", p.op, pad, left)
}

type numberExpr struct {
	value *big.Float
}

func (n numberExpr) Stringify(indent int) string {
	return fmt.Sprintf("(number %s)", n.value.String())
}

type identifierExpr struct {
	value string
}

func (i identifierExpr) Stringify(indent int) string {
	return fmt.Sprintf("(identifier %s)", i.value)
}

type parser struct {
	tokens []token
	pos    int
	ops    []string
	fns    map[string]int
}

func parse(input string) (expression, error) {
	return newParser(input).parse()
}

func newParser(input string) *parser {
	return &parser{
		tokens: tokenize(input),
		// XXX don't hardcode this
		ops: []string{"+"},
		fns: map[string]int{
			"abs":    1,
			"select": 2,
			"from":   1,
		},
	}
}

func (p *parser) isOp(op string) bool {
	for _, x := range p.ops {
		if x == op {
			return true
		}
	}
	return false
}

func (p *parser) isFn(fn string) (int, bool) {
	argc, ok := p.fns[fn]
	return argc, ok
}

func (p *parser) lookahead(n int) token {
	if p.pos+n < len(p.tokens) {
		return p.tokens[p.pos+n]
	}
	return tokenEOF
}

func (p *parser) peek() token {
	return p.lookahead(0)
}

func (p *parser) eat() token {
	t := p.peek()
	p.pos++
	return t
}

func (p *parser) done() bool {
	return p.pos >= len(p.tokens)
}

func (p *parser) parse() (expression, error) {
	if p.done() {
		return nil, errors.New("unexpected eof")
	}
	return p.expr()
}

// expr = app
//      | op
//      | unit
//      ;
func (p *parser) expr() (expression, error) {
	next := p.peek()
	if argc, ok := p.isFn(next.lexeme); ok {
		op := p.eat()
		var args []expression
		for ; argc > 0; argc-- {
			arg, err := p.parse()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return &appExpr{op: op.lexeme, args: args}, nil
	}

	expr, err := p.unit()
	if err != nil {
		return nil, err
	}

	if p.isOp(p.peek().lexeme) {
		op := p.eat()
		rhs, err := p.parse()
		if err != nil {
			return nil, err
		}
		expr = &opExpr{op: op.lexeme, lhs: expr, rhs: rhs}
		return expr, nil
	}

	return expr, err
}

// unit = group
//      | num
//      | id
//      ;
func (p *parser) unit() (expression, error) {
	next := p.peek()
	if next.eqv(tokenCloseParen) {
		return nil, nil
	} else if next.eqv(tokenOpenParen) {
		return p.group()
	} else if next.is(tokNum) {
		return p.num()
	}
	return p.id()
}

// group = "(" expr ")"
//       ;
func (p *parser) group() (expression, error) {
	next := p.eat()
	if !next.eqv(tokenOpenParen) {
		return nil, fmt.Errorf("expecting an open paren but got %s instead", next)
	}

	sub, err := p.parse()
	if err != nil {
		return nil, err
	}

	next = p.eat()
	if !next.eqv(tokenCloseParen) {
		return nil, fmt.Errorf("expecting a closing paren but got %s instead", next)
	}

	return &groupExpr{sub: sub}, nil
}

func (p *parser) id() (expression, error) {
	id := p.eat()
	return &identifierExpr{id.lexeme}, nil
}

func (p *parser) num() (expression, error) {
	next := p.eat()
	if !next.is(tokNum) {
		return nil, fmt.Errorf("expecting a number but got %s instead", next)
	}

	value, _, err := big.ParseFloat(next.lexeme, 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("unable to parse number: %v", err)
	}
	return &numberExpr{value: value}, nil
}
