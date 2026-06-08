package field

// Kind identifies which field a column, expression, or DAG node lives in.
type Kind uint8

const (
	Base Kind = iota
	Ext
)

func (k Kind) String() string {
	switch k {
	case Base:
		return "base"
	case Ext:
		return "ext"
	default:
		return "unknown"
	}
}

// Join returns the field needed to contain values from both inputs.
func Join(a, b Kind) Kind {
	if a == Ext || b == Ext {
		return Ext
	}
	return Base
}

// JoinAll returns the field needed to contain every input.
func JoinAll(fields ...Kind) Kind {
	res := Base
	for _, f := range fields {
		res = Join(res, f)
	}
	return res
}
