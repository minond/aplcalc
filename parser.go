package main

import (
	"errors"
	"fmt"
	"math/big"
	"unicode"
)

type tok uint8

const (
	tokEOF tok = iota
	tokInvalid
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
		return "(eof)"
	case tokInvalid:
		return fmt.Sprintf("(invalid `%s`)", t.lexeme)
	case tokNum:
		return fmt.Sprintf("(num `%s`)", t.lexeme)
	case tokWord:
		return fmt.Sprintf("(word `%s`)", t.lexeme)
	default:
		return fmt.Sprintf("(unknown `%s`)", t.lexeme)
	}
}

var (
	tokenEOF = token{tok: tokEOF}

	tokenCloseParen = token{tok: tokWord, lexeme: ")"}
	tokenDiv        = token{tok: tokWord, lexeme: "/"}
	tokenEq         = token{tok: tokWord, lexeme: "="}
	tokenExp        = token{tok: tokWord, lexeme: "^"}
	tokenMinus      = token{tok: tokWord, lexeme: "-"}
	tokenMod        = token{tok: tokWord, lexeme: "%"}
	tokenMult       = token{tok: tokWord, lexeme: "*"}
	tokenOpenParen  = token{tok: tokWord, lexeme: "("}
	tokenPlus       = token{tok: tokWord, lexeme: "+"}

	tokenMap = map[rune]token{
		'%': tokenMod,
		'(': tokenOpenParen,
		')': tokenCloseParen,
		'*': tokenMult,
		'+': tokenPlus,
		'-': tokenMinus,
		'/': tokenDiv,
		'=': tokenEq,
		'^': tokenExp,
	}
)

func tokenize(input string) []token {
	runes := []rune(input)
	max := len(runes)
	var curr rune
	var tokens []token

	for pos := 0; pos < max; {
		curr = runes[pos]
		if token, known := tokenMap[curr]; known {
			tokens = append(tokens, token)
			pos++
			continue
		}
		switch {
		case unicode.IsSpace(curr):
			pos++
		case unicode.IsNumber(curr):
			num, size := eat(runes, pos, max, unicode.IsNumber)
			pos += size
			tokens = append(tokens, token{tok: tokNum, lexeme: string(num)})
		case unicode.IsLetter(curr):
			word, size := eat(runes, pos, max, not(unicode.IsSpace))
			pos += size
			tokens = append(tokens, token{tok: tokWord, lexeme: string(word)})
		default:
			pos++
			tokens = append(tokens, token{tok: tokInvalid, lexeme: string(curr)})
		}
	}

	return tokens
}

type runePred func(rune) bool

func not(fn runePred) runePred {
	return func(r rune) bool {
		return !fn(r)
	}
}

func or(fns ...runePred) runePred {
	return func(r rune) bool {
		for _, fn := range fns {
			if fn(r) {
				return true
			}
		}
		return false
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

// entry = statement ( ";" statement ) *
//       ;
//
// statement = assignment
//           | expression
//           ;
//
// assignment = identifier "=" expression
//            ;
//
// expression = "(" expression ")"
//            | operator expression
//            | expression operator expression
//            | value
//
// value = identifier
//       | number
//       ;
//
// identifier = ?? valid identifier characters ??
//
// number = ?? valid number characters ??
type statement interface{}

type expression interface {
	fmt.Stringer
}

type group struct {
	sub expression
}

func (g group) String() string {
	return fmt.Sprintf("(group %s)", g.sub)
}

type infix struct {
	op  token
	lhs expression
	rhs expression
}

func (b infix) String() string {
	return fmt.Sprintf("(infix-app %s %s %s)", b.op, b.lhs, b.rhs)
}

type prefix struct {
	op      token
	subject expression
}

func (p prefix) String() string {
	return fmt.Sprintf("(prefix-app %s %s)", p.op, p.subject)
}

type number struct {
	value *big.Float
}

func (n number) String() string {
	return fmt.Sprintf("(number %s)", n.value.String())
}

type identifier struct {
	value string
}

func (i identifier) String() string {
	return fmt.Sprintf("(identifier %s)", i.value)
}

type parser struct {
	tokens []token
	pos    int
}

func parse(input string) (expression, error) {
	return newParser(input).parse()
}

func newParser(input string) *parser {
	return &parser{tokens: tokenize(input)}
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
	return p.parseInfix()
}

func (p *parser) parseUnit() (expression, error) {
	next := p.peek()
	if next.eqv(tokenOpenParen) {
		return p.parseGroup()
	} else if next.is(tokNum) {
		return p.parseNumber()
	}
	return p.parsePrefix()
}

func (p *parser) parseGroup() (expression, error) {
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

	return &group{sub: sub}, nil
}

func (p *parser) parseNumber() (expression, error) {
	next := p.eat()
	if !next.is(tokNum) {
		return nil, fmt.Errorf("expecting a number but got %s instead", next)
	}

	value, _, err := big.ParseFloat(next.lexeme, 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("unable to parse number: %v", err)
	}
	return &number{value: value}, nil
}

func (p *parser) parsePrefix() (expression, error) {
	id := p.eat()
	if p.done() {
		return &identifier{id.lexeme}, nil
	}
	subject, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &prefix{op: id, subject: subject}, nil
}

func (p *parser) parseInfix() (expression, error) {
	expr, err := p.parseUnit()
	if err != nil {
		return nil, err
	}
	if !p.done() && !p.peek().eqv(tokenCloseParen) {
		op := p.eat()
		rhs, err := p.parse()
		if err != nil {
			return nil, err
		}
		expr = &infix{op: op, lhs: expr, rhs: rhs}
	}
	return expr, err
}
