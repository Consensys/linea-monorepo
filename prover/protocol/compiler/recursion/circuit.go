package recursion

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/hasher_factory"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
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
	X                  ExtFrontendVariable            `gnark:",public"`
	Ys                 []ExtFrontendVariable          `gnark:",public"`
	Commitments        [][blockSize]frontend.Variable `gnark:",public"`
	Pubs               []frontend.Variable            `gnark:",public"`
	WizardVerifier     *wizard.VerifierCircuit
	withoutGkr         bool                       `gnark:"-"`
	withExternalHasher bool                       `gnark:"-"`
	PolyQuery          query.UnivariateEval       `gnark:"-"`
	MerkleRoots        [][blockSize]ifaces.Column `gnark:"-"`
}

// ExtFrontendVariable allows storing the extension as a 4-element array of frontend variables in Plonk public inputs (contrary to WrappedVariable/ koalagnark.Ext, which takes more space).
type ExtFrontendVariable = [4]frontend.Variable

// E4Gen is a helper function for converting an ExtFrontendVariable to a koalagnark.Ext
func E4Gen(x ExtFrontendVariable) koalagnark.Ext {
	return koalagnark.NewExtFrom4FrontendVars(x[0], x[1], x[2], x[3])
}

// Ext4FV is a helper function for converting a koalagnark.Ext to an ExtFrontendVariable
func Ext4FV(x koalagnark.Ext) ExtFrontendVariable {
	return ExtFrontendVariable{
		x.B0.A0.Native(),
		x.B0.A1.Native(),
		x.B1.A0.Native(),
		x.B1.A1.Native(),
	}
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

	// Count public input slots: base field = 1 slot, extension field = 4 slots (coordinates)
	numPubSlots := 0
	for i := range comp.PublicInputs {
		if !comp.PublicInputs[i].Acc.IsBase() {
			numPubSlots += 4 // extension field: 4 base field coordinates
		} else {
			numPubSlots++ // base field: 1 element
		}
	}

	return &RecursionCircuit{
		withoutGkr:         withoutGkr,
		withExternalHasher: withExternalHasher,
		PolyQuery:          polyQuery,
		MerkleRoots:        merkleRoots,
		WizardVerifier:     wizard.AllocateWizardCircuit(comp, numRound, false),
		Pubs:               make([]frontend.Variable, numPubSlots),
		Commitments:        make([][blockSize]frontend.Variable, len(merkleRoots)),
		Ys:                 make([]ExtFrontendVariable, len(polyQuery.Pols)),
	}
}

// Define implements the [frontend.Circuit] interface.
func (r *RecursionCircuit) Define(api frontend.API) error {
	koalaAPI := koalagnark.NewAPI(api)
	w := r.WizardVerifier

	// Setup HasherFactory if needed (for GKR or external hasher)
	if r.withExternalHasher {
		w.HasherFactory = hasher_factory.NewKoalaBearHasherFactory(api)
	}

	// Initialize Fiat-Shamir with external hasher if enabled
	// This must happen BEFORE calling Verify() which would otherwise overwrite it
	if r.withExternalHasher {
		w.HasherFactory = hasher_factory.NewKoalaBearHasherFactory(api)
	}

	w.Verify(api)

	for i := range r.Pubs {
		pub := w.Spec.PublicInputs[i].Acc.GetFrontendVariable(api, w)
		api.AssertIsEqual(r.Pubs[i], pub.Native())
	}

	polyParams := w.GetUnivariateParams(r.PolyQuery.Name())
	koalaAPI.AssertIsEqualExt(E4Gen(r.X), polyParams.ExtX)

	for i := range polyParams.ExtYs {
		koalaAPI.AssertIsEqualExt(E4Gen(r.Ys[i]), polyParams.ExtYs[i])
	}

	for i := range r.Commitments {
		for j := 0; j < blockSize; j++ {
			mr := r.MerkleRoots[i][j].GetColAssignmentGnarkAt(w, 0)
			api.AssertIsEqual(r.Commitments[i][j], mr.Native())
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
		wizardVerifier = wizard.AssignVerifierCircuit(comp, proof, numRound, false)
		params         = wizardVerifier.GetUnivariateParams(polyQuery.Name())
		circuit        = &RecursionCircuit{
			WizardVerifier: wizardVerifier,
			X:              Ext4FV(&params.ExtX),
			Ys:             make([]ExtFrontendVariable, len(params.ExtYs)),
			Pubs:           make([]frontend.Variable, len(pubs)),
			PolyQuery:      polyQuery,
		}
	)

	for i := range pubs {
		circuit.Pubs[i] = pubs[i]
	}

	for i := range params.ExtYs {
		circuit.Ys[i] = Ext4FV(params.ExtYs[i])
	}

	if pcsCtx.Items.Precomputeds.MerkleRoot[0] != nil {
		mRoot := pcsCtx.Items.Precomputeds.MerkleRoot
		circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
		octuplet := [8]frontend.Variable{}
		for j := 0; j < blockSize; j++ {
			a := mRoot[j].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0)
			octuplet[j] = a.Native()
		}
		circuit.Commitments = append(circuit.Commitments, octuplet)

	}

	for i := range pcsCtx.Items.MerkleRoots {
		if pcsCtx.Items.MerkleRoots[i][0] != nil {
			mRoot := pcsCtx.Items.MerkleRoots[i]
			circuit.MerkleRoots = append(circuit.MerkleRoots, mRoot)
			octuplet := [8]frontend.Variable{}
			for j := 0; j < blockSize; j++ {
				a := mRoot[j].GetColAssignmentGnarkAt(circuit.WizardVerifier, 0)
				octuplet[j] = a.Native()
			}
			circuit.Commitments = append(circuit.Commitments, octuplet)
		}
	}

	return circuit
}

// SplitPublicInputs parses a vector of field elements and returns the
// parsed arguments.
// @azam x, ys stored as field extension (4 field elements), mRoot 8 field elements, pubs stored as field element.
func SplitPublicInputs[T any](r *Recursion, allPubs []T) (x, ys, mRoots, pubs []T) {
	var (
		numPubSlots = 0
		pcsCtx      = r.PcsCtx[0]
		numYs       = len(pcsCtx.Query.Pols)
		numMRoots   = 0
		allPubDrain = allPubs
	)

	// Count public input slots: base field = 1 slot, extension field = 4 slots (coordinates)
	for i := range r.InputCompiledIOP.PublicInputs {
		if !r.InputCompiledIOP.PublicInputs[i].Acc.IsBase() {
			numPubSlots += 4 // extension field: 4 base field coordinates
		} else {
			numPubSlots++ // base field: 1 element
		}
	}

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
	pubs, _ = allPubDrain[:numPubSlots], allPubDrain[numPubSlots:]

	return x, ys, mRoots, pubs
}
