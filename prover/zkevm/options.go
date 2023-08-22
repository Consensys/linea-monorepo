package zkevm

import (
	"github.com/consensys/accelerated-crypto-monorepo/glue"
)

/*
Options allows to customize the wrapper allowing to
*/
type Option func(*Builder)

/*
This option ensures that the commitment size is limited to a given value
*/
func ColDepthLimit(x int) Option {
	return func(b *Builder) {
		b.ColDepthLimit = x
	}
}

/*
This option ensures that the zkevm does not use more than n columns in the proving
*/
func NumColLimit(n int) Option {
	return func(b *Builder) {
		b.NumColLimit = n
	}
}

// TODO @alex move the keccak option in a sub-package of zkevm

// Max number of keccak to prove
const NUM_KECCAKF int = 1 << 13

// If passed to `WrapDefine`, this will enable keccak proving
func WithKeccak() Option {
	return func(b *Builder) {
		if b.keccak.enabled {
			panic("the keccak option has already been enabled")
		}
		b.keccak.enabled = true
	}
}

// Give the maximal number of signatures that we can accept
const NUM_ECDSA = 1 << 9

// WithECDSA allows validating ECDSA signatures. The maximum number of
// signatures is given by numSigs. If there are total signatures in a block
// (number of transcations + calls to ECRecover precompile) is larger than the
// capacity, then the prover panics.
//
// The transactions signatures are obtained from txExtractor. If it is nil, then
// the function is not called.
func WithECDSA(txExtractor glue.TxSignatureExtractor) Option {
	return func(b *Builder) {
		if b.ecdsa.numSigs != 0 {
			panic("ECDSA glue option already set")
		}
		b.ecdsa.numSigs = NUM_ECDSA
		b.ecdsa.txExtractor = txExtractor
	}
}

// Registers the statemanager options
func WithStateManager(b *Builder) {
	if b.statemanager.enabled {
		panic("the state manager option has already been enabled")
	}
	b.statemanager.enabled = true
}
