package iokeccakf

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type OutputKeccakF struct {
	hash    ifaces.Column
	isHash  ifaces.Column
	hashNum ifaces.Column

	stateCurr [5][5][8]ifaces.Column
	isBase2   ifaces.Column // indicate where the state should be interpreted as output
}

// it first applies to-basex to get laneX, then a projection query to map lanex to blocks
func NewOutputKeccakF(comp *wizard.CompiledIOP, inputs OutputKeccakF) {

	stateCols := make([][]ifaces.Column, 4*8)
	filterA := make([]ifaces.Column, 4*8)
	hashCol := make([][]ifaces.Column, 1)
	// extract the hash result from the state

	j := 0
	for j < len(stateCols) {
		for x := 0; x < 3; x++ {
			for z := 0; z < 8; z++ {
				stateCols[j] = []ifaces.Column{inputs.stateCurr[x][0][z]}
				filterA[j] = inputs.isBase2
				j++

			}
		}

	}

	hashCol[0] = []ifaces.Column{inputs.hash}

	comp.InsertProjection(ifaces.QueryIDf("OUTPUT_KECCAKF_HASH"),
		query.ProjectionMultiAryInput{
			ColumnsA: stateCols,
			ColumnsB: hashCol,
			FiltersA: filterA,
			FiltersB: []ifaces.Column{inputs.isHash},
		},
	)

}
