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

var (
	builtinAdd = &Op{
		Apply: func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Num)
			rhs := vals[1].(*Num)
			res := big.NewFloat(0).Add(lhs.Value, rhs.Value)
			return &Num{Value: res}, nil
		},
	}

	builtinMul = &Op{
		Apply: func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Num)
			rhs := vals[1].(*Num)
			res := big.NewFloat(0).Mul(lhs.Value, rhs.Value)
			return &Num{Value: res}, nil
		},
	}

	builtinNeg = &Fn{
		Argc: 1,
		Apply: func(env *Environment, vals ...Value) (Value, error) {
			sub := vals[0].(*Num)
			res := big.NewFloat(0).Neg(sub.Value)
			return &Num{Value: res}, nil
		},
	}

	builtinAbs = &Fn{
		Argc: 1,
		Apply: func(env *Environment, vals ...Value) (Value, error) {
			sub := vals[0].(*Num)
			res := big.NewFloat(0).Abs(sub.Value)
			return &Num{Value: res}, nil
		},
	}
)
