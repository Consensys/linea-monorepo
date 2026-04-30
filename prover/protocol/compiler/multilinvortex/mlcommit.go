package multilinvortex

import (
	"fmt"
	"math/bits"
	"strings"

	"github.com/consensys/gnark/frontend"
	smt_koalabear "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	vortex_common "github.com/consensys/linea-monorepo/prover/crypto/vortex"
	vortex_koalabear "github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	poseidon2 "github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const (
	mlCommitExtraKey        = "mlvortex_committed"
	mlCommitEntriesKey      = "mlvortex_commit_entries"
	mlDefaultNumOpen        = 128
)

// mlCommitEntry records the proof columns produced by CommitMLColumns for one
// group of UALPHA/ROWEVAL columns.  CommitOriginalMLColumns reads these to
// locate the UAlpha opened-data column for Check 1.
type mlCommitEntry struct {
	cols          []ifaces.Column // the UALPHA or ROWEVAL columns that were committed
	entryList     []int
	nbRows        int
	depth         int
	openedDataCol ifaces.Column
	pathsCol      ifaces.Column
	rootCol       ifaces.Column
}

// CommitMLColumns is a compile pass to call immediately after each [Compile].
// It finds the MLVORTEX_UALPHA_* / MLVORTEX_ROWEVAL_* proof columns added by
// that Compile call and, for each group:
//
//   - Inserts proof columns for the Merkle root, the t opened column vectors,
//     and the t Merkle sibling paths.
//   - Registers a prover action (Phase 1: commit; Phase 2: open) that assigns
//     all proof columns.
//   - Registers a verifier action (Check 2) that re-derives the Poseidon2 leaf
//     hash for each opened column and verifies its Merkle path against the root.
//
// The default number of opened columns is 128; use [CommitMLColumnsWithOpening]
// to override.  The function is idempotent: repeated calls only process newly
// added columns.
func CommitMLColumns(comp *wizard.CompiledIOP) {
	commitMLColumnsImpl(comp, mlDefaultNumOpen)
}

// CommitMLColumnsWithOpening returns a compile pass like [CommitMLColumns] but
// with a configurable number of opened columns per committed group.
func CommitMLColumnsWithOpening(numOpen int) func(*wizard.CompiledIOP) {
	return func(comp *wizard.CompiledIOP) {
		commitMLColumnsImpl(comp, numOpen)
	}
}

func commitMLColumnsImpl(comp *wizard.CompiledIOP, numOpen int) {
	if comp.ExtraData == nil {
		comp.ExtraData = make(map[string]any)
	}
	seen, _ := comp.ExtraData[mlCommitExtraKey].(map[ifaces.ColID]bool)
	if seen == nil {
		seen = make(map[ifaces.ColID]bool)
		comp.ExtraData[mlCommitExtraKey] = seen
	}

	type groupKey struct {
		round int
		size  int
	}
	groups := make(map[groupKey][]ifaces.Column)

	for _, name := range comp.Columns.AllKeysProof() {
		if seen[name] {
			continue
		}
		sname := string(name)
		if !strings.HasPrefix(sname, "MLVORTEX_UALPHA_") &&
			!strings.HasPrefix(sname, "MLVORTEX_ROWEVAL_") {
			continue
		}
		col := comp.Columns.GetHandle(name)
		key := groupKey{round: col.Round(), size: col.Size()}
		groups[key] = append(groups[key], col)
		seen[name] = true
	}

	for key, cols := range groups {
		K := len(cols)
		nbRows := 4 * K // 4 base-field components per fext column
		nbCodewordCols := key.size * 2   // RS rate = 2
		t := min(numOpen, nbCodewordCols)
		// depth = log2(nbCodewordCols); nbCodewordCols is always a power of 2
		depth := bits.Len(uint(nbCodewordCols)) - 1
		entryList := evenlySpacedIndices(nbCodewordCols, t)

		// Each fext column (degree-4 over KoalaBear) decomposes into 4 base-field
		// polynomials of length key.size.
		// Rate=2, SIS params required by NewParams but unused by CommitMerkleWithoutSIS.
		params := vortex_koalabear.NewParams(2, key.size, nbRows, 6, 16)

		// Unique suffix for proof column IDs: round + fext size.
		sfx := fmt.Sprintf("r%d_s%d", key.round, key.size)

		// Proof column for the Merkle root: 1 field.Octuplet = 8 field.Elements.
		rootCol := comp.InsertProof(key.round,
			ifaces.ColID("MLVORTEX_ROOT_"+sfx), 8, false)

		// Proof column for all t opened column vectors (each has nbRows elements).
		openedSize := utils.NextPowerOfTwo(t * nbRows)
		openedDataCol := comp.InsertProof(key.round,
			ifaces.ColID("MLVORTEX_OPENED_"+sfx), openedSize, false)

		// Proof column for all t Merkle sibling paths (each has depth octuplets
		// = depth*8 field elements).
		pathsSize := utils.NextPowerOfTwo(t * depth * 8)
		pathsCol := comp.InsertProof(key.round,
			ifaces.ColID("MLVORTEX_PATHS_"+sfx), pathsSize, false)

		comp.RegisterProverAction(key.round, &mlCommitProverAction{
			cols:          cols,
			params:        &params,
			entryList:     entryList,
			nbRows:        nbRows,
			depth:         depth,
			openedSize:    openedSize,
			pathsSize:     pathsSize,
			rootCol:       rootCol,
			openedDataCol: openedDataCol,
			pathsCol:      pathsCol,
		})

		comp.RegisterVerifierAction(comp.NumRounds()-1, &mlCheck2VerifierAction{
			rootCol:       rootCol,
			openedDataCol: openedDataCol,
			pathsCol:      pathsCol,
			nbRows:        nbRows,
			entryList:     entryList,
			depth:         depth,
		})

		entry := mlCommitEntry{
			cols:          cols,
			entryList:     entryList,
			nbRows:        nbRows,
			depth:         depth,
			openedDataCol: openedDataCol,
			pathsCol:      pathsCol,
			rootCol:       rootCol,
		}
		entries, _ := comp.ExtraData[mlCommitEntriesKey].([]mlCommitEntry)
		comp.ExtraData[mlCommitEntriesKey] = append(entries, entry)
	}
}

// mlCommitProverAction performs the two phases of the ML Vortex prover for
// one group of same-size fext columns:
//
//  1. Decompose, RS-encode, Poseidon2-hash columns, build Merkle tree.
//  2. Select t opened positions, extract column vectors + sibling paths,
//     and assign them to the proof columns declared at compile time.
type mlCommitProverAction struct {
	cols      []ifaces.Column
	params    *vortex_koalabear.Params
	entryList []int
	nbRows    int
	depth     int
	openedSize int
	pathsSize  int
	rootCol       ifaces.Column
	openedDataCol ifaces.Column
	pathsCol      ifaces.Column
}

func (a *mlCommitProverAction) Run(run *wizard.ProverRuntime) {
	K := len(a.cols)
	size := a.cols[0].Size()

	// ── Phase 1: decompose fext → 4K base-field polynomials ──────────────────
	pols := make([]smartvectors.SmartVector, 4*K)
	parallel.Execute(K, func(start, stop int) {
		for k := start; k < stop; k++ {
			sv := smartvectors.IntoRegVecExt(run.GetColumn(a.cols[k].GetColID()))
			b0a0 := make([]field.Element, size)
			b0a1 := make([]field.Element, size)
			b1a0 := make([]field.Element, size)
			b1a1 := make([]field.Element, size)
			for i, e := range sv {
				b0a0[i] = e.B0.A0
				b0a1[i] = e.B0.A1
				b1a0[i] = e.B1.A0
				b1a1[i] = e.B1.A1
			}
			pols[k] = smartvectors.NewRegular(b0a0)
			pols[K+k] = smartvectors.NewRegular(b0a1)
			pols[2*K+k] = smartvectors.NewRegular(b1a0)
			pols[3*K+k] = smartvectors.NewRegular(b1a1)
		}
	})

	// RS-encode all 4K rows, Poseidon2-hash each codeword column, build Merkle tree.
	encodedMatrix, _, tree, _ := a.params.CommitMerkleWithoutSIS(pols)

	// ── Assign rootCol ────────────────────────────────────────────────────────
	rootOct := tree.Root
	rootVec := make([]field.Element, 8)
	for i, e := range rootOct {
		rootVec[i] = e
	}
	run.AssignColumn(a.rootCol.GetColID(), smartvectors.NewRegular(rootVec))

	// ── Phase 2: open t columns ───────────────────────────────────────────────
	// SelectColumnsAndMerkleProofs extracts the t column vectors from the
	// encoded matrix and generates a Merkle sibling path for each.
	dummy := &vortex_common.OpeningProof{}
	merkleProofs := vortex_koalabear.SelectColumnsAndMerkleProofs(
		dummy,
		a.entryList,
		[]vortex_koalabear.EncodedMatrix{encodedMatrix},
		[]*smt_koalabear.Tree{tree},
	)

	// Pack opened column vectors into openedDataCol.
	// Layout: [col0_row0, col0_row1, ..., col0_rowN-1, col1_row0, ...].
	openedVec := make([]field.Element, a.openedSize)
	for j, colVec := range dummy.Columns[0] {
		for row, val := range colVec {
			openedVec[j*a.nbRows+row] = val
		}
	}
	run.AssignColumn(a.openedDataCol.GetColID(), smartvectors.NewRegular(openedVec))

	// Pack Merkle sibling paths into pathsCol.
	// Layout: [col0_depth0_oct0..7, col0_depth1_oct0..7, ..., col1_depth0_...].
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

// mlCheck2VerifierAction implements Check 2 of the ML Vortex protocol:
// for each opened column position, re-derive the Poseidon2 leaf hash from the
// provided column data and verify the Merkle sibling path against the root.
type mlCheck2VerifierAction struct {
	rootCol       ifaces.Column
	openedDataCol ifaces.Column
	pathsCol      ifaces.Column
	nbRows        int
	entryList     []int
	depth         int
}

func (v *mlCheck2VerifierAction) Run(run wizard.Runtime) error {
	// Read the Merkle root (8 field elements → 1 field.Octuplet).
	var root field.Octuplet
	for i := range root {
		root[i] = run.GetColumnAt(v.rootCol.GetColID(), i)
	}

	h := poseidon2.NewMDHasher()

	for j, colIdx := range v.entryList {
		// Re-derive the Poseidon2 leaf hash from the opened column data.
		// The prover hashes the column using the same MDHasher fallback used by
		// noSisTransversalHash when nbRows is not divisible by 8. For consistency,
		// we always use MDHasher here; the hash is identical to the SIMD path.
		colData := make([]field.Element, v.nbRows)
		for row := range colData {
			colData[row] = run.GetColumnAt(v.openedDataCol.GetColID(), j*v.nbRows+row)
		}
		h.WriteElements(colData...)
		leaf := h.SumElement()
		h.Reset()

		// Reconstruct the Merkle proof from pathsCol.
		siblings := make([]types.KoalaOctuplet, v.depth)
		for d := range siblings {
			for i := range siblings[d] {
				siblings[d][i] = run.GetColumnAt(v.pathsCol.GetColID(), j*v.depth*8+d*8+i)
			}
		}
		proof := smt_koalabear.Proof{Path: colIdx, Siblings: siblings}

		if err := smt_koalabear.Verify(&proof, leaf, root); err != nil {
			return fmt.Errorf("ML Vortex Check 2 failed at opened column %d: %w", colIdx, err)
		}
	}
	return nil
}

func (v *mlCheck2VerifierAction) RunGnark(_ frontend.API, _ wizard.GnarkRuntime) {
	panic("mlCheck2VerifierAction.RunGnark: not implemented")
}

func (v *mlCheck2VerifierAction) Skip()           {}
func (v *mlCheck2VerifierAction) IsSkipped() bool { return false }

// evenlySpacedIndices returns t evenly-spaced indices in [0, n).
func evenlySpacedIndices(n, t int) []int {
	if t >= n {
		idx := make([]int, n)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	idx := make([]int, t)
	for i := range idx {
		idx[i] = i * n / t
	}
	return idx
}
