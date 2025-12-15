package emulated

import (
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Evaluation represents an emulated polynomial evaluation module
// that computes
//
//	\sum_i \prod_j Terms[i][j] = modulus * quotient + carry * (2^nbBitsPerLimb - challenge)
//
// where all Terms[i][j], modulus, quotient, and carry are given limb-wise
// in little-endian order (smallest limb first).
// The module computes the quotient and carry limbs at prover time
// and assigns them to the respective columns automatically.
//
// The returned evaluation module can be used to reference the computed
// auxiliary columns.
type Evaluation struct {
	// nbBitsPerLimb is the number of bits per limb
	nbBitsPerLimb int
	// round is the maximum round number of the input columns
	round int
	// name of the module
	name string
	// maxTermDegree is the maximum degree of all terms
	maxTermDegree int
	// nbLimbs is the maximum number of limbs seen over terms and modulus
	nbLimbs int

	// Terms are the evaluation terms
	// such that \sum_i \prod_j Terms[i][j] == 0
	Terms [][]Limbs
	// Modulus is the modulus for the evaluation
	Modulus Limbs

	// Quotient is the computed quotient limbs
	Quotient Limbs
	// Carry are the computed carry limbs
	Carry Limbs

	// Challenge is the random challenge used for the polynomial evaluation
	Challenge *coin.Info
	// ChallengePowers are the powers of the challenge used for the polynomial evaluation
	ChallengePowers []ifaces.Column
}

// NewEval creates a new emulated evaluation module that computes
//
//	\sum_i \prod_j Terms[i][j] = modulus * quotient + carry * (2^nbBitsPerLimb - challenge)
//
// We assume that the inputs are already given limb-wise in little-endian order
// (smalles limb first) and that the limbs are already range-checked to be
// within [0, 2^nbBitsPerLimb).
//
// The module computes the quotient and carry limbs at prover time and assigns
// them to the respective columns automatically.
//
// The returned evaluation module can be used to reference the computed
// auxiliary columns.
func NewEval(comp *wizard.CompiledIOP, name string, nbBitsPerLimb int, modulus Limbs, terms [][]Limbs) *Evaluation {
	round := 0
	nbRows := modulus.Columns[0].Size()
	maxTermDegree := 0
	nbLimbs := len(modulus.Columns)
	nbRangecheckBits := 16
	nbRangecheckLimbs := (nbBitsPerLimb + nbRangecheckBits - 1) / nbRangecheckBits
	for i := range terms {
		maxTermDegree = max(maxTermDegree, len(terms[i]))
		for j := range terms[i] {
			for k := range terms[i][j].Columns {
				round = max(round, terms[i][j].Columns[k].Round())
			}
		}
	}
	nbQuoBits := 0
	for i := range terms {
		nbTermQuoLimbsBits := len(terms[i][0].Columns) * nbBitsPerLimb
		for j := 1; j < len(terms[i]); j++ {
			nbTermQuoLimbsBits = nbMultiplicationResLimbs(nbTermQuoLimbsBits, len(terms[i][j].Columns)*nbBitsPerLimb)
			nbLimbs = max(nbLimbs, len(terms[i][j].Columns))
		}
		nbQuoBits = max(nbQuoBits, nbTermQuoLimbsBits)
	}
	nbQuoBits += utils.DivCeil(utils.Log2Ceil(len(terms)), nbBitsPerLimb) // add some slack for the addition of terms
	nbCarryBits := nbQuoBits
	nbQuoBits = max(0, nbQuoBits-len(modulus.Columns)*nbBitsPerLimb+1) // we divide by modulus of nbLimbs size
	nbQuoLimbs := utils.DivCeil(nbQuoBits, nbBitsPerLimb)
	for i := range modulus.Columns {
		round = max(round, modulus.Columns[i].Round())
	}

	quotient := Limbs{
		Columns: make([]ifaces.Column, nbQuoLimbs),
	}
	for i := range quotient.Columns {
		quotient.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_EVAL_QUO_LIMB_%d", name, i),
			nbRows,
		)
	}
	nbCarryLimbs := utils.DivCeil(nbCarryBits, nbBitsPerLimb)
	carry := Limbs{
		Columns: make([]ifaces.Column, nbCarryLimbs),
	}
	for i := range carry.Columns {
		carry.Columns[i] = comp.InsertCommit(
			round,
			ifaces.ColIDf("%s_EMUL_EVAL_CARRY_LIMB_%d", name, i),
			nbRows,
		)
	}
	// define the challenge and challenge powers for polynomial evaluation at random point
	challenge := comp.InsertCoin(round+1, coin.Namef("%s_EMUL_CHALLENGE", name), coin.Field)
	challengePowers := make([]ifaces.Column, len(carry.Columns))
	for i := range challengePowers {
		challengePowers[i] = comp.InsertCommit(
			round+1,
			ifaces.ColIDf("%s_EMUL_CHALLENGE_POWER_%d", name, i),
			nbRows,
		)
	}

	pa := &Evaluation{
		Terms:           terms,
		Modulus:         modulus,
		Quotient:        quotient,
		Carry:           carry,
		Challenge:       &challenge,
		ChallengePowers: challengePowers,
		nbBitsPerLimb:   nbBitsPerLimb,
		round:           round,
		name:            name,
		maxTermDegree:   maxTermDegree,
		nbLimbs:         nbLimbs,
	}

	// we need to register the prover actions for assigning the emulated columns before doing range checks
	// to ensure the values are available (prover action is FIFO)
	comp.RegisterProverAction(round, &proverActionFn{pa.assignEmulatedColumns})
	comp.RegisterProverAction(round+1, &proverActionFn{pa.assignChallengePowers})

	// range check the quotient limbs
	for i := range quotient.Columns {
		bigrange.BigRange(
			comp,
			ifaces.ColumnAsVariable(pa.Quotient.Columns[i]), int(nbRangecheckLimbs), nbRangecheckBits,
			fmt.Sprintf("%s_EMUL_QUOTIENT_LIMB_RANGE_%d", name, i),
		)
	}

	// define the constraints for the emulated evaluation
	pa.csEval(comp)
	// define constraints for the correctness of challenge powers
	csChallengePowers(comp, pa.Challenge, pa.ChallengePowers, round, name)
	return pa
}

func (cs *Evaluation) csEval(comp *wizard.CompiledIOP) {
	// TODO(ivokub): should we write the evaluation results in a limb?

	// this method computes the following constraint:
	//
	//  \sum_i \prod_j Terms[i][j](x) - modulus(x) * quotient(x) - carry(x) * (2^nbBitsPerLimb - challenge) = 0

	// we first compute all unique limb polynomials. This is to ensure that
	// if the same limb appears multiple times in the terms, we only compute
	// its polynomial evaluation once.
	uniqueLimbs := make(map[string]*symbolic.Expression)
	for i := range cs.Terms {
		for j := range cs.Terms[i] {
			name := cs.Terms[i][j].String()
			if _, ok := uniqueLimbs[name]; !ok {
				uniqueLimbs[name] = csPolyEval(cs.Terms[i][j], cs.ChallengePowers)
			}
		}
	}

	// now we start computing the evaluation sum
	//
	//  \sum_i \prod_j Terms[i][j](x)
	evalSum := symbolic.NewConstant(0)
	for i := range cs.Terms {
		// compute \prod_j Terms[i][j](x)
		termProd := symbolic.NewConstant(1)
		for j := range cs.Terms[i] {
			name := cs.Terms[i][j].String()
			termProd = symbolic.Mul(termProd, uniqueLimbs[name])
		}
		evalSum = symbolic.Add(evalSum, termProd)
	}
	// now we compute the other polynomials

	// modulus(x)
	modulusEval := csPolyEval(cs.Modulus, cs.ChallengePowers)
	// quotient(x)
	quotientEval := csPolyEval(cs.Quotient, cs.ChallengePowers)
	// carry(x)
	carryEval := csPolyEval(cs.Carry, cs.ChallengePowers)
	// 2^nbBitsPerLimb - challenge
	coef := big.NewInt(0).Lsh(big.NewInt(1), uint(cs.nbBitsPerLimb))
	carryCoef := symbolic.Sub(
		symbolic.NewConstant(coef),
		cs.Challenge.AsVariable(),
	)

	// (2^nbBitsPerLimb - challenge) * carry(x)
	carryCoefEval := symbolic.Mul(carryEval, carryCoef)
	// modulus(x) * quotient(x)
	qmEval := symbolic.Mul(quotientEval, modulusEval)

	// finally define the constraint:
	// \sum_i \prod_j Terms[i][j](x) - modulus(x) * quotient(x) - carry(x) * (2^nbBitsPerLimb - challenge) = 0
	comp.InsertGlobal(
		cs.round+1,
		ifaces.QueryIDf("%s_EMUL_EVAL", cs.name),
		symbolic.Sub(
			evalSum,
			qmEval,
			carryCoefEval,
		),
	)
}
