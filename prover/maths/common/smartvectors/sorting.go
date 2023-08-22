package smartvectors

import (
	"sort"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func Sort(v SmartVector) SmartVector {

	switch x := v.(type) {
	case *Constant:
		// Then, there is nothing to sort
		return v.DeepCopy()
	case *PaddedCircularWindow:
		// First we sort the window
		res := x.DeepCopy().(*PaddedCircularWindow)
		// First, sort the table
		table := fr.Vector(res.window)
		sort.Sort(table)
		// Find the smallest element larger or equal than paddingVal,
		// via binary search
		pos := sort.Search(len(table), func(i int) bool {
			return table[i].Cmp(&res.paddingVal) >= 0
		})

		switch pos {
		case 0:
			// The whole sorted window is larger than the padding value
			res.offset = res.Len() - len(table)
			res.window = table // Unsure this is necessary
			return res
		case len(table):
			// The whole table is smaller than the sorted value
			res.offset = 0
			res.window = table // Unsure this is necessary
			return res
		default:
			// The part at the left of pos in win is smaller than the padding
			// Thus, we rotate the window
			win := NewRegular(res.window)
			win = win.RotateRight(-pos).(*Regular)
			res.window = *win
			res.offset = res.Len() - pos
			return res
		}
	case *Regular:
		// Deep-copy the slice to prevent side effects
		resSlice := vector.DeepCopy(*x)
		table := fr.Vector(resSlice)
		sort.Sort(table)
		return NewRegular(table)
	default:
		utils.Panic("unsupported %t", v)
	}

	panic("unreachable")
}
