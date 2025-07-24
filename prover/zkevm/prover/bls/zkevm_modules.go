package bls

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func NewG1AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(comp, G1, limits, newAddDataSource(comp, G1))
}

func NewG2AddZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsAdd {
	return newAdd(comp, G2, limits, newAddDataSource(comp, G2))
}

func NewG1MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(comp, G1, limits, newMsmDataSource(comp, G1))
}

func NewG2MsmZkEvm(comp *wizard.CompiledIOP, limits *Limits) *BlsMsm {
	return newMsm(comp, G2, limits, newMsmDataSource(comp, G2))
}

func NewPairing(comp *wizard.CompiledIOP, limits *Limits) *BlsPair {
	return newPair(comp, limits, newPairDataSource(comp))
}
