package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

// EncodedMatrix represents the witness of a Vortex matrix commitment, it is
// represented as an array of rows.
type EncodedMatrix []smartvectors.SmartVector

// Commit to a sequence of columns and Merkle hash on top of that. Returns the
// tree and an array containing the concatenated columns hashes. The final
// short commitment can be obtained from the returned tree as:
//
//	tree.Root()
//
// And can be safely converted to a field Element via
// [field.Element.SetBytesCanonical]
func (p *Params) CommitMerkle(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	logrus.Infof("Vortex compiler: RS encoding nrows=%v of ncol=%v to codeword-size=%v", len(ps), p.NbColumns, p.NbColumns*p.BlowUpFactor)

	numRows := len(ps)
	numCols := p.NbColumns

	input := make([][]field.Element, numRows)
	parallel.Execute(numRows, func(start, end int) {
		for i := start; i < end; i++ {
			input[i] = make([]field.Element, numCols)
			for j := 0; j < numCols; j++ {
				input[i][j] = ps[i].Get(j)
			}
		}
	})

	// In Commit phase, it's not used, so set to 0 as a placeholder. numSelectedColumns is only used in the Open phase
	numSelectedColumns := 0
	params, err := vortex.NewParams(p.NbColumns, p.MaxNbRows, p.Key.GnarkInternal, p.BlowUpFactor, numSelectedColumns)
	if err != nil {
		utils.Panic(err.Error())
	}
	logrus.Infof("Vortex compiler: Commit START")
	proverState, err := vortex.Commit(params, input)
	if err != nil {
		utils.Panic(err.Error())
	}

	print("Vortex compiler: Commit DONE\n", proverState)

	// Format Encoded Matrix for Return
	encodedMatrix = make([]smartvectors.SmartVector, p.MaxNbRows)
	for i := range encodedMatrix {
		encodedMatrix[i] = smartvectors.NewRegular(proverState.EncodedMatrix[i*p.NbColumns : (i+1)*p.NbColumns])
	}

	//TODO: check if we need to consider the case where noSisTransversalHash is called
	return encodedMatrix, proverState.MerkleTree, proverState.SisHashes
}
