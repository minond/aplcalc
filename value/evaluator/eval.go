package evaluator

import (
	"errors"
	"fmt"

	"github.com/minond/calc/parser"
	"github.com/minond/calc/value"
)

// func fold(env *value.Environment, call *parser.App) (value.Value, error) {
// 	if len(call.Args) != 2 {
// 		return nil, fmt.Errorf("expecting 2 arguments but got %d", len(call.Args))
// 	}
//
// 	fnEx, argEx := call.Args[0], call.Args[1]
// 	fnId, ok := fnEx.(*parser.Id)
// 	if !ok {
// 		return nil, errors.New("expecting an identifier")
// 	} else if !env.HasFn(fnId.Value) {
// 		return nil, fmt.Errorf("`%s` is not a function", fnId.Value)
// 	}
//
// 	fn := env.GetFn(fnId.Value)
// 	arg, err := Eval(env, argEx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return fn.Apply(env, arg)
// }

func define(env *value.Environment, def *parser.Op) (value.Value, error) {
	name, ok := def.Lhs.(*parser.Id)
	if !ok {
		return nil, errors.New("invalid identifier")
	}

	value, err := Eval(env, def.Rhs)
	if err != nil {
		return nil, err
	}

	env.SetVal(name.Value, value)
	return value, nil
}

func Eval(env *value.Environment, expr parser.Expr) (value.Value, error) {
	switch e := expr.(type) {
	case *parser.Num:
		return &value.Num{Value: e.Value}, nil
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
		var args []value.Value
		for _, arg := range e.Args {
			val, err := Eval(env, arg)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}

		return fn.Apply(env, args...)

	case *parser.Op:
		if e.Op == ":=" {
			return define(env, e)
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
