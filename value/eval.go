package value

import (
	"errors"
	"fmt"

	"github.com/minond/calc/parser"
)

func Eval(env *Environment, expr parser.Expr) (Value, error) {
	switch e := expr.(type) {
	case *parser.Num:
		return &Num{Value: e.Value}, nil
	case *parser.Id:
		if !env.HasVal(e.Value) {
			return nil, fmt.Errorf("%s is not defined", e.Value)
		}
		return env.GetVal(e.Value), nil
	case *parser.Group:
		return Eval(env, e.Sub)

	case *parser.App:
		if !env.HasFn(e.Op) {
			return nil, fmt.Errorf("%s is not defined", e.Op)
		}
		fn := env.GetFn(e.Op)
		var args []Value
		for _, arg := range e.Args {
			val, err := Eval(env, arg)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}

		return fn.Apply(env, args...)

	case *parser.Op:
		if e.Op == "=" {
			var key string
			switch id := e.Lhs.(type) {
			case *parser.Id:
				key = id.Value
			default:
				return nil, errors.New("invalid identifier")
			}
			val, err := Eval(env, e.Rhs)
			if err != nil {
				return nil, err
			}
			env.SetVal(key, val)
			return val, nil
		}

		if !env.HasOp(e.Op) {
			return nil, fmt.Errorf("%s is not defined", e.Op)
		}
		op := env.GetOp(e.Op)
		lhs, err := Eval(env, e.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := Eval(env, e.Rhs)
		if err != nil {
			return nil, err
		}

		return op.Apply(env, lhs, rhs)
	}

	return nil, errors.New("bad expression")
}
