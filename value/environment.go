package value

type Environment struct {
	ops map[string]*Op
	fns map[string]*Fn
	val map[string]Value
}

func (env *Environment) HasOp(id string) bool {
	_, ok := env.ops[id]
	return ok
}

func (env *Environment) HasFn(id string) bool {
	_, ok := env.fns[id]
	return ok
}

func (env *Environment) HasVal(id string) bool {
	_, ok := env.val[id]
	return ok
}

func (env *Environment) GetOp(id string) *Op {
	return env.ops[id]
}

func (env *Environment) GetFn(id string) *Fn {
	return env.fns[id]
}

func (env *Environment) GetVal(id string) Value {
	return env.val[id]
}

func (env *Environment) SetVal(id string, val Value) {
	env.val[id] = val
}

func (env *Environment) SetFn(id string, fn *Fn) {
	env.fns[id] = fn
}

func (env *Environment) SetOp(id string, op *Op) {
	env.ops[id] = op
}

func NewEnvironment() *Environment {
	return &Environment{
		val: make(map[string]Value),
		ops: map[string]*Op{
			"!=":  set,
			"*":   mul,
			"+":   add,
			"..":  range_,
			"@":   access,
			"---": g_take,

			// Placeholders for special operators
			":=": &Op{},
		},
		fns: map[string]*Fn{
			"...":  until,
			"...$": g_until,
			"abs":  abs,
			"len":  len_,
			"neg":  neg,
		},
	}
}
