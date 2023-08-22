package bigrange

import (
	"math/big"
	"reflect"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"

	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizardutils"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

func BigRange(comp *wizard.CompiledIOP, expr *symbolic.Expression, numLimbs, bitPerLimbs int, name string) {

	limbs := make([]ifaces.Column, numLimbs)
	round := wizardutils.LastRoundToEval(comp, expr)
	size := wizardutils.ExprIsOnSameLengthHandles(comp, expr)

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

	// Build the linear combination with power ofs 2^bitPerLimbs.
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

	comp.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {

		// Evaluate the expression we wish to range proof
		boarded := expr.Board()
		metadatas := boarded.ListVariableMetadata()

		/*
			Collects the relevant datas into a slice for the evaluation
		*/
		evalInputs := make([]sv.SmartVector, len(metadatas))

		/*
			Omega is a root of unity which generates the domain of evaluation
			of the constraint. Its size coincide with the size of the domain
			of evaluation. For each value of `i`, X will evaluate to omega^i.
		*/
		omega := fft.GetOmega(size)
		omegaI := field.One()

		// precomputations of the powers of omega, can be optimized if useful
		omegas := make([]field.Element, size)
		for i := 0; i < size; i++ {
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
				evalInputs[k] = sv.NewConstant(x, size)
			case variables.X:
				evalInputs[k] = meta.EvalCoset(size, 0, 1, false)
			case variables.PeriodicSample:
				evalInputs[k] = meta.EvalCoset(size, 0, 1, false)
			case *ifaces.Accessor:
				evalInputs[k] = sv.NewConstant(meta.GetVal(run), size)
			default:
				utils.Panic("Not a variable type %v in sub-wizard %v", reflect.TypeOf(metadataInterface), name)
			}
		}

		// This panics if the global constraints doesn't use any commitment
		resWitness := boarded.Evaluate(evalInputs)

		limbsWitness := make([][]field.Element, numLimbs)
		for i := range limbsWitness {
			limbsWitness[i] = make([]field.Element, size)
		}

		for j := 0; j < size; j++ {
			x := resWitness.Get(j)
			var tmp big.Int
			x.BigInt(&tmp)

			for i := 0; i < numLimbs; i++ {
				l := uint64(0)
				for k := i * bitPerLimbs; k < (i+1)*bitPerLimbs; k++ {
					extractedBit := tmp.Bit(k)
					l |= uint64(extractedBit) << (k % bitPerLimbs)
				}
				limbsWitness[i][j].SetUint64(l)
			}
		}

		// Then assigns the limbs
		for i := range limbsWitness {
			run.AssignColumn(limbs[i].GetColID(), sv.NewRegular(limbsWitness[i]))
		}

	})

}
