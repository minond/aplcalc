package value

import "math/big"

var add = &Op{
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		lhs := vals[0].(*Num)
		rhs := vals[1].(*Num)
		res := big.NewFloat(0).Add(lhs.Value, rhs.Value)
		return &Num{Value: res}, nil
	},
}

var mul = &Op{
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		lhs := vals[0].(*Num)
		rhs := vals[1].(*Num)
		res := big.NewFloat(0).Mul(lhs.Value, rhs.Value)
		return &Num{Value: res}, nil
	},
}

var neg = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		sub := vals[0].(*Num)
		res := big.NewFloat(0).Neg(sub.Value)
		return &Num{Value: res}, nil
	},
}

var abs = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		sub := vals[0].(*Num)
		res := big.NewFloat(0).Abs(sub.Value)
		return &Num{Value: res}, nil
	},
}
