package main

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/minond/calc/parser"
)

type environment struct {
	values map[string]Value
}

func (env *environment) has(id string) bool {
	_, ok := env.values[id]
	return ok
}

func (env *environment) get(id string) Value {
	return env.values[id]
}

func (env *environment) set(id string, val Value) {
	env.values[id] = val
}

func newEnvironment() *environment {
	return &environment{
		values: map[string]Value{
			"*":   builtinMul,
			"+":   builtinAdd,
			"abs": builtinAbs,
			"neg": builtinNeg,
		},
	}
}

type ty interface {
	fmt.Stringer
	Eq(ty) bool
}

type tyNumberS struct{}

func (tyNumberS) String() string {
	return "number"
}

func (tyNumberS) Eq(ty ty) bool {
	return ty == tyNumber
}

type tyFunctionS struct{}

func (tyFunctionS) String() string {
	return "function"
}

func (tyFunctionS) Eq(ty ty) bool {
	return ty == tyFunction
}

var (
	tyNumber   = tyNumberS{}
	tyFunction = tyFunctionS{}
)

func validTypes(expected []ty, args ...Value) error {
	if len(expected) != len(args) {
		return fmt.Errorf("expected %d arguments but got %d",
			len(expected), len(args))
	}
	for i, t := range expected {
		if !t.Eq(args[i].Type()) {
			return fmt.Errorf("expected %s in position %d but got a %s instead",
				t, i, args[i].Type())
		}
	}
	return nil
}

type Value interface {
	Stringify() string
	Type() ty
}

type numberValue struct {
	Value *big.Float
}

func (n *numberValue) Stringify() string {
	return n.Value.Text('g', -1)
}

func (n *numberValue) Type() ty {
	return tyNumber
}

type builtinFuncValue struct {
	Args  []ty
	apply func(env *environment, vals ...Value) (Value, error)
}

func (b *builtinFuncValue) Stringify() string {
	return fmt.Sprintf("builtin/%d", len(b.Args))
}

func (b *builtinFuncValue) Type() ty {
	return tyFunction
}

var (
	builtinAdd = &builtinFuncValue{
		Args: []ty{tyNumber, tyNumber},
		apply: func(env *environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Add(lhs.Value, rhs.Value)
			return &numberValue{Value: res}, nil
		},
	}

	builtinMul = &builtinFuncValue{
		Args: []ty{tyNumber, tyNumber},
		apply: func(env *environment, vals ...Value) (Value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Mul(lhs.Value, rhs.Value)
			return &numberValue{Value: res}, nil
		},
	}

	builtinNeg = &builtinFuncValue{
		Args: []ty{tyNumber},
		apply: func(env *environment, vals ...Value) (Value, error) {
			sub := vals[0].(*numberValue)
			res := big.NewFloat(0).Neg(sub.Value)
			return &numberValue{Value: res}, nil
		},
	}

	builtinAbs = &builtinFuncValue{
		Args: []ty{tyNumber},
		apply: func(env *environment, vals ...Value) (Value, error) {
			sub := vals[0].(*numberValue)
			res := big.NewFloat(0).Abs(sub.Value)
			return &numberValue{Value: res}, nil
		},
	}
)

func eval(env *environment, expr parser.Expr) (Value, error) {
	switch e := expr.(type) {
	case *parser.Num:
		return &numberValue{Value: e.Value}, nil
	case *parser.Id:
		if !env.has(e.Value) {
			return nil, fmt.Errorf("%s is not defined", e.Value)
		}
		return env.get(e.Value), nil
	case *parser.Group:
		return eval(env, e.Sub)

	case *parser.App:
		if !env.has(e.Op) {
			return nil, fmt.Errorf("%s is not defined", e.Op)
		}
		fn, valid := env.get(e.Op).(*builtinFuncValue)
		if !valid {
			return nil, fmt.Errorf("%s is not a function", e.Op)
		}

		var args []Value
		for _, arg := range e.Args {
			val, err := eval(env, arg)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}

		if err := validTypes(fn.Args, args...); err != nil {
			return nil, err
		}

		return fn.apply(env, args...)

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
			env.set(key, val)
			return val, nil
		}

		if !env.has(e.Op) {
			return nil, fmt.Errorf("%s is not defined", e.Op)
		}
		fn, valid := env.get(e.Op).(*builtinFuncValue)
		if !valid {
			return nil, fmt.Errorf("%s is not a function", e.Op)
		}

		lhs, err := eval(env, e.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := eval(env, e.Rhs)
		if err != nil {
			return nil, err
		}

		if err := validTypes(fn.Args, lhs, rhs); err != nil {
			return nil, err
		}

		return fn.apply(env, lhs, rhs)
	}

	return nil, errors.New("bad expression")
}
