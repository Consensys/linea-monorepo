// CPU fallback for Commit/Prove/LinComb/ExtractColumns when CUDA is not available.

//go:build !cuda

package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
)

// CommitState holds prover state after commitment.
// On the CPU build, it wraps either a gnark-crypto ProverState (for Commit)
// or an encodedMatrix (for CommitSIS).
type CommitState struct {
	inner         *vortex.ProverState
	encodedMatrix []smartvectors.SmartVector
	nRows         int
}

// NRows returns the number of rows in this commit.
func (cs *CommitState) NRows() int { return cs.nRows }

// FreeGPU is a no-op on CPU builds.
func (cs *CommitState) FreeGPU() {}

// IsDeviceResident reports whether the encoded matrix is resident on device.
// CPU builds never keep Vortex state on device.
func (cs *CommitState) IsDeviceResident() bool { return false }

// Commit encodes the input matrix using Reed-Solomon, hashes columns
// via SIS + Poseidon2, builds a Merkle tree, and returns the commitment root.
func (p *Params) Commit(rows [][]koalabear.Element) (*CommitState, Hash, error) {
	ps, err := vortex.Commit(p.inner, rows)
	if err != nil {
		return nil, Hash{}, err
	}
	return &CommitState{inner: ps, nRows: len(rows)}, ps.GetCommitment(), nil
}

// Prove generates an opening proof for the committed matrix.
func (cs *CommitState) Prove(alpha fext.E4, selectedCols []int) (*Proof, error) {
	cs.inner.OpenLinComb(alpha)
	vp, err := cs.inner.OpenColumns(selectedCols)
	if err != nil {
		return nil, err
	}
	return &Proof{
		UAlpha:       vp.UAlpha,
		Columns:      vp.OpenedColumns,
		MerkleProofs: vp.MerkleProofOpenedColumns,
	}, nil
}

// LinComb computes UAlpha[j] = Σᵢ αⁱ · rows[i].Get(j) on CPU.
func (cs *CommitState) LinComb(alpha fext.E4) ([]fext.E4, error) {
	if cs.encodedMatrix == nil {
		panic("vortex: CommitState has no encodedMatrix for CPU LinComb")
	}
	n := cs.encodedMatrix[0].Len()
	result := make([]fext.E4, n)
	var pow fext.E4
	pow.SetOne()
	for _, row := range cs.encodedMatrix {
		for j := range n {
			v := row.Get(j)
			var term fext.E4
			term.B0.A0 = v
			term.Mul(&term, &pow)
			result[j].Add(&result[j], &term)
		}
		pow.Mul(&pow, &alpha)
	}
	return result, nil
}

// ExtractColumns gathers selected columns from host-side SmartVectors.
func (cs *CommitState) ExtractColumns(selectedCols []int) ([][]koalabear.Element, error) {
	if cs.encodedMatrix == nil {
		panic("vortex: CommitState has no encodedMatrix for CPU ExtractColumns")
	}
	columns := make([][]koalabear.Element, len(selectedCols))
	for i, c := range selectedCols {
		col := make([]koalabear.Element, len(cs.encodedMatrix))
		for r, row := range cs.encodedMatrix {
			col[r] = row.Get(c)
		}
		columns[i] = col
	}
	return columns, nil
}

// GetEncodedMatrix returns the host-side encoded matrix as SmartVectors.
func (cs *CommitState) GetEncodedMatrix() []smartvectors.SmartVector {
	return cs.encodedMatrix
}

func (cs *CommitState) ExtractAllRows() ([][]koalabear.Element, error) {
	panic("gpu: cuda required")
}

func (cs *CommitState) MerkleTree() any {
	panic("gpu: cuda required")
}

func (cs *CommitState) ExtractSISHashes() ([]koalabear.Element, error) {
	panic("gpu: cuda required")
}

func (cs *CommitState) ExtractLeaves() ([]Hash, error) {
	panic("gpu: cuda required")
}
