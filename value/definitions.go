package value

import (
	"errors"
	"fmt"
	"math/big"
)

var add = &Op{
	Impl: map[ty]handler{
		TNum: func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Num)
			rhs := vals[1].(*Num)
			res := big.NewFloat(0).Add(lhs.Value, rhs.Value)
			return &Num{Value: res}, nil
		},
		TArr: func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Arr)
			rhs := vals[1].(*Arr)
			if len(lhs.Values) != len(rhs.Values) {
				return nil, fmt.Errorf("array sizes do not match, left has %d items but right has %d",
					len(lhs.Values), len(rhs.Values))
			}
			res := &Arr{Values: make([]*Num, len(lhs.Values))}
			for i := range lhs.Values {
				res.Values[i] = &Num{
					Value: big.NewFloat(0).Add(lhs.Values[i].Value, rhs.Values[i].Value),
				}
			}
			return res, nil
		},
		TArr | TNum: func(env *Environment, vals ...Value) (Value, error) {
			var arr *Arr
			var num *Num

			switch vals[0].(type) {
			case *Arr:
				arr = vals[0].(*Arr)
				num = vals[1].(*Num)
			default:
				arr = vals[1].(*Arr)
				num = vals[0].(*Num)
			}

			res := &Arr{Values: make([]*Num, len(arr.Values))}
			for i := range arr.Values {
				res.Values[i] = &Num{
					Value: big.NewFloat(0).Add(arr.Values[i].Value, num.Value),
				}
			}
			return res, nil
		},
	},
}

var mul = &Op{
	Impl: map[ty]handler{
		TNum: func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Num)
			rhs := vals[1].(*Num)
			res := big.NewFloat(0).Mul(lhs.Value, rhs.Value)
			return &Num{Value: res}, nil
		},
	},
}

var range_ = &Op{
	Impl: map[ty]handler{
		TNum: func(env *Environment, vals ...Value) (Value, error) {
			a1, a2 := vals[0].(*Num), vals[1].(*Num)
			min64, _ := a1.Value.Int64()
			min := int(min64)
			max64, _ := a2.Value.Int64()
			max := int(max64)
			res := &Arr{Values: make([]*Num, max-min)}
			for i := 0; i < max-min; i++ {
				res.Values[i] = &Num{Value: big.NewFloat(float64(min + i))}
			}
			return res, nil
		},
	},
}

var access = &Op{
	Impl: map[ty]handler{
		TArr: func(env *Environment, vals ...Value) (Value, error) {
			orig := vals[0].(*Arr)
			idxs := vals[1].(*Arr)
			res := &Arr{Values: make([]*Num, len(idxs.Values))}
			size := len(orig.Values)
			for i, nidx := range idxs.Values {
				f64, _ := nidx.Value.Float64()
				idx := int(f64)
				if idx >= size {
					return nil, errors.New("index out of bounds")
				}
				res.Values[i] = orig.Values[idx]
			}
			return res, nil
		},
		TNum | TArr: func(env *Environment, vals ...Value) (Value, error) {
			var arr *Arr
			var num *Num

			switch vals[0].(type) {
			case *Arr:
				arr = vals[0].(*Arr)
				num = vals[1].(*Num)
			default:
				arr = vals[1].(*Arr)
				num = vals[0].(*Num)
			}

			f64, _ := num.Value.Float64()
			idx := int(f64)
			if idx >= len(arr.Values) {
				return nil, errors.New("index out of bounds")
			}
			return arr.Values[idx], nil
		},
	},
}

var set = &Op{
	Impl: map[ty]handler{
		TNum | TArr: func(env *Environment, vals ...Value) (Value, error) {
			var arr *Arr
			var num *Num

			switch vals[0].(type) {
			case *Arr:
				arr = vals[0].(*Arr)
				num = vals[1].(*Num)
			default:
				arr = vals[1].(*Arr)
				num = vals[0].(*Num)
			}

			res := &Arr{Values: make([]*Num, len(arr.Values))}
			for i := range arr.Values {
				res.Values[i] = &Num{Value: num.Value}
			}
			return res, nil
		},
	},
}

var neg = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		arg := vals[0].(*Num)
		res := big.NewFloat(0).Neg(arg.Value)
		return &Num{Value: res}, nil
	},
}

var abs = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		arg := vals[0].(*Num)
		res := big.NewFloat(0).Abs(arg.Value)
		return &Num{Value: res}, nil
	},
}

var until = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		arg := vals[0].(*Num)
		max64, _ := arg.Value.Int64()
		max := int(max64)
		res := &Arr{Values: make([]*Num, max)}
		for i := 0; i < max; i++ {
			res.Values[i] = &Num{Value: big.NewFloat(float64(i))}
		}
		return res, nil
	},
}

var len_ = &Fn{
	Argc: 1,
	Apply: func(env *Environment, vals ...Value) (Value, error) {
		arg := vals[0].(*Arr)
		return &Num{Value: big.NewFloat(float64(len(arg.Values)))}, nil
	},
}
