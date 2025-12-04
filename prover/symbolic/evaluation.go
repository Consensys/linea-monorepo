package symbolic

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"

	"github.com/consensys/gnark/frontend"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
)

// evaluation is a helper to evaluate an expression board on chunks of data.
type evaluation[T any] struct {
	values    []T // contiguous memory to store all the vectors
	scratch   []T
	chunkSize int
	board     *ExpressionBoard
}

// Evaluate evaluates the expression board on the provided inputs.
func (b *ExpressionBoard) Evaluate(inputs []sv.SmartVector) sv.SmartVector {
	// essentially, we can see the inputs as "columns", and the chunks as "rows"
	// the relations between the columns are defined by the expression board
	// we evaluate the expression board chunk by chunk, in parallel.

	if len(inputs) == 0 {
		panic("no input provided")
	}
	// sanity check the inputs
	totalSize := inputs[0].Len()
	for i := 1; i < len(inputs); i++ {
		if inputs[i].Len() != totalSize {
			utils.Panic("mismatch in the size: len(v0) %v, len(v%v) %v", totalSize, i, inputs[i].Len())
		}
	}
	if totalSize == 0 {
		panic("inputs all have size 0")
	}

	chunkSize := min(1<<5, totalSize) // default chunk size
	if totalSize%chunkSize != 0 {
		panic("chunk size should divide total size")
	}

	numChunks := totalSize / chunkSize

	// if all inputs are base, the result is also base
	isBase := sv.AreAllBase(inputs)
	if isBase {
		res := make([]field.Element, totalSize)

		parallel.Execute(numChunks, func(start, stop int) {

			eval := newEvaluation[field.Element](b, chunkSize)
			for chunkID := start; chunkID < stop; chunkID++ {

				chunkStart := chunkID * chunkSize
				chunkStop := (chunkID + 1) * chunkSize

				eval.evaluate(inputs, chunkStart, chunkStop)

				lastNodeOffset := (len(b.Nodes) - 1) * chunkSize
				chunkLen := chunkStop - chunkStart
				copy(res[chunkStart:chunkStop], eval.values[lastNodeOffset:lastNodeOffset+chunkLen])
			}

		})

		if areAllConstants(inputs) {
			// TODO @gbotrel we are wasting compute if this is a common case.
			// in standard_benchmark, this happens only once (not with base, with ext)
			return sv.NewConstant(res[0], totalSize)
		}

		return sv.NewRegular(res)
	}

	// The result is an extension vector
	res := make([]fext.Element, totalSize)

	parallel.Execute(numChunks, func(start, stop int) {

		eval := newEvaluation[fext.Element](b, chunkSize)
		for chunkID := start; chunkID < stop; chunkID++ {

			chunkStart := chunkID * chunkSize
			chunkStop := (chunkID + 1) * chunkSize

			eval.evaluate(inputs, chunkStart, chunkStop)

			lastNodeOffset := (len(b.Nodes) - 1) * chunkSize
			chunkLen := chunkStop - chunkStart
			copy(res[chunkStart:chunkStop], eval.values[lastNodeOffset:lastNodeOffset+chunkLen])
		}

	})
	if areAllConstants(inputs) {
		// TODO @gbotrel we are wasting compute if this is a common case.
		// in standard_benchmark, this happens only once.
		return sv.NewConstantExt(res[0], totalSize)
	}

	return sv.NewRegularExt(res)
}

func areAllConstants(inp []sv.SmartVector) bool {
	for _, v := range inp {
		// must be sv.Constant or sv.ConstantExt
		switch v.(type) {
		case *sv.Constant:
		case *sv.ConstantExt:
		default:
			return false
		}
	}
	return true
}

func newEvaluation[T any](b *ExpressionBoard, chunkSize int) evaluation[T] {
	eval := evaluation[T]{
		values:    make([]T, len(b.Nodes)*chunkSize),
		scratch:   make([]T, chunkSize),
		chunkSize: chunkSize,
		board:     b,
	}

	// Init the constants
	for i, node := range b.Nodes {
		if op, ok := node.Operator.(Constant); ok {
			fill(eval.values[i*chunkSize:(i+1)*chunkSize], op.Val)
		}
	}

	return eval
}

func fill[T any](dst []T, v fext.GenericFieldElem) {
	switch casted := any(dst).(type) {
	case []fext.Element:
		ev := v.GetExt()
		for i := range casted {
			casted[i] = ev
		}
	case []field.Element:
		ev, _ := v.GetBase()
		for i := range casted {
			casted[i] = ev
		}
	default:
		utils.Panic("unknown type %T", dst)
	}
}

func (e *evaluation[T]) evaluate(inputs []sv.SmartVector, chunkStart, chunkStop int) {
	switch casted := any(e).(type) {
	case *evaluation[field.Element]:
		evaluateBase(casted, inputs, chunkStart, chunkStop)
	case *evaluation[fext.Element]:
		evaluateExt(casted, inputs, chunkStart, chunkStop)
	default:
		utils.Panic("unknown type %T", casted)
	}
}

func evaluateBase(e *evaluation[field.Element], inputs []sv.SmartVector, chunkStart, chunkStop int) {
	inputCursor := 0
	chunkLen := chunkStop - chunkStart

	for i, node := range e.board.Nodes {
		offset := i * e.chunkSize
		dst := e.values[offset : offset+chunkLen]

		switch op := node.Operator.(type) {
		case Constant:
			continue
		case Variable:
			input := inputs[inputCursor]
			switch rv := input.(type) {
			case *sv.Regular:
				copy(dst, (*rv)[chunkStart:chunkStop])
			default:
				sb := input.SubVector(chunkStart, chunkStop)
				sb.WriteInSlice(dst)
			}
			inputCursor++
		case Product:
			vRes := field.Vector(dst)
			vTmp := field.Vector(e.scratch[:chunkLen])
			for k, childID := range node.Children {
				childOffset := int(childID) * e.chunkSize
				vInput := field.Vector(e.values[childOffset : childOffset+chunkLen])
				if k == 0 {
					vRes.Exp(vInput, int64(op.Exponents[k]))
				} else {
					vTmp.Exp(vInput, int64(op.Exponents[k]))
					vRes.Mul(vRes, vTmp)
				}
			}
		case LinComb:
			vRes := field.Vector(dst)
			vTmp := field.Vector(e.scratch[:chunkLen])
			var t0 field.Element
			for k, childID := range node.Children {
				childOffset := int(childID) * e.chunkSize
				vInput := field.Vector(e.values[childOffset : childOffset+chunkLen])
				coeff := op.Coeffs[k]

				if k == 0 {
					switch coeff {
					case 0:
						for j := range vRes {
							vRes[j].SetZero()
						}
					case 1:
						copy(vRes, vInput)
					case 2:
						vRes.Add(vInput, vInput)
					case -1:
						for j := range vRes {
							vRes[j].SetZero()
						}
						vRes.Sub(vRes, vInput)
					default:
						t0.SetInt64(int64(coeff))
						vRes.ScalarMul(vInput, &t0)
					}
					continue
				}

				switch coeff {
				case 0:
					continue
				case 1:
					vRes.Add(vRes, vInput)
				case 2:
					vRes.Add(vRes, vInput)
					vRes.Add(vRes, vInput)
				case -1:
					vRes.Sub(vRes, vInput)
				default:
					t0.SetInt64(int64(op.Coeffs[k]))
					vTmp.ScalarMul(vInput, &t0)
					vRes.Add(vRes, vTmp)
				}
			}
		case PolyEval:
			// result = input[0] + input[1]·x + input[2]·x² + input[3]·x³ + ...
			// i.e., ∑_{i=0}^{n} input[i]·x^i
			// input[0] is x (constant)
			// input[1] is coeff 0
			// input[2] is coeff 1 ...

			xOffset := int(node.Children[0]) * e.chunkSize
			x := e.values[xOffset] // First element of chunk is enough as it is constant

			vRes := field.Vector(dst)

			lastChildID := node.Children[len(node.Children)-1]
			lastChildOffset := int(lastChildID) * e.chunkSize
			copy(vRes, field.Vector(e.values[lastChildOffset:lastChildOffset+chunkLen]))

			for k := len(node.Children) - 2; k >= 1; k-- {
				childID := node.Children[k]
				childOffset := int(childID) * e.chunkSize
				vTmp := field.Vector(e.values[childOffset : childOffset+chunkLen])
				vRes.ScalarMul(vRes, &x)
				vRes.Add(vRes, vTmp)
			}
		default:
			utils.Panic("unknown op %T", op)
		}
	}
}

func evaluateExt(e *evaluation[fext.Element], inputs []sv.SmartVector, chunkStart, chunkStop int) {
	inputCursor := 0
	chunkLen := chunkStop - chunkStart

	for i, node := range e.board.Nodes {
		offset := i * e.chunkSize
		dst := e.values[offset : offset+chunkLen]

		switch op := node.Operator.(type) {
		case Constant:
			continue
		case Variable:
			input := inputs[inputCursor]
			// if input is rotated, SubVector is expensive so we check for that
			switch rv := input.(type) {
			case *sv.RotatedExt:
				rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, dst)
			case *sv.Rotated:
				rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, dst)
			case *sv.Regular:
				for i := 0; i < chunkLen; i++ {
					// rest of e4 should be at zero since the whole vector is a regular one
					// and we didn't mutate the input.
					dst[i].B0.A0 = (*rv)[chunkStart+i]
				}
			case *sv.RegularExt:
				copy(dst, (*rv)[chunkStart:chunkStop])
			case *sv.ConstantExt:
				for i := 0; i < chunkLen; i++ {
					dst[i].Set(&rv.Value)
				}
			default:
				sb := input.SubVector(chunkStart, chunkStop)
				sb.WriteInSliceExt(dst)
			}
			inputCursor++
		case Product:
			vRes := extensions.Vector(dst)
			vTmp := extensions.Vector(e.scratch[:chunkLen])

			childOffset0 := int(node.Children[0]) * e.chunkSize
			vInput0 := extensions.Vector(e.values[childOffset0 : childOffset0+chunkLen])
			vRes.Exp(vInput0, int64(op.Exponents[0]))

			for k := 1; k < len(node.Children); k++ {
				childID := node.Children[k]
				childOffset := int(childID) * e.chunkSize
				vInput := extensions.Vector(e.values[childOffset : childOffset+chunkLen])
				if op.Exponents[k] == 1 {
					vRes.Mul(vRes, vInput)
					continue
				}
				vTmp.Exp(vInput, int64(op.Exponents[k]))
				vRes.Mul(vRes, vTmp)
			}
		case LinComb:
			vRes := extensions.Vector(dst)
			vTmp := extensions.Vector(e.scratch[:chunkLen])
			var t0 extensions.E4
			for k, childID := range node.Children {
				childOffset := int(childID) * e.chunkSize
				vInput := extensions.Vector(e.values[childOffset : childOffset+chunkLen])
				coeff := op.Coeffs[k]

				if k == 0 {
					switch coeff {
					case 0:
						for j := range vRes {
							vRes[j].SetZero()
						}
					case 1:
						copy(vRes, vInput)
					case 2:
						vRes.Add(vInput, vInput)
					case -1:
						for j := range vRes {
							vRes[j].SetZero()
						}
						vRes.Sub(vRes, vInput)
					default:
						t0.B0.A0.SetInt64(int64(coeff))
						vRes.ScalarMul(vInput, &t0)
					}
					continue
				}

				switch coeff {
				case 0:
					continue
				case 1:
					vRes.Add(vRes, vInput)
				case 2:
					vRes.Add(vRes, vInput)
					vRes.Add(vRes, vInput)
				case -1:
					vRes.Sub(vRes, vInput)
				default:
					t0.B0.A0.SetInt64(int64(op.Coeffs[k]))
					vTmp.ScalarMul(vInput, &t0)
					vRes.Add(vRes, vTmp)
				}
			}
		case PolyEval:
			xOffset := int(node.Children[0]) * e.chunkSize
			x := e.values[xOffset]

			vRes := extensions.Vector(dst)

			lastChildID := node.Children[len(node.Children)-1]
			lastChildOffset := int(lastChildID) * e.chunkSize
			copy(vRes, extensions.Vector(e.values[lastChildOffset:lastChildOffset+chunkLen]))

			for k := len(node.Children) - 2; k >= 1; k-- {
				childID := node.Children[k]
				childOffset := int(childID) * e.chunkSize
				vTmp := extensions.Vector(e.values[childOffset : childOffset+chunkLen])
				vRes.ScalarMul(vRes, &x)
				vRes.Add(vRes, vTmp)
			}
		default:
			utils.Panic("unknown op %T", op)
		}
	}
}

type GetDegree = func(m interface{}) int

// ListVariableMetadata return the list of the metadata of the variables
// into the board. Importantly, the order in which the metadata is returned
// matches the order in which the assignments to the boarded expression must be
// provided.
func (b *ExpressionBoard) ListVariableMetadata() []Metadata {
	res := []Metadata{}
	for i := range b.Nodes {
		if vari, ok := b.Nodes[i].Operator.(Variable); ok {
			res = append(res, vari.Metadata)
		}
	}
	return res
}

// Degree returns the overall degree of the expression board. It admits a custom
// function `getDeg` which is used to assign a degree to the [Variable] leaves
// of the ExpressionBoard.
func (b *ExpressionBoard) Degree(getdeg GetDegree) int {
	if len(b.Nodes) == 0 {
		return 0
	}
	degrees := make([]int, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch v := node.Operator.(type) {
		case Constant:
			degrees[i] = 0
		case Variable:
			degrees[i] = getdeg(v.Metadata)
			inputCursor++
		default:
			childrenDeg := make([]int, len(node.Children))
			for k, childID := range node.Children {
				childrenDeg[k] = degrees[childID]
			}
			degrees[i] = node.Operator.Degree(childrenDeg)
		}
	}
	return degrees[len(b.Nodes)-1]
}

/*
GnarkEval evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEval(api frontend.API, inputs []zk.WrappedVariable) zk.WrappedVariable {
	if len(b.Nodes) == 0 {
		panic("empty board")
	}
	results := make([]zk.WrappedVariable, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch op := node.Operator.(type) {
		case Constant:
			tmp := op.Val.GetExt()
			results[i] = zk.ValueOf(tmp.B0.A0.String()) // @thomas ext or base ?
		case Variable:
			results[i] = inputs[inputCursor]
			inputCursor++
		default:
			nodeInputs := make([]zk.WrappedVariable, len(node.Children))
			for k, childID := range node.Children {
				nodeInputs[k] = results[childID]
			}
			results[i] = node.Operator.GnarkEval(api, nodeInputs)
		}
	}
	return results[len(b.Nodes)-1]
}

/*
GnarkEvalExt evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen) gnarkfext.E4Gen {
	if len(b.Nodes) == 0 {
		panic("empty board")
	}
	results := make([]gnarkfext.E4Gen, len(b.Nodes))
	inputCursor := 0

	for i, node := range b.Nodes {
		switch op := node.Operator.(type) {
		case Constant:
			results[i] = gnarkfext.NewE4Gen(op.Val.GetExt())
		case Variable:
			results[i] = inputs[inputCursor]
			inputCursor++
		default:
			nodeInputs := make([]gnarkfext.E4Gen, len(node.Children))
			for k, childID := range node.Children {
				nodeInputs[k] = results[childID]
			}
			results[i] = node.Operator.GnarkEvalExt(api, nodeInputs)
		}
	}
	return results[len(b.Nodes)-1]
}

// DumpToString is a debug utility which print out the expression in a readable
// format.
func (b *ExpressionBoard) DumpToString() string {
	res := ""
	for i, node := range b.Nodes {
		res += fmt.Sprintf("%d: (%T) %++v\n", i, node.Operator, node)
	}
	return res
}

// CountNodes returns the node count of the expression
func (b *ExpressionBoard) CountNodes() int {
	return len(b.Nodes)
}
