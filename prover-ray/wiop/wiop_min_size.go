package wiop

import "fmt"

// MinRuntimeSize returns the smallest size the module may take at runtime:
// the smallest power of two N with N > 2·max|s|, the maximum being over the
// shifts s at which the system's [LagrangeEval] queries open the module's
// columns (1 when there is no shift).
//
// Above that size, no two shifts of a column coincide modulo N, and their
// order modulo N (non-negative shifts ascending, then negative ones) is the
// same whatever N. The openings of the module, and so the proof, then have the
// same shape at every size, which a size-oblivious verifier of dynamic modules
// relies on (#4246). [System.Prove] and [System.Verify] reject a dynamic
// module below its minimum.
//
// The shifts are read from the Lagrange evaluations, the openings left after
// compilation, so the minimum is final once the system is compiled.
func (m *Module) MinRuntimeSize() int {
	return minSizeForShift(m.system.maxShifts()[m.index])
}

// MinRuntimeSizes returns the [Module.MinRuntimeSize] of every dynamic module
// of the system, by module index, computed in one pass over the Lagrange
// evaluations.
func (sys *System) MinRuntimeSizes() map[int]int {
	maxShifts := sys.maxShifts()
	res := make(map[int]int)
	for k, mod := range sys.Modules {
		if mod.isDynamic {
			res[k] = minSizeForShift(maxShifts[k])
		}
	}
	return res
}

// maxShifts returns, per module index, the largest |s| over the shifts s at
// which the system's Lagrange evaluations open that module's columns.
func (sys *System) maxShifts() map[int]int {
	res := make(map[int]int)
	for _, eval := range sys.LagrangeEvals {
		for _, view := range eval.Polynomials {
			s := view.ShiftingOffset
			if s < 0 {
				s = -s
			}
			idx := view.Column.Module.index
			res[idx] = max(res[idx], s)
		}
	}
	return res
}

// minSizeForShift returns the smallest power of two N > 2·maxShift.
func minSizeForShift(maxShift int) int {
	n := 1
	for n <= 2*maxShift {
		n *= 2
	}
	return n
}

// checkMinRuntimeSizes returns an error naming the first dynamic module (in
// module order) whose size in sizes is below its [Module.MinRuntimeSize].
// sizes maps module indices to sizes, as [Runtime] and [Proof] hold them; a
// module without a size is not checked.
func (sys *System) checkMinRuntimeSizes(sizes map[int]int) error {
	maxShifts := sys.maxShifts()
	for k, mod := range sys.Modules {
		if !mod.isDynamic {
			continue
		}
		size, ok := sizes[k]
		if !ok {
			continue
		}
		if minSize := minSizeForShift(maxShifts[k]); size < minSize {
			return fmt.Errorf(
				"wiop: dynamic module %q has size %d, below its minimum %d (shifts up to ±%d need N > %d)",
				mod.Context.Path(), size, minSize, maxShifts[k], 2*maxShifts[k],
			)
		}
	}
	return nil
}
