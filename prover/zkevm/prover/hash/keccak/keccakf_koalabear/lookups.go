package keccakfkoalabear

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

type lookupTables struct {
	// ccBase8Theta is the 9 digit tuple representation of cc used in
	// the theta step of keccakf. it contains (a0 + a1*BaseX + a2*BaseX^2
	// +...+ a8*BaseX^8). Here, each ai belongs to [0,5]. For theta X=8.
	ccBase8Theta ifaces.Column
	// ccBitConvertedTheta is the bit converted version of cc used in
	// the theta step of keccakf.
	ccBitConvertedTheta ifaces.Column
}
