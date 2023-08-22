package ringsis

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Standard parameter that we use for ring-SIS
var StdParams = Params{LogTwoBound_: 4, LogTwoDegree: 4}

// Params encapsulates the parameters of a ring SIS instance
type Params struct {
	LogTwoBound_, LogTwoDegree int
}

// Number of limbs to represent a field element with the current representation
func (p *Params) NumLimbs() int {
	return utils.DivCeil(field.Bits, p.LogTwoBound_)
}

// Returns the number of field element to characterize the output
func (p *Params) OutputSize() int {
	return 1 << p.LogTwoDegree
}

// Returns the number of polynomial used to encode a field
// will panic if smaller than one. (Used for self-recursion
// and we have this requirement anyway.)
func (p *Params) NumPolyPerField() int {
	res := p.NumLimbs() / (1 << p.LogTwoDegree)
	if res == 0 {
		panic("The lattice instance needs less than a poly per instance")
	}
	return res
}
