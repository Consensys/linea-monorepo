package circuits

import (
	"context"

	"github.com/consensys/gnark-crypto/kzg"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/test/unsafekzg"
)

type SRSProvider interface {
	GetSRS(ctx context.Context, ccs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error)
}

type UnsafeSRSProvider struct {
}

// NewUnsafeSRSProvider returns a new UnsafeSRSProvider
// if tau is provided, it will be used as the tau value for the SRS (slow path)
// otherwise, a random tau will be generated (fast path)
func NewUnsafeSRSProvider() UnsafeSRSProvider {
	return UnsafeSRSProvider{}
}

func (u UnsafeSRSProvider) GetSRS(ctx context.Context, ccs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error) {
	return unsafekzg.NewSRS(ccs, unsafekzg.WithFSCache())
}
