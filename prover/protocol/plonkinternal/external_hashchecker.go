package plonkinternal

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// addHashConstraint adds the constraints the hashes of the tagged witness.
// It checks that the assignment of the hash column is consistent with the
// LRO columns using a lookup and then it uses a Poseidon2 query to enforce the
// hash.
func (ctx *CompilationCtx) addHashConstraint() {

	// Get the hash claims once and cache them (hashedGetter reads from a channel)
	hashClaims := ctx.Plonk.hashedGetter()

	// Return if there are no hash claims
	if hashClaims == nil || len(hashClaims) == 0 {
		return
	}

	logrus.Infof("[external-hash-checker] starting addHashConstraint for %v instances with %v hash claims", ctx.MaxNbInstances, len(hashClaims))

	var (
		numRowLRO = ctx.DomainSize()
		round     = ctx.Columns.L[0].Round()
		eho       = &ctx.ExternalHasherOption

		// Records the positions of the hash claims in the Plonk rows.
		_ = "breakpoint" // for debugging
	)

	logrus.Infof("[external-hash-checker] calling getHashCheckedPositionSV...")
	posOsSv, posBlSv, posNsSv := ctx.getHashCheckedPositionSV(hashClaims)
	chunkSize := posOsSv[0].Len()
	logrus.Infof("[external-hash-checker] chunkSize=%v, numRowLRO=%v", chunkSize, numRowLRO)

	var (

		// Declare the L, R, O position columns. These will be cached and
		// reused by the fixed permutation compiler.
		posL = dedicated.CounterPrecomputed(ctx.comp, 0, numRowLRO)
		posR = dedicated.CounterPrecomputed(ctx.comp, numRowLRO, 2*numRowLRO)
		posO = dedicated.CounterPrecomputed(ctx.comp, 2*numRowLRO, 3*numRowLRO)
	)

	for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
		if j == 0 {
			logrus.Infof("[external-hash-checker] inserting %v position precomputed columns...", poseidon2_koalabear.BlockSize)
		}
		eho.PosOldState[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionOS_%v", j), posOsSv[j])
		eho.PosBlock[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionBL_%v", j), posBlSv[j])
		eho.PosNewState[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionNS_%v", j), posNsSv[j])
	}
	eho.OldStates = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)
	eho.Blocks = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)
	eho.NewStates = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)

	logrus.Infof("[external-hash-checker] starting loop for %v instances, this will create 3*8*%v=%v lookup queries", ctx.MaxNbInstances, ctx.MaxNbInstances, 3*8*ctx.MaxNbInstances)

	for i := 0; i < ctx.MaxNbInstances; i++ {
		if i%10 == 0 || i < 5 {
			logrus.Infof("[external-hash-checker] processing instance %v/%v", i, ctx.MaxNbInstances)
		}

		var (
			selector = verifiercol.NewRepeatedAccessor(
				accessors.NewFromPublicColumn(ctx.Columns.Activators[i], 0),
				chunkSize,
			)

			lookupTables = [][]ifaces.Column{
				{
					posL, ctx.Columns.L[i],
				},
				{
					posR, ctx.Columns.R[i],
				},
				{
					posO, ctx.Columns.O[i],
				},
			}
		)

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			eho.OldStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckOldState_%v_%v", i, j), chunkSize, true)
			eho.Blocks[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckBlock_%v_%v", i, j), chunkSize, true)
			eho.NewStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckNewState_%v_%v", i, j), chunkSize, true)

			// Those are lookups checking that the LRO columns are consistent with
			// the hash claims.

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportOldState_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosOldState[j], eho.OldStates[i][j],
				},
				// no filters
				nil, nil,
			)

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportBlock_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosBlock[j], eho.Blocks[i][j],
				},
				// no filters
				nil, nil,
			)

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportNewState_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosNewState[j], eho.NewStates[i][j],
				},
				// no filters
				nil, nil,
			)
		}

		ctx.comp.InsertPoseidon2(
			round,
			ctx.queryIDf("HashCheckPoseidon_%v", i),
			eho.Blocks[i],
			eho.OldStates[i],
			eho.NewStates[i],
			selector,
		)
	}

	logrus.Infof("[external-hash-checker] completed addHashConstraint, created %v total lookup queries", 3*8*ctx.MaxNbInstances)
}

// assignHashColumns assigns the hash c olumns.
func (ctx *GenericPlonkProverAction) assignHashColumns(run *wizard.ProverRuntime) {

	logrus.Infof("[external-hash-checker-prover] starting assignHashColumns for %v instances", ctx.MaxNbInstances)

	var (
		eho   = &ctx.ExternalHasherOption
		posOs [poseidon2_koalabear.BlockSize][]field.Element
		posBl [poseidon2_koalabear.BlockSize][]field.Element
		posNs [poseidon2_koalabear.BlockSize][]field.Element
	)

	for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
		posOs[j] = eho.PosOldState[j].GetColAssignment(run).IntoRegVecSaveAlloc()
		posBl[j] = eho.PosBlock[j].GetColAssignment(run).IntoRegVecSaveAlloc()
		posNs[j] = eho.PosNewState[j].GetColAssignment(run).IntoRegVecSaveAlloc()
	}
	chunkSize := len(posOs[0])
	sizeHashing := chunkSize * poseidon2_koalabear.BlockSize
	logrus.Infof("[external-hash-checker-prover] chunkSize=%v, sizeHashing=%v", chunkSize, sizeHashing)

	for i := 0; i < ctx.MaxNbInstances; i++ {
		if i%10 == 0 || i < 5 {
			logrus.Infof("[external-hash-checker-prover] assigning instance %v/%v", i, ctx.MaxNbInstances)
		}

		var (
			src = []smartvectors.SmartVector{
				ctx.Columns.L[i].GetColAssignment(run),
				ctx.Columns.R[i].GetColAssignment(run),
				ctx.Columns.O[i].GetColAssignment(run),
			}
			sizeLRO  = ctx.Columns.L[i].Size()
			oldState = make([]field.Element, sizeHashing)
			block    = make([]field.Element, sizeHashing)
			newState = make([]field.Element, sizeHashing)
		)

		for j := range oldState {

			var (
				osID = int(posOs[j%poseidon2_koalabear.BlockSize][j/poseidon2_koalabear.BlockSize].Uint64())
				blID = int(posBl[j%poseidon2_koalabear.BlockSize][j/poseidon2_koalabear.BlockSize].Uint64())
				nsID = int(posNs[j%poseidon2_koalabear.BlockSize][j/poseidon2_koalabear.BlockSize].Uint64())
				os   = src[osID/sizeLRO].Get(osID % sizeLRO)
				bl   = src[blID/sizeLRO].Get(blID % sizeLRO)
				ns   = src[nsID/sizeLRO].Get(nsID % sizeLRO)
			)
			oldState[j] = os
			block[j] = bl
			newState[j] = ns
		}

		var vecOldState, vecBlock, vecNewState [poseidon2_koalabear.BlockSize][]field.Element

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			vecOldState[j] = make([]field.Element, chunkSize)
			vecBlock[j] = make([]field.Element, chunkSize)
			vecNewState[j] = make([]field.Element, chunkSize)
		}

		for k := range oldState {
			vecOldState[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = oldState[k]
			vecBlock[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = block[k]
			vecNewState[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = newState[k]
		}

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			run.AssignColumn(eho.OldStates[i][j].GetColID(), smartvectors.NewRegular(vecOldState[j]))
			run.AssignColumn(eho.Blocks[i][j].GetColID(), smartvectors.NewRegular(vecBlock[j]))
			run.AssignColumn(eho.NewStates[i][j].GetColID(), smartvectors.NewRegular(vecNewState[j]))
		}

	}

	logrus.Infof("[external-hash-checker-prover] completed assignHashColumns")
}

// getHashCheckedPositionSV returns the smartvectors containing the position
// of the hash claims in the LRO columns.
func (ctx *CompilationCtx) getHashCheckedPositionSV(sls [][3][2]int) (posOS, posBl, posNS [poseidon2_koalabear.BlockSize]smartvectors.SmartVector) {

	var (
		size        = utils.NextPowerOfTwo(len(sls))
		numRowPlonk = ctx.DomainSize()
	)
	logrus.Infof("[external-hash-checker] getHashCheckedPositionSV: found %v hash claims, nextPowerOfTwo=%v, numRowPlonk=%v", len(sls), size, numRowPlonk)
	if len(sls) == 0 {
		panic("no hash claims found")
	}

	if ctx.ExternalHasherOption.FixedNbRows > 0 {
		fixedNbRow := ctx.ExternalHasherOption.FixedNbRows
		if fixedNbRow < size {
			utils.Panic("the fixed number of rows %v is smaller than the number of hash claims %v", fixedNbRow, len(sls))
		}
		logrus.Infof("[external-hash-checker] using FixedNbRows=%v (was calculated as %v)", fixedNbRow, size)
		size = ctx.ExternalHasherOption.FixedNbRows
	}
	chunkSize := size / poseidon2_koalabear.BlockSize
	logrus.Infof("[external-hash-checker] final size=%v, chunkSize=%v, will pad from %v claims", size, chunkSize, len(sls))

	var (
		ost = make([]field.Element, size)
		blk = make([]field.Element, size)
		nst = make([]field.Element, size)
	)

	for i, ss := range sls {
		ost[i].SetUint64(uint64(ss[0][0] + ss[0][1]*numRowPlonk))
		blk[i].SetUint64(uint64(ss[1][0] + ss[1][1]*numRowPlonk))
		nst[i].SetUint64(uint64(ss[2][0] + ss[2][1]*numRowPlonk))
	}
	for i := len(sls); i < size; i++ {
		ost[i] = ost[i-1]
		blk[i] = blk[i-1]
		nst[i] = nst[i-1]
	}

	var ostOct, blkOct, nstOct [poseidon2_koalabear.BlockSize][]field.Element

	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		ostOct[i] = make([]field.Element, chunkSize)
		blkOct[i] = make([]field.Element, chunkSize)
		nstOct[i] = make([]field.Element, chunkSize)
	}

	for i := 0; i < size; i++ {
		ostOct[i%poseidon2_koalabear.BlockSize][i/poseidon2_koalabear.BlockSize] = ost[i]
		blkOct[i%poseidon2_koalabear.BlockSize][i/poseidon2_koalabear.BlockSize] = blk[i]
		nstOct[i%poseidon2_koalabear.BlockSize][i/poseidon2_koalabear.BlockSize] = nst[i]
	}

	for i := 0; i < poseidon2_koalabear.BlockSize; i++ {
		posOS[i] = smartvectors.NewRegular(ostOct[i])
		posBl[i] = smartvectors.NewRegular(blkOct[i])
		posNS[i] = smartvectors.NewRegular(nstOct[i])
	}

	logrus.Infof("[external-hash-checker] getHashCheckedPositionSV completed")
	return posOS, posBl, posNS
}
