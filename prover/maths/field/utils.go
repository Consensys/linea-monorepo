package field

import (
	"math"
)

func ToInt(e *Element) int {
	n := e.Uint64()
	if !e.IsUint64() || n > math.MaxInt {
		panic("out of range")
	}
	return int(n) // #nosec G115 -- Checked for overflow
}
