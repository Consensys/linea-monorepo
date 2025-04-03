package smartvectors

import (
	"iter"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// IterateSmart returns an iterator over the elements of the smartvectors,
// the iterator will iterate depending on the type of smartvectors:
//
//   - Regular, Regular, Rotated: will iterate over the elements of the
//     vector as it would be expected by a normal iterator.
//   - Constant: the iterator will only return a single value.
//   - PaddedCircularWindow (left-padded): the iterator will first return
//     the filling value and then the elements of the windows.
//   - PaddedCircularWindow (right-padded): the iterator will first return
//     the elements of the window and then the filling value
//   - PaddedCircularWindow (bi-directionally-padded): the iterator will return
//     (1) one element for the padding value, (2) the elements of the window
//     and (3) one element for the padding value again
func IterateSmart(sv SmartVector) iter.Seq[field.Element]
