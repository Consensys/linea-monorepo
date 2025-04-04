package bigrange

import (
	"math/big"
	"reflect"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
)

// BigRange enforces that an input sympolic expression evaluates to values that
// can be expressed in `numLimbs` limbs of `bitPerLimbs` bits. This is equivalent
// to enforcing that the values taken by `expr` are within the range
// [0; 2 ** (numLimbs * bitPerLimbs)]. The string `name` is used to be appended
// to all the generated queries and columns names to provide some context and
// to distinguish between different calls to `BigRange`. `BigRange` should never
// be called twice with the same `name` or it will result in an error stemming
// from the fact that it would try to create two columns with the same name,
// which is forbidden.
//
// Example:
//
// ```
//
//	// Enforce that all the values of `col` are within range [0, 2**64]
//	bigrange.BigRange(comp, ifaces.ColAsVariable(col), 4, 16, "my_context")
//
// ```
func BigRange(comp *wizard.CompiledIOP, expr *symbolic.Expression, numLimbs, bitPerLimbs int, name string) {

	var (
		limbs        = make([]ifaces.Column, numLimbs)
		round        = wizardutils.LastRoundToEval(expr)
		boarded      = expr.Board() // Still computed here, but we'll pass it as a pointer to the struct
		size         = wizardutils.ExprIsOnSameLengthHandles(&boarded)
		totalNumBits = numLimbs * bitPerLimbs
	)

	for i := range limbs {
		// Declare the limbs for the number
		limbs[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("BIGRANGE_%v_LIMB_%v", name, i),
			size,
		)
		// Enforces the range over the limbs
		comp.InsertRange(
			round,
			ifaces.QueryIDf("BIGRANGE_%v_LIMB_%v", name, i),
			limbs[i],
			1<<bitPerLimbs,
		)
	}

	// Build the linear combination with powers of 2^bitPerLimbs.
	// The limbs are in "little-endian" order. Namely, the first
	// limb encodes the least significant bits first.
	pow2 := symbolic.NewConstant(1 << bitPerLimbs)
	acc := ifaces.ColumnAsVariable(limbs[numLimbs-1])
	for i := numLimbs - 2; i >= 0; i-- {
		acc = acc.Mul(pow2)
		acc = acc.Add(ifaces.ColumnAsVariable(limbs[i]))
	}

	// Declare the global constraint
	comp.InsertGlobal(round, ifaces.QueryIDf("GLOBAL_BIGRANGE_%v", name), acc.Sub(expr))

	// The below prover steps assign the limb values
	comp.RegisterProverAction(round, &bigRangeAssignProverAction{
		expr:         expr,
		boarded:      &boarded, // Pass as a pointer
		limbs:        limbs,
		size:         size,
		numLimbs:     numLimbs,
		bitPerLimbs:  bitPerLimbs,
		totalNumBits: totalNumBits,
		name:         name,
	})
}

// bigRangeAssignProverAction assigns the limb values for the BigRange check.
// It implements the [wizard.ProverAction] interface.
type bigRangeAssignProverAction struct {
	expr         *symbolic.Expression
	boarded      *symbolic.ExpressionBoard // Changed to a pointer
	limbs        []ifaces.Column
	size         int
	numLimbs     int
	bitPerLimbs  int
	totalNumBits int
	name         string
}

// Run executes the assignment of the limb values.
func (a *bigRangeAssignProverAction) Run(run *wizard.ProverRuntime) {
	// Evaluate the expression we wish to range proof
	metadatas := a.boarded.ListVariableMetadata() // Use the pointer directly

	/*
		Collects the relevant datas into a slice for the evaluation
	*/
	evalInputs := make([]sv.SmartVector, len(metadatas))

	/*
		Omega is a root of unity which generates the domain of evaluation
		of the constraint. Its size coincide with the size of the domain
		of evaluation. For each value of `i`, X will evaluate to omega^i.
	*/
	omega := fft.GetOmega(a.size)
	omegaI := field.One()

	// precomputations of the powers of omega, can be optimized if useful
	omegas := make([]field.Element, a.size)
	for i := 0; i < a.size; i++ {
		omegas[i] = omegaI
		omegaI.Mul(&omegaI, &omega)
	}

	/*
		Collect the relevants inputs for evaluating the constraint
	*/
	for k, metadataInterface := range metadatas {
		switch meta := metadataInterface.(type) {
		case ifaces.Column:
			w := meta.GetColAssignment(run)
			evalInputs[k] = w
		case coin.Info:
			// Implicitly, the coin has to be a field element in the expression
			// It will panic if not
			x := run.GetRandomCoinField(meta.Name)
			evalInputs[k] = sv.NewConstant(x, a.size)
		case variables.X:
			evalInputs[k] = meta.EvalCoset(a.size, 0, 1, false)
		case variables.PeriodicSample:
			evalInputs[k] = meta.EvalCoset(a.size, 0, 1, false)
		case ifaces.Accessor:
			evalInputs[k] = sv.NewConstant(meta.GetVal(run), a.size)
		default:
			utils.Panic("Not a variable type %v in sub-wizard %v", reflect.TypeOf(metadataInterface), a.name)
		}
	}

	// This panics if the global constraints doesn't use any commitment
	resWitness := a.boarded.Evaluate(evalInputs) // Use the pointer directly

	limbsWitness := make([][]field.Element, a.numLimbs)
	for i := range limbsWitness {
		limbsWitness[i] = make([]field.Element, a.size)
	}

	for j := 0; j < a.size; j++ {
		x := resWitness.Get(j)
		var tmp big.Int
		x.BigInt(&tmp)

		if tmp.BitLen() > a.totalNumBits {
			utils.Panic(
				"BigRange: cannot prove that the bitLen is smaller than %v : the provided witness has %v bits on position %v (%v)",
				a.totalNumBits, tmp.BitLen(), j, x.String(),
			)
		}

		for i := 0; i < a.numLimbs; i++ {
			l := uint64(0)
			for k := i * a.bitPerLimbs; k < (i+1)*a.bitPerLimbs; k++ {
				extractedBit := tmp.Bit(k)
				l |= uint64(extractedBit) << (k % a.bitPerLimbs)
			}
			limbsWitness[i][j].SetUint64(l)
		}
	}

	// Then assigns the limbs
	for i := range limbsWitness {
		run.AssignColumn(a.limbs[i].GetColID(), sv.NewRegular(limbsWitness[i]))
	}
}
