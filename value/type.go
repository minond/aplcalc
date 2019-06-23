package value

type ty uint8

const (
	TUnknown ty = 1 << iota
	TArr
	TNum
)

func (ty ty) String() string {
	switch ty {
	case TArr:
		return "<array>"
	case TNum:
		return "<number>"
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
	default:
		return TUnknown
	}
}
