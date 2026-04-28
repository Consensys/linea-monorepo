package vortex

import (
	"os"
	"time"

	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/utils/types"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	gnarkvortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	vortex_bls12377 "github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/gpu"
	gpuvortex "github.com/consensys/linea-monorepo/prover/gpu/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// fextE4 alias used for clarity in the GPU LinComb glue. The underlying
// type is github.com/consensys/gnark-crypto/field/koalabear/extensions.E4
// (the same type the rest of the protocol uses via prover/maths/field/fext).
type fextE4 = extensions.E4

// useGPUVortex gates GPU dispatch on LIMITLESS_GPU_VORTEX=1. Default off.
func useGPUVortex() bool { return os.Getenv("LIMITLESS_GPU_VORTEX") == "1" }

type commitmentMode int

const (
	// Denotes the Vortex mode when we don't apply
	// self recursion
	NonSelfRecursion commitmentMode = iota
	// Denotes the Vortex mode when we apply
	// self recursion and commit using SIS
	SelfRecursionSIS
	// Denotes the Vortex mode when we apply
	// self recursion and commit using only Poseidon2
	SelfRecursionPoseidon2Only
)

// ReassignPrecomputedRootAction is a [wizard.ProverAction] that assigns the
// precomputed Merkle root of the Vortex invokation. The action is defined
// for round 0 only and only if the AddPrecomputedMerkleRootToPublicInputsOpt
// is enabled.
type ReassignPrecomputedRootAction struct {
	*Ctx
}

func (r ReassignPrecomputedRootAction) Run(run *wizard.ProverRuntime) {
	if r.IsBLS {
		for i := 0; i < encoding.KoalabearChunks; i++ {
			run.AssignColumn(
				r.Items.Precomputeds.MerkleRoot[i].GetColID(),
				smartvectors.NewConstant(r.AddPrecomputedMerkleRootToPublicInputsOpt.PrecomputedBLSValue[i], 1),
			)
		}
	} else {
		for i := 0; i < blockSize; i++ {
			run.AssignColumn(
				r.Items.Precomputeds.MerkleRoot[i].GetColID(),
				smartvectors.NewConstant(r.AddPrecomputedMerkleRootToPublicInputsOpt.PrecomputedValue[i], 1),
			)
		}
	}

}

// ColumnAssignmentProverAction is a [wizard.ProverAction] that assigns the
// the columns at a given round.
type ColumnAssignmentProverAction struct {
	*Ctx
	Round int
}

// Prover steps of Vortex that is run in place of committing to polynomials
func (ctx *ColumnAssignmentProverAction) Run(run *wizard.ProverRuntime) {

	round := ctx.Round

	// Check if that is a dry round
	if ctx.RoundStatus[round] == IsEmpty {
		// Nothing special to do.
		return
	}

	var (
		committedMatrix vortex_bls12377.EncodedMatrix
		sisColHashes    []field.Element // column hashes generated from SisTransversalHash
		noSisColHashes  []field.Element // column hashes generated from noSisTransversalHash, using LeafHashFunc
	)

	pols := ctx.getPols(run, round)

	// If there are no polynomials to commit to, we don't need to do anything
	if len(pols) == 0 {
		logrus.Infof("Vortex AssignColumn at round %v: No polynomials to commit to", round)
		return
	}

	// We commit to the polynomials with SIS hashing if the number of polynomials
	// is greater than the [ApplyToSISThreshold].

	if ctx.IsBLS {
		var (
			tree      *smt_bls12377.Tree
			colHashes []bls12377.Element
		)
		committedMatrix, _, tree, colHashes = ctx.VortexBLSParams.CommitMerkleWithoutSIS(pols)

		run.State.InsertNew(ctx.VortexProverStateName(round), committedMatrix)
		run.State.InsertNew(ctx.MerkleTreeName(round), tree)

		if ctx.IsSelfrecursed {
			// We need to store the SIS and non-SIS column hashes in the prover state
			// so that we can use them in the self-recursion compiler.
			if ctx.RoundStatus[round] == IsNoSis {
				run.State.InsertNew(ctx.NoSisHashName(round), colHashes)
			}
		}
		roots := encoding.EncodeBLS12RootToKoalabear(tree.Root)

		for i := 0; i < encoding.KoalabearChunks; i++ {
			run.AssignColumn(ifaces.ColID(ctx.MerkleRootName(round, i)), smartvectors.NewConstant(roots[i], 1))
		}
	} else {
		var (
			tree    *smt_koalabear.Tree
			gpuCS   *gpuvortex.CommitState // device-resident handle when SIS+GPU
		)

		if ctx.RoundStatus[round] == IsNoSis {
			committedMatrix, _, tree, noSisColHashes = ctx.VortexKoalaParams.CommitMerkleWithoutSIS(pols)
		} else if ctx.RoundStatus[round] == IsSISApplied {
			// SIS-applied Koala rounds are the GPU sweet spot: most of the
			// segment-prover wall-clock comes from these.
			//
			// Path selection:
			//
			//   GPU device-resident (CommitSIS): keeps the encoded matrix on
			//     device. Downstream LinComb runs as a single device kernel
			//     that produces a small UAlpha vector (~16 MiB at 2^20 scw),
			//     and OpenSelectedColumns extracts only the verifier's
			//     selected columns (~few MiB). Net: we save ~1.4 s of host
			//     reconstruction at production size and 4.7× the AVX-512 CPU
			//     vortex_koalabear baseline. Requires the goroutine to have
			//     bound a GPU via runtime.LockOSThread + dev.Bind() — that's
			//     what limitless.pinGPU(slot) does.
			//
			//   GPU drop-in (CommitMerkleWithSIS): D2Hs the full encoded
			//     matrix and rebuilds []SmartVector in Go-managed memory.
			//     The reconstruction overhead alone (~1.4 s at 2^20×2^11)
			//     wipes out the GPU compute win. NOT used here — kept around
			//     for legacy callers that still expect EncodedMatrix.
			//
			//   CPU AVX-512 (vortex_koalabear): production fallback. Used
			//     when GPU is unavailable, or LIMITLESS_GPU_VORTEX is unset.
			if useGPUVortex() && gpu.CurrentDevice() != nil {
				start := time.Now()
				gpuCS, tree, sisColHashes = gpuvortex.CommitSIS(
					ctx.VortexKoalaParams, pols, ctx.IsSelfrecursed)
				gpu.TraceEvent("vortex_commit_sis", gpu.CurrentDeviceID(), time.Since(start), map[string]any{
					"round": round,
					"rows":  len(pols),
					"cols":  ctx.VortexKoalaParams.NbColumns,
				})
			} else {
				committedMatrix, _, tree, sisColHashes = ctx.VortexKoalaParams.CommitMerkleWithSIS(pols)
			}
		}

		// Store the per-round commit handle. For GPU SIS rounds we wrap the
		// *CommitState; otherwise we wrap the host EncodedMatrix. Downstream
		// actions read both via asHandle() and dispatch on the variant.
		var handle *committedHandle
		if gpuCS != nil {
			handle = newGPUHandle(gpuCS)
		} else {
			handle = newHostHandle(committedMatrix)
		}
		run.State.InsertNew(ctx.VortexProverStateName(round), handle)
		run.State.InsertNew(ctx.MerkleTreeName(round), tree)

		// Only to be read by the self-recursion compiler.
		if ctx.IsSelfrecursed {
			// We need to store the SIS and non-SIS column hashes in the prover state
			// so that we can use them in the self-recursion compiler.
			if ctx.RoundStatus[round] == IsNoSis {
				run.State.InsertNew(ctx.NoSisHashName(round), noSisColHashes)
			} else if ctx.RoundStatus[round] == IsSISApplied {
				run.State.InsertNew(ctx.SisHashName(round), sisColHashes)
			}
		}
		for i := 0; i < blockSize; i++ {
			run.AssignColumn(ifaces.ColID(ctx.MerkleRootName(round, i)), smartvectors.NewConstant(tree.Root[i], 1))
		}
	}

}

type LinearCombinationComputationProverAction struct {
	*Ctx
}

// Run computes UAlpha = Σᵢ αⁱ · row[i] over the global stack of all
// committed rows (NoSIS rounds first, then SIS rounds; precomputeds are
// inserted on either side per ctx.IsSISAppliedToPrecomputed).
//
// Hybrid host + GPU path
// ──────────────────────
// Host rows (precomputeds, NoSIS rounds, SIS rounds without GPU) are
// concatenated and fed to the standard parallel host LinearCombination.
//
// GPU rows (SIS-applied Koala rounds with a *CommitState handle) are not
// D2H'd. Instead, for each GPU matrix M with cumulative row offset
// off_M relative to the start of the SIS section, we call
//
//	partial_M = cs_M.LinComb(α)            // single device kernel
//	UAlpha   += α^(off_NoSIS_total + off_M) · partial_M
//
// where off_NoSIS_total is the total row count of all host SIS+NoSIS
// rows that come before the GPU section. partial_M is a small E4 vector
// (~scw elements ≈ 16 MiB at 2^20 scw); the only D2H is this final
// vector, not the full encoded matrix.
//
// Stacking order is preserved: every GPU matrix's offset accounts for
// the rows that *would have been* before it in the flat stack.
func (ctx *LinearCombinationComputationProverAction) Run(pr *wizard.ProverRuntime) {
	var (
		committedSVNoSIS = []smartvectors.SmartVector{}
		hostSISRows      = []smartvectors.SmartVector{}
		gpuSISStates     = []*gpuvortex.CommitState{}
		gpuSISOffsets    = []int{} // row offset within the SIS section for each GPU matrix
	)

	// Precomputeds: always host. Add to NoSIS or SIS section per the SIS flag.
	if ctx.IsNonEmptyPrecomputed() {
		if ctx.IsSISAppliedToPrecomputed() {
			hostSISRows = append(hostSISRows, ctx.Items.Precomputeds.CommittedMatrix...)
		} else {
			committedSVNoSIS = append(committedSVNoSIS, ctx.Items.Precomputeds.CommittedMatrix...)
		}
	}

	// Walk rounds. SIS rounds may be GPU-resident (*committedHandle wrapping
	// *CommitState) or host (*committedHandle wrapping EncodedMatrix, or a
	// raw EncodedMatrix from legacy callers).
	sisRowOffset := len(hostSISRows) // running offset within SIS section
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		if ctx.RoundStatus[round] == IsEmpty {
			continue
		}
		raw := pr.State.MustGet(ctx.VortexProverStateName(round))
		h := asHandle(raw)
		if h == nil {
			utils.Panic("vortex linComb: unexpected state type at round %v: %T", round, raw)
		}

		switch ctx.RoundStatus[round] {
		case IsNoSis:
			// NoSIS: always host (no GPU NoSIS path yet).
			committedSVNoSIS = append(committedSVNoSIS, h.hostMatrix()...)
		case IsSISApplied:
			if h.isGPU() {
				gpuSISStates = append(gpuSISStates, h.gpu)
				gpuSISOffsets = append(gpuSISOffsets, sisRowOffset)
				sisRowOffset += h.numRows()
			} else {
				rows := h.hostMatrix()
				hostSISRows = append(hostSISRows, rows...)
				sisRowOffset += len(rows)
			}
		}
	}

	randomCoinLC := pr.GetRandomCoinFieldExt(ctx.Items.Alpha.Name)

	// Compute the host-only linear combination over [NoSIS..., hostSIS...].
	// NoSIS comes first so its rows occupy α^[0, |NoSIS|). Then within the
	// SIS section, host SIS rows come at offsets [0, |hostSIS|) relative
	// to the SIS section start; GPU SIS rows come *after* host SIS rows
	// in the same flat stack — but only as far as offset accounting goes;
	// they're not present in committedSV here.
	//
	// We fix this by ordering the host stack as
	//   [NoSIS..., hostSIS...]
	// and adding GPU partials with an offset of |NoSIS| + sisRowOffset_M
	// where sisRowOffset_M is the offset within the SIS section.
	committedSV := append(committedSVNoSIS, hostSISRows...)

	proof := &vortex.OpeningProof{}
	if len(committedSV) > 0 {
		vortex.LinearCombination(proof, committedSV, randomCoinLC)
	}

	// If there are no GPU SIS matrices, we're done — single-shot host path.
	if len(gpuSISStates) == 0 {
		pr.AssignColumn(ctx.Items.Ualpha.GetColID(), proof.LinearCombination)
		return
	}

	// All-GPU case: initialise the result vector to zeroes of length scw
	// before accumulating partials. We infer scw from the first state's
	// row count × rate via NumEncodedCols, but the simplest correct path
	// is to ask the first CommitState for its sizeCodeword via LinComb's
	// output length. We do that lazily inside addGPUPartialsToLinComb.
	if proof.LinearCombination == nil {
		// Defer allocation: addGPUPartialsToLinComb will allocate based
		// on the first GPU partial's length.
	}

	// Add α^global_offset · cs.LinComb(α) for each GPU matrix.
	noSISCount := len(committedSVNoSIS)
	addGPUPartialsToLinComb(proof, gpuSISStates, gpuSISOffsets, noSISCount, randomCoinLC)
	pr.AssignColumn(ctx.Items.Ualpha.GetColID(), proof.LinearCombination)
}

// addGPUPartialsToLinComb computes, for each (cs, sisOffset_M) pair:
//
//	contribution = α^(noSISCount + sisOffset_M) · cs.LinComb(α)
//
// and adds it elementwise to proof.LinearCombination (which must already
// be a *RegularExt SmartVector of length sizeCodeWord).
//
// Why the offsets are split: the global stack is
//
//	[NoSIS rows ........... | hostSIS rows ........ | GPU SIS rows ........]
//	 ↑ off=0                  ↑ off=noSISCount       ↑ off=noSISCount+|hostSIS|
//
// gpuSISOffsets[i] is sisOffset_M (within the SIS section), so the
// global offset is noSISCount + gpuSISOffsets[i].
func addGPUPartialsToLinComb(proof *vortex.OpeningProof,
	states []*gpuvortex.CommitState, sisOffsets []int, noSISCount int,
	alpha fextE4,
) {
	for i, cs := range states {
		partial, err := cs.LinComb(alpha)
		if err != nil {
			utils.Panic("vortex GPU lincomb partial[%d]: %v", i, err)
		}
		// Lazy-allocate the result vector when the host LinComb didn't
		// run (all-GPU case): partial has length sizeCodeword and that's
		// the global LinComb length too.
		if proof.LinearCombination == nil {
			zeros := make([]fextE4, len(partial))
			proof.LinearCombination = smartvectors.NewRegularExt(zeros)
		}
		dst := proof.LinearCombination.(*smartvectors.RegularExt)

		// Scale partial by α^(noSISCount + sisOffsets[i]).
		// Use sequential multiply rather than big-int Exp; the offset
		// is bounded by the total committed-row count (typically small).
		var scale fextE4
		scale.SetOne()
		for k := noSISCount + sisOffsets[i]; k > 0; k-- {
			scale.Mul(&scale, &alpha)
		}
		// (*dst)[j] += scale · partial[j]
		for j := range partial {
			var t fextE4
			t.Mul(&partial[j], &scale)
			(*dst)[j].Add(&(*dst)[j], &t)
		}
	}
}

// ComputeLinearCombFromRsMatrix is the same as ComputeLinearComb but uses
// the RS encoded matrix instead of using the basic one. It is slower than
// the later but is recommended.
func (ctx *Ctx) ComputeLinearCombFromRsMatrix(run *wizard.ProverRuntime) {

	var (
		committedSVSIS   = []smartvectors.SmartVector{}
		committedSVNoSIS = []smartvectors.SmartVector{}
	)

	// Add the precomputed columns to commitedSVSIS or commitedSVNoSIS
	if ctx.IsSISAppliedToPrecomputed() {
		committedSVSIS = append(committedSVSIS, ctx.Items.Precomputeds.CommittedMatrix...)
	} else {
		committedSVNoSIS = append(committedSVNoSIS, ctx.Items.Precomputeds.CommittedMatrix...)
	}

	// Collect all the committed polynomials : round by round
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// There are not included in the commitments so there
		// is no need to proceed.
		if ctx.RoundStatus[round] == IsEmpty {
			continue
		}

		committedMatrix := run.State.MustGet(ctx.VortexProverStateName(round)).(vortex_koalabear.EncodedMatrix)

		// Push pols to the right stack
		if ctx.RoundStatus[round] == IsNoSis {
			committedSVNoSIS = append(committedSVNoSIS, committedMatrix...)
		} else if ctx.RoundStatus[round] == IsSISApplied {
			committedSVSIS = append(committedSVSIS, committedMatrix...)
		}
	}

	// Construct committedSV by stacking the No SIS round
	// matrices before the SIS round matrices
	committedSV := append(committedSVNoSIS, committedSVSIS...)

	// And get the randomness
	randomCoinLC := run.GetRandomCoinFieldExt(ctx.Items.Alpha.Name)

	// and compute and assign the random linear combination of the rows
	proof := &vortex.OpeningProof{}
	vortex.LinearCombination(proof, committedSV, randomCoinLC)

	run.AssignColumn(ctx.Items.Ualpha.GetColID(), proof.LinearCombination)
}

// Prover steps of Vortex where he opens the columns selected by the verifier
// We stack the no SIS round matrices before the SIS round matrices in the committed matrix stack.
// The same is done for the tree.
type OpenSelectedColumnsProverAction struct {
	*Ctx
}

// Run extracts the selected columns + Merkle proofs and writes them into
// the proof under the appropriate column IDs.
//
// GPU integration
// ───────────────
// For SIS-applied Koala rounds with a *committedHandle wrapping a
// gpuvortex.CommitState, columns are extracted directly via
// cs.ExtractColumns — a small D2H of only the selected entries. The
// host-side Merkle tree (already stored under MerkleTreeName(round)) is
// used unchanged for proof generation. After all columns are extracted
// the GPU buffers are freed via h.free().
//
// Host SIS rounds, NoSIS rounds, BLS rounds, and precomputeds all stay
// on the existing host path.
func (ctx *OpenSelectedColumnsProverAction) Run(run *wizard.ProverRuntime) {

	var (
		committedMatricesSIS   = []vortex_bls12377.EncodedMatrix{}
		committedMatricesNoSIS = []vortex_bls12377.EncodedMatrix{}
		treesSIS               = []*smt_koalabear.Tree{}
		treesNoSIS             = []*smt_koalabear.Tree{}
		blsTrees               = []*smt_bls12377.Tree{}
		// GPU SIS rounds: handle (for ExtractColumns + free) and the index
		// into the final committedMatrices list. We pass an empty
		// EncodedMatrix as a placeholder so SelectColumnsAndMerkleProofs's
		// per-matrix iteration still indexes correctly; we overwrite
		// proof.Columns[idx] afterward.
		gpuSISHandles  []*committedHandle
		gpuSISMatrixIdx []int // global index into committedMatrices = NoSIS+SIS
	)

	// Append the precomputed committedMatrices and trees to the SIS or no SIS matrices
	// or trees as per the number of precomputed columns are more than the [ApplyToSISThreshold]
	if ctx.IsNonEmptyPrecomputed() {
		if ctx.IsSISAppliedToPrecomputed() {
			committedMatricesSIS = append(committedMatricesSIS, ctx.Items.Precomputeds.CommittedMatrix)
			treesSIS = append(treesSIS, ctx.Items.Precomputeds.Tree)
		} else {
			if ctx.IsBLS {
				committedMatricesNoSIS = append(committedMatricesNoSIS, ctx.Items.Precomputeds.CommittedMatrix)
				blsTrees = append(blsTrees, ctx.Items.Precomputeds.BLSTree)
			} else {
				committedMatricesNoSIS = append(committedMatricesNoSIS, ctx.Items.Precomputeds.CommittedMatrix)
				treesNoSIS = append(treesNoSIS, ctx.Items.Precomputeds.Tree)
			}

		}
	}

	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		// There are not included in the commitments so there
		// is no need to proceed.
		if ctx.RoundStatus[round] == IsEmpty {
			continue
		}
		// Fetch the round's commit handle (legacy callers may store a raw
		// EncodedMatrix; new callers store a *committedHandle that may
		// wrap either a host EncodedMatrix or a GPU CommitState).
		raw := run.State.MustGet(ctx.VortexProverStateName(round))
		// Delete eagerly: the encoded matrix may be very large and the
		// state map keeps references alive longer than necessary.
		run.State.Del(ctx.VortexProverStateName(round))

		h := asHandle(raw)
		if h == nil {
			utils.Panic("vortex open: unexpected state type at round %v: %T", round, raw)
		}

		// Also fetches the trees from the prover state
		if ctx.IsBLS {
			// BLS path has no GPU implementation yet — must be host-resident.
			committedMatrix := h.hostMatrix()
			tree := run.State.MustGet(ctx.MerkleTreeName(round)).(*smt_bls12377.Tree)
			if ctx.RoundStatus[round] == IsNoSis {
				committedMatricesNoSIS = append(committedMatricesNoSIS, committedMatrix)
				blsTrees = append(blsTrees, tree)
			} else if ctx.RoundStatus[round] == IsSISApplied {
				utils.Panic("IsSISApplied is not supported in BLS mode (round %v)", round)
			}

		} else {
			tree := run.State.MustGet(ctx.MerkleTreeName(round)).(*smt_koalabear.Tree)
			if ctx.RoundStatus[round] == IsNoSis {
				// NoSIS rounds are host-resident.
				committedMatricesNoSIS = append(committedMatricesNoSIS, h.hostMatrix())
				treesNoSIS = append(treesNoSIS, tree)
			} else if ctx.RoundStatus[round] == IsSISApplied {
				if h.isGPU() {
					// Reserve a slot in the SIS section with an empty
					// placeholder; we'll overwrite proof.Columns[idx] after
					// SelectColumnsAndMerkleProofs runs. The host path
					// iterates `range committedMatrices[i]` so an empty
					// slice is safe — proof.Columns[idx] just becomes a
					// list of empty []field.Element which we replace.
					gpuSISHandles = append(gpuSISHandles, h)
					committedMatricesSIS = append(committedMatricesSIS, vortex_bls12377.EncodedMatrix{})
					treesSIS = append(treesSIS, tree)
				} else {
					committedMatricesSIS = append(committedMatricesSIS, h.hostMatrix())
					treesSIS = append(treesSIS, tree)
				}
			}
		}

	}

	// Pre-compute the global index of each GPU SIS matrix in the final
	// committedMatrices list (NoSIS first, then SIS).
	if len(gpuSISHandles) > 0 {
		// SIS section starts at len(committedMatricesNoSIS).
		// gpu handles were appended to committedMatricesSIS in the order
		// we encountered them, so their SIS-section offsets are
		// (len(committedMatricesSIS) - len(gpuSISHandles)) + i, but a
		// single sweep is simpler:
		nNoSIS := len(committedMatricesNoSIS)
		gpuSISMatrixIdx = make([]int, 0, len(gpuSISHandles))
		gpuPtr := 0
		for sisIdx := 0; sisIdx < len(committedMatricesSIS); sisIdx++ {
			if len(committedMatricesSIS[sisIdx]) == 0 {
				gpuSISMatrixIdx = append(gpuSISMatrixIdx, nNoSIS+sisIdx)
				gpuPtr++
			}
		}
		_ = gpuPtr // assertion: gpuPtr should equal len(gpuSISHandles)
	}

	// Free original committed columns from run.Columns — their data has been
	// encoded into the Vortex matrices and is no longer needed in raw form.
	for round := 0; round <= ctx.MaxCommittedRound; round++ {
		if ctx.RoundStatus[round] == IsEmpty {
			continue
		}
		for _, colName := range ctx.CommitmentsByRounds.MustGet(round) {
			run.Columns.TryDel(colName)
		}
	}

	// Stack the no SIS matrices and trees before the SIS matrices and trees
	committedMatrices := append(committedMatricesNoSIS, committedMatricesSIS...)
	trees := append(treesNoSIS, treesSIS...)

	entryList := run.GetRandomCoinIntegerVec(ctx.Items.Q.Name)
	proof := vortex.OpeningProof{}

	// Amend the Vortex proof with the Merkle proofs and registers
	// the Merkle proofs in the prover runtime

	if ctx.IsBLS {
		merkleProofs := vortex_bls12377.SelectColumnsAndMerkleProofs(&proof, entryList, committedMatrices, blsTrees)

		packedMProofs := ctx.packBLSMerkleProofs(merkleProofs)

		for i := range ctx.Items.BLSMerkleProofs {
			run.AssignColumn(ctx.Items.BLSMerkleProofs[i].GetColID(), packedMProofs[i])
		}

	} else {

		merkleProofs := vortex_koalabear.SelectColumnsAndMerkleProofs(&proof, entryList, committedMatrices, trees)

		// For GPU SIS rounds, SelectColumnsAndMerkleProofs above produced
		// empty placeholder slots in proof.Columns. Replace them with the
		// real columns extracted directly from the device buffers.
		// This is the small D2H — only |entryList| × nRows × 4 bytes per
		// matrix, never the full encoded matrix.
		for i, h := range gpuSISHandles {
			matrixIdx := gpuSISMatrixIdx[i]
			proof.Columns[matrixIdx] = h.extractColumns(entryList)
		}

		packedMProofs := ctx.packMerkleProofs(merkleProofs)

		for i := range ctx.Items.MerkleProofs {
			run.AssignColumn(ctx.Items.MerkleProofs[i].GetColID(), packedMProofs[i])
		}

	}
	// Release GPU buffers now that columns have been extracted. Subsequent
	// recursion / self-recursion paths read host data only.
	for _, h := range gpuSISHandles {
		h.free()
	}
	selectedCols := proof.Columns

	// Assign the opened columns
	ctx.assignOpenedColumns(run, entryList, selectedCols, NonSelfRecursion)

	// Assign the SIS and non SIS selected columns for the self-recursion
	// compiler. Instead of calling SelectColumnsAndMerkleProofs again, we
	// partition the already-computed selectedCols by the NoSIS/SIS boundary.
	// The combined matrices were built as append(NoSIS, SIS...), so the
	// first len(committedMatricesNoSIS) entries are NoSIS.
	numNoSIS := len(committedMatricesNoSIS)

	if len(committedMatricesSIS) > 0 {
		sisSelectedCols := selectedCols[numNoSIS:]
		ctx.assignOpenedColumns(run, entryList, sisSelectedCols, SelfRecursionSIS)
	}
	if len(committedMatricesNoSIS) > 0 {
		nonSisSelectedCols := selectedCols[:numNoSIS]
		ctx.assignOpenedColumns(run, entryList, nonSisSelectedCols, SelfRecursionPoseidon2Only)
		ctx.storeSelectedColumnsForNonSisRounds(run, nonSisSelectedCols)
	}
}

// returns the list of all committed smartvectors for the given round
// so that we can commit to them
func (ctx *Ctx) getPols(run *wizard.ProverRuntime, round int) (pols []smartvectors.SmartVector) {
	names := ctx.CommitmentsByRounds.MustGet(round)
	pols = make([]smartvectors.SmartVector, len(names))
	for i := range names {
		pols[i] = run.Columns.MustGet(names[i])
	}
	return pols
}

// pack a list of merkle-proofs in a vector as used in the merkle proof module
func (ctx *Ctx) packMerkleProofs(proofs [][]smt_koalabear.Proof) [8]smartvectors.SmartVector {

	depth := len(proofs[0][0].Siblings) // depth of the Merkle-tree
	res := [8][]field.Element{}
	for i := range res {
		res[i] = make([]field.Element, ctx.MerkleProofSize())
	}
	numProofWritten := 0

	// Sanity-checks

	if depth != utils.Log2Ceil(ctx.NumEncodedCols()) {
		utils.Panic(
			"expected depth to be equal to Log2(NumEncodedCols()), got %v, %v",
			depth, utils.Log2Ceil(ctx.NumEncodedCols()),
		)
	}

	// When we commit to the precomputeds, len(proofs) = ctx.NumCommittedRounds + 1,
	// otherwise len(proofs) = ctx.NumCommittedRounds
	if len(proofs) != ctx.NumCommittedRounds() && !ctx.IsNonEmptyPrecomputed() {
		utils.Panic(
			"inconsitent proofs length %v, %v",
			len(proofs), ctx.NumCommittedRounds(),
		)
	}

	if len(proofs[0]) != ctx.NbColsToOpen() {
		utils.Panic(
			"expected proofs[0] and NbColsToOpen to be equal: %v, %v",
			len(proofs[0]), ctx.NbColsToOpen(),
		)
	}

	for i := range proofs {
		for j := range proofs[i] {
			p := proofs[i][j]
			for k := range p.Siblings {
				// The proof stores the sibling bottom-up but we want to pack
				// the proof in top-down order.
				hashOct := p.Siblings[depth-1-k]
				for coord := range res {
					res[coord][numProofWritten*depth+k] = hashOct[coord]
				}
			}
			numProofWritten++
		}
	}

	resSV := [8]smartvectors.SmartVector{}
	for i := range res {
		resSV[i] = smartvectors.NewRegular(res[i])
	}

	return resSV
}

// pack a list of merkle-proofs in a vector as used in the merkle proof module
func (ctx *Ctx) packBLSMerkleProofs(proofs [][]smt_bls12377.Proof) [encoding.KoalabearChunks]smartvectors.SmartVector {

	depth := len(proofs[0][0].Siblings) // depth of the Merkle-tree
	res := [encoding.KoalabearChunks][]field.Element{}
	for i := range res {
		res[i] = make([]field.Element, ctx.MerkleProofSize())
	}

	numProofWritten := 0

	// Sanity-checks
	if depth != utils.Log2Ceil(ctx.NumEncodedCols()) {
		utils.Panic(
			"expected depth to be equal to Log2(NumEncodedCols()), got %v, %v",
			depth, utils.Log2Ceil(ctx.NumEncodedCols()),
		)
	}

	// When we commit to the precomputeds, len(proofs) = ctx.NumCommittedRounds + 1,
	// otherwise len(proofs) = ctx.NumCommittedRounds
	if len(proofs) != ctx.NumCommittedRounds() && !ctx.IsNonEmptyPrecomputed() {
		utils.Panic(
			"inconsitent proofs length %v, %v",
			len(proofs), ctx.NumCommittedRounds(),
		)
	}

	if len(proofs) != (ctx.NumCommittedRounds()+1) && ctx.IsNonEmptyPrecomputed() {
		utils.Panic(
			"inconsitent proofs length %v, %v",
			len(proofs), ctx.NumCommittedRounds()+1,
		)
	}

	if len(proofs[0]) != ctx.NbColsToOpen() {
		utils.Panic(
			"expected proofs[0] and NbColsToOpen to be equal: %v, %v",
			len(proofs[0]), ctx.NbColsToOpen(),
		)
	}

	for i := range proofs {
		for j := range proofs[i] {
			p := proofs[i][j]
			for k := range p.Siblings {
				// The proof stores the sibling bottom-up but
				// we want to pack the proof in top-down order.
				koalaElems := encoding.EncodeBLS12RootToKoalabear(p.Siblings[depth-1-k])

				for coord := range res {
					res[coord][numProofWritten*depth+k] = koalaElems[coord]
				}
			}
			numProofWritten++
		}
	}

	// return smartvectors.NewRegular(res)
	resSV := [encoding.KoalabearChunks]smartvectors.SmartVector{}
	for i := range res {
		resSV[i] = smartvectors.NewRegular(res[i])
	}

	return resSV
}

// unpack a list of merkle proofs from a vector as in
func (ctx *Ctx) unpackMerkleProofs(sv [8]smartvectors.SmartVector, entryList []int) (proofs [][]smt_koalabear.Proof) {

	depth := utils.Log2Ceil(ctx.NumEncodedCols()) // depth of the Merkle-tree
	numComs := ctx.NumCommittedRounds()
	if ctx.IsNonEmptyPrecomputed() {
		numComs = ctx.NumCommittedRounds() + 1 // Need to consider the precomputed commitments
	}
	numEntries := len(entryList)

	proofs = make([][]smt_koalabear.Proof, numComs)
	curr := 0 // tracks the position in sv that we are parsing.

	for i := range proofs {
		proofs[i] = make([]smt_koalabear.Proof, numEntries)
		for j := range proofs[i] {
			// initialize the proof that we are parsing
			proof := smt_koalabear.Proof{
				Path:     entryList[j],
				Siblings: make([]types.KoalaOctuplet, depth),
			}

			// parse the siblings accounting for the fact that we
			// are inversing the order.
			for k := range proof.Siblings {
				var v gnarkvortex.Hash
				for coord := 0; coord < len(v); coord++ {
					v[coord] = sv[coord].Get(curr)
				}
				proof.Siblings[depth-k-1] = v
				curr++
			}

			proofs[i][j] = proof
		}
	}
	return proofs
}

// unpack a list of merkle proofs from a vector as in
func (ctx *Ctx) unpackBLSMerkleProofs(sv [encoding.KoalabearChunks]smartvectors.SmartVector, entryList []int) (proofs [][]smt_bls12377.Proof) {

	depth := utils.Log2Ceil(ctx.NumEncodedCols()) // depth of the Merkle-tree
	numComs := ctx.NumCommittedRounds()
	if ctx.IsNonEmptyPrecomputed() {
		numComs = ctx.NumCommittedRounds() + 1 // Need to consider the precomputed commitments
	}
	numEntries := len(entryList)

	proofs = make([][]smt_bls12377.Proof, numComs)
	curr := 0 // tracks the position in sv that we are parsing.

	for i := range proofs {
		proofs[i] = make([]smt_bls12377.Proof, numEntries)
		for j := range proofs[i] {
			// initialize the proof that we are parsing
			proof := smt_bls12377.Proof{
				Path:     entryList[j],
				Siblings: make([]bls12377.Element, depth),
			}

			// parse the siblings accounting for the fact that we
			// are inversing the order.
			for k := range proof.Siblings {
				var v [encoding.KoalabearChunks]field.Element
				for coord := 0; coord < len(v); coord++ {
					v[coord] = sv[coord].Get(curr)
				}
				proof.Siblings[depth-k-1] = encoding.DecodeKoalabearToBLS12Root(v)
				curr++
			}

			proofs[i][j] = proof
		}
	}
	return proofs
}

// assignOpenedColumns assign the opened columns for
// both normal and self-recursion compilers
func (ctx *Ctx) assignOpenedColumns(
	pr *wizard.ProverRuntime,
	entryList []int,
	selectedCols [][][]field.Element,
	mode commitmentMode) {
	// The columns are split by commitment round. So we need to
	// restick them when we commit them.
	totalLen := 0
	for i := range selectedCols {
		if len(selectedCols[i]) > 0 {
			totalLen += len(selectedCols[i][0])
		}
	}
	for j := range entryList {
		fullCol := make([]field.Element, 0, totalLen)
		for i := range selectedCols {
			fullCol = append(fullCol, selectedCols[i][j]...)
		}

		// Converts it into a smart-vector and zero-pad it if necessary
		var assignable smartvectors.SmartVector = smartvectors.NewRegular(fullCol)
		if assignable.Len() < utils.NextPowerOfTwo(len(fullCol)) {
			assignable = smartvectors.RightZeroPadded(fullCol, utils.NextPowerOfTwo(len(fullCol)))
		}
		if mode == NonSelfRecursion {
			pr.AssignColumn(ctx.Items.OpenedColumns[j].GetColID(), assignable)
		} else if mode == SelfRecursionSIS {
			pr.AssignColumn(ctx.Items.OpenedSISColumns[j].GetColID(), assignable)
		} else if mode == SelfRecursionPoseidon2Only {
			pr.AssignColumn(ctx.Items.OpenedNonSISColumns[j].GetColID(), assignable)
		}
	}

}

// storeSelectedColumnsForNonSisRound stores the selected columns in the prover state
// for the non SIS rounds which is to be used in the self-recursion compilers
func (ctx *Ctx) storeSelectedColumnsForNonSisRounds(
	pr *wizard.ProverRuntime,
	selectedCols [][][]field.Element) {
	numNonSisRound := ctx.NumCommittedRoundsNoSis()
	if ctx.IsNonEmptyPrecomputed() && !ctx.IsSISAppliedToPrecomputed() {
		numNonSisRound++
	}
	// selectedColsQ[i][j][k] stores the jth selected
	// column of the ith non SIS round
	selectedColsQ := make([][][]field.Element, numNonSisRound)
	// Sanity check
	if len(selectedCols) != numNonSisRound {
		utils.Panic(
			"expected selectedCols to be of length %v, got %v",
			numNonSisRound, len(selectedCols),
		)
	}
	for i := range selectedCols {
		// Sanity check
		if len(selectedCols[i]) != ctx.NbColsToOpen() {
			utils.Panic(
				"expected selectedCols[%v] to be of length %v, got %v",
				i, ctx.NbColsToOpen(), len(selectedCols[i]),
			)
		}
		selectedColsQ[i] = make([][]field.Element, ctx.NbColsToOpen())
		for j := range selectedCols[i] {
			selectedColsQ[i][j] = make([]field.Element, len(selectedCols[i][j]))
			copy(selectedColsQ[i][j], selectedCols[i][j])
		}
	}
	// Store the selected columns in the prover state
	pr.State.InsertNew(
		ctx.SelectedColumnNonSISName(),
		selectedColsQ)
}
