package parser

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"unicode"

	"github.com/minond/calc/value"
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

type Expr interface {
	Stringify(indent int) string
}

type Group struct {
	Sub Expr
}

func (g Group) Stringify(indent int) string {
	if g.Sub == nil {
		return "(group empty)"
	}
	return fmt.Sprintf("(group\n%s%s)",
		strings.Repeat(" ", indent+2),
		g.Sub.Stringify(indent+2))
}

type Op struct {
	Op  string
	Lhs Expr
	Rhs Expr
}

func (o Op) Stringify(indent int) string {
	return fmt.Sprintf("(op %s\n%s%s\n%s%s)",
		o.Op,
		strings.Repeat(" ", indent+2),
		o.Lhs.Stringify(indent+2),
		strings.Repeat(" ", indent+2),
		o.Rhs.Stringify(indent+2))
}

type App struct {
	Op   string
	Args []Expr
}

func (p App) Stringify(indent int) string {
	var args []string
	for _, arg := range p.Args {
		args = append(args, arg.Stringify(indent+2))
	}

	pad := "\n" + strings.Repeat(" ", indent+2)
	left := strings.Join(args, pad)
	return fmt.Sprintf("(app %s%s%s)", p.Op, pad, left)
}

type Arr struct {
	Values []*Num
}

func (a Arr) Stringify(indent int) string {
	var vals []string
	for _, val := range a.Values {
		vals = append(vals, val.Stringify(indent+2))
	}

	pad := "\n" + strings.Repeat(" ", indent+2)
	left := strings.Join(vals, pad)
	return fmt.Sprintf("(array%s%s)", pad, left)
}

type Num struct {
	Value *big.Float
}

func (n Num) Stringify(indent int) string {
	return fmt.Sprintf("(num %s)", n.Value.String())
}

type Id struct {
	Value string
}

func (i Id) Stringify(indent int) string {
	return fmt.Sprintf("(id %s)", i.Value)
}

type parser struct {
	env *value.Environment

	mux    sync.Mutex
	tokens []token
	pos    int
}

func NewParser(env *value.Environment) *parser {
	return &parser{env: env}
}

func (p *parser) Parse(input string) (Expr, error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.tokens = tokenize(input)
	p.pos = 0
	return p.expr()
}

func (p *parser) isOp(op string) bool {
	return p.env.HasOp(op)
}

func (p *parser) isFn(fn string) (int, bool) {
	if p.env.HasFn(fn) {
		return p.env.GetFn(fn).Argc, true
	}
	return 0, false
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

// expr = app
//      | op
//      | unit
//      ;
func (p *parser) expr() (Expr, error) {
	if p.done() {
		return nil, errors.New("unexpected eof")
	}

	next := p.peek()
	if argc, ok := p.isFn(next.lexeme); ok {
		op := p.eat()
		var args []Expr
		for ; argc > 0; argc-- {
			arg, err := p.expr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return &App{Op: op.lexeme, Args: args}, nil
	}

	expr, err := p.unit()
	if err != nil {
		return nil, err
	}

	if p.isOp(p.peek().lexeme) {
		op := p.eat()
		rhs, err := p.expr()
		if err != nil {
			return nil, err
		}
		expr = &Op{Op: op.lexeme, Lhs: expr, Rhs: rhs}
		return expr, nil
	}

	return expr, err
}

// unit = group
//      | num
//      | arr
//      | id
//      ;
func (p *parser) unit() (Expr, error) {
	next := p.peek()
	if next.eqv(tokenCloseParen) {
		return nil, nil
	} else if next.eqv(tokenOpenParen) {
		return p.group()
	} else if next.is(tokNum) && p.lookahead(1).is(tokNum) {
		return p.arr()
	} else if next.is(tokNum) {
		return p.num()
	}
	return p.id()
}

// group = "(" expr ")"
//       ;
func (p *parser) group() (Expr, error) {
	next := p.eat()
	if !next.eqv(tokenOpenParen) {
		return nil, fmt.Errorf("expecting an open paren but got %s instead", next)
	}

	sub, err := p.expr()
	if err != nil {
		return nil, err
	}

	next = p.eat()
	if !next.eqv(tokenCloseParen) {
		return nil, fmt.Errorf("expecting a closing paren but got %s instead", next)
	}

	return &Group{Sub: sub}, nil
}

func (p *parser) id() (Expr, error) {
	id := p.eat()
	return &Id{id.lexeme}, nil
}

func (p *parser) arr() (Expr, error) {
	arr := &Arr{}
	for p.peek().is(tokNum) {
		val, err := p.num()
		if err != nil {
			return nil, err
		}
		arr.Values = append(arr.Values, val)
	}
	return arr, nil
}

func (p *parser) num() (*Num, error) {
	next := p.eat()
	if !next.is(tokNum) {
		return nil, fmt.Errorf("expecting a number but got %s instead", next)
	}

	value, _, err := big.ParseFloat(next.lexeme, 10, 0, big.ToNearestEven)
	if err != nil {
		return nil, fmt.Errorf("unable to parse number: %v", err)
	}
	return &Num{Value: value}, nil
}
