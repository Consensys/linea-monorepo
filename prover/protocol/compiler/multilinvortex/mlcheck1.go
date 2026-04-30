package multilinvortex

// CommitOriginalMLColumns fixes the soundness gap in the ML Vortex protocol
// for the first-round columns (RegisterCommit columns at round 0).
//
// Without this pass, the Fiat-Shamir coins (evaluation point and α) are derived
// from a transcript that contains no commitment to the witness columns, letting
// a cheating prover choose UAlpha freely.
//
// This pass adds, for each round-0 MultilinearEval query group:
//
//   - A Merkle root proof column at round (ctx.Round-1), so the root enters the
//     FS transcript before α is drawn.
//   - t opened column vectors + Merkle sibling paths at round (ctx.Round+1).
//   - Check 2 on original: verifier re-hashes opened columns and checks the
//     Merkle path against the root.
//   - Check 1: verifier checks UAlpha[k][j] = Σ_b α^b · orig_enc[k·2ⁿᴿᵒʷ+b][j]
//     for every opened position j, confirming that UAlpha was computed correctly
//     from the committed original data.
//
// Call CommitOriginalMLColumns after each (multilinvortex.Compile + CommitMLColumns)
// pair in the compile chain.  Repeated calls are idempotent (each context is
// processed once).

import (
	"fmt"
	"math/bits"

	"github.com/consensys/gnark/frontend"
	poseidon2 "github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	smt_koalabear "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	vortex_common "github.com/consensys/linea-monorepo/prover/crypto/vortex"
	vortex_koalabear "github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const mlOrigCommitSeenKey = "mlvortex_orig_seen"

func CommitOriginalMLColumns(comp *wizard.CompiledIOP) {
	if comp.ExtraData == nil {
		return
	}
	ctxs, _ := comp.ExtraData[mlvortexContextsKey].([]*Context)
	entries, _ := comp.ExtraData[mlCommitEntriesKey].([]mlCommitEntry)

	seen, _ := comp.ExtraData[mlOrigCommitSeenKey].(map[ifaces.QueryID]bool)
	if seen == nil {
		seen = make(map[ifaces.QueryID]bool)
		comp.ExtraData[mlOrigCommitSeenKey] = seen
	}

	for _, ctx := range ctxs {
		if seen[ctx.InputQuery.QueryID] {
			continue
		}
		// Only handle contexts whose input polys are round-0 committed (not proof).
		if len(ctx.InputQuery.Pols) == 0 {
			continue
		}
		if comp.Columns.Status(ctx.InputQuery.Pols[0].GetColID()) != column.Committed {
			continue
		}

		// Find the matching UAlpha mlCommitEntry.
		uAlphaEntry := findUAlphaEntry(entries, ctx)
		if uAlphaEntry == nil {
			continue
		}

		seen[ctx.InputQuery.QueryID] = true

		K := len(ctx.InputQuery.Pols)
		nRow := ctx.NRow
		nCol := ctx.NCol
		nRowSize := 1 << nRow
		nColSize := 1 << nCol
		nbRowsOrig := K * nRowSize
		nbCodewordCols := nColSize * 2
		t := len(uAlphaEntry.entryList)
		depth := bits.Len(uint(nbCodewordCols)) - 1

		params := vortex_koalabear.NewParams(2, nColSize, nbRowsOrig, 6, 16)

		sfx := fmt.Sprintf("q%s", string(ctx.InputQuery.QueryID))

		// Root at round (ctx.Round-1) so it enters the FS transcript before α.
		rootRound := ctx.Round - 1
		if rootRound < 0 {
			rootRound = 0
		}
		rootCol := comp.InsertProof(rootRound,
			ifaces.ColID("MLVORTEX_ORIG_ROOT_"+sfx), 8, false)

		// Opened data and paths at round (ctx.Round+1) — same round as UAlpha.
		openedSize := utils.NextPowerOfTwo(t * nbRowsOrig)
		openedDataCol := comp.InsertProof(ctx.Round+1,
			ifaces.ColID("MLVORTEX_ORIG_OPENED_"+sfx), openedSize, false)
		pathsSize := utils.NextPowerOfTwo(t * depth * 8)
		pathsCol := comp.InsertProof(ctx.Round+1,
			ifaces.ColID("MLVORTEX_ORIG_PATHS_"+sfx), pathsSize, false)

		shared := &mlOrigShared{}

		comp.RegisterProverAction(rootRound, &mlOrigCommitProverAction{
			cols:    ctx.InputQuery.Pols,
			params:  &params,
			K:       K,
			nRow:    nRow,
			nCol:    nCol,
			rootCol: rootCol,
			shared:  shared,
		})

		comp.RegisterProverAction(ctx.Round+1, &mlOrigOpenProverAction{
			shared:        shared,
			entryList:     uAlphaEntry.entryList,
			nbRowsOrig:    nbRowsOrig,
			depth:         depth,
			openedSize:    openedSize,
			pathsSize:     pathsSize,
			openedDataCol: openedDataCol,
			pathsCol:      pathsCol,
		})

		// Find the index of each UAlpha column within the mixed group so the
		// verifier can read the correct row offsets from the opened UAlpha data.
		kTotal := len(uAlphaEntry.cols) // may include RowEvals if nRow == nCol
		uAlphaIndices := make([]int, K)
		for k, ua := range ctx.UAlpha {
			for idx, c := range uAlphaEntry.cols {
				if c.GetColID() == ua.GetColID() {
					uAlphaIndices[k] = idx
					break
				}
			}
		}

		comp.RegisterVerifierAction(comp.NumRounds()-1, &mlCheck1VerifierAction{
			rootCol:             rootCol,
			origOpenedDataCol:   openedDataCol,
			origPathsCol:        pathsCol,
			uAlphaOpenedDataCol: uAlphaEntry.openedDataCol,
			alphaCoin:           ctx.AlphaCoin,
			K:                   K,
			kTotal:              kTotal,
			nRow:                nRow,
			nbRowsOrig:          nbRowsOrig,
			nbRowsUA:            uAlphaEntry.nbRows, // 4*kTotal
			entryList:           uAlphaEntry.entryList,
			depth:               depth,
			uAlphaIndices:       uAlphaIndices,
		})
	}
}

// findUAlphaEntry finds the mlCommitEntry whose cols match ctx.UAlpha.
func findUAlphaEntry(entries []mlCommitEntry, ctx *Context) *mlCommitEntry {
	if len(ctx.UAlpha) == 0 {
		return nil
	}
	target := ctx.UAlpha[0].GetColID()
	for i := range entries {
		for _, c := range entries[i].cols {
			if c.GetColID() == target {
				return &entries[i]
			}
		}
	}
	return nil
}

// mlOrigShared holds the encoded matrix built at commit-time for use at open-time.
type mlOrigShared struct {
	encodedMatrix vortex_koalabear.EncodedMatrix
	tree          *smt_koalabear.Tree
}

// mlOrigCommitProverAction runs at round (ctx.Round-1): RS-encodes + Merkle-commits
// the original round-0 columns and assigns the root proof column.
type mlOrigCommitProverAction struct {
	cols    []ifaces.Column
	params  *vortex_koalabear.Params
	K       int
	nRow    int
	nCol    int
	rootCol ifaces.Column
	shared  *mlOrigShared
}

func (a *mlOrigCommitProverAction) Run(run *wizard.ProverRuntime) {
	nRowSize := 1 << a.nRow
	nColSize := 1 << a.nCol

	// Build the flat matrix: K * 2^nRow rows of base-field elements, each of
	// length 2^nCol.  Row (k * nRowSize + b) = the b-th chunk of poly P[k].
	pols := make([]smartvectors.SmartVector, a.K*nRowSize)
	for k, col := range a.cols {
		sv := run.GetColumn(col.GetColID())
		for b := 0; b < nRowSize; b++ {
			rowData := make([]field.Element, nColSize)
			for j := 0; j < nColSize; j++ {
				rowData[j] = sv.Get(b*nColSize + j)
			}
			pols[k*nRowSize+b] = smartvectors.NewRegular(rowData)
		}
	}

	encodedMatrix, _, tree, _ := a.params.CommitMerkleWithoutSIS(pols)
	a.shared.encodedMatrix = encodedMatrix
	a.shared.tree = tree

	rootVec := make([]field.Element, 8)
	for i, e := range tree.Root {
		rootVec[i] = e
	}
	run.AssignColumn(a.rootCol.GetColID(), smartvectors.NewRegular(rootVec))
}

// mlOrigOpenProverAction runs at round (ctx.Round+1): opens t columns of the
// original encoded matrix and assigns the opened-data and paths proof columns.
type mlOrigOpenProverAction struct {
	shared        *mlOrigShared
	entryList     []int
	nbRowsOrig    int
	depth         int
	openedSize    int
	pathsSize     int
	openedDataCol ifaces.Column
	pathsCol      ifaces.Column
}

func (a *mlOrigOpenProverAction) Run(run *wizard.ProverRuntime) {
	dummy := &vortex_common.OpeningProof{}
	merkleProofs := vortex_koalabear.SelectColumnsAndMerkleProofs(
		dummy,
		a.entryList,
		[]vortex_koalabear.EncodedMatrix{a.shared.encodedMatrix},
		[]*smt_koalabear.Tree{a.shared.tree},
	)

	openedVec := make([]field.Element, a.openedSize)
	for j, colVec := range dummy.Columns[0] {
		for row, val := range colVec {
			openedVec[j*a.nbRowsOrig+row] = val
		}
	}
	run.AssignColumn(a.openedDataCol.GetColID(), smartvectors.NewRegular(openedVec))

	pathsVec := make([]field.Element, a.pathsSize)
	for j, proof := range merkleProofs[0] {
		for d, sib := range proof.Siblings {
			for i, e := range sib {
				pathsVec[j*a.depth*8+d*8+i] = e
			}
		}
	}
	run.AssignColumn(a.pathsCol.GetColID(), smartvectors.NewRegular(pathsVec))
}

// mlCheck1VerifierAction implements:
//   - Check 2 on original: Poseidon2-hash each opened column, verify Merkle path.
//   - Check 1: for each k and each opened position j,
//     UAlpha_enc[k][j] == Σ_b α^b · orig_enc[k·2^nRow+b][j]
type mlCheck1VerifierAction struct {
	rootCol             ifaces.Column
	origOpenedDataCol   ifaces.Column
	origPathsCol        ifaces.Column
	uAlphaOpenedDataCol ifaces.Column
	alphaCoin           coin.Info
	K                   int
	kTotal              int // total cols in UAlpha CommitMLColumns group (may include RowEvals)
	nRow                int
	nbRowsOrig          int   // K * 2^nRow
	nbRowsUA            int   // 4*kTotal
	entryList           []int
	depth               int
	uAlphaIndices       []int // index of UAlpha[k] within the group's cols slice
}

func (v *mlCheck1VerifierAction) Run(run wizard.Runtime) error {
	// Read Merkle root.
	var root field.Octuplet
	for i := range root {
		root[i] = run.GetColumnAt(v.rootCol.GetColID(), i)
	}

	h := poseidon2.NewMDHasher()

	// Pre-compute α powers for Check 1.
	alpha := run.GetRandomCoinFieldExt(v.alphaCoin.Name)
	nRowSize := 1 << v.nRow
	alphaPow := make([]fext.Element, nRowSize)
	alphaPow[0].SetOne()
	for b := 1; b < nRowSize; b++ {
		alphaPow[b].Mul(&alphaPow[b-1], &alpha)
	}

	for jIdx, colIdx := range v.entryList {
		// ── Check 2 on original: Merkle path ──────────────────────────────────
		colData := make([]field.Element, v.nbRowsOrig)
		for row := range colData {
			colData[row] = run.GetColumnAt(v.origOpenedDataCol.GetColID(), jIdx*v.nbRowsOrig+row)
		}
		h.WriteElements(colData...)
		leaf := h.SumElement()
		h.Reset()

		siblings := make([]types.KoalaOctuplet, v.depth)
		for d := range siblings {
			for i := range siblings[d] {
				siblings[d][i] = run.GetColumnAt(v.origPathsCol.GetColID(), jIdx*v.depth*8+d*8+i)
			}
		}
		proof := smt_koalabear.Proof{Path: colIdx, Siblings: siblings}
		if err := smt_koalabear.Verify(&proof, leaf, root); err != nil {
			return fmt.Errorf("ML Vortex Check 2 on original failed at column %d: %w", colIdx, err)
		}

		// ── Check 1: UAlpha = α-combination of original rows ──────────────────
		// The UAlpha opened data layout (nbRowsUA = 4*K):
		//   offset 0*K+k → B0.A0 of UAlpha[k]
		//   offset 1*K+k → B0.A1 of UAlpha[k]
		//   offset 2*K+k → B1.A0 of UAlpha[k]
		//   offset 3*K+k → B1.A1 of UAlpha[k]
		for k := 0; k < v.K; k++ {
			var expected fext.Element
			for b := 0; b < nRowSize; b++ {
				origVal := run.GetColumnAt(v.origOpenedDataCol.GetColID(),
					jIdx*v.nbRowsOrig+k*nRowSize+b)
				// α^b * origVal (base-field scalar → scale each fext component).
				var t fext.Element
				t.B0.A0.Mul(&alphaPow[b].B0.A0, &origVal)
				t.B0.A1.Mul(&alphaPow[b].B0.A1, &origVal)
				t.B1.A0.Mul(&alphaPow[b].B1.A0, &origVal)
				t.B1.A1.Mul(&alphaPow[b].B1.A1, &origVal)
				expected.Add(&expected, &t)
			}

			var actual fext.Element
			uIdx := v.uAlphaIndices[k]
			actual.B0.A0 = run.GetColumnAt(v.uAlphaOpenedDataCol.GetColID(), jIdx*v.nbRowsUA+0*v.kTotal+uIdx)
			actual.B0.A1 = run.GetColumnAt(v.uAlphaOpenedDataCol.GetColID(), jIdx*v.nbRowsUA+1*v.kTotal+uIdx)
			actual.B1.A0 = run.GetColumnAt(v.uAlphaOpenedDataCol.GetColID(), jIdx*v.nbRowsUA+2*v.kTotal+uIdx)
			actual.B1.A1 = run.GetColumnAt(v.uAlphaOpenedDataCol.GetColID(), jIdx*v.nbRowsUA+3*v.kTotal+uIdx)

			if !expected.Equal(&actual) {
				return fmt.Errorf("ML Vortex Check 1 failed: column k=%d, opened position %d", k, colIdx)
			}
		}
	}
	return nil
}

func (v *mlCheck1VerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("mlCheck1VerifierAction.RunGnark: not implemented")
}

func (v *mlCheck1VerifierAction) Skip()           {}
func (v *mlCheck1VerifierAction) IsSkipped() bool { return false }
