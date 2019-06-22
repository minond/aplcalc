package main

import (
	"errors"
	"fmt"
	"math/big"
)

type environment struct {
	values map[string]value
}

func (env *environment) has(id string) bool {
	_, ok := env.values[id]
	return ok
}

func (env *environment) get(id string) value {
	return env.values[id]
}

func (env *environment) set(id string, val value) {
	env.values[id] = val
}

func newEnvironment() *environment {
	return &environment{
		values: map[string]value{
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

func validTypes(expected []ty, args ...value) error {
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

type value interface {
	Stringify() string
	Type() ty
}

type numberValue struct {
	value *big.Float
}

func (n *numberValue) Stringify() string {
	return n.value.Text('g', -1)
}

func (n *numberValue) Type() ty {
	return tyNumber
}

type builtinFuncValue struct {
	args  []ty
	apply func(env *environment, vals ...value) (value, error)
}

func (b *builtinFuncValue) Stringify() string {
	return fmt.Sprintf("builtin/%d", len(b.args))
}

func (b *builtinFuncValue) Type() ty {
	return tyFunction
}

var (
	builtinAdd = &builtinFuncValue{
		args: []ty{tyNumber, tyNumber},
		apply: func(env *environment, vals ...value) (value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Add(lhs.value, rhs.value)
			return &numberValue{value: res}, nil
		},
	}

	builtinMul = &builtinFuncValue{
		args: []ty{tyNumber, tyNumber},
		apply: func(env *environment, vals ...value) (value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Mul(lhs.value, rhs.value)
			return &numberValue{value: res}, nil
		},
	}

	builtinNeg = &builtinFuncValue{
		args: []ty{tyNumber},
		apply: func(env *environment, vals ...value) (value, error) {
			sub := vals[0].(*numberValue)
			res := big.NewFloat(0).Neg(sub.value)
			return &numberValue{value: res}, nil
		},
	}

	builtinAbs = &builtinFuncValue{
		args: []ty{tyNumber},
		apply: func(env *environment, vals ...value) (value, error) {
			sub := vals[0].(*numberValue)
			res := big.NewFloat(0).Abs(sub.value)
			return &numberValue{value: res}, nil
		},
	}
)

func eval(env *environment, expr expression) (value, error) {
	switch e := expr.(type) {
	case *numberExpr:
		return &numberValue{value: e.value}, nil
	case *identifierExpr:
		if !env.has(e.value) {
			return nil, fmt.Errorf("%s is not defined", e.value)
		}
		return env.get(e.value), nil
	case *groupExpr:
		return eval(env, e.sub)

	case *appExpr:
		if !env.has(e.op) {
			return nil, fmt.Errorf("%s is not defined", e.op)
		}
		fn, valid := env.get(e.op).(*builtinFuncValue)
		if !valid {
			return nil, fmt.Errorf("%s is not a function", e.op)
		}

		var args []value
		for _, arg := range e.args {
			val, err := eval(env, arg)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}

		if err := validTypes(fn.args, args...); err != nil {
			return nil, err
		}

		return fn.apply(env, args...)

	case *opExpr:
		if e.op == "=" {
			var key string
			switch id := e.lhs.(type) {
			case *identifierExpr:
				key = id.value
			default:
				return nil, errors.New("invalid identifier")
			}
			val, err := eval(env, e.rhs)
			if err != nil {
				return nil, err
			}
			env.set(key, val)
			return val, nil
		}

		if !env.has(e.op) {
			return nil, fmt.Errorf("%s is not defined", e.op)
		}
		fn, valid := env.get(e.op).(*builtinFuncValue)
		if !valid {
			return nil, fmt.Errorf("%s is not a function", e.op)
		}

		lhs, err := eval(env, e.lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := eval(env, e.rhs)
		if err != nil {
			return nil, err
		}

		if err := validTypes(fn.args, lhs, rhs); err != nil {
			return nil, err
		}

		return fn.apply(env, lhs, rhs)
	}

	return nil, errors.New("bad expression")
}
