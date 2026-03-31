package fiatshamir

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/koalagnark"
)

// FiatShamir represents a Fiat-Shamir engine. The engine holds the transcript
// of a protocol instance: e.g. the list of all the messages sent by the
// prover (we disregard the verifier message because they are all public-coins).
type FiatShamir interface {
	Update(vec ...field.Element)
	UpdateExt(vec ...field.Ext)
	UpdateGeneric(vec ...field.FieldElem)
	UpdateSV(sv smartvectors.SmartVector)
	RandomField() field.Octuplet
	RandomFext() field.Ext
	RandomFieldFromSeed(seed field.Octuplet, name string) field.Ext
	RandomManyIntegers(num, upperBound int) []int
	SetState(s field.Octuplet)
	State() field.Octuplet
}

// FiatShamirGnark is as [FiatShamir] but for a gnark circuit context.
type FiatShamirGnark interface {
	Update(vec ...koalagnark.Element)
	UpdateExt(vec ...koalagnark.Ext)
	UpdateVec(mat ...[]koalagnark.Element)
	RandomField() koalagnark.Octuplet
	RandomFieldExt() koalagnark.Ext
	RandomManyIntegers(num, upperBound int) []frontend.Variable
	RandomFieldFromSeed(seed koalagnark.Octuplet, name string) koalagnark.Ext
	SetState(state koalagnark.Octuplet)
	State() koalagnark.Octuplet
}
