package symbolic

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Evaluate the board for a batch  of inputs in parallel
func (b *ExpressionBoard) EvaluateExt(inputs []sv.SmartVector, p ...mempool.MemPool) sv.SmartVector {
	return b.EvaluateMixed(inputs, p...)
}

// Evaluate the board for a batch  of inputs in parallel
func (b *ExpressionBoard) EvaluateMixed(inputs []sv.SmartVector, p ...mempool.MemPool) sv.SmartVector {

	/*
		Find the size of the vector
	*/
	totalSize := 0
	for i, inp := range inputs {
		if totalSize > 0 && totalSize != inp.Len() {
			// Expects that all vector inputs have the same size
			utils.Panic("mismatch in the size: len(v) %v, totalsize %v, pos %v", inp.Len(), totalSize, i)
		}

		if totalSize == 0 {
			totalSize = inp.Len()
		}
	}

	if totalSize == 0 {
		utils.Panic("either there is no input or the inputs all have size 0")
	}

	if len(p) > 0 && p[0].Size() != MaxChunkSize {
		utils.Panic("the pool should be a pool of vectors of size=%v but it is %v", MaxChunkSize, p[0].Size())
	}

	isBase := sv.AreAllBase(inputs)
	if isBase {
		return b.Evaluate(inputs, p...)
	}

	// The result is an extension vector
	res := make([]fext.Element, totalSize)

	if totalSize < MaxChunkSize {
		// return b.evaluateSingleThreadMixed(inputs, p...).DeepCopy()
		solver := newChunkSolver(b)

		solver.reset(inputs, 0, totalSize, b)
		solver.evaluate()

		copy(res[:totalSize], solver.assignment[len(b.Nodes)-1][0].ConcreteValue[:totalSize])
		if totalSize == 1 {
			return sv.NewConstantExt(res[0], 1)
		}
		return sv.NewRegularExt(res)
	}

	if totalSize%MaxChunkSize != 0 {
		utils.Panic("totalSize %v is not divided by the chunk size %v", totalSize, MaxChunkSize)
	}

	numChunks := totalSize / MaxChunkSize

	parallel.Execute(numChunks, func(start, stop int) {

		solver := newChunkSolver(b)
		for chunkID := start; chunkID < stop; chunkID++ {

			chunkStart := chunkID * MaxChunkSize
			chunkStop := (chunkID + 1) * MaxChunkSize

			solver.reset(inputs, chunkStart, chunkStop, b)
			solver.evaluate()

			copy(res[chunkStart:chunkStop], solver.assignment[len(b.Nodes)-1][0].ConcreteValue[:])
		}

	})
	return sv.NewRegularExt(res)
}

func newChunkSolver(b *ExpressionBoard) chunkSolver {
	boardAssignmentsignment := chunkSolver{b: big.NewInt(0)}
	boardAssignmentsignment.assignment = make([][]nodeAssignmentGautam, len(b.Nodes))
	for lvl := range boardAssignmentsignment.assignment {
		boardAssignmentsignment.assignment[lvl] = make([]nodeAssignmentGautam, len(b.Nodes[lvl]))
	}

	// Init the constants and inputs.
	for i := range b.Nodes {
		for j := range b.Nodes[i] {
			na := &boardAssignmentsignment.assignment[i][j]
			node := &b.Nodes[i][j]
			na.Node = node
			na.HasValue = false
			na.IsConstant = false
			na.Inputs = make([]sv.SmartVector, len(node.Children))

			switch op := node.Operator.(type) {
			case Constant:
				// The constants are identified to constant vectors
				na.IsConstant = true
				na.HasValue = true
				cst := op.Val.GetExt()
				for k := range na.ConcreteValue {
					na.ConcreteValue[k] = cst
				}
				if len(node.Children) != 0 {
					panic("constant with children")
				}
			}

			// set the inputs pointers
			for k := range node.Children {
				id := node.Children[k]
				s := sv.NewRegularExt(boardAssignmentsignment.assignment[id.level()][id.posInLevel()].ConcreteValue[:])
				na.Inputs[k] = s
			}
		}
	}

	return boardAssignmentsignment
}

func (solver *chunkSolver) reset(inputs []sv.SmartVector, chunkStart, chunkStop int, b *ExpressionBoard) {
	inputCursor := 0
	chunkLen := chunkStop - chunkStart // can be < MaxChunkSize

	// Init the constants and inputs.
	for i := range b.Nodes {
		for j := range b.Nodes[i] {
			na := &solver.assignment[i][j]
			if !na.IsConstant {
				na.HasValue = false
			}

			if i == 0 {
				switch na.Node.Operator.(type) {
				case Variable:
					// leaves, we set the input.
					// Sanity-check the input should have the correct length
					sb := inputs[inputCursor].SubVector(chunkStart, chunkStop)
					sb.WriteInSliceExt(na.ConcreteValue[:chunkLen])
					// nodeAssignments[0][i].Value = inputs[inputCursor]
					inputCursor++
					na.HasValue = true
				}
			}
		}
	}

}

func (solver *chunkSolver) evaluate() {
	// Evaluate values level by level
	for level := 1; level < len(solver.assignment); level++ {
		for pil := range solver.assignment[level] {
			solver.evalNode(&solver.assignment[level][pil])
		}
	}
}

func (solver *chunkSolver) evalNode(na *nodeAssignmentGautam) {

	if na.HasValue {
		return
	}
	// TODO do we need to sanity check all inputs have values ?
	switch op := na.Node.Operator.(type) {
	case Product:
		res := extensions.Vector(na.ConcreteValue[:])
		// TODO @gbotrel if all exponents are 0, concreteValue can be dirty
		first := true
		for i := range na.Inputs {
			vTmp := extensions.Vector(solver.tmp[:])
			vInput := extensions.Vector(*na.Inputs[i].(*sv.RegularExt))
			vTmp.Exp(vInput, int64(op.Exponents[i]))
			if first {
				first = false
				copy(res, vTmp)
			} else {
				res.Mul(res, vTmp)
			}
		}
	case LinComb:
		var t0 extensions.E4
		vTmp := extensions.Vector(solver.tmp[:])
		res := extensions.Vector(na.ConcreteValue[:])
		for i := range na.Inputs {
			vInput := extensions.Vector(*na.Inputs[i].(*sv.RegularExt))
			coeff := op.Coeffs[i]

			if i == 0 {
				switch coeff {
				case 0:
					for j := range res {
						res[j].SetZero()
					}
				case 1:
					copy(res, vInput)
				case 2:
					res.Add(vInput, vInput)
				case -1:
					for j := range res {
						res[j].SetZero()
					}
					res.Sub(res, vInput)
				default:
					t0.B0.A0.SetInt64(int64(coeff))
					res.ScalarMul(vInput, &t0)
				}
				continue
			}

			switch coeff {
			case 0:
				continue
			case 1:
				res.Add(res, vInput)
			case 2:
				res.Add(res, vInput)
				res.Add(res, vInput)
			case -1:
				res.Sub(res, vInput)
			default:
				t0.B0.A0.SetInt64(int64(op.Coeffs[i]))
				vTmp.ScalarMul(vInput, &t0)
				res.Add(res, vTmp)
			}
		}
	default:
		eval := na.Node.Operator.EvaluateExt(na.Inputs)
		eval.WriteInSliceExt(na.ConcreteValue[:])
	}

	na.HasValue = true

}
