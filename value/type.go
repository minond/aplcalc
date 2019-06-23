package value

import "strings"

type signature string

func sig(tys ...ty) signature {
	args := make([]string, len(tys))
	for i := range tys {
		args[i] = tys[i].String()
	}
	return signature(strings.Join(args, "/"))
}

type ty uint8

const (
	TUnknown ty = 1 << iota
	TArr
	TNum
	TGen
)

func (ty ty) String() string {
	switch ty {
	case TArr:
		return "<array>"
	case TNum:
		return "<number>"
	case TGen:
		return "<generator>"
	default:
		return "<unknown>"
	}
}

func Ty(v Value) ty {
	switch v.(type) {
	case *Arr:
		return TArr
	case *Num:
		return TNum
	case *Gen:
		return TGen
	default:
		return TUnknown
	}
}
