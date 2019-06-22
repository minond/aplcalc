package value

import (
	"fmt"
	"math/big"
)

type Value interface {
	Stringify() string
}

type Num struct {
	Value *big.Float
}

func (n *Num) Stringify() string {
	return n.Value.Text('g', -1)
}

type Op struct {
	Apply func(env *Environment, vals ...Value) (Value, error)
}

func (*Op) Stringify() string {
	return "op"
}

type Fn struct {
	Argc  int
	Apply func(env *Environment, vals ...Value) (Value, error)
}

func (fn *Fn) Stringify() string {
	return fmt.Sprintf("fn/%d", fn.Argc)
}
