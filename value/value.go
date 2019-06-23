package value

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type handler func(*Environment, ...Value) (Value, error)

type Value interface {
	Stringify() string
}

type Num struct {
	Value *big.Float
}

func (n *Num) Stringify() string {
	return n.Value.Text('g', -1)
}

type Arr struct {
	Values []*Num
}

func (a *Arr) Stringify() string {
	var vals []string
	var max float64
	for _, val := range a.Values {
		v, _ := val.Value.Float64()
		if v > max {
			max = v
		}
	}
	formatter := fmt.Sprintf("%% %ds", len(strconv.Itoa(int(max))))
	for i, val := range a.Values {
		vals = append(vals, fmt.Sprintf(formatter, val.Stringify()))
		if (i+1)%10 == 0 {
			vals = append(vals, "\n ")
		}
	}
	return strings.Join(vals, " ")
}

type Op struct {
	Impl map[ty]handler
}

func (op *Op) Dispatch(env *Environment, vals ...Value) (Value, error) {
	if len(vals) != 2 {
		return nil, fmt.Errorf("expecting 2 arguments but got %d", len(vals))
	}

	a1, a2 := vals[0], vals[1]
	t1, t2 := Ty(a1), Ty(a2)
	handler, ok := op.Impl[t1|t2]
	if !ok {
		return nil, fmt.Errorf("operator does not implement %s|%s", t1, t2)
	}

	return handler(env, a1, a2)
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
