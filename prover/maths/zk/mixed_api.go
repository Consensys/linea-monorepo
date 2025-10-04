package zk

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

type VType uint

const (
	Native VType = iota
	Emulated
)

type WrappedVariable struct {
	V  frontend.Variable
	EV emulated.Element[emulated.KoalaBear]
}

func (w *WrappedVariable) AsNative() frontend.Variable {
	return w.V
}

func (w *WrappedVariable) AsEmulated() *emulated.Element[emulated.KoalaBear] {
	return &w.EV
}

func (w *WrappedVariable) Initialize(field *big.Int) {
	if field.Cmp(koalabear.Modulus()) == 0 {
		return
	} else {
		w.EV.Initialize(field)
	}
}

func ValueOf(v any) WrappedVariable {
	var res WrappedVariable
	res.EV = emulated.ValueOf[emulated.KoalaBear](v)
	res.V = v
	return res
}

type GenericApi struct {
	nativeApi   frontend.API
	emulatedApi *emulated.Field[emulated.KoalaBear]
}

func NewGenericApi(api frontend.API) (GenericApi, error) {
	ff := api.Compiler().Field()
	kf := koalabear.Modulus()
	if ff.Cmp(kf) == 0 {
		return GenericApi{nativeApi: api}, nil
	} else {
		f, err := emulated.NewField[emulated.KoalaBear](api)
		if err != nil {
			return GenericApi{}, err
		}
		return GenericApi{nativeApi: api, emulatedApi: f}, nil
	}
}

func (g *GenericApi) Type() VType {
	if g.emulatedApi == nil {
		return Native
	}
	return Emulated
}

func (g *GenericApi) Mul(a, b *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Mul(a.AsNative(), b.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Mul(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) MulConst(a *WrappedVariable, b *big.Int) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Mul(a.AsNative(), b)}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.MulConst(a.AsEmulated(), b)}
	}
}

func (g *GenericApi) Add(a, b *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Add(a.AsNative(), b.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Add(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) Neg(a *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Neg(a.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Neg(a.AsEmulated())}
	}
}

func (g *GenericApi) Sub(a, b *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Sub(a.AsNative(), b.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Sub(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) Inverse(a *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Inverse(a.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Inverse(a.AsEmulated())}
	}
}

func (g *GenericApi) Div(a, b *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Div(a.AsNative(), b.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Div(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) ToBinary(a *WrappedVariable, n ...int) []frontend.Variable {
	if g.Type() == Native {
		return g.nativeApi.ToBinary(a.AsNative(), n...)
	} else {
		return g.emulatedApi.ToBits(a.AsEmulated())
	}
}

func (g *GenericApi) FromBinary(b ...frontend.Variable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.FromBinary(b)}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.FromBits(b)}
	}
}

func (g *GenericApi) And(a, b frontend.Variable) frontend.Variable {
	return g.nativeApi.And(a, b)
}

func (g *GenericApi) Xor(a, b frontend.Variable) frontend.Variable {
	return g.nativeApi.Xor(a, b)
}

func (g *GenericApi) Or(a, b frontend.Variable) frontend.Variable {
	return g.nativeApi.Or(a, b)
}

// Select(b frontend.Variable, i1, i2 *T) *T
func (g *GenericApi) Select(b frontend.Variable, i1, i2 *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Select(b, i1.AsNative(), i2.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.FromBits(b)}
	}
}

func (g *GenericApi) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 *WrappedVariable) *WrappedVariable {
	if g.Type() == Native {
		return &WrappedVariable{V: g.nativeApi.Lookup2(
			b0, b1,
			i0.AsNative(), i1.AsNative(), i1.AsNative(), i3.AsNative())}
	} else {
		return &WrappedVariable{EV: *g.emulatedApi.Lookup2(
			b0, b1,
			i0.AsEmulated(), i1.AsEmulated(), i1.AsEmulated(), i3.AsEmulated())}
	}
}

func (g *GenericApi) IsZero(i1 *WrappedVariable) frontend.Variable {
	if g.Type() == Native {
		return g.nativeApi.IsZero(i1.AsNative())
	} else {
		return g.emulatedApi.IsZero(i1.AsEmulated())
	}
}

func (g *GenericApi) AssertIsEqual(a, b *WrappedVariable) {
	if g.Type() == Native {
		g.nativeApi.AssertIsEqual(a.AsNative(), b.AsNative())
	} else {
		g.emulatedApi.AssertIsEqual(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) AssertIsDifferent(a, b *WrappedVariable) {
	if g.Type() == Native {
		g.nativeApi.AssertIsDifferent(a.AsNative(), b.AsNative())
	} else {
		g.emulatedApi.AssertIsDifferent(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) AssertIsLessOrEqual(a, b *WrappedVariable) {
	if g.Type() == Native {
		g.nativeApi.AssertIsLessOrEqual(a.AsNative(), b.AsNative())
	} else {
		g.emulatedApi.AssertIsLessOrEqual(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) NewHint(f solver.Hint, nbOutputs int, inputs ...*WrappedVariable) ([]*WrappedVariable, error) {
	if g.Type() == Native {
		_inputs := make([]frontend.Variable, len(inputs))
		for i, r := range inputs {
			_inputs[i] = r.AsNative()
		}
		_r, err := g.nativeApi.NewHint(f, nbOutputs, _inputs...)
		if err != nil {
			return nil, err
		}
		res := make([]*WrappedVariable, nbOutputs)
		for i, r := range _r {
			res[i] = &WrappedVariable{V: r}
		}
		return res, nil
	} else {
		_inputs := make([]*emulated.Element[emulated.KoalaBear], len(inputs))
		for i, r := range inputs {
			_inputs[i] = r.AsEmulated()
		}
		_r, err := g.emulatedApi.NewHint(f, nbOutputs, _inputs...)
		if err != nil {
			return nil, err
		}
		res := make([]*WrappedVariable, nbOutputs)
		for i, r := range _r {
			res[i] = &WrappedVariable{EV: *r}
		}
		return res, nil
	}
}

func (g *GenericApi) Println(a ...*WrappedVariable) {
	if g.Type() == Native {
		for i := 0; i < len(a); i++ {
			g.nativeApi.Println(a[i].AsNative())
		}
	} else {
		for i := 0; i < len(a); i++ {
			for j := 0; j < len(a[i].EV.Limbs); j++ {
				g.nativeApi.Println(a[i].EV.Limbs[j])
			}
		}
	}
}
