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
			"+": builtinAdd,
			"*": builtinMul,
		},
	}
}

type value interface {
	Stringify() string
}

type numberValue struct {
	value *big.Float
}

func (n *numberValue) Stringify() string {
	return n.value.String()
}

type builtinFuncValue struct {
	args  int
	apply func(env *environment, vals ...value) (value, error)
}

func (b *builtinFuncValue) Stringify() string {
	return fmt.Sprintf("builtin/%d", b.args)
}

var (
	builtinAdd = &builtinFuncValue{
		args: 2,
		apply: func(env *environment, vals ...value) (value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Add(lhs.value, rhs.value)
			return &numberValue{value: res}, nil
		},
	}
	builtinMul = &builtinFuncValue{
		args: 2,
		apply: func(env *environment, vals ...value) (value, error) {
			lhs := vals[0].(*numberValue)
			rhs := vals[1].(*numberValue)
			res := big.NewFloat(0).Mul(lhs.value, rhs.value)
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
	case *infixExpr:
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

		return fn.apply(env, lhs, rhs)
	}

	return nil, errors.New("bad expression")
}
