package ifaces

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/gnark/frontend"
)

// ID for a query
type QueryID string

// Utility function to format names
func QueryIDf(s string, args ...interface{}) QueryID {
	return QueryID(fmt.Sprintf(s, args...))
}

// A query that can be made by the verifier to the wizard oracle
type Query interface {
	Check(run Runtime) error
	CheckGnark(api frontend.API, run GnarkRuntime)
}

// Represents the parameters of a query
type QueryParams interface {
	// Update fiat-shamir with the query parameters
	UpdateFS(*fiatshamir.State)
}

// Gnark query params
type GnarkQueryParams interface {
	// Update fiat-shamir with the query parameters in a circuit
	UpdateFS(*fiatshamir.GnarkFiatShamir)
}
