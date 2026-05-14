//go:build !cuda

package globalcs

import "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"

func tryGPUQuotientReevalCoset(
	_ int,
	_ field.Element,
	_ [][]field.Element,
	_ [][]field.Element,
) bool {
	return false
}
