package smartvectors

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/maths/v2/field"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// It's a slice - zero padded up to a certain length - and rotated
type PaddedCircularWindow[T anyField] struct {
	Window_     []T
	PaddingVal_ T
	// Totlen is the length of the represented vector
	TotLen_, Offset_ int
}

// Create a new padded circular window vector
func NewPaddedCircularWindow[T anyField](window []T, paddingVal T, offset, totLen int) *PaddedCircularWindow[T] {
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
	return &PaddedCircularWindow[T]{
		Window_:     window,
		PaddingVal_: paddingVal,
		Offset_:     offset,
		TotLen_:     totLen,
	}
}

// Returns the length of the vector
func (p *PaddedCircularWindow[T]) Len() int {
	return p.TotLen_
}

// Offset returns the offset of the PCW
func (p *PaddedCircularWindow[T]) Offset() int {
	return p.Offset_
}

// Windows returns the length of the window of the PCQ
func (p *PaddedCircularWindow[T]) Window() []T {
	return p.Window_
}

// PaddingVal returns the value used for padding the window
func (p *PaddedCircularWindow[T]) PaddingVal() T {
	return p.PaddingVal_
}

func (p *PaddedCircularWindow[T]) Get(n int) field.Gen {
	// Check if the queried index is in the window
	posFromWindowsPoV := utils.PositiveMod(n-p.Offset_, p.TotLen_)
	if posFromWindowsPoV < len(p.Window_) {
		return field.NewGen(p.Window_[posFromWindowsPoV])
	}
	// Else, return the padding value
	return field.NewGen(p.PaddingVal_)
}

// Extract a subvector from p[Start:Stop), the subvector cannot "roll-over".
// i.e, we enforce that Start < Stop
func (p *PaddedCircularWindow[T]) SubVector(start, stop int) SmartVector {
	// negative Start value is not allowed
	if start < 0 {
		panic("negative Start value is not allowed")
	}

	// Sanity checks for all subvectors
	assertValidRange(start, stop)
	assertInBound(stop, p.TotLen_+1) // The +1 is because we accept if "Stop = length"

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
		return NewConstant(p.PaddingVal_, b)
	}

	// Case 2 : return a regular vector
	if b <= d && d < c {
		reg := Regular[T](p.Window_[n-c : n-c+b])
		return &reg
	}

	// Case 2* : same as 2 but c == 0
	if b <= d && c == 0 {
		reg := Regular[T](p.Window_[:b])
		return &reg
	}

	// Case 3 : the window is fully contained in the subvector
	if c < d && d <= b {
		return NewPaddedCircularWindow(p.Window_, p.PaddingVal_, c, b)
	}

	// Case 4 : left-ended
	if c < b && c <= d {
		return NewPaddedCircularWindow(p.Window_[:b-c], p.PaddingVal_, c, b)
	}

	// Case 5 : the window is double ended (we skip some element in the center of the window)
	if d < c && c < b {
		left := p.Window_[:b-c]
		right := p.Window_[n-c:]

		// The deep-copy of left ensures that we do not append
		// on the same concrete slice.
		w := append(field.VecDeepCopy(left), right...)
		return NewPaddedCircularWindow(w, p.PaddingVal_, c, b)
	}

	// Case 6 : right-ended
	if 0 < d && d < b && b <= c {
		return NewPaddedCircularWindow(p.Window_[n-c:], p.PaddingVal_, 0, b)
	}

	// Case 7 : d == 0 and c is out
	if d == 0 && b <= c {
		return NewConstant(p.PaddingVal_, b)
	}

	panic(fmt.Sprintf("unsupported case : b %v, c %v, d %v", b, c, d))

}

// Rotate the vector
func (p *PaddedCircularWindow[T]) RotateRight(offset int) SmartVector {
	return NewPaddedCircularWindow[T](
		field.VecDeepCopy(p.Window_),
		p.PaddingVal_,
		p.Offset_+offset,
		p.TotLen_,
	)
}

func (p *PaddedCircularWindow[T]) interval() CircularInterval {
	return IvalWithStartLen(p.Offset_, len(p.Window_), p.TotLen_)
}

// normalize converts the (circle) coordinator x to another coordinate by changing
// the origin point on the discret circle. mod denotes the number of points in
// the circle.
func normalize(x, newRef, mod int) int {
	return utils.PositiveMod(x-newRef, mod)
}

func (w *PaddedCircularWindow[T]) DeepCopy() SmartVector {
	return NewPaddedCircularWindow(
		field.VecDeepCopy(w.Window_),
		w.PaddingVal_,
		w.Offset_,
		w.TotLen_,
	)
}

func (w *PaddedCircularWindow[T]) GetPtr(n int) *field.Gen {
	x := w.Get(n)
	return &x
}
