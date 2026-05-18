package multilinearvortex

import (
	"sort"

	"github.com/consensys/linea-monorepo/prover-ray/wiop"
)

// sortedLayout is a layout where the columns are sorted by size. It implements
// the [sort.Interface]. This is a prover runtime structure and should not be
// included in the system in anyway (including as a field of prover actions).
//
// The layout is derived by sorting the columns in decreasing order. This yields
// the optimal packing of the columns in the Vortex matrix.
type sortedLayout struct {
	ColumnIDs   []wiop.ObjectID
	ColumnSizes []int
	// EmbeddingHypercubeSize is the total-size of the Q polynomial packing.
	// Must be a power of two.
	EmbeddingHypercubeSize int
	// locators[i] indicates where the column i is stored in the hypercube Q.
	//
	// The locator take the form of an unsigned integer. They can be converted
	// into a list of boolean values as follows:
	//
	//   - log2Q = log2(EmbeddingHypercubeSize)
	//   - locator = locator[i]
	//   - log2Size = size[i]
	//     locatorVariables = bitsOf(locator)[64-log2Q:log2Size]
	//
	// These do not need to be user-provided as they are computed internally
	locators []uint64
}

// Len implements [sort.Interface] and returns the number of columns.
func (s *sortedLayout) Len() int {
	return len(s.ColumnIDs)
}

// Less implements [sort.Interface] and returns whether column i is smaller than
// column j.
func (s *sortedLayout) Less(i, j int) bool {
	return s.ColumnSizes[i] < s.ColumnSizes[j]
}

// Swap implements the [sort.Interface] and swaps column i and j.
func (s *sortedLayout) Swap(i, j int) {
	s.ColumnIDs[i], s.ColumnIDs[j] = s.ColumnIDs[j], s.ColumnIDs[i]
	s.ColumnSizes[i], s.ColumnSizes[j] = s.ColumnSizes[j], s.ColumnSizes[i]
}

// Multilinear returns a list of columns, sizes and locators for the columns.
// The locator is a sequence of boolean values corresponding to the location
// of the sub-hypercube storing the column in the hypercube Q.
//

// The layout algorithm is implemented by sorting the columns by decreasing size
// , the locator is computed as the cumulative sum sequence of the sizes.
func (s *sortedLayout) Resolves() (
	sortedIDs []wiop.ObjectID,
	sortedSizes []int,
	locators []uint64,
) {

	sort.Stable(sort.Reverse(s))

	s.locators = make([]uint64, len(s.ColumnIDs))
	currLocator := uint64(0)

	for i := range s.ColumnIDs {
		s.locators[i] = currLocator
		currLocator += uint64(s.ColumnSizes[i])
	}

	return s.ColumnIDs, s.ColumnSizes, nil
}
