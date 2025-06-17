package bls

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func NewG1AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G1,
		limits,
		newAddDataSource(comp, G1),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}

func NewG2AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(
		comp,
		G2,
		limits,
		newAddDataSource(comp, G2),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}
