package value

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type handler func(*Environment, ...Value) (Value, error)
type fntable map[signature]handler

type Value interface {
	Stringify() string
}

type Num struct {
	Value *big.Float
}

func (n *Num) Stringify() string {
	return n.Value.Text('g', -1)
}

type stepper func(Value, int) (next Value, done bool, ok bool)

type Gen struct {
	ty    ty
	done  bool
	curr  Value
	next  stepper
	steps []stepper
}

func (g *Gen) Stringify() string {
	return fmt.Sprintf("generator%s", g.ty)
}

func (g *Gen) With(step stepper) *Gen {
	return &Gen{
		ty:    g.ty,
		done:  g.done,
		curr:  g.curr,
		next:  g.next,
		steps: append(g.steps, step),
	}
}

func (g *Gen) Next() (Value, bool) {
	res := g.curr
	if g.done {
		return res, false
	}

	curr, done, ok := g.next(g.curr, 1)
	if !ok {
		return nil, false
	}

	g.curr = curr
	g.done = done
	if g.done {
		return res, false
	}

	for _, step := range g.steps {
		if g.done {
			break
		}

		val, done, ok := step(res, 1)
		if !ok {
			return nil, false
		}

		res = val
		if done {
			return res, false
		}

		g.done = done
	}

	if g.done {
		return res, false
	}
	return res, true
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
	Impl fntable
}

func (op *Op) Dispatch(env *Environment, vals ...Value) (Value, error) {
	if len(vals) != 2 {
		return nil, fmt.Errorf("expecting 2 arguments but got %d", len(vals))
	}

	a1, a2 := vals[0], vals[1]
	t1, t2 := Ty(a1), Ty(a2)
	handler, ok := op.Impl[sig(t1, t2)]
	if !ok {
		return nil, fmt.Errorf("operator does not implement %s", sig(t1, t2))
	}

	return handler(env, a1, a2)
}

func (*Op) Stringify() string {
	return "op"
}

type Fn struct {
	Argc int
	Impl fntable
}

func (fn *Fn) Stringify() string {
	return fmt.Sprintf("fn/%d", fn.Argc)
}

func (fn *Fn) Dispatch(env *Environment, vals ...Value) (Value, error) {
	if len(vals) != fn.Argc {
		return nil, fmt.Errorf("expecting %d arguments but got %d",
			fn.Argc, len(vals))
	}

	var tys []ty
	for _, arg := range vals {
		tys = append(tys, Ty(arg))
	}

	handler, ok := fn.Impl[sig(tys...)]
	if !ok {
		return nil, fmt.Errorf("function does not implement %s", sig(tys...))
	}

	return handler(env, vals...)
}
