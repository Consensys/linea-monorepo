package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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
	X                  gnarkfext.E4Gen                 `gnark:",public"`
	Ys                 []gnarkfext.E4Gen               `gnark:",public"`
	Commitments        [][blockSize]zk.WrappedVariable `gnark:",public"`
	Pubs               []zk.WrappedVariable            `gnark:",public"`
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
		Pubs:               make([]zk.WrappedVariable, len(comp.PublicInputs)),
		Commitments:        make([][blockSize]zk.WrappedVariable, len(merkleRoots)),
		Ys:                 make([]gnarkfext.E4Gen, len(polyQuery.Pols)),
	}
}

// Define implements the [frontend.Circuit] interface.
func (r *RecursionCircuit) Define(api frontend.API) error {
	eapi, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}
	apiGen, err := zk.NewGenericApi(api)
		if err != nil {
			panic(err)
		}
	w := r.WizardVerifier

	if !r.withoutGkr {
		w.HasherFactory = hasher_factory.NewKoalaBearHasherFactory(apiGen.NativeApi)
		w.BLSFS = fiatshamir.NewGnarkFSKoalabear(apiGen.NativeApi)
	}

	if r.withExternalHasher {
		w.HasherFactory = hasher_factory.NewKoalaBearHasherFactory(apiGen.NativeApi)
	}

	w.Verify(apiGen.NativeApi)

	for i := range r.Pubs {
		pub := w.Spec.PublicInputs[i].Acc.GetFrontendVariable(apiGen.NativeApi, w)
		apiGen.AssertIsEqual(r.Pubs[i], pub)
	}

	polyParams := w.GetUnivariateParams(r.PolyQuery.Name())
	eapi.AssertIsEqual(&r.X, &polyParams.ExtX)

	for i := range polyParams.ExtYs {
		eapi.AssertIsEqual(&r.Ys[i], &polyParams.ExtYs[i])
	}

	for i := range r.Commitments {
		for j := 0; j < blockSize; j++ {
			apiGen.AssertIsEqual(r.Commitments[i][j], r.MerkleRoots[i][j].GetColAssignmentGnarkAt(w, 0))
		}
	}

	return nil
}

// AssignRecursionCircuit assigns a recursion based on a compiled-IOP
// and a proof.
func AssignRecursionCircuit(comp *wizard.CompiledIOP, proof wizard.Proof, pubs []field.Element, finalFsState field.Octuplet) *RecursionCircuit {

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
			circuit.Commitments[0][i] = mRoot[i].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0)
		}
	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i][0] != nil {
			mRoot := pcsCtx.Items.MerkleRoots[i]
			circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
			for j := 0; j < blockSize; j++ {
				if pcsCtx.IsNonEmptyPrecomputed() {
					circuit.Commitments[i+1][j] = mRoot[j].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0)
				} else {
					circuit.Commitments[i][j] = mRoot[j].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0)
				}
			}
		}
	}

	return circuit
}

// SplitPublicInputs parses a vector of field elements and returns the
// parsed arguments.
// TODO@yao : check
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
	// X                          [4]frontend.Variable   `gnark:",public"`
	// Ys                         [4*numYs]frontend.Variable `gnark:",public"`
	// Commitments/merkleRoots    [8*numMRoots]frontend.Variable `gnark:",public"`
	// Pubs                       []frontend.Variable `gnark:",public"`

	//
	x, allPubDrain = allPubDrain[:4], allPubDrain[4:]
	ys, allPubDrain = allPubDrain[:4*numYs], allPubDrain[4*numYs:]
	mRoots, allPubDrain = allPubDrain[:8*numMRoots], allPubDrain[8*numMRoots:]
	pubs, _ = allPubDrain[:numPubs], allPubDrain[numPubs:]

	return x, ys, mRoots, pubs
}
