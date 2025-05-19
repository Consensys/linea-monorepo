package fft

import (
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
)

/*
Creates a domain without a coset
*/
func NewDomain(m int) *Domain {

	// gnark-crypto's NewDomain takes the domain cardinality directly.
	gnarkD := gnarkfft.NewDomain(uint64(m))
	return &Domain{GnarkDomain: gnarkD}
}
