package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// assertInBound panics if pos >= length
func assertInBound(pos, length int) {
	if pos < 0 || pos >= length {
		logrus.Panicf("Bound assertion failed, cannot access pos %v for vector of length %v", pos, length)
	}
}

// assertValidLen panics if l is not a power of two
func assertValidLen(l int) {
	if l <= 0 {
		utils.Panic("Slice should have a strictly positive length but has %v", l)
	}
	if !utils.IsPowerOfTwo(l) {
		utils.Panic("Slice should have a power of two length but has %v", l)
	}
}

// assertValidRange checks that 0 <= start < stop
func assertValidRange(start, stop int) {
	if start < 0 || start >= stop {
		utils.Panic("Invalid range: [%v, %v]", start, stop)
	}
}
