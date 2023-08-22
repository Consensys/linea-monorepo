package glue

import "github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/plonk"

type option struct {
	PlonkOpts []plonk.Option
	BatchSize int
}

// Option allows changing glue implementation parameters
type Option func(*option)

// WithBatchSize splits larger circuit into smaller circuits to bound the number
// of constraints.
func WithBatchSize(size int) Option {
	return func(o *option) {
		o.BatchSize = size
	}
}

// WithPlonkOption defines a PLONK prover option to be passed.
func WithPlonkOption(opt plonk.Option) Option {
	return func(o *option) {
		o.PlonkOpts = append(o.PlonkOpts, opt)
	}
}
