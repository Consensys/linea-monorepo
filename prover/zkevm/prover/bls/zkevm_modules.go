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

func NewG1MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G1,
		limits,
		newMsmDataSource(comp, G1),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}

func NewG2MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(
		comp,
		G2,
		limits,
		newMsmDataSource(comp, G2),
		[]query.PlonkOption{query.PlonkRangeCheckOption(16, 6, true)},
	)
}
