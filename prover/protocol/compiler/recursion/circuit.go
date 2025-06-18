package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
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
	withoutGkr         bool                 `gnark:"-"`
	withExternalHasher bool                 `gnark:"-"`
	PolyQuery          query.UnivariateEval `gnark:"-"`
	MerkleRoots        []ifaces.Column      `gnark:"-"`
}

// AllocRecursionCircuit allocates a new RecursionCircuit with the
// given parameters.
func AllocRecursionCircuit(comp *wizard.CompiledIOP, withoutGkr bool, withExternalHasher bool) *RecursionCircuit {

	var (
		pcsCtx      = comp.PcsCtxs.(*vortex.Ctx)
		polyQuery   = pcsCtx.Query
		numRound    = comp.QueriesParams.Round(polyQuery.QueryID) + 1
		merkleRoots = []ifaces.Column{}
	)

	if pcsCtx.Items.Precomputeds.MerkleRoot != nil {
		merkleRoots = append(merkleRoots, pcsCtx.Items.Precomputeds.MerkleRoot)
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i] != nil {
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
		api.AssertIsEqual(r.Commitments[i], r.MerkleRoots[i].GetColAssignmentGnarkAt(w, 0))
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
			X:              params.X,
			Ys:             params.Ys,
			Pubs:           vector.IntoGnarkAssignment(pubs),
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
// allPubs in Base,
// x,y in fext,
// mRoot in [8]Base

// TODO@yao check the type, x, ys in fext, mRoots in [8]field?
// they are all from allPubs, if they share different data types, what type does allPubs have
func SplitPublicInputs(r *Recursion, allPubs []field.Element) (x fext.Element, ys []fext.Element, mRoots, pubs []field.Element) {

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
	// X              fext.Element
	// Ys             []fext.Element
	// mRoots         []field.Element, each root is [8]field.Element
	// Pubs           []field.Element
	//
	x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1 = allPubDrain[0], allPubDrain[1], allPubDrain[2], allPubDrain[3]
	allPubDrain = allPubDrain[4:]
	for i := 0; i < numYs; i++ {
		ys[i].B0.A0, ys[i].B0.A1, ys[i].B1.A0, ys[i].B1.A1 = allPubDrain[4*i], allPubDrain[4*i+1], allPubDrain[4*i+2], allPubDrain[4*i+3]

	}
	allPubDrain = allPubDrain[4*numYs:]
	mRoots, allPubDrain = allPubDrain[:8*numMRoots], allPubDrain[8*numMRoots:]
	pubs, _ = allPubDrain[:numPubs], allPubDrain[numPubs:]

	return x, ys, mRoots, pubs
}

func SplitPublicInputsGnark(r *Recursion, allPubs []frontend.Variable) (x gnarkfext.Element, ys []gnarkfext.Element, mRoots, pubs []frontend.Variable) {

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
	// X              gnarkfext.Element   `gnark:",public"`
	// Ys             []gnarkfext.Element `gnark:",public"`
	// Commitments    []frontend.Variable `gnark:",public"`
	// Pubs           []frontend.Variable `gnark:",public"`
	//
	x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1 = allPubDrain[0], allPubDrain[1], allPubDrain[2], allPubDrain[3]
	allPubDrain = allPubDrain[4:]
	for i := 0; i < numYs; i++ {
		ys[i].B0.A0, ys[i].B0.A1, ys[i].B1.A0, ys[i].B1.A1 = allPubDrain[4*i], allPubDrain[4*i+1], allPubDrain[4*i+2], allPubDrain[4*i+3]

	}
	allPubDrain = allPubDrain[4*numYs:]
	mRoots, allPubDrain = allPubDrain[:8*numMRoots], allPubDrain[8*numMRoots:]
	pubs, _ = allPubDrain[:numPubs], allPubDrain[numPubs:]

	return x, ys, mRoots, pubs
}
