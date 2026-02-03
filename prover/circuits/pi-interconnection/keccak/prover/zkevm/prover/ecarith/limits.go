package ecarith

import "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"

// Limits defines the upper limits on the size of the circuit and the number of
// gnark circuits. The total number of allowed EC_MUL precompile calls is
// product of the fields.
type Limits struct {
	// how many ecmul can we do in a single circuit
	NbInputInstances int
	// how many circuit instances can we have
	NbCircuitInstances int
}

func (l *Limits) sizeEcMulIntegration() int {
	return utils.NextPowerOfTwo(l.NbInputInstances*nbRowsPerEcMul) * utils.NextPowerOfTwo(l.NbCircuitInstances)
}

func (l *Limits) sizeEcAddIntegration() int {
	return utils.NextPowerOfTwo(l.NbInputInstances*nbRowsPerEcAdd) * utils.NextPowerOfTwo(l.NbCircuitInstances)
}
