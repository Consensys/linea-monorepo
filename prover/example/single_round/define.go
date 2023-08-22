package single_round

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
)

var (
	P1, P2, P3, P4 ifaces.ColID = "P1", "P2", "P3", "P4"
)

func GenWizard(smallSize, size int) (define func(build *zkevm.Builder), prover func(*wizard.ProverRuntime)) {
	return func(build *zkevm.Builder) {
			Define(build, smallSize, size)
		}, func(run *wizard.ProverRuntime) {
			Prove(run, smallSize, size)
		}
}

/*
The example aims at showing a mixture of all the different features of the
wizard.IOPs
*/
func Define(build *zkevm.Builder, smallSize, size int) {

	/*
		Registration of the polynomials
	*/

	P1 := build.RegisterCommit("P1", size)
	P2 := build.RegisterCommit("P2", size)
	P3 := build.RegisterCommit("P3", size)
	P4 := build.RegisterCommit("P4", smallSize)

	/*
		Advanced references to already committed polynomials
	*/

	_ = P1.Shift(1)
	_ = P4.Repeat(2)
	_ = P3.Shift(-3).Repeat(4)
	_ = zkevm.Interleave(P1, P2)

	/*
		It is also possible to instantiate symbolic expressions to formulate
		more complex queries.
	*/

	var1 := P4.Repeat(2).AsVariable()
	var2 := zkevm.Interleave(P1, P2).AsVariable()
	constant := symbolic.NewConstant(12)
	_ = symbolic.NewConstant("45")
	_ = var1.Add(var2)               // var1 + var2
	_ = var2.Mul(var1)               // var1 * var2
	_ = constant.Sub(var1).Add(var1) // (cnst - var1) + var1
	_ = var1.Neg()                   // - var1
	_ = var2.Square()                // var2 ^ 2

	/*
		Create a global constraint between variables. "Shift" indicates that
		we reference P(wX) in place of P(X). In order to use a committed polynomial
		in an expression, it must be converted in `Variable` using `AsVariable`. It
		is possible to pass an advanced reference as well.

		The constraint below enforces that

			`P1(x) * P2(w^3.x) = 0 for all x such that x^size = 1`
	*/

	expr := P1.AsVariable().Mul(
		P2.Shift(3).AsVariable(),
	)

	build.GlobalConstraint("GLOBAL", expr)

	/*
		A local constraint is formatted in the same way as a global one.
		The evaluation point is always 0 (the global constraint will check
		on the entire domain). In that case, the shift is used to point to
		another point.

		The constraint below enforces that

			`P1(1) * P2(w^3) = 0`
	*/

	expr = P1.AsVariable().Mul(
		P2.Shift(3).AsVariable(),
	)

	build.LocalConstraint("NAME_OF_THE_GLOBAL_CONSTRAINT", expr)

	/*
		A permutation constraint enforces that two vector of polynomials
		evaluates to row-permuted matrices (considering that each polynomial)
		evaluates to a column.

		For instance, assume that N=4 is the size of the domain of roots of unity
		here. Let A1(X), A2(X), B1(X), B2(X) with the following evaluations on the
		domain. The example below works because (A1, A2) and (B1, B2) works.

						A1(X)	A2(X)	|	B1(X)	B2(X)
										|
				1		a		e		|	b		f
				w		b		f		|	d		h
				w^2		c		g		|	a		e
				w^3		d		h		|	c		g

		In case, each side contains one column (A1, B1). Then the check consists in
		asserting that the evaluation vectors of A1 and B1 are permutation of each
		other. Equivalently, this means that both evaluation vectors contains the
		same elements the same number of time.

		In the example below, we are adding the constraint that P2 and P3 are permutation
		of each other.
	*/

	build.Permutation("NAME_OF_THE_PERMUTATION_CONSTRAINT",
		[]zkevm.Handle{P2}, []zkevm.Handle{P3})

	/*
		In the example below, we are adding the constraint that P3 evaluates only to
		elements that are contained in P4. The difference with the permutation constraint
		is that P3 does not have to contains every element of P4. P3 can contain has many
		time the same element of P4 as it wants.

		As for the permutations, it is possible to register multi-inclusions by passing
		several polynomials in both sides of the inclusion constraint. In that case, the
		constraint will asserted of the rows of the evaluation matrix of both sides (with
		the same conventions as for the permutation).
	*/

	build.Inclusion("NAME_OF_THE_INCLUSION_CONSTRAINT",
		[]zkevm.Handle{P4}, []zkevm.Handle{P3})

	/*
		In the example below, we are adding the constraint that P4
		has only values contained within [0, smallSize)
	*/
	build.Range("NAME_OF_THE_RANGE_CONSTRAINT", P4, smallSize)

}
