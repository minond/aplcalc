package value

import (
	"errors"
	"fmt"
	"math/big"
)

func numbinop(operation func(*Num, *Num) *Num) fntable {
	return map[signature]handler{
		sig(TArr, TArr): func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Arr)
			rhs := vals[1].(*Arr)
			if len(lhs.Values) != len(rhs.Values) {
				return nil, fmt.Errorf("array sizes do not match, left has %d items but right has %d",
					len(lhs.Values), len(rhs.Values))
			}
			res := &Arr{Values: make([]*Num, len(lhs.Values))}
			for i := range lhs.Values {
				res.Values[i] = operation(lhs.Values[i], rhs.Values[i])
			}
			return res, nil
		},
		sig(TArr, TNum): func(env *Environment, vals ...Value) (Value, error) {
			arr := vals[0].(*Arr)
			num := vals[1].(*Num)
			res := &Arr{Values: make([]*Num, len(arr.Values))}
			for i := range arr.Values {
				res.Values[i] = operation(arr.Values[i], num)
			}
			return res, nil
		},
		sig(TNum, TArr): func(env *Environment, vals ...Value) (Value, error) {
			num := vals[0].(*Num)
			arr := vals[1].(*Arr)
			res := &Arr{Values: make([]*Num, len(arr.Values))}
			for i := range arr.Values {
				res.Values[i] = operation(num, arr.Values[i])
			}
			return res, nil
		},
		sig(TNum, TNum): func(env *Environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*Num)
			rhs := vals[1].(*Num)
			return operation(lhs, rhs), nil
		},
	}
}

var add = &Op{
	Impl: numbinop(func(lhs *Num, rhs *Num) *Num {
		return &Num{Value: big.NewFloat(0).Add(lhs.Value, rhs.Value)}
	}),
}

var mul = &Op{
	Impl: numbinop(func(lhs *Num, rhs *Num) *Num {
		return &Num{Value: big.NewFloat(0).Mul(lhs.Value, rhs.Value)}
	}),
}

var range_ = &Op{
	Impl: fntable{
		sig(TNum, TNum): func(env *Environment, vals ...Value) (Value, error) {
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
	Impl: fntable{
		sig(TArr, TArr): func(env *Environment, vals ...Value) (Value, error) {
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
		sig(TArr, TNum): func(env *Environment, vals ...Value) (Value, error) {
			arr := vals[0].(*Arr)
			num := vals[1].(*Num)
			f64, _ := num.Value.Float64()
			idx := int(f64)
			if idx >= len(arr.Values) {
				return nil, errors.New("index out of bounds")
			}
			return arr.Values[idx], nil
		},
		sig(TNum, TArr): func(env *Environment, vals ...Value) (Value, error) {
			arr := vals[1].(*Arr)
			num := vals[0].(*Num)
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
	Impl: fntable{
		sig(TArr, TNum): func(env *Environment, vals ...Value) (Value, error) {
			arr := vals[0].(*Arr)
			num := vals[1].(*Num)
			res := &Arr{Values: make([]*Num, len(arr.Values))}
			for i := range arr.Values {
				res.Values[i] = &Num{Value: num.Value}
			}
			return res, nil
		},
		sig(TNum, TArr): func(env *Environment, vals ...Value) (Value, error) {
			num := vals[0].(*Num)
			arr := vals[1].(*Arr)
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
	Impl: fntable{
		sig(TNum): func(env *Environment, vals ...Value) (Value, error) {
			arg := vals[0].(*Num)
			res := big.NewFloat(0).Neg(arg.Value)
			return &Num{Value: res}, nil
		},
	},
}

var abs = &Fn{
	Argc: 1,
	Impl: fntable{
		sig(TNum): func(env *Environment, vals ...Value) (Value, error) {
			arg := vals[0].(*Num)
			res := big.NewFloat(0).Abs(arg.Value)
			return &Num{Value: res}, nil
		},
	},
}

var until = &Fn{
	Argc: 1,
	Impl: fntable{
		sig(TNum): func(env *Environment, vals ...Value) (Value, error) {
			arg := vals[0].(*Num)
			max64, _ := arg.Value.Int64()
			max := int(max64)
			res := &Arr{Values: make([]*Num, max)}
			for i := 0; i < max; i++ {
				res.Values[i] = &Num{Value: big.NewFloat(float64(i))}
			}
			return res, nil
		},
	},
}

var g_until = &Fn{
	Argc: 1,
	Impl: fntable{
		sig(TNum): func(env *Environment, vals ...Value) (Value, error) {
			arg := vals[0].(*Num)
			max64, _ := arg.Value.Int64()
			max := int(max64)
			res := &Gen{
				ty:   TNum,
				done: false,
				curr: &Num{Value: big.NewFloat(0)},
				step: func(curr Value, size int) (Value, bool, bool) {
					num, _ := curr.(*Num)
					curr64, _ := num.Value.Int64()
					if int(curr64)+size > max {
						return nil, true, false
					}
					stepped := big.NewFloat(float64(int(curr64) + size))
					return &Num{Value: stepped}, false, true
				},
			}
			return res, nil
		},
	},
}

var g_take = &Op{
	Impl: fntable{
		sig(TGen, TNum): func(env *Environment, vals ...Value) (Value, error) {
			gen := vals[0].(*Gen)
			num := vals[1].(*Num)

			max64, _ := num.Value.Int64()
			max := int(max64)

			res := &Arr{Values: make([]*Num, max)}
			for i := 0; i < max; i++ {
				val, ok := gen.Next()
				if !ok {
					return nil, fmt.Errorf("error on step %d", i)
				}
				res.Values[i] = val.(*Num)
			}

			return res, nil
		},
	},
}

var len_ = &Fn{
	Argc: 1,
	Impl: fntable{
		sig(TArr): func(env *Environment, vals ...Value) (Value, error) {
			arg := vals[0].(*Arr)
			return &Num{Value: big.NewFloat(float64(len(arg.Values)))}, nil
		},
	},
}
