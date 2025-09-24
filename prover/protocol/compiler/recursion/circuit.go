package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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
type RecursionCircuit struct {
	X                  gnarkfext.Element   `gnark:",public"`
	Ys                 []gnarkfext.Element `gnark:",public"`
	Commitments        []frontend.Variable `gnark:",public"`
	Pubs               []frontend.Variable `gnark:",public"`
	WizardVerifier     *wizard.VerifierCircuit
	withoutGkr         bool                       `gnark:"-"`
	withExternalHasher bool                       `gnark:"-"`
	PolyQuery          query.UnivariateEval       `gnark:"-"`
	MerkleRoots        [][blockSize]ifaces.Column `gnark:"-"`
}

// AllocRecursionCircuit allocates a new RecursionCircuit with the
// given parameters.
func AllocRecursionCircuit(comp *wizard.CompiledIOP, withoutGkr bool, withExternalHasher bool) *RecursionCircuit {

	var (
		pcsCtx      = comp.PcsCtxs.(*vortex.Ctx)
		polyQuery   = pcsCtx.Query
		numRound    = comp.QueriesParams.Round(polyQuery.QueryID) + 1
		merkleRoots = [][blockSize]ifaces.Column{}
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot[0] != nil {
		merkleRoots = append(merkleRoots, pcsCtx.Items.Precomputeds.MerkleRoot)
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i][0] != nil {
			merkleRoots = append(merkleRoots, pcsCtx.Items.MerkleRoots[i])
		}
	}

	return &RecursionCircuit{
		withoutGkr:         withoutGkr,
		withExternalHasher: withExternalHasher,
		PolyQuery:          polyQuery,
		MerkleRoots:        merkleRoots,
		WizardVerifier:     wizard.AllocateWizardCircuit(comp, numRound),
		Pubs:               make([]frontend.Variable, len(comp.PublicInputs)),
		Commitments:        make([]frontend.Variable, len(merkleRoots)),
		Ys:                 make([]gnarkfext.Element, len(polyQuery.Pols)),
	}
}

// Define implements the [frontend.Circuit] interface.
func (r *RecursionCircuit) Define(api frontend.API) error {

	w := r.WizardVerifier

	if !r.withoutGkr {
		temp := gkrmimc.NewHasherFactory(api)
		w.HasherFactory = temp
		w.FS = fiatshamir.NewGnarkFiatShamir(api, w.HasherFactory)
	}

	if r.withExternalHasher {
		w.HasherFactory = &mimc.ExternalHasherFactory{Api: api} // TODO: fix in crypto/mimc/factories.go
	}

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
		api.AssertIsEqual(r.Commitments[i], r.MerkleRoots[i/blockSize][i%blockSize].GetColAssignmentGnarkAt(w, 0))
	}

	return nil
}

// AssignRecursionCircuit assigns a recursion based on a compiled-IOP
// and a proof.
func AssignRecursionCircuit(comp *wizard.CompiledIOP, proof wizard.Proof, pubs []field.Element, finalFsState field.Element) *RecursionCircuit {

	var (
		pcsCtx         = comp.PcsCtxs.(*vortex.Ctx)
		polyQuery      = pcsCtx.Query
		numRound       = comp.QueriesParams.Round(polyQuery.QueryID) + 1
		wizardVerifier = wizard.AssignVerifierCircuit(comp, proof, numRound)
		params         = wizardVerifier.GetUnivariateParams(polyQuery.Name())
		circuit        = &RecursionCircuit{
			WizardVerifier: wizard.AssignVerifierCircuit(comp, proof, numRound),
			X:              params.ExtX,
			Ys:             params.ExtYs,
			Pubs:           vector.IntoGnarkAssignment(pubs),
			PolyQuery:      polyQuery,
		}
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot[0] != nil {
		mRoot := pcsCtx.Items.Precomputeds.MerkleRoot
		circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
		for i := 0; i < blockSize; i++ {
			circuit.Commitments = append(circuit.Commitments, mRoot[i].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0))
		}
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i][0] != nil {
			mRoot := pcsCtx.Items.MerkleRoots[i]
			circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
			for j := 0; j < blockSize; j++ {
				circuit.Commitments = append(circuit.Commitments, mRoot[j].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0))
			}
		}
	}

	return circuit
}

// SplitPublicInputs parses a vector of field elements and returns the
// parsed arguments.
func SplitPublicInputs[T any](r *Recursion, allPubs []T) (x, ys, mRoots, pubs []T) {

	var (
		numPubs     = len(r.InputCompiledIOP.PublicInputs)
		pcsCtx      = r.PcsCtx[0]
		numYs       = len(pcsCtx.Query.Pols)
		numMRoots   = 0
		allPubDrain = allPubs
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot[0] != nil {
		numMRoots++
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i][0] != nil {
			numMRoots++
		}
	}

	// The order below is based on the field declaration order for the
	// circuit struct.
	//
	// X              frontend.Variable   `gnark:",public"`
	// Ys             []frontend.Variable `gnark:",public"`
	// Commitments    []frontend.Variable `gnark:",public"`
	// Pubs           []frontend.Variable `gnark:",public"`
	//
	x, allPubDrain = allPubDrain[:4], allPubDrain[4:]
	ys, allPubDrain = allPubDrain[:4*numYs], allPubDrain[4*numYs:]
	mRoots, allPubDrain = allPubDrain[:8*numMRoots], allPubDrain[8*numMRoots:]
	pubs, _ = allPubDrain[:numPubs], allPubDrain[numPubs:]

	return x, ys, mRoots, pubs
}
