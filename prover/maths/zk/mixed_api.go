package zk

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/selector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// cachedKoalabearModulus caches koalabear.Modulus() to avoid repeated allocations
var cachedKoalabearModulus = koalabear.Modulus()

type VType uint

const (
	Native VType = iota
	Emulated
)

type WrappedVariable struct {
	V frontend.Variable

	// The non pointer version is needed when instantiating a circuit
	EV        emulated.Element[emulated.KoalaBear]
	EVpointer *emulated.Element[emulated.KoalaBear]
}

type Octuplet [8]WrappedVariable

func (w *WrappedVariable) AsNative() frontend.Variable {
	if w.V != nil {
		return w.V
	}
	if len(w.EV.Limbs) == 1 {
		return w.EV.Limbs[0]
	}
	if len(w.EVpointer.Limbs) == 1 {
		return w.EVpointer.Limbs[0]
	}
	utils.Panic("unexpected shape for wrapped variable: %++v", w)
	return nil // unreachable
}

func (w *WrappedVariable) AsEmulated() *emulated.Element[emulated.KoalaBear] {
	if w.EVpointer == nil {
		return &w.EV
	} else {
		return w.EVpointer
	}
}

func WrapFrontendVariable(v frontend.Variable) WrappedVariable {

	switch v.(type) {
	case WrappedVariable, *WrappedVariable:
		panic("attempted to wrap a zk.WrappedVariable into a zk.WrappedVariable")
	}

	var res WrappedVariable
	res.V = v
	res.EV = emulated.Element[emulated.KoalaBear]{Limbs: []frontend.Variable{v}}
	return res
}

func (w *WrappedVariable) IsEmpty() bool {
	return w.AsNative() == nil && w.V == nil && len(w.EV.Limbs) == 0
}

func (w *WrappedVariable) Initialize(field *big.Int) {
	if field.Cmp(cachedKoalabearModulus) == 0 {
		return
	} else {
		w.EV.Initialize(field)
		// w.EVpointer = &w.EV
	}
}

func ValueOf(v any) WrappedVariable {
	var res WrappedVariable
	res.EV = emulated.ValueOf[emulated.KoalaBear](v)
	res.V = v
	return res
}

func ValueFromKoala(v field.Element) WrappedVariable {
	return ValueOf(v.Uint64())
}

type GenericApi struct {
	NativeApi   frontend.API
	EmulatedApi *emulated.Field[emulated.KoalaBear]
}

func NewGenericApi(api frontend.API) (GenericApi, error) {
	ff := api.Compiler().Field()
	if ff.Cmp(cachedKoalabearModulus) == 0 {
		return GenericApi{NativeApi: api}, nil
	} else {
		f, err := emulated.NewField[emulated.KoalaBear](api)
		if err != nil {
			return GenericApi{}, err
		}
		return GenericApi{NativeApi: api, EmulatedApi: f}, nil
	}
}

func MustMakeGenericApi(api frontend.API) GenericApi {
	g, err := NewGenericApi(api)
	if err != nil {
		panic(err)
	}
	return g
}

func (g *GenericApi) Type() VType {
	if g.EmulatedApi == nil {
		return Native
	}
	return Emulated
}

func (g *GenericApi) GetFrontendVariable(v WrappedVariable) frontend.Variable {
	if g.EmulatedApi == nil {
		return v.V
	}
	return v.EV.Limbs[0]
}

func (g *GenericApi) Mul(a, b WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Mul(a.AsNative(), b.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Mul(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) MulConst(a WrappedVariable, b *big.Int) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Mul(a.AsNative(), b)}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.MulConst(a.AsEmulated(), b)}
	}
}

func (g *GenericApi) Add(a, b WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Add(a.AsNative(), b.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Add(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) Neg(a WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Neg(a.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Neg(a.AsEmulated())}
	}
}

func (g *GenericApi) Sub(a, b WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Sub(a.AsNative(), b.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Sub(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) Inverse(a WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Inverse(a.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Inverse(a.AsEmulated())}
	}
}

func (g *GenericApi) Div(a, b WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Div(a.AsNative(), b.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Div(a.AsEmulated(), b.AsEmulated())}
	}
}

func (g *GenericApi) ToBinary(a WrappedVariable, n ...int) []frontend.Variable {
	if g.Type() == Native {
		return g.NativeApi.ToBinary(a.AsNative(), n...)
	} else {
		return g.EmulatedApi.ToBits(a.AsEmulated())
	}
}

func (g *GenericApi) FromBinary(b ...frontend.Variable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.FromBinary(b...)}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.FromBits(b...)}
	}
}

func (g *GenericApi) And(a, b frontend.Variable) frontend.Variable {
	return g.NativeApi.And(a, b)
}

func (g *GenericApi) Xor(a, b frontend.Variable) frontend.Variable {
	return g.NativeApi.Xor(a, b)
}

func (g *GenericApi) Or(a, b frontend.Variable) frontend.Variable {
	return g.NativeApi.Or(a, b)
}

// Select(b frontend.Variable, i1, i2 *T) *T
func (g *GenericApi) Select(b frontend.Variable, i1, i2 WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Select(b, i1.AsNative(), i2.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Select(b, i1.AsEmulated(), i2.AsEmulated())}
	}
}

func (g *GenericApi) Lookup2(b0, b1 frontend.Variable, i0, i1, i2, i3 WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		return WrappedVariable{V: g.NativeApi.Lookup2(
			b0, b1,
			i0.AsNative(), i1.AsNative(), i2.AsNative(), i3.AsNative())}
	} else {
		return WrappedVariable{EVpointer: g.EmulatedApi.Lookup2(
			b0, b1,
			i0.AsEmulated(), i1.AsEmulated(), i2.AsEmulated(), i3.AsEmulated())}
	}
}

func (g *GenericApi) IsZero(i1 WrappedVariable) frontend.Variable {
	if g.Type() == Native {
		return g.NativeApi.IsZero(i1.AsNative())
	} else {
		return g.EmulatedApi.IsZero(i1.AsEmulated())
	}
}

func (g *GenericApi) AssertIsEqual(a, b WrappedVariable) {
	if g.Type() == Native {
		g.NativeApi.AssertIsEqual(a.AsNative(), b.AsNative())
	} else {
		g.EmulatedApi.AssertIsEqual(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) AssertIsDifferent(a, b WrappedVariable) {
	if g.Type() == Native {
		g.NativeApi.AssertIsDifferent(a.AsNative(), b.AsNative())
	} else {
		g.EmulatedApi.AssertIsDifferent(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) AssertIsLessOrEqual(a, b WrappedVariable) {
	if g.Type() == Native {
		g.NativeApi.AssertIsLessOrEqual(a.AsNative(), b.AsNative())
	} else {
		g.EmulatedApi.AssertIsLessOrEqual(a.AsEmulated(), b.AsEmulated())
	}
}

func (g *GenericApi) NewHint(f solver.Hint, nbOutputs int, inputs ...WrappedVariable) ([]WrappedVariable, error) {
	if g.Type() == Native {
		_inputs := make([]frontend.Variable, len(inputs))
		for i, r := range inputs {
			_inputs[i] = r.AsNative()
		}
		_r, err := g.NativeApi.NewHint(f, nbOutputs, _inputs...)
		if err != nil {
			return nil, err
		}
		res := make([]WrappedVariable, nbOutputs)
		for i, r := range _r {
			res[i] = WrappedVariable{V: r}
		}
		return res, nil
	} else {
		_inputs := make([]*emulated.Element[emulated.KoalaBear], len(inputs))
		for i, r := range inputs {
			_inputs[i] = r.AsEmulated()
		}
		_r, err := g.EmulatedApi.NewHint(f, nbOutputs, _inputs...)
		if err != nil {
			return nil, err
		}
		res := make([]WrappedVariable, nbOutputs)
		for i, r := range _r {
			res[i] = WrappedVariable{EVpointer: r}
		}
		return res, nil
	}
}

// @thomas fixme this function does not print constants
func (g *GenericApi) Println(a ...WrappedVariable) {
	if g.Type() == Native {
		for i := 0; i < len(a); i++ {
			g.NativeApi.Println(a[i].AsNative())
		}
	} else {
		for i := 0; i < len(a); i++ {
			if a[i].EVpointer == nil {
				g.EmulatedApi.Reduce(&a[i].EV)
				for j := 0; j < len(a[i].EV.Limbs); j++ {
					g.NativeApi.Println(a[i].EV.Limbs[j])
				}
			} else {
				g.EmulatedApi.Reduce(a[i].EVpointer)
				for j := 0; j < len(a[i].EVpointer.Limbs); j++ {
					g.NativeApi.Println(a[i].EVpointer.Limbs[j])
				}
			}
		}
	}
}

func (g *GenericApi) Mux(sel frontend.Variable, inputs ...WrappedVariable) WrappedVariable {
	if g.Type() == Native {
		_inputs := make([]frontend.Variable, len(inputs))
		for i := 0; i < len(_inputs); i++ {
			_inputs[i] = inputs[i].AsNative()
		}
		res := selector.Mux(g.NativeApi, sel, _inputs...)
		return WrappedVariable{V: res}
	} else {
		_inputs := make([]*emulated.Element[emulated.KoalaBear], len(inputs))
		for i := 0; i < len(_inputs); i++ {
			_inputs[i] = inputs[i].AsEmulated()
		}
		res := g.EmulatedApi.Mux(sel, _inputs...)
		return WrappedVariable{EVpointer: res}
	}
}

func (o Octuplet) AsNative() [8]frontend.Variable {
	res := [8]frontend.Variable{}
	for i := range res {
		res[i] = o[i].AsNative()
		if res[i] == nil {
			panic("wrapped variable is nil")
		}
	}
	return res
}
