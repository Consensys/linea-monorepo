// Stub file for circuit.go during koalabear migration.
// Remove this file and the //go:build ignore tag from circuit.go when migration is complete.
package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// RecursionCircuit is a stub type during migration
type RecursionCircuit struct {
	X                  gnarkfext.E4Gen      `gnark:",public"`
	Ys                 []gnarkfext.E4Gen    `gnark:",public"`
	Commitments        []zk.WrappedVariable `gnark:",public"`
	Pubs               []zk.WrappedVariable `gnark:",public"`
	WizardVerifier     *wizard.VerifierCircuit
	withoutGkr         bool                       `gnark:"-"`
	withExternalHasher bool                       `gnark:"-"`
	PolyQuery          query.UnivariateEval       `gnark:"-"`
	MerkleRoots        [][blockSize]ifaces.Column `gnark:"-"`
}

// AllocRecursionCircuit is a stub during migration
func AllocRecursionCircuit(comp *wizard.CompiledIOP, withoutGkr bool, withExternalHasher bool) *RecursionCircuit {
	panic("recursion/circuit.go is stubbed during migration")
}

// Define is a stub during migration
func (r *RecursionCircuit) Define(api frontend.API) error {
	panic("recursion/circuit.go is stubbed during migration")
}

// AssignRecursionCircuit is a stub during migration
func AssignRecursionCircuit(comp *wizard.CompiledIOP, proof wizard.Proof, pubs []field.Element, finalFsState field.Octuplet) *RecursionCircuit {
	panic("recursion/circuit.go is stubbed during migration")
}

// SplitPublicInputs is a stub during migration
func SplitPublicInputs[T any](r *Recursion, allPubs []T) (x, ys, mRoots, pubs []T) {
	panic("recursion/circuit.go is stubbed during migration")
}

