package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// RecursionCircuit is a gnark-circuit doing the recursion of a
// wizard-IOP. It embeds the [wizard.VerifierCircuit] and it
// additionally exposes the final FS state, the vortex commitments
// and the vortex statement (in addition to the "regular" public
// inputs of the protocol).
//
// It implements the [frontend.Circuit] interface.
//
// Alex: please don't change the ordering of the arguments as this
// affects the parsing of the witness.
type RecursionCircuit[T zk.Element] struct {
	X                  gnarkfext.E4Gen[T]   `gnark:",public"`
	Ys                 []gnarkfext.E4Gen[T] `gnark:",public"`
	Commitments        []T                  `gnark:",public"`
	Pubs               []T                  `gnark:",public"`
	WizardVerifier     *wizard.VerifierCircuit
	withoutGkr         bool                    `gnark:"-"`
	withExternalHasher bool                    `gnark:"-"`
	PolyQuery          query.UnivariateEval[T] `gnark:"-"`
	MerkleRoots        []ifaces.Column[T]      `gnark:"-"`
}

// AllocRecursionCircuit allocates a new RecursionCircuit with the
// given parameters.
func AllocRecursionCircuit[T zk.Element](comp *wizard.CompiledIOP[T][T], withoutGkr bool, withExternalHasher bool) *RecursionCircuit[T] {

	var (
		pcsCtx      = comp.PcsCtxs.(*vortex.Ctx)
		polyQuery   = pcsCtx.Query
		numRound    = comp.QueriesParams.Round(polyQuery.QueryID) + 1
		merkleRoots = []ifaces.Column[T]{}
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot != nil {
		merkleRoots = append(merkleRoots, pcsCtx.Items.Precomputeds.MerkleRoot)
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i] != nil {
			merkleRoots = append(merkleRoots, pcsCtx.Items.MerkleRoots[i])
		}
	}

	return &RecursionCircuit[T]{
		withoutGkr:         withoutGkr,
		withExternalHasher: withExternalHasher,
		PolyQuery:          polyQuery,
		MerkleRoots:        merkleRoots,
		WizardVerifier:     wizard.AllocateWizardCircuit(comp, numRound),
		Pubs:               make([]T, len(comp.PublicInputs)),
		Commitments:        make([]T, len(merkleRoots)),
		Ys:                 make([]gnarkfext.E4Gen[T], len(polyQuery.Pols)),
	}
}

// Define implements the [frontend.Circuit] interface.
func (r *RecursionCircuit[T]) Define(api frontend.API) error {

	w := r.WizardVerifier

	if !r.withoutGkr {
		temp := gkrmimc.NewHasherFactory(api)
		w.HasherFactory = temp
		w.FS = fiatshamir.NewGnarkFiatShamir(api, w.HasherFactory)
	}

	// if r.withExternalHasher {
	// 	w.HasherFactory = &mimc.ExternalHasherFactory{Api: api} // TODO: fix in crypto/mimc/factories.go
	// }

	w.Verify(api)

	for i := range r.Pubs {
		pub := w.Spec.PublicInputs[i].Acc.GetFrontendVariable(api, w)
		api.AssertIsEqual(r.Pubs[i], pub)
	}

	polyParams := w.GetUnivariateParams(r.PolyQuery.Name())
	api.AssertIsEqual(r.X, polyParams.X)

	for i := range polyParams.Ys {
		api.AssertIsEqual(r.Ys[i], polyParams.Ys[i])
	}

	for i := range r.Commitments {
		api.AssertIsEqual(r.Commitments[i], r.MerkleRoots[i].GetColAssignmentGnarkAt(w, 0))
	}

	return nil
}

// AssignRecursionCircuit assigns a recursion based on a compiled-IOP
// and a proof.
func AssignRecursionCircuit[T zk.Element](comp *wizard.CompiledIOP[T][T], proof wizard.Proof, pubs []field.Element, finalFsState field.Element) *RecursionCircuit[T] {

	var (
		pcsCtx         = comp.PcsCtxs.(*vortex.Ctx)
		polyQuery      = pcsCtx.Query
		numRound       = comp.QueriesParams.Round(polyQuery.QueryID) + 1
		wizardVerifier = wizard.AssignVerifierCircuit(comp, proof, numRound)
		params         = wizardVerifier.GetUnivariateParams(polyQuery.Name())
		circuit        = &RecursionCircuit[T]{
			WizardVerifier: wizard.AssignVerifierCircuit(comp, proof, numRound),
			X:              params.ExtX,
			Ys:             params.ExtYs,
			Pubs:           vector.IntoGnarkAssignment[T](pubs),
			PolyQuery:      polyQuery,
		}
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot != nil {
		mRoot := pcsCtx.Items.Precomputeds.MerkleRoot
		circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
		circuit.Commitments = append(circuit.Commitments, mRoot.GetColAssignmentGnarkAt(circuit.WizardVerifier, 0))
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i] != nil {
			mRoot := pcsCtx.Items.MerkleRoots[i]
			circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
			circuit.Commitments = append(circuit.Commitments, mRoot.GetColAssignmentGnarkAt(circuit.WizardVerifier, 0))
		}
	}

	return circuit
}

// SplitPublicInputs parses a vector of field elements and returns the
// parsed arguments.
func SplitPublicInputs[R any, T zk.Element](r *Recursion[T], allPubs []R) (x R, ys, mRoots, pubs []R) {

	var (
		numPubs     = len(r.InputCompiledIOP.PublicInputs)
		pcsCtx      = r.PcsCtx[0]
		numYs       = len(pcsCtx.Query.Pols)
		numMRoots   = 0
		allPubDrain = allPubs
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot != nil {
		numMRoots++
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i] != nil {
			numMRoots++
		}
	}

	// The order below is based on the field declaration order for the
	// circuit struct.
	//
	// X              T   `gnark:",public"`
	// Ys             []T `gnark:",public"`
	// Commitments    []T `gnark:",public"`
	// Pubs           []T `gnark:",public"`
	//
	x, allPubDrain = allPubDrain[0], allPubDrain[1:]
	ys, allPubDrain = allPubDrain[:numYs], allPubDrain[numYs:]
	mRoots, allPubDrain = allPubDrain[:numMRoots], allPubDrain[numMRoots:]
	pubs, _ = allPubDrain[:numPubs], allPubDrain[numPubs:]

	return x, ys, mRoots, pubs
}
