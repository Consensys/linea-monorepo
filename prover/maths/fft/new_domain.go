package fft

import (
	gnarkfft "github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/utils"
)

/*
Creates a domain without a coset
*/
func NewDomain(m int) *Domain {
	// gnark-crypto's NewDomain takes the domain cardinality directly.
	gnarkD := gnarkfft.NewDomain(uint64(m))
	return &Domain{gnarkDomain: gnarkD}
}

/*
Equip the current domain with a coset shifted by the multiplicative generator
*/
func (dom *Domain) WithCoset() *Domain {
	return dom.WithCustomCoset(1, 0)
}

/*
Equipe the current domain with a custom coset obtained as explained in
the doc of `GetCoset`
*/
func (dom *Domain) WithCustomCoset(r, numcoset int) *Domain {
	n := utils.ToInt(dom.gnarkDomain.Cardinality)
	dom.gnarkDomain.CosetTable(),
		dom.gnarkDomain.cosetTableInv,
		dom.CosetTableReversed,
		dom.CosetTableInvReversed = GetCoset(n, r, numcoset)

	return dom
}
