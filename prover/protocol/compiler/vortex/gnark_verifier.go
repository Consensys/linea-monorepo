package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	vortex_bls12377 "github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func (a *ExplicitPolynomialEval) RunGnark(api frontend.API, c wizard.GnarkRuntime) {
	fmt.Printf("gnark vortex ExplicitPolynomialEval ...\n")
	a.gnarkExplicitPublicEvaluation(api, c) // Adjust based on context; see note below
}

func (ctx *VortexVerifierAction) RunGnark(api frontend.API, vr wizard.GnarkRuntime) {
	fmt.Printf("verifying VortexVerifierAction ...\n")
	// The skip verification flag may be on, if the current vortex
	// context get self-recursed. In this case, the verifier does
	// not need to
	if ctx.IsSelfrecursed {
		return
	}

	// In non-Merkle mode, this is left as empty
	roots := []frontend.Variable{}

	// Append the precomputed roots when IsCommitToPrecomputed is true
	if ctx.IsNonEmptyPrecomputed() {
		preRoots := [vortex_bls12377.GnarkKoalabearNumElements]zk.WrappedVariable{}
		// apiGen, _ := zk.NewGenericApi(api)

		for i := 0; i < vortex_bls12377.GnarkKoalabearNumElements; i++ {
			precompRootSv := vr.GetColumn(ctx.Items.Precomputeds.GnarkMerkleRoot[i].GetColID())
			preRoots[i] = precompRootSv[0]
		}
		fmt.Printf("Gnark precomputed Merkle root verifier: \n")
		api.Println(vortex_bls12377.Encode11WVsToFV(api, preRoots))

		roots = append(roots, vortex_bls12377.Encode11WVsToFV(api, preRoots))
	}

	// Collect all the commitments : rounds by rounds
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		if ctx.RoundStatus[round] == IsEmpty {
			continue // skip the dry rounds
		}
		//TODO@yao: check if this is correct, roots should be frontend.Variable, change all blockSize to vortex_bls12377.GnarkKoalabearNumElements
		preRoots := [vortex_bls12377.GnarkKoalabearNumElements]zk.WrappedVariable{}

		for i := 0; i < vortex_bls12377.GnarkKoalabearNumElements; i++ {
			rootSv := vr.GetColumn(ctx.MerkleRootName(round, i))
			preRoots[i] = rootSv[0]
		}
		// apiGen, _ := zk.NewGenericApi(api)
		// apiGen.Println(preRoots[:]...)
		roots = append(roots, vortex_bls12377.Encode11WVsToFV(api, preRoots))

	}

	randomCoin := vr.GetRandomCoinFieldExt(ctx.LinCombRandCoinName())

	// Collect the linear combination
	proof := vortex_bls12377.GProof{}
	proof.Rate = uint64(ctx.BlowUpFactor)
	proof.RsDomain = fft.NewDomain(uint64(ctx.NumEncodedCols()), fft.WithCache())
	proof.LinearCombination = vr.GetColumnExt(ctx.LinCombName()) //TODO@yao: this is correct, remove it
	// ext4, _ := gnarkfext.NewExt4(api)
	// ext4.Println(proof.LinearCombination...)

	// Collect the random entry List and the random coin
	entryList := vr.GetRandomCoinIntegerVec(ctx.RandColSelectionName())

	// Collect the opened columns and split them "by-commitment-rounds"
	proof.Columns = ctx.GnarkRecoverSelectedColumns(api, vr)
	x := vr.GetUnivariateParams(ctx.Query.QueryID).ExtX

	// function that will defer the hashing to gkr
	makePoseidon2Hasherfunc := func(_ frontend.API) (poseidon2_bls12377.GnarkMDHasher, error) {
		// factory := vr.GetHasherFactory()
		// if factory != nil {
		// 	h := vr.GetHasherFactory().NewHasher() //TODO@yao: here the hasher facroty should be able to create poseidon2_bls12377.GnarkMDHasher to be consistent with the hasher
		// 	return h, nil
		// }
		h, err := poseidon2_bls12377.NewGnarkMDHasher(api)
		return h, err
	}

	packedMProofs := [vortex_bls12377.GnarkKoalabearNumElements][]zk.WrappedVariable{}
	for i := range packedMProofs {
		packedMProofs[i] = vr.GetColumn(ctx.MerkleProofName(i))
	}

	proof.MerkleProofs = ctx.unpackMerkleProofsGnark(api, packedMProofs, entryList)

	// pass the parameters for a merkle-mode sis verification
	params := vortex_bls12377.GParams{}
	params.HasherFunc = makePoseidon2Hasherfunc
	params.NoSisHasher = makePoseidon2Hasherfunc
	params.Key = ctx.VortexParams.Key

	vortex_bls12377.GnarkVerifyOpeningWithMerkleProof(
		api,
		params,
		roots,
		proof,
		x,
		ctx.gnarkGetYs(api, vr),
		randomCoin,
		entryList,
	)

}

// returns the Ys as a vector
func (ctx *Ctx) gnarkGetYs(_ frontend.API, vr wizard.GnarkRuntime) (ys [][]gnarkfext.E4Gen) {

	query := ctx.Query
	params := vr.GetUnivariateParams(ctx.Query.QueryID)

	// Build an index table to efficiently lookup an alleged
	// prover evaluation from its colID.
	ysMap := make(map[ifaces.ColID]gnarkfext.E4Gen, len(params.ExtYs))
	for i := range query.Pols {
		ysMap[query.Pols[i].GetColID()] = params.ExtYs[i]
	}

	// Also add the shadow evaluations into ysMap. Since the shadow columns
	// are full-zeroes. We know that the evaluation will also always be zero.
	//
	// The sorting is necessary to ensure that the iteration below happens in
	// deterministic order over the [ShadowCols] map.
	shadowIDs := utils.SortedKeysOf(ctx.ShadowCols, func(a, b ifaces.ColID) bool {
		return a < b
	})

	for _, shadowID := range shadowIDs {
		ysMap[shadowID] = gnarkfext.E4Gen{}
	}

	ys = [][]gnarkfext.E4Gen{}

	// add ys for precomputed when IsCommitToPrecomputed is true
	if ctx.IsNonEmptyPrecomputed() {
		names := make([]ifaces.ColID, len(ctx.Items.Precomputeds.PrecomputedColums))
		for i, poly := range ctx.Items.Precomputeds.PrecomputedColums {
			names[i] = poly.GetColID()
		}
		ysPrecomputed := make([]gnarkfext.E4Gen, len(names))
		for i, name := range names {
			y, yFound := ysMap[name]
			if !yFound {
				utils.Panic("was not found: %v", name)
			}
			if y.B0.A0.IsEmpty() {
				utils.Panic("found Y but it was nil: %v", name)
			}
			ysPrecomputed[i] = y
		}
		ys = append(ys, ysPrecomputed)
	}

	// Get the list of the polynomials
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		if ctx.RoundStatus[round] == IsEmpty {
			continue // skip the dry rounds
		}
		names := ctx.CommitmentsByRounds.MustGet(round)
		ysRounds := make([]gnarkfext.E4Gen, len(names))
		for i, name := range names {
			y, yFound := ysMap[name]
			if !yFound {
				utils.Panic("was not found: %v", name)
			}

			if y.B0.A0.IsEmpty() {
				utils.Panic("found Y but it was nil: %v", name)
			}

			ysRounds[i] = ysMap[name]
		}

		ys = append(ys, ysRounds)
	}

	return ys
}

// Returns the opened columns from the messages. The returned columns are
// split "by-commitment-round".
func (ctx *Ctx) GnarkRecoverSelectedColumns(api frontend.API, vr wizard.GnarkRuntime) [][][]zk.WrappedVariable {

	// Collect the columns : first extract the full columns
	// Bear in mind that the prover messages are zero-padded
	fullSelectedCols := make([][]zk.WrappedVariable, ctx.NbColsToOpen())
	for j := 0; j < ctx.NbColsToOpen(); j++ {
		fullSelectedCols[j] = vr.GetColumn(ctx.SelectedColName(j))
	}

	// Split the columns per commitment for the verification
	openedSubColumns := [][][]zk.WrappedVariable{}
	roundStartAt := 0

	// Process precomputed
	if ctx.IsNonEmptyPrecomputed() {
		openedPrecompCols := make([][]zk.WrappedVariable, ctx.NbColsToOpen())
		numPrecomputeds := len(ctx.Items.Precomputeds.PrecomputedColums)
		for j := 0; j < ctx.NbColsToOpen(); j++ {
			openedPrecompCols[j] = fullSelectedCols[j][roundStartAt : roundStartAt+numPrecomputeds]
		}
		// update the start counter to ensure we do not pass twice the same row
		roundStartAt += numPrecomputeds
		openedSubColumns = append(openedSubColumns, openedPrecompCols)

	}

	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		if ctx.RoundStatus[round] == IsEmpty {
			continue // skip the dry rounds
		}
		openedSubColumnsForRound := make([][]zk.WrappedVariable, ctx.NbColsToOpen())
		numRowsForRound := ctx.getNbCommittedRows(round)
		for j := 0; j < ctx.NbColsToOpen(); j++ {
			openedSubColumnsForRound[j] = fullSelectedCols[j][roundStartAt : roundStartAt+numRowsForRound]
		}

		// update the start counter to ensure we do not pass twice the same row
		roundStartAt += numRowsForRound
		openedSubColumns = append(openedSubColumns, openedSubColumnsForRound)
	}

	// sanity-check : make sure we have not forgotten any column
	if roundStartAt != ctx.CommittedRowsCount {
		utils.Panic("we have a mistmatch in the row count : %v != %v", roundStartAt, ctx.CommittedRowsCount)
	}

	return openedSubColumns
}

// Evaluates explicitly the public polynomials (proof, vk, public inputs)
func (ctx *Ctx) gnarkExplicitPublicEvaluation(api frontend.API, vr wizard.GnarkRuntime) {

	var (
		params     = vr.GetUnivariateParams(ctx.Query.QueryID)
		polys      = make([][]zk.WrappedVariable, 0)
		expectedYs = make([]gnarkfext.E4Gen, 0)
	)

	for i, pol := range ctx.Query.Pols {

		// If the column is a VerifierDefined column, then it is directly
		// concerned by direct verification but we cannot access its status.
		// status so we need a hierarchical check to make sure we can access
		// its status.
		if _, isVerifierCol := pol.(verifiercol.VerifierCol); !isVerifierCol {
			status := ctx.Comp.Columns.Status(pol.GetColID())
			if !status.IsPublic() {
				// then, its not concerned by direct evaluation because the
				// evaluation is implicitly checked by the invokation of the
				// Vortex protocol.
				continue
			}
		}

		polys = append(polys, pol.GetColAssignmentGnark(vr))
		expectedYs = append(expectedYs, params.ExtYs[i])
	}

	ys := fastpoly.BatchEvaluateLagrangeGnarkMixed(api, polys, params.ExtX)

	ext4, _ := gnarkfext.NewExt4(api)
	for i := range expectedYs {
		ext4.AssertIsEqual(&ys[i], &expectedYs[i])
	}
}

// unpack a list of merkle proofs from a vector as in
func (ctx *Ctx) unpackMerkleProofsGnark(api frontend.API, sv [vortex_bls12377.GnarkKoalabearNumElements][]zk.WrappedVariable, entryList []zk.WrappedVariable) (proofs [][]smt_bls12377.GnarkProof) {

	depth := utils.Log2Ceil(ctx.NumEncodedCols()) // depth of the Merkle-tree
	numComs := ctx.NumCommittedRounds()
	if ctx.IsNonEmptyPrecomputed() {
		numComs += 1
	}

	numEntries := len(entryList)

	proofs = make([][]smt_bls12377.GnarkProof, numComs)
	curr := 0 // tracks the position in sv that we are parsing.
	for i := range proofs {
		proofs[i] = make([]smt_bls12377.GnarkProof, numEntries)
		for j := range proofs[i] {
			// initialize the proof that we are parsing
			proof := smt_bls12377.GnarkProof{
				Path:     entryList[j].AsNative(),
				Siblings: make([]frontend.Variable, depth),
			}

			// parse the siblings accounting for the fact that we
			// are inversing the order.
			for k := range proof.Siblings {

				var v [vortex_bls12377.GnarkKoalabearNumElements]zk.WrappedVariable
				for coord := 0; coord < vortex_bls12377.GnarkKoalabearNumElements; coord++ {
					v[coord] = sv[coord][curr]
				}
				// apiGen, _ := zk.NewGenericApi(api)
				// apiGen.Println(v[:]...)
				proof.Siblings[depth-k-1] = vortex_bls12377.Encode11WVsToFV(api, v)
				curr++
			}

			proofs[i][j] = proof
		}
	}
	return proofs
}
