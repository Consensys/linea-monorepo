package poseidon2

import (
	"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// numGKRLayers is the total number of SBox layers in Poseidon2 (= fullRounds - 1 because the
// initial matMulExternal has no SBox).
// Rounds 1-3 (external), 4-24 (internal), 25-27 (external) = 27 layers total.
const numGKRLayers = fullRounds - 1 // = 27

// isGKRExternalLayer returns true if GKR layer l uses the full (width-16) SBox.
// Layers 0,1,2 = rounds 1,2,3 (external); layers 3..23 = rounds 4..24 (internal);
// layers 24,25,26 = rounds 25,26,27 (external).
func isGKRExternalLayer(l int) bool {
	return l < partialRounds || l >= numGKRLayers-partialRounds
}

// gkrTranscriptPerLayer returns the number of field elements per layer in the GKR transcript.
// Each layer stores: logN*5 (sumcheck polys) + 16 (oracle evals).
func gkrTranscriptPerLayer(logN int) int {
	return logN*5 + width
}

// gkrTranscriptSize returns the unpadded transcript size for a given logN.
func gkrTranscriptSize(logN int) int {
	return numGKRLayers * gkrTranscriptPerLayer(logN)
}

// GKRPoseidon2Context stores compilation artefacts for the GKR-based Poseidon2 compiler.
type GKRPoseidon2Context struct {
	// Same stacked I/O columns as Poseidon2Context
	CompiledQueries  []*query.Poseidon2
	StackedOldStates [blockSize]ifaces.Column
	StackedBlocks    [blockSize]ifaces.Column
	StackedNewStates [blockSize]ifaces.Column

	// EqR0: committed column of size TotalSize, stores eq_{r0}[i].
	EqR0 ifaces.Column
	// EqRFinal: committed column of size TotalSize, stores eq_{r_final}[i].
	EqRFinal ifaces.Column
	// GKRTranscript: proof column (directly visible to verifier), size TranscriptPaddedSize.
	// Contains the flat GKR sumcheck transcript for all 27 SBox layers.
	GKRTranscript ifaces.Column

	// InputIPID: name of InnerProduct(EqR0, [os_0..os_7, block_0..block_7]).
	InputIPID ifaces.QueryID
	// OutputIPID: name of InnerProduct(EqRFinal, [ns_0..ns_7, block_0..block_7]).
	OutputIPID ifaces.QueryID

	// R0CoinNames: names of the FieldExt coins for r0.
	R0CoinNames []coin.Name

	// LogN: log2(TotalSize).
	LogN int
	// TotalSize: size of the stacked columns (power of 2).
	TotalSize int
	// TranscriptPaddedSize: padded size of the GKR transcript column.
	TranscriptPaddedSize int

	// GKRRound: the wizard protocol round at which GKR columns/queries are registered.
	GKRRound int

	// stackedIOCache caches the stacked I/O arrays computed at round p for use at round p+1.
	stackedIOCache *gkrStackedIO
}

// gkrStackedIO caches stacked I/O assignments between the two prover actions.
type gkrStackedIO struct {
	oldStates [blockSize][]field.Element
	blocks    [blockSize][]field.Element
	newStates [blockSize][]field.Element
}

// gkrIOProverAction is the round-p ProverAction that stacks the I/O columns.
type gkrIOProverAction struct{ ctx *GKRPoseidon2Context }

// CompileGKRPoseidon2 compiles all Poseidon2 queries in comp using the GKR approach.
// This reduces committed column cells from ~230×N to ~28×N compared to CompilePoseidon2.
func CompileGKRPoseidon2(comp *wizard.CompiledIOP) {
	_ = defineGKRContext(comp)
}

// defineGKRContext constructs the GKR-based Poseidon2 compilation context.
func defineGKRContext(comp *wizard.CompiledIOP) *GKRPoseidon2Context {
	var (
		ctx             = &GKRPoseidon2Context{}
		protocolRoundID = 0
		allQueries      = comp.QueriesNoParams.AllUnignoredKeys()
		totalSize       = 0
	)

	for _, qName := range allQueries {
		if q_, ok := comp.QueriesNoParams.Data(qName).(query.Poseidon2); ok {
			comp.QueriesNoParams.MarkAsIgnored(qName)
			protocolRoundID = max(protocolRoundID, comp.QueriesNoParams.Round(qName))
			totalSize += q_.Blocks[0].Size()
			ctx.CompiledQueries = append(ctx.CompiledQueries, &q_)
		}
	}

	if len(ctx.CompiledQueries) == 0 {
		return nil
	}

	totalSize = utils.NextPowerOfTwo(totalSize)
	logN := utils.Log2Ceil(totalSize)
	transcriptSize := gkrTranscriptSize(logN)
	transcriptPaddedSize := utils.NextPowerOfTwo(transcriptSize)

	ctx.TotalSize = totalSize
	ctx.LogN = logN
	ctx.TranscriptPaddedSize = transcriptPaddedSize

	// Round p: define stacked I/O columns (same as CompilePoseidon2)
	for block := 0; block < blockSize; block++ {
		ctx.StackedOldStates[block] = comp.InsertCommit(protocolRoundID,
			ifaces.ColIDf("GKR_Poseidon2_OLD_STATES_%v_%v_%v", comp.SelfRecursionCount, uniqueID(comp), block),
			totalSize, true)
		ctx.StackedBlocks[block] = comp.InsertCommit(protocolRoundID,
			ifaces.ColIDf("GKR_Poseidon2_BLOCKS_%v_%v_%v", comp.SelfRecursionCount, uniqueID(comp), block),
			totalSize, true)
		ctx.StackedNewStates[block] = comp.InsertCommit(protocolRoundID,
			ifaces.ColIDf("GKR_Poseidon2_NEW_STATES_%v_%v_%v", comp.SelfRecursionCount, uniqueID(comp), block),
			totalSize, true)
	}

	pragmas.AddModuleRef(ctx.StackedBlocks[0], "GKR_POSEIDON2_COMPILER")

	// Register inclusion queries linking each Poseidon2 query to the stacked columns
	for i := range ctx.CompiledQueries {
		stackedCols := make([]ifaces.Column, 0, blockSize*3)
		queryCols := make([]ifaces.Column, 0, blockSize*3)
		for block := 0; block < blockSize; block++ {
			stackedCols = append(stackedCols,
				ctx.StackedBlocks[block],
				ctx.StackedOldStates[block],
				ctx.StackedNewStates[block],
			)
			queryCols = append(queryCols,
				ctx.CompiledQueries[i].Blocks[block],
				ctx.CompiledQueries[i].OldState[block],
				ctx.CompiledQueries[i].NewState[block],
			)
		}
		comp.GenericFragmentedConditionalInclusion(
			protocolRoundID,
			ifaces.QueryIDf("GKR_Poseidon2_QUERY_%v_INCLUSION_%v_%v", i, comp.SelfRecursionCount, uniqueID(comp)),
			[][]ifaces.Column{stackedCols},
			queryCols,
			nil,
			ctx.CompiledQueries[i].Selector,
		)
	}

	// Round p+1: register r0 coins, GKR columns, and InnerProduct queries.
	gkrRound := protocolRoundID + 1
	ctx.GKRRound = gkrRound

	// Register r0 coins: ceil(logN/4) FieldExt coins (each gives 4 base field elements).
	// Always register at least 1 coin to ensure gkrRound is counted in comp.NumRounds().
	numR0Coins := max(1, (logN+3)/4)
	ctx.R0CoinNames = make([]coin.Name, numR0Coins)
	for k := 0; k < numR0Coins; k++ {
		name := coin.Name(fmt.Sprintf("GKR_Poseidon2_R0_%v_%v_%v", comp.SelfRecursionCount, uniqueID(comp), k))
		comp.InsertCoin(gkrRound, name, coin.FieldExt)
		ctx.R0CoinNames[k] = name
	}

	// EqR0: committed column (will be assigned by prover at round gkrRound)
	ctx.EqR0 = comp.InsertCommit(gkrRound,
		ifaces.ColIDf("GKR_Poseidon2_EQR0_%v_%v", comp.SelfRecursionCount, uniqueID(comp)),
		totalSize, true)

	// EqRFinal: committed column (assigned by prover at round gkrRound)
	ctx.EqRFinal = comp.InsertCommit(gkrRound,
		ifaces.ColIDf("GKR_Poseidon2_EQRFINAL_%v_%v", comp.SelfRecursionCount, uniqueID(comp)),
		totalSize, true)

	// GKRTranscript: proof column (directly visible to verifier, part of proof)
	ctx.GKRTranscript = comp.InsertColumn(gkrRound,
		ifaces.ColIDf("GKR_Poseidon2_TRANSCRIPT_%v_%v", comp.SelfRecursionCount, uniqueID(comp)),
		transcriptPaddedSize, column.Proof, true)

	// InputInnerProduct: <EqR0, [os_0..os_7, block_0..block_7]> → 16 claims
	inputCols := make([]ifaces.Column, width)
	copy(inputCols[:blockSize], ctx.StackedOldStates[:])
	copy(inputCols[blockSize:], ctx.StackedBlocks[:])
	inputIPID := ifaces.QueryIDf("GKR_Poseidon2_INPUT_IP_%v_%v", comp.SelfRecursionCount, uniqueID(comp))
	comp.InsertInnerProduct(gkrRound, inputIPID, ctx.EqR0, inputCols)
	ctx.InputIPID = inputIPID

	// OutputInnerProduct: <EqRFinal, [ns_0..ns_7, block_0..block_7]> → 16 claims
	outputCols := make([]ifaces.Column, width)
	copy(outputCols[:blockSize], ctx.StackedNewStates[:])
	copy(outputCols[blockSize:], ctx.StackedBlocks[:])
	outputIPID := ifaces.QueryIDf("GKR_Poseidon2_OUTPUT_IP_%v_%v", comp.SelfRecursionCount, uniqueID(comp))
	comp.InsertInnerProduct(gkrRound, outputIPID, ctx.EqRFinal, outputCols)
	ctx.OutputIPID = outputIPID

	// Round p: assign stacked I/O columns (must be in the same round as those columns).
	comp.RegisterProverAction(protocolRoundID, &gkrIOProverAction{ctx: ctx})
	// Round p+1: compute GKR proof, assign eq/transcript/IP columns.
	comp.RegisterProverAction(gkrRound, ctx)
	comp.RegisterVerifierAction(gkrRound, &gkrPoseidon2VerifierAction{ctx: ctx})

	return ctx
}

// ---------- ProverAction (round p) ----------

// Run implements wizard.ProverAction for gkrIOProverAction.
// It stacks the I/O columns at round p (the same round they are committed in).
func (a *gkrIOProverAction) Run(run *wizard.ProverRuntime) {
	ctx := a.ctx
	var (
		zeroBlock       field.Octuplet
		totalSize       = ctx.TotalSize
		poseidon2OfZero = vortex.CompressPoseidon2(zeroBlock, zeroBlock)
	)

	var (
		stackedOldStates [blockSize][]field.Element
		stackedBlocks    [blockSize][]field.Element
		stackedNewStates [blockSize][]field.Element
	)

	for i := range ctx.CompiledQueries {
		var (
			q         = ctx.CompiledQueries[i]
			sel       smartvectors.SmartVector
			os, b, ns [8]smartvectors.SmartVector
		)
		if q.Selector != nil {
			sel = q.Selector.GetColAssignment(run)
		}
		for block := 0; block < blockSize; block++ {
			os[block] = q.OldState[block].GetColAssignment(run)
			b[block] = q.Blocks[block].GetColAssignment(run)
			ns[block] = q.NewState[block].GetColAssignment(run)
		}
		allState := slices.Concat(os[:], b[:], ns[:])
		allState = append(allState, sel)
		start, stop := smartvectors.CoWindowRange(allState...)

		for block := 0; block < blockSize; block++ {
			pushToStacked := func(j int) {
				if sel != nil && sel.GetPtr(j).IsZero() {
					return
				}
				stackedOldStates[block] = append(stackedOldStates[block], os[block].Get(j))
				stackedBlocks[block] = append(stackedBlocks[block], b[block].Get(j))
				stackedNewStates[block] = append(stackedNewStates[block], ns[block].Get(j))
			}
			for j := start; j < stop; j++ {
				pushToStacked(j)
			}
			if start > 0 {
				pushToStacked(0)
				continue
			}
			if stop < os[block].Len() {
				pushToStacked(os[block].Len() - 1)
				continue
			}
		}
	}

	for block := 0; block < blockSize; block++ {
		if len(stackedOldStates[block]) < totalSize {
			stackedOldStates[block] = append(stackedOldStates[block], field.NewElement(0))
			stackedBlocks[block] = append(stackedBlocks[block], field.NewElement(0))
			stackedNewStates[block] = append(stackedNewStates[block], poseidon2OfZero[block])
		}
		run.AssignColumn(ctx.StackedOldStates[block].GetColID(),
			smartvectors.RightZeroPadded(stackedOldStates[block], totalSize))
		run.AssignColumn(ctx.StackedBlocks[block].GetColID(),
			smartvectors.RightZeroPadded(stackedBlocks[block], totalSize))
		run.AssignColumn(ctx.StackedNewStates[block].GetColID(),
			smartvectors.RightPadded(stackedNewStates[block], poseidon2OfZero[block], totalSize))
	}

	// Cache the raw stacked arrays for use by the round-(p+1) GKR ProverAction.
	ctx.stackedIOCache = &gkrStackedIO{
		oldStates: stackedOldStates,
		blocks:    stackedBlocks,
		newStates: stackedNewStates,
	}
}

// ---------- ProverAction (round p+1) ----------

// Run implements wizard.ProverAction for GKRPoseidon2Context.
// It runs at round gkrRound = p+1, after r0 coins have been sampled.
func (ctx *GKRPoseidon2Context) Run(run *wizard.ProverRuntime) {
	var (
		zeroBlock       field.Octuplet
		totalSize       = ctx.TotalSize
		poseidon2OfZero = vortex.CompressPoseidon2(zeroBlock, zeroBlock)
	)

	// Retrieve stacked I/O arrays from the round-p cache.
	cache := ctx.stackedIOCache
	stackedOldStates := cache.oldStates
	stackedBlocks := cache.blocks
	stackedNewStates := cache.newStates

	effectiveSize := len(stackedOldStates[0])

	// --- Step 2: compute all post-SBox states for GKR ---
	// postSBoxStates[l][w][i] = state element w for Poseidon2 instance i, AFTER SBox at layer l.
	// For external layers all 16 cols are cubed; for internal only col 0 is cubed.
	postSBoxStates := gkrComputePostSBoxStates(stackedOldStates, stackedBlocks, effectiveSize)

	// --- Step 3: derive r0 from wizard coins ---
	r0 := gkrExtractR0(run, ctx.R0CoinNames, ctx.LogN)

	// --- Step 4: compute GKR transcript ---
	transcript, rFinal := gkrProveAll(postSBoxStates, r0, ctx.LogN, totalSize)

	// --- Step 5: assign EqR0 and EqRFinal ---
	// computeEq uses big-endian bit ordering: evalMLE(computeEq(r, n), f) = MLE_f(reverse(r)).
	// To evaluate MLE_f at r0 (resp. rFinal), we must pass reverse(r0) (resp. reverse(rFinal)).
	eqR0 := computeEq(reverseSlice(r0), totalSize)
	eqRFinal := computeEq(reverseSlice(rFinal), totalSize)

	run.AssignColumn(ctx.EqR0.GetColID(),
		smartvectors.NewRegular(eqR0))
	run.AssignColumn(ctx.EqRFinal.GetColID(),
		smartvectors.NewRegular(eqRFinal))

	// --- Step 6: assign GKRTranscript column (flat, padded to TranscriptPaddedSize) ---
	run.AssignColumn(ctx.GKRTranscript.GetColID(),
		smartvectors.RightZeroPadded(transcript, ctx.TranscriptPaddedSize))

	// --- Step 7: assign InnerProduct params ---
	// Pad to totalSize before evalMLE (raw cache arrays may be shorter).
	inputIPYs := make([]field.Element, width)
	for w := 0; w < blockSize; w++ {
		padded := make([]field.Element, totalSize)
		copy(padded, stackedOldStates[w])
		inputIPYs[w] = evalMLE(eqR0, padded)
	}
	for w := 0; w < blockSize; w++ {
		padded := make([]field.Element, totalSize)
		copy(padded, stackedBlocks[w])
		inputIPYs[blockSize+w] = evalMLE(eqR0, padded)
	}
	run.AssignInnerProduct(ctx.InputIPID, gkrFieldToFext(inputIPYs)...)

	outputIPYs := make([]field.Element, width)
	for w := 0; w < blockSize; w++ {
		col := make([]field.Element, totalSize)
		for i := range col {
			if i < len(stackedNewStates[w]) {
				col[i] = stackedNewStates[w][i]
			} else {
				col[i] = poseidon2OfZero[w]
			}
		}
		outputIPYs[w] = evalMLE(eqRFinal, col)
	}
	for w := 0; w < blockSize; w++ {
		blkCol := make([]field.Element, totalSize)
		for i := range blkCol {
			if i < len(stackedBlocks[w]) {
				blkCol[i] = stackedBlocks[w][i]
			} else {
				blkCol[i] = field.NewElement(0)
			}
		}
		outputIPYs[blockSize+w] = evalMLE(eqRFinal, blkCol)
	}
	run.AssignInnerProduct(ctx.OutputIPID, gkrFieldToFext(outputIPYs)...)
}

// ---------- GKR prover helpers ----------

// gkrExtractR0 extracts logN base field scalars from the FieldExt coins.
func gkrExtractR0(run *wizard.ProverRuntime, coinNames []coin.Name, logN int) []field.Element {
	r0 := make([]field.Element, logN)
	k := 0
	for _, name := range coinNames {
		fext := run.GetRandomCoinFieldExt(name)
		comps := [4]field.Element{fext.B0.A0, fext.B0.A1, fext.B1.A0, fext.B1.A1}
		for _, c := range comps {
			if k >= logN {
				break
			}
			r0[k] = c
			k++
		}
	}
	return r0
}

// gkrComputePostSBoxStates computes the post-SBox state for all 27 layers and all instances.
// For external layers: result[l][w][i] = preSBox[l][w][i]^3  (full SBox applied to all 16 cols).
// For internal layers: result[l][0][i] = preSBox[l][0][i]^3, result[l][w][i] = preSBox[l][w][i] (w>0).
// These post-SBox values are what the GKR bookkeeping tables track.
// Storing post-SBox (rather than pre-SBox) enables a linear oracle check in the verifier:
//   MLE(postSBox[l])(r) → matMul → MLE(preSBox[l+1])(r)   (valid because matMul is linear).
func gkrComputePostSBoxStates(
	oldStates, blocks [blockSize][]field.Element,
	effectiveSize int,
) [numGKRLayers][width][]field.Element {

	var result [numGKRLayers][width][]field.Element
	for l := 0; l < numGKRLayers; l++ {
		for w := 0; w < width; w++ {
			result[l][w] = make([]field.Element, effectiveSize)
		}
	}

	parallel.Execute(effectiveSize, func(start, stop int) {
		for i := start; i < stop; i++ {
			var state [width]field.Element
			for w := 0; w < blockSize; w++ {
				state[w] = oldStates[w][i]
				state[w+blockSize] = blocks[w][i]
			}

			matMulExternalInPlace(&state)

			// Rounds 1-3 (external SBox): save post-SBox state
			for round := 1; round <= partialRounds; round++ {
				addRoundKeyCompute(round-1, &state)
				l := round - 1
				for w := 0; w < width; w++ {
					result[l][w][i] = sBoxCompute(w, round, state[:])
				}
				// advance state through sBox + matMul
				for w := 0; w < width; w++ {
					state[w] = result[l][w][i]
				}
				matMulExternalInPlace(&state)
			}

			// Rounds 4-24 (internal SBox)
			for round := partialRounds + 1; round <= fullRounds-partialRounds-1; round++ {
				addRoundKeyCompute(round-1, &state)
				l := round - 1
				for w := 0; w < width; w++ {
					result[l][w][i] = sBoxCompute(w, round, state[:])
				}
				for w := 0; w < width; w++ {
					state[w] = result[l][w][i]
				}
				matMulInternalInPlace(&state)
			}

			// Rounds 25-27 (external SBox)
			for round := fullRounds - partialRounds; round < fullRounds; round++ {
				addRoundKeyCompute(round-1, &state)
				l := round - 1
				for w := 0; w < width; w++ {
					result[l][w][i] = sBoxCompute(w, round, state[:])
				}
				for w := 0; w < width; w++ {
					state[w] = result[l][w][i]
				}
				matMulExternalInPlace(&state)
			}
		}
	})

	return result
}

// gkrProveAll runs the GKR prover for all 27 SBox layers and returns the flat transcript
// and the final evaluation point r_final.
func gkrProveAll(
	postSBoxStates [numGKRLayers][width][]field.Element,
	r0 []field.Element,
	logN int,
	totalSize int,
) (transcript []field.Element, rFinal []field.Element) {

	fs := fiatshamir.NewFSKoalabear()
	fs.Update(r0...)

	perLayer := gkrTranscriptPerLayer(logN)
	transcript = make([]field.Element, 0, numGKRLayers*perLayer)

	// Initialize bookkeeping tables
	bookEq := computeEq(r0, totalSize)
	var bookState [width][]field.Element
	for w := 0; w < width; w++ {
		bookState[w] = make([]field.Element, totalSize)
	}

	// Precompute the canonical pre-SBox states for the all-zeros input instance.
	// These are used to pad bookState at positions beyond effectiveSize, ensuring
	// consistency with the newStates column which pads with poseidon2OfZero.
	var zeroOS, zeroBlk [blockSize][]field.Element
	for w := 0; w < blockSize; w++ {
		zeroOS[w] = make([]field.Element, 1) // single zero element
		zeroBlk[w] = make([]field.Element, 1)
	}
	zeroPadPreSBox := gkrComputePostSBoxStates(zeroOS, zeroBlk, 1)

	currentR := r0

	for l := 0; l < numGKRLayers; l++ {
		isExt := isGKRExternalLayer(l)
		n := len(bookEq)

		// Copy post-SBox state for layer l into bookState.
		// Pad positions beyond effectiveSize with the canonical zero-instance postSBox
		// state, matching the poseidon2OfZero padding in newStates.
		effectiveSize := len(postSBoxStates[l][0])
		for w := 0; w < width; w++ {
			copy(bookState[w][:effectiveSize], postSBoxStates[l][w])
			for i := effectiveSize; i < n; i++ {
				bookState[w][i] = zeroPadPreSBox[l][w][0]
			}
		}

		// Re-initialize bookEq from currentR (needed because each layer starts fresh)
		freshEq := computeEq(currentR, totalSize)
		copy(bookEq, freshEq)

		// Derive alpha for external rounds
		var alpha field.Element
		if isExt {
			alpha = gkrDeriveAlpha(fs)
		}


		// Run sumcheck prover
		polys, newPoint := gkrSumcheckProveLayer(bookEq, &bookState, isExt, alpha, logN, fs)

		// Extract oracle evals (bookState[w][0] after folding)
		var oracleEvals [width]field.Element
		for w := 0; w < width; w++ {
			oracleEvals[w] = bookState[w][0]
		}

		// Append to transcript: polys (logN*5) + oracle evals (16)
		for k := 0; k < logN; k++ {
			transcript = append(transcript, polys[k][:]...)
		}
		transcript = append(transcript, oracleEvals[:]...)

		// Absorb oracle evals into FS (already done inside gkrSumcheckProveLayer via Update)
		// Actually we need to update FS with oracle evals here
		fs.Update(oracleEvals[:]...)

		// Update currentR to newPoint for next layer
		currentR = newPoint
	}

	rFinal = currentR
	return transcript, rFinal
}

// ---------- VerifierAction ----------

type gkrPoseidon2VerifierAction struct {
	ctx *GKRPoseidon2Context
}

// Run implements wizard.VerifierAction.Run (plain verifier).
func (va *gkrPoseidon2VerifierAction) Run(run wizard.Runtime) error {
	ctx := va.ctx
	logN := ctx.LogN

	// Extract r0 from coins
	r0 := gkrExtractR0FromRuntime(run, ctx.R0CoinNames, logN)

	// Get input InnerProduct params
	inputParams := run.GetParams(ctx.InputIPID).(query.InnerProductParams)
	outputParams := run.GetParams(ctx.OutputIPID).(query.InnerProductParams)

	// Read GKR transcript from the Proof column
	transcriptSV := ctx.GKRTranscript.GetColAssignment(run)
	transcriptSize := gkrTranscriptSize(logN)
	transcript := make([]field.Element, transcriptSize)
	for i := 0; i < transcriptSize; i++ {
		transcript[i] = transcriptSV.Get(i)
	}

	// Initialize GKR claims: start from <eqR0, [os, block]>, then apply the
	// initial matMulExt and addRoundKey[0] to get <eqR0, preSBoxState[0]>.
	var claims [width]field.Element
	for w := 0; w < width; w++ {
		claims[w] = inputParams.Ys[w].B0.A0
	}
	applyMatMulExternalClaims(&claims)
	applyAddRoundKeyClaims(0, &claims)

	// Initialize internal FS
	fs := fiatshamir.NewFSKoalabear()
	fs.Update(r0...)

	currentR := r0
	perLayer := gkrTranscriptPerLayer(logN)

	for l := 0; l < numGKRLayers; l++ {
		isExt := isGKRExternalLayer(l)
		offset := l * perLayer

		// Extract transcript for this layer
		var polys [][5]field.Element
		polys = make([][5]field.Element, logN)
		for k := 0; k < logN; k++ {
			for t := 0; t < 5; t++ {
				polys[k][t] = transcript[offset+k*5+t]
			}
		}
		var oracleEvals [width]field.Element
		oracleBase := offset + logN*5
		for w := 0; w < width; w++ {
			oracleEvals[w] = transcript[oracleBase+w]
		}

		// Derive alpha for external rounds
		var alpha field.Element
		if isExt {
			alpha = gkrDeriveAlpha(fs)
		}

		// Compute the initial sum for this layer's sumcheck.
		// For logN>0 the sum is read directly from polys[0][0]+polys[0][1] (the
		// first-round polynomial already encodes the full sum over all instances).
		// For logN=0 there are no polys; the claim reduces to the single-instance
		// SBox evaluation, which equals Σ_w α^w × claims[w]^3 (correct for n=1).
		var currentSum field.Element
		if len(polys) > 0 {
			currentSum.Add(&polys[0][0], &polys[0][1])
		} else if isExt {
			var alphaPow field.Element
			alphaPow.SetOne()
			for w := 0; w < width; w++ {
				var term field.Element
				cube := sBoxCubeF(claims[w])
				term.Mul(&alphaPow, &cube)
				currentSum.Add(&currentSum, &term)
				alphaPow.Mul(&alphaPow, &alpha)
			}
		} else {
			currentSum = sBoxCubeF(claims[0])
		}

		// Verify sumcheck
		newPoint, err := gkrSumcheckVerifyLayer(polys, oracleEvals, currentSum, currentR, isExt, alpha, fs)
		if err != nil {
			return fmt.Errorf("GKR Poseidon2 layer %d: %w", l, err)
		}

		// Oracle holds post-SBox values. Apply matMul to get state after this round,
		// then add the next round key to obtain preSBox[l+1].
		claims = oracleEvals
		if isExt {
			applyMatMulExternalClaims(&claims)
		} else {
			applyMatMulInternalClaims(&claims)
		}
		if l < numGKRLayers-1 {
			applyAddRoundKeyClaims(l+1, &claims)
		}

		currentR = newPoint
	}

	// Final check: claims at r_final should match the output InnerProduct params
	// Output: ns[j] = state[8+j] (after round 27 matMul) + block[j] (feedforward)
	// After the GKR, claims[0..15] represent state[0..15] at r_final.
	// Output check: claims[8+j] + outputParams.Ys[width/2+j] == outputParams.Ys[j] for j=0..7
	// Where outputParams.Ys[j] = <eqRFinal, ns_j> and outputParams.Ys[8+j] = <eqRFinal, block_j>
	for j := 0; j < blockSize; j++ {
		var expected field.Element
		expected.Add(&claims[blockSize+j], &outputParams.Ys[blockSize+j].B0.A0)
		if expected != outputParams.Ys[j].B0.A0 {
			return fmt.Errorf("GKR Poseidon2 output check j=%d: got %v, expected %v (claims[%d]=%v + block=%v)",
				j, expected, outputParams.Ys[j].B0.A0, blockSize+j, claims[blockSize+j], outputParams.Ys[blockSize+j].B0.A0)
		}
	}

	return nil
}

// RunGnark implements wizard.VerifierAction.RunGnark (gnark circuit verifier).
// TODO: implement full gnark circuit GKR verification.
func (va *gkrPoseidon2VerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	// The gnark circuit verification of the GKR transcript requires:
	// 1. Reading r0 from the gnark circuit coins
	// 2. Reading the GKR transcript from the proof column (Proof status → available in VerifierCircuit)
	// 3. Running the GKR verifier using gnark field arithmetic + gnark FS
	// This is deferred to a future implementation.
	panic("GKR Poseidon2 gnark circuit verifier not yet implemented")
}

// ---------- Runtime helpers ----------

// gkrExtractR0FromRuntime extracts logN base field scalars from the FieldExt coins in a runtime.
func gkrExtractR0FromRuntime(run ifaces.Runtime, coinNames []coin.Name, logN int) []field.Element {
	r0 := make([]field.Element, logN)
	k := 0
	for _, name := range coinNames {
		fextVal := run.GetRandomCoinFieldExt(name)
		comps := [4]field.Element{fextVal.B0.A0, fextVal.B0.A1, fextVal.B1.A0, fextVal.B1.A1}
		for _, c := range comps {
			if k >= logN {
				break
			}
			r0[k] = c
			k++
		}
	}
	return r0
}

// gkrFieldToFext converts base field elements to fext.Element by base-field embedding
// (only B0.A0 is set; all other components are zero).
func gkrFieldToFext(vals []field.Element) []fext.Element {
	res := make([]fext.Element, len(vals))
	for i, v := range vals {
		res[i].B0.A0 = v
	}
	return res
}
