package smartvectors

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// It's a slice - zero padded up to a certain length - and rotated
type PaddedCircularWindow struct {
	window     []field.Element
	paddingVal field.Element
	// Totlen is the length of the represented vector
	totLen, offset int
}

// Create a new padded circular window vector
func NewPaddedCircularWindow(window []field.Element, paddingVal field.Element, offset, totLen int) *PaddedCircularWindow {
	// The window should not be larger than the total length
	if len(window) > totLen {
		utils.Panic("The window size is too large %v because totlen is %v", len(window), totLen)
	}

	if len(window) == totLen {
		utils.Panic("Forbidden : the window should not take the full length")
	}

	if len(window) == 0 {
		utils.Panic("Forbidden : empty window")
	}

	// Normalize the offset to be in range [0:totlen)
	offset = utils.PositiveMod(offset, totLen)
	return &PaddedCircularWindow{
		window:     window,
		paddingVal: paddingVal,
		offset:     offset,
		totLen:     totLen,
	}
}

// Returns the length of the vector
func (p *PaddedCircularWindow) Len() int {
	return p.totLen
}

// Returns a queries position
func (p *PaddedCircularWindow) GetBase(n int) (field.Element, error) {
	// Check if the queried index is in the window
	posFromWindowsPoV := utils.PositiveMod(n-p.offset, p.totLen)
	if posFromWindowsPoV < len(p.window) {
		return p.window[posFromWindowsPoV], nil
	}
	// Else, return the padding value
	return p.paddingVal, nil
}

func (p *PaddedCircularWindow) GetExt(n int) fext.Element {
	elem, _ := p.GetBase(n)
	return *new(fext.Element).SetFromBase(&elem)
}

func (r *PaddedCircularWindow) Get(n int) field.Element {
	res, err := r.GetBase(n)
	if err != nil {
		panic(err)
	}
	return res
}

// Extract a subvector from p[Start:Stop), the subvector cannot "roll-over".
// i.e, we enforce that Start < Stop
func (p *PaddedCircularWindow) SubVector(start, stop int) SmartVector {
	// negative Start value is not allowed
	if start < 0 {
		panic("negative Start value is not allowed")
	}
	// Sanity checks for all subvectors
	assertCorrectBound(start, p.totLen)
	// The +1 is because we accept if "Stop = length"
	assertCorrectBound(stop, p.totLen+1)

	if start > stop {
		panic("rollover are forbidden")
	}

	if start == stop {
		panic("zero length subvector is forbidden")
	}

	/*
		This function has a high-combinatoric complexity and in order to reason about
		each case, we represent them as follows:
			[a, b) is the interval of the subvector. We can assume that [a, b) does not
			roll-over the vector.
			[c,d) is the interval of the window of `p`. It can roll-over.

		We use 'a' as the origin for other coordinates and reduce
		the ongoing combinatoric when listing all the cases. Let b_ = b-a

			0xxxxxxxxxxxxxxxxxxxxxb                      N
			|                     |                      |
		1)	|                     |   c------------d     |
		2)	|-------------------------d     c------------| (including b == d)
		2*)	c-------------------------d                  | (including b == d)
		3)	|       c-------d     |                      |
		4)	|       c------------------------d           | (including b == d)
		5)	|-----d         c----------------------------|
		6)	|-------d             |           c----------| (including c == b)
		7)	d                     |           c----------| (including c == b)


		For consistency, with the above picture, we rename as the offset coordinates
	*/

	n := p.Len()
	b := stop - start
	c := normalize(p.interval().Start(), start, n)
	d := normalize(p.interval().Stop(), start, n)

	// Case 1 : return a constant vector
	if b <= c && c < d {
		return NewConstant(p.paddingVal, b)
	}

	// Case 2 : return a regular vector
	if b <= d && d < c {
		reg := Regular(p.window[n-c : n-c+b])
		return &reg
	}

	// Case 2* : same as 2 but c == 0
	if b <= d && c == 0 {
		reg := Regular(p.window[:b])
		return &reg
	}

	// Case 3 : the window is fully contained in the subvector
	if c < d && d <= b {
		return NewPaddedCircularWindow(p.window, p.paddingVal, c, b)
	}

	// Case 4 : left-ended
	if c < b && c <= d {
		return NewPaddedCircularWindow(p.window[:b-c], p.paddingVal, c, b)
	}

	// Case 5 : the window is double ended (we skip some element in the center of the window)
	if d < c && c < b {
		left := p.window[:b-c]
		right := p.window[n-c:]

		// The deep-copy of left ensures that we do not append
		// on the same concrete slice.
		w := append(vector.DeepCopy(left), right...)
		return NewPaddedCircularWindow(w, p.paddingVal, c, b)
	}

	// Case 6 : right-ended
	if 0 < d && d < b && b <= c {
		return NewPaddedCircularWindow(p.window[n-c:], p.paddingVal, 0, b)
	}

	// Case 7 : d == 0 and c is out
	if d == 0 && b <= c {
		return NewConstant(p.paddingVal, b)
	}

	panic(fmt.Sprintf("unsupported case : b %v, c %v, d %v", b, c, d))

}

// Rotate the vector
func (p *PaddedCircularWindow) RotateRight(offset int) SmartVector {
	return NewPaddedCircularWindow(vector.DeepCopy(p.window), p.paddingVal, p.offset+offset, p.totLen)
}

func (p *PaddedCircularWindow) WriteInSlice(buff []field.Element) {
	assertHasLength(len(buff), p.totLen)

	for i := range p.window {
		pos := utils.PositiveMod(i+p.offset, p.totLen)
		buff[pos] = p.window[i]
	}

	for i := len(p.window); i < p.totLen; i++ {
		pos := utils.PositiveMod(i+p.offset, p.totLen)
		buff[pos] = p.paddingVal
	}
}

func (p *PaddedCircularWindow) WriteInSliceExt(buff []fext.Element) {
	temp := make([]field.Element, len(buff))
	p.WriteInSlice(temp)
	for i := 0; i < len(buff); i++ {
		elem := temp[i]
		buff[i].SetFromBase(&elem)
	}

}

func (p *PaddedCircularWindow) Pretty() string {
	return fmt.Sprintf("Windowed[totlen=%v offset=%v, paddingVal=%v, window=%v]", p.totLen, p.offset, p.paddingVal.String(), vector.Prettify(p.window))
}

func (p *PaddedCircularWindow) interval() CircularInterval {
	return IvalWithStartLen(p.offset, len(p.window), p.totLen)
}

// normalize converts the (circle) coordinator x to another coordinate by changing
// the origin point on the discret circle. mod denotes the number of points in
// the circle.
func normalize(x, newRef, mod int) int {
	return utils.PositiveMod(x-newRef, mod)
}

// processWindowedOnly applies the operator `op` to all the smartvectors
// contained in `svecs` with `coeffs` that have the type [PaddedCircularWindow]
//
// The function does so by attempting to fit result on the smallest possible
// window.
//
// In case, this is not possible. The function will "give up" and convert all
// the paddedCircularWindow into Regulars and pretend it did not find any.
//
// The function returns the partial result of the operation and the number of
// padded circular windows SmartVector that it found.
func processWindowedOnly(op operator, svecs []SmartVector, coeffs_ []int) (res SmartVector, numMatches int) {

	// First we compute the union windows.
	length := svecs[0].Len()
	windows := []PaddedCircularWindow{}
	intervals := []CircularInterval{}
	coeffs := []int{}

	// Gather all the windows into a slice
	for i, svec := range svecs {
		if pcw, ok := svec.(*PaddedCircularWindow); ok {
			windows = append(windows, *pcw)
			intervals = append(intervals, pcw.interval())
			coeffs = append(coeffs, coeffs_[i]) // collect the coeffs related to each window
			// Sanity-check : all vectors must have the same length
			assertHasLength(svec.Len(), length)
			numMatches++
		}
	}

	if numMatches == 0 {
		return nil, numMatches
	}

	// has the dimension of the cover with garbage values in it
	smallestCover := SmallestCoverInterval(intervals)

	// Edge-case: in case the smallest-cover of the pcw found in svecs is the
	// full-circle the code below will not work as it assumes that is possible
	if smallestCover.IsFullCircle() {
		for i, svec := range svecs {
			if _, ok := svec.(*PaddedCircularWindow); ok {
				temp, _ := svec.IntoRegVecSaveAllocBase()
				svecs[i] = NewRegular(temp)
			}
		}
		return nil, 0
	}

	// Sanity-check : normally all offset are normalized, this should ensure that Start
	// is positive. This is critical here because if some of the offset are not normalized
	// then we may end up with a union windows that does not make sense.
	if smallestCover.Start() < 0 {
		utils.Panic("All offset should be normalized, but Start is %v", smallestCover.Start())
	}

	// Ensures we do not reuse an input vector here to limit the risk of overwriting one
	// of the input. This can happen if there is only a single window or if one windows
	// covers all the other.
	unionWindow := make([]field.Element, smallestCover.IntervalLen)
	var paddedTerm field.Element
	offset := smallestCover.Start()

	/*
		Now we actually compute the linear combinations for all offsets
	*/

	isFirst := true
	for i, pcw := range windows {
		interval := intervals[i]

		// Find the intersection with the larger window
		start_ := normalize(interval.Start(), offset, length)
		stop_ := normalize(interval.Stop(), offset, length)
		if stop_ == 0 {
			stop_ = length
		}

		// For the first match, we can save the operations by copying instead of
		// multiplying / adding
		if isFirst {
			isFirst = false
			op.vecIntoTerm(unionWindow[start_:stop_], pcw.window, coeffs[i])
			// #nosec G601 -- Deliberate pass by reference. (We trust the pointed object is not mutated)
			op.constIntoTerm(&paddedTerm, &pcw.paddingVal, coeffs[i])
			vector.Fill(unionWindow[:start_], paddedTerm)
			vector.Fill(unionWindow[stop_:], paddedTerm)
			continue
		}

		// sanity-check : Start and Stop are consistent with the size of pcw
		if stop_-start_ != len(pcw.window) {
			utils.Panic(
				"sanity-check failed. The renormalized coordinates (Start=%v, Stop=%v) are inconsistent with pcw : (len=%v)",
				start_, stop_, len(pcw.window),
			)
		}

		op.vecIntoVec(unionWindow[start_:stop_], pcw.window, coeffs[i])

		// Update the padded term
		// #nosec G601 -- Deliberate pass by reference. (We trust the pointed object is not mutated)
		op.constIntoConst(&paddedTerm, &pcw.paddingVal, coeffs[i])

		// Complete the left and the right-side of the window (i.e) the part
		// of unionWindow that does not overlap with pcw.window.
		// #nosec G601 -- Deliberate pass by reference. (We trust the pointed object is not mutated)
		op.constIntoVec(unionWindow[:start_], &pcw.paddingVal, coeffs[i])
		// #nosec G601 -- Deliberate pass by reference. (We trust the pointed object is not mutated)
		op.constIntoVec(unionWindow[stop_:], &pcw.paddingVal, coeffs[i])
	}

	if smallestCover.IsFullCircle() {
		return NewRegular(unionWindow), numMatches
	}

	return NewPaddedCircularWindow(unionWindow, paddedTerm, offset, length), numMatches
}

func (w *PaddedCircularWindow) DeepCopy() SmartVector {
	window := vector.DeepCopy(w.window)
	return NewPaddedCircularWindow(window, w.paddingVal, w.offset, w.totLen)
}

// Converts a smart-vector into a normal vec. The implementation minimizes
// then number of copies.
func (w *PaddedCircularWindow) IntoRegVecSaveAlloc() []field.Element {
	res, err := w.IntoRegVecSaveAllocBase()
	if err != nil {
		panic(conversionError)
	}
	return res

}

func (w *PaddedCircularWindow) IntoRegVecSaveAllocBase() ([]field.Element, error) {
	return IntoRegVec(w), nil
}

func (w *PaddedCircularWindow) IntoRegVecSaveAllocExt() []fext.Element {
	temp, _ := w.IntoRegVecSaveAllocBase()
	res := make([]fext.Element, len(temp))
	for i := 0; i < len(temp); i++ {
		elem := temp[i]
		res[i].SetFromBase(&elem)
	}
	return res
}
