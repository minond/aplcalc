package main

import (
	"errors"
	"fmt"

	"github.com/minond/calc/parser"
	"github.com/minond/calc/value"
)

func eval(env *value.Environment, expr parser.Expr) (value.Value, error) {
	switch e := expr.(type) {
	case *parser.Num:
		return &value.Num{Value: e.Value}, nil
	case *parser.Id:
		if !env.HasVal(e.Value) {
			return nil, fmt.Errorf("%s is not defined", e.Value)
		}
		return env.GetVal(e.Value), nil
	case *parser.Group:
		return eval(env, e.Sub)

	case *parser.App:
		if !env.HasFn(e.Op) {
			return nil, fmt.Errorf("%s is not defined", e.Op)
		}
		fn := env.GetFn(e.Op)
		var args []value.Value
		for _, arg := range e.Args {
			val, err := eval(env, arg)
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
			val, err := eval(env, e.Rhs)
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
		lhs, err := eval(env, e.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := eval(env, e.Rhs)
		if err != nil {
			return nil, err
		}

		return op.Apply(env, lhs, rhs)
	}

	return nil, errors.New("bad expression")
}
