package smartvectors

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/sirupsen/logrus"
)

func assertCorrectBound(pos, length int) {
	if pos >= length {
		logrus.Panicf("Bound assertion failed, cannot access pos %v for vector of length %v", pos, length)
	}
}

func assertHasLength(a, b int) {
	if a != b {
		utils.Panic("the two slices should have the same length (found %v and %v)", a, b)
	}
}

func assertPowerOfTwoLen(l int) {
	if !utils.IsPowerOfTwo(l) {
		logrus.Panicf("Slice should have a power of two length but has %v", l)
	}
}

func assertStrictPositiveLen(l int) {

	if l == 0 {
		logrus.Panicf("FORBIDDEN : Got a null length vector")
	}

	if l <= 0 {
		logrus.Panicf("FORBIDDEN : Got a negative length %v", l)
	}
}

// Prints a summary of the vector
func PrettyShort(sv SmartVector) string {
	x, y, z := sv.Get(0), sv.Get(1), sv.Get(2)
	return fmt.Sprintf("(%T, size=%v) [%v, %v, %v ..]", sv, sv.Len(), x.String(), y.String(), z.String())
}
