package column

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

/*
Represents a vector as the result of interleaved vectors of the same size.
*/
type Interleaved struct {
	Parents []ifaces.Column
}

/*
Instantiates an interleaved handle.

Will panic, if
  - Non-powers of two parents
  - The parents do not have the same size
*/
func Interleave(parents ...ifaces.Column) ifaces.Column {
	if !utils.IsPowerOfTwo(len(parents)) {
		utils.Panic("Expected power of two parents, got %v", len(parents))
	}

	size0 := parents[0].Size()
	for _, parent := range parents {
		if size0 != parent.Size() {
			utils.Panic("Mismatch of the sizes of the parents when creating an interleaving")
		}
	}

	return Interleaved{Parents: parents}
}

/*
Does not change the size
*/
func (i Interleaved) Size() int {
	return i.Parents[0].Size() * len(i.Parents)
}

/*
String repr of a shifted handle
*/
func (s Interleaved) GetColID() ifaces.ColID {
	msg := INTERLEAVED
	for _, h := range s.Parents {
		msg = fmt.Sprintf("%v_%v", msg, h.GetColID())
	}
	return ifaces.ColID(msg)
}

/*
Defers to the parents
*/
func (s Interleaved) MustExists() {
	for _, h := range s.Parents {
		h.MustExists()
	}
}

/*
Defers to the parent, returns the max
*/
func (s Interleaved) Round() int {
	curr := 0
	for _, h := range s.Parents {
		curr = utils.Max(curr, h.Round())
	}
	return curr
}

/*
Get the witness from the parents and manually interleave them in a new vector.
*/
func (s Interleaved) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {

	parents := []ifaces.ColAssignment{}
	for _, h := range s.Parents {
		parentWit := h.GetColAssignment(run)
		parents = append(parents, parentWit)
	}

	res := make([]field.Element, 0, s.Size())
	parentLength := parents[0].Len()

	for i := 0; i < parentLength; i++ {
		for j := range parents {
			res = append(res, parents[j].Get(i))
		}
	}

	return smartvectors.NewRegular(res)
}

/*
Given evaluations wraps different evaluations of the parents to get an
evaluation of the interleaved columns.

Panics if len(parentXs) != len(parentYs)

If the parents are P0, P1, P2, P3, ... P_{k-1} each of size n
Let ω be an "nk" root of unity

  - Then, we assert that, for all j parentXs[j] = ω^j * x, where x = parentXs[0]
    Panic if this fails.
  - We assume that parentYs[j] = P(parentXs[j])

- The resulting evaluation of I, is obtained as

	I(x) = Σ_{j=0..k-1} c_j.parentYs[j], where c_j satisfies
	c_j * k * [parentXs[j]^n - 1] = (x^{nk} - 1)
*/
func (il *Interleaved) recoverEvaluationFromParents(
	parentXs, parentYs []field.Element,
) (y field.Element) {

	if len(parentXs) != len(parentYs) || len(parentXs) != len(il.Parents) {
		utils.Panic("mismatch in the size of parentXs and parentYs")
	}

	n := il.Parents[0].Size()
	k := len(il.Parents)
	nk := il.Size()           // which obviously is assumed to equals n * k
	omega := fft.GetOmega(nk) // such that to ω^{nk} = 1
	x := parentXs[0]          // the evaluation point for I (i.e `il`, the interleaved polynomial)
	one := field.One()

	/*
		checks the parentXs
	*/
	for j := 1; j < k; j++ {
		var expected field.Element
		expected.Mul(&parentXs[j], &omega)
		if expected != parentXs[j-1] {
			// It is a legit error case, but we should check it earlier.
			utils.Panic("mismatch for the xs")
		}
	}

	/*
		computes the resulting y

		`xNKminOneDivK` = (x^(nk) - 1) / k is a common term for all j
	*/
	var xNKminOneDivK field.Element // common for all j
	xNKminOneDivK.Exp(x, big.NewInt(int64(nk)))
	xNKminOneDivK.Sub(&xNKminOneDivK, &one)
	kFr := field.NewElement(uint64(k))
	xNKminOneDivK.Div(&xNKminOneDivK, &kFr)

	y = field.Zero()

	for j := range il.Parents {
		/*
			c_j = (x^(nk) - 1) * (x_j^n - 1)^-1 / k
		*/
		var cj field.Element
		cj.Exp(parentXs[j], big.NewInt(int64(n)))
		cj.Sub(&cj, &one)
		cj.Inverse(&cj)
		cj.Mul(&cj, &xNKminOneDivK)
		/*
			y += parentYs[j] * c_j

			Note that we reuse c_j to store the product `parentYs[j] * c_j`
		*/
		cj.Mul(&cj, &parentYs[j])
		y.Add(&y, &cj)
	}

	return y
}

/*
Same as `recoverEvaluationFromParents` but in gnark circuit
*/
func (il *Interleaved) gnarkRecoverEvaluationFromParents(api frontend.API, parentXs, parentYs []frontend.Variable) frontend.Variable {
	if len(parentXs) != len(parentYs) || len(parentXs) != len(il.Parents) {
		utils.Panic("mismatch in the size of parentXs and parentYs")
	}

	n := il.Parents[0].Size()
	k := len(il.Parents)
	nk := il.Size()  // which obviously is assumed to equals n * k
	x := parentXs[0] // the evaluation point for I (i.e `il`, the interleaved polynomial)
	one := frontend.Variable(field.One())

	/*
		No need to check the parents here
	*/

	/*
		computes the resulting y

		`xNKminOneDivK` = (x^(nk) - 1) / k is a common term for all j
	*/
	xNKminOneDivK := gnarkutil.Exp(api, x, nk)
	xNKminOneDivK = api.Sub(xNKminOneDivK, one)
	kFr := frontend.Variable(k)
	xNKminOneDivK = api.Div(xNKminOneDivK, kFr)

	y := frontend.Variable(0)

	for j := range il.Parents {
		/*
			c_j = (x^(nk) - 1) * (x_j^n - 1)^-1 / k
		*/
		cj := gnarkutil.Exp(api, parentXs[j], n)
		cj = api.Sub(cj, one)
		cj = api.Inverse(cj)
		cj = api.Mul(cj, xNKminOneDivK)
		/*
			y += parentYs[j] * c_j

			Note that we reuse c_j to store the product `parentYs[j] * c_j`
		*/
		cj = api.Mul(cj, parentYs[j])
		y = api.Add(y, cj)
	}

	return y
}

func (i Interleaved) IsComposite() bool { return true }

/*
Get the witness from the parents and manually interleave them in a new vector.
*/
func (s Interleaved) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {

	parents := [][]frontend.Variable{}
	for _, h := range s.Parents {
		parentWit := h.GetColAssignmentGnark(run)
		parents = append(parents, parentWit)
	}

	res := make([]frontend.Variable, 0, s.Size())
	parentLength := len(parents[0])

	for i := 0; i < parentLength; i++ {
		for j := range parents {
			res = append(res, parents[j][i])
		}
	}

	return res
}

/*
Get a particular position of the column
*/
func (s Interleaved) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	// index of the underlying parent
	noParent := pos % len(s.Parents)
	posInParent := pos / len(s.Parents)
	return s.Parents[noParent].GetColAssignmentAt(run, posInParent)
}

/*
Get the witness from the parents and manually interleave them in a new vector.
*/
func (s Interleaved) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
	// index of the underlying parent
	noParent := pos % len(s.Parents)
	posInParent := pos / len(s.Parents)
	return s.Parents[noParent].GetColAssignmentGnarkAt(run, posInParent)
}

/*
Returns the name of the column as a string
*/
func (s Interleaved) String() string {
	return string(s.GetColID())
}
