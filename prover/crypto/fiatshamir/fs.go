package fiatshamir

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// circuit
type GnarkFS interface {
	Update(vec ...zk.WrappedVariable)
	UpdateExt(vec ...gnarkfext.E4Gen)
	UpdateVec(mat ...[]zk.WrappedVariable)
	RandomField() zk.Octuplet
	RandomFieldExt() gnarkfext.E4Gen
	RandomManyIntegers(num, upperBound int) []frontend.Variable
	SetState(state zk.Octuplet)
	State() zk.Octuplet
}

func NewGnarkFSKoalabear(api frontend.API) GnarkFS {
	return fiatshamir_koalabear.NewGnarkFSWV(api)
}

// NewGnarkFSKoalabearWithFactory creates a Fiat-Shamir instance using the provided
// HasherFactory. This enables external hasher optimization when configured.
func NewGnarkFSKoalabearWithFactory(api frontend.API, factory hasher_factory.HasherFactory) GnarkFS {
	hasher := factory.NewHasher()
	return fiatshamir_koalabear.NewGnarkFSWVWithHasher(api, hasher)
}

func NewGnarkFSBLS12377(api frontend.API) GnarkFS {
	return fiatshamir_bls12377.NewGnarkFS(api)
}

// non circuit
type FS interface {
	Update(vec ...field.Element)
	UpdateExt(vec ...fext.Element)
	UpdateGeneric(vec ...fext.GenericFieldElem)
	UpdateSV(sv smartvectors.SmartVector)
	RandomField() field.Octuplet
	RandomFext() fext.Element
	RandomFieldFromSeed(seed field.Octuplet, name string) fext.Element
	RandomManyIntegers(num, upperBound int) []int
	SetState(s field.Octuplet)
	State() field.Octuplet
}

func NewFSKoalabear() FS {
	return fiatshamir_koalabear.NewFS()
}

func NewFSBls12377() FS {
	return fiatshamir_bls12377.NewFS()
}
