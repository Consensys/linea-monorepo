package symbolic

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/arena"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"

	"github.com/consensys/gnark/frontend"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
)

// evaluation is a helper to evaluate an expression board on chunks of data.
type evaluation[T any] struct {
	nodes       [][]evaluationNode[T]
	scratch     []T
	chunkSize   int
	vectorArena *arena.VectorArena // contiguous memory to store all the vectors
}

// evaluationNode is a node in the evaluation graph.
type evaluationNode[T any] struct {
	value      []T      // value of the node after evaluation
	inputs     [][]T    // inputs values (they are slices with underlying memory in the arena)
	op         Operator // op(inputs) = value
	hasValue   bool     // true if value is valid
	isConstant bool     // true if the node is a constant
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

				eval.reset(inputs, chunkStart, chunkStop, b)
				eval.evaluate()

				copy(res[chunkStart:chunkStop], eval.nodes[len(b.Nodes)-1][0].value[:])
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

			eval.reset(inputs, chunkStart, chunkStop, b)
			eval.evaluate()

			copy(res[chunkStart:chunkStop], eval.nodes[len(b.Nodes)-1][0].value[:])
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
		scratch:     make([]T, chunkSize),
		chunkSize:   chunkSize,
		vectorArena: arena.NewVectorArena[T](b.CountNodes() * chunkSize),
		nodes:       make([][]evaluationNode[T], len(b.Nodes)),
	}
	for lvl := range eval.nodes {
		eval.nodes[lvl] = make([]evaluationNode[T], len(b.Nodes[lvl]))
	}

	// Init the constants and inputs.
	for i := range b.Nodes {
		for j := range b.Nodes[i] {
			na := &eval.nodes[i][j]
			node := &b.Nodes[i][j]
			na.op = node.Operator
			na.hasValue = false
			na.isConstant = false
			na.inputs = make([][]T, len(node.Children))
			na.value = arena.Get[T](eval.vectorArena, chunkSize)

			switch op := node.Operator.(type) {
			case Constant:
				// The constants are identified to constant vectors
				na.isConstant = true
				na.hasValue = true
				fill(na.value, op.Val)
				if len(node.Children) != 0 {
					panic("constant with children")
				}
			}

			// set the inputs pointers
			for k := range node.Children {
				id := node.Children[k]
				na.inputs[k] = eval.nodes[id.level()][id.posInLevel()].value
			}
		}
	}

	// we can propagate constants here, since it will be useful for all chunks and done only once.
	// starting from level 1 since level 0 are inputs/constants,
	// if all my inputs are constant, I can compute my value, and be a constant too.
	// but in practice it never happens so we omit this part.
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

func (e *evaluation[T]) reset(inputs []sv.SmartVector, chunkStart, chunkStop int, b *ExpressionBoard) {
	inputCursor := 0
	chunkLen := chunkStop - chunkStart // can be < MaxChunkSize

	// Init the constants and inputs.
	for i := range b.Nodes {
		for j := range b.Nodes[i] {
			na := &e.nodes[i][j]
			if !na.isConstant {
				na.hasValue = false
			}

			if i == 0 {
				switch na.op.(type) {
				case Variable:
					input := inputs[inputCursor]
					switch casted := any(&na.value).(type) {
					case *[]field.Element:
						switch rv := input.(type) {
						case *sv.Regular:
							copy((*casted)[:chunkLen], (*rv)[chunkStart:chunkStop])
						default:
							sb := input.SubVector(chunkStart, chunkStop)
							sb.WriteInSlice((*casted)[:chunkLen])
						}
					case *[]fext.Element:

						// if input is rotated, SubVector is expensive so we check for that
						switch rv := input.(type) {
						case *sv.RotatedExt:
							rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, (*casted)[:chunkLen])
						case *sv.Rotated:
							rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, (*casted)[:chunkLen])
						case *sv.Regular:
							for i := 0; i < chunkLen; i++ {
								// rest of e4 should be at zero since the whole vector is a regular one
								// and we didn't mutate the input.
								(*casted)[i].B0.A0 = (*rv)[chunkStart+i]
							}
						case *sv.RegularExt:
							copy((*casted)[:chunkLen], (*rv)[chunkStart:chunkStop])
						case *sv.ConstantExt:
							for i := 0; i < chunkLen; i++ {
								(*casted)[i].Set(&rv.Value)
							}
						default:
							sb := input.SubVector(chunkStart, chunkStop)
							sb.WriteInSliceExt((*casted)[:chunkLen])
						}
					}

					inputCursor++
					na.hasValue = true
				}
			}
		}
	}
}

func (e *evaluation[T]) evaluate() {
	// Evaluate values level by level
	switch casted := any(e).(type) {
	case *evaluation[field.Element]:
		evaluateBase(casted)
	case *evaluation[fext.Element]:
		evaluateExt(casted)
	default:
		utils.Panic("unknown type %T", casted)
	}
}

func evaluateBase(s *evaluation[field.Element]) {
	// level 0 are inputs/constants
	// we start at level 1
	for level := 1; level < len(s.nodes); level++ {
		for pil := range s.nodes[level] {
			evalNodeBase(s, &s.nodes[level][pil])
		}
	}
}

func evaluateExt(s *evaluation[fext.Element]) {
	// level 0 are inputs/constants
	// we start at level 1
	for level := 1; level < len(s.nodes); level++ {
		for pil := range s.nodes[level] {
			evalNodeExt(s, &s.nodes[level][pil])
		}
	}
}

// evalNodeBase and evalNodeExt are specialized and identical; could use function operator
// to make more generic.

func evalNodeExt(solver *evaluation[fext.Element], na *evaluationNode[fext.Element]) {
	if na.hasValue {
		return
	}

	vRes := extensions.Vector(na.value[:])
	vTmp := extensions.Vector(solver.scratch[:])

	switch op := na.op.(type) {
	case Product:
		vRes.Exp(na.inputs[0][:], int64(op.Exponents[0]))
		for i := 1; i < len(na.inputs); i++ {
			vInput := extensions.Vector(na.inputs[i][:])
			if op.Exponents[i] == 1 {
				// common case
				vRes.Mul(vRes, vInput)
				continue
			}
			vTmp.Exp(vInput, int64(op.Exponents[i]))
			vRes.Mul(vRes, vTmp)
		}
	case LinComb:
		var t0 extensions.E4
		for i := range na.inputs {
			vInput := extensions.Vector(na.inputs[i][:])
			coeff := op.Coeffs[i]

			if i == 0 {
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
				t0.B0.A0.SetInt64(int64(op.Coeffs[i]))
				vTmp.ScalarMul(vInput, &t0)
				vRes.Add(vRes, vTmp)
			}
		}
	case PolyEval:
		// result = input[0] + input[1]·x + input[2]·x² + input[3]·x³ + ...
		// i.e., ∑_{i=0}^{n} input[i]·x^i
		x := na.inputs[0][0] // we assume that the first input is always a constant
		copy(vRes, extensions.Vector(na.inputs[len(na.inputs)-1][:]))

		for i := len(na.inputs) - 2; i >= 1; i-- {
			vInput := extensions.Vector(na.inputs[i][:])
			vRes.ScalarMul(vRes, &x)
			vRes.Add(vRes, vInput)
		}
	default:
		utils.Panic("unknown op %T", na.op)
	}

	na.hasValue = true

}

func evalNodeBase(solver *evaluation[field.Element], na *evaluationNode[field.Element]) {
	if na.hasValue {
		return
	}

	vRes := field.Vector(na.value[:])
	vTmp := field.Vector(solver.scratch[:])

	switch op := na.op.(type) {
	case Product:
		for i := range na.inputs {
			vInput := field.Vector(na.inputs[i][:])
			if i == 0 {
				vRes.Exp(vInput, int64(op.Exponents[i]))
			} else {
				vTmp.Exp(vInput, int64(op.Exponents[i]))
				vRes.Mul(vRes, vTmp)
			}
		}
	case LinComb:
		var t0 field.Element
		for i := range na.inputs {
			vInput := field.Vector(na.inputs[i][:])
			coeff := op.Coeffs[i]

			if i == 0 {
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
				t0.SetInt64(int64(op.Coeffs[i]))
				vTmp.ScalarMul(vInput, &t0)
				vRes.Add(vRes, vTmp)
			}
		}
	case PolyEval:
		// result = input[0] + input[1]·x + input[2]·x² + input[3]·x³ + ...
		// i.e., ∑_{i=0}^{n} input[i]·x^i
		x := na.inputs[0][0] // we assume that the first input is always a constant
		copy(vRes, field.Vector(na.inputs[len(na.inputs)-1][:]))

		for i := len(na.inputs) - 2; i >= 1; i-- {
			vTmp := field.Vector(na.inputs[i][:])
			vRes.ScalarMul(vRes, &x)
			vRes.Add(vRes, vTmp)
		}
	default:
		utils.Panic("unknown op %T", na.op)
	}

	na.hasValue = true

}

type GetDegree = func(m interface{}) int

// ListVariableMetadata return the list of the metadata of the variables
// into the board. Importantly, the order in which the metadata is returned
// matches the order in which the assignments to the boarded expression must be
// provided.
func (b *ExpressionBoard) ListVariableMetadata() []Metadata {
	res := []Metadata{}
	for i := range b.Nodes[0] {
		if vari, ok := b.Nodes[0][i].Operator.(Variable); ok {
			res = append(res, vari.Metadata)
		}
	}
	return res
}

// Degree returns the overall degree of the expression board. It admits a custom
// function `getDeg` which is used to assign a degree to the [Variable] leaves
// of the ExpressionBoard.
func (b *ExpressionBoard) Degree(getdeg GetDegree) int {

	/*
		First, build a buffer to store the intermediate results
	*/
	intermediateRes := make([][]int, len(b.Nodes))
	for level := range b.Nodes {
		intermediateRes[level] = make([]int, len(b.Nodes[level]))
	}

	/*
		Then, store the initial values in the level entries of the vector
	*/
	inputCursor := 0

	for i, node := range b.Nodes[0] {
		switch v := node.Operator.(type) {
		case Constant:
			intermediateRes[0][i] = 0
		case Variable:
			intermediateRes[0][i] = getdeg(v.Metadata)
			inputCursor++
		}
	}

	/*
		Then computes the levels one by one
	*/
	for level := 1; level < len(b.Nodes); level++ {
		for pos, node := range b.Nodes[level] {
			/*
				Collect the inputs of the current node from the intermediateRes
			*/
			childrenDeg := make([]int, len(node.Children))
			for i, childID := range node.Children {
				childrenDeg[i] = intermediateRes[childID.level()][childID.posInLevel()]
			}

			/*
				Run the evaluation
			*/
			nodeDeg := node.Operator.Degree(childrenDeg)

			/*
				Registers the result in the intermediate results
			*/
			intermediateRes[level][pos] = nodeDeg
		}
	}

	// Deep-copy the result from the last level (which assumedly contains only one node)
	if len(intermediateRes[len(intermediateRes)-1]) > 1 {
		panic("multiple heads")
	}
	return intermediateRes[len(b.Nodes)-1][0]
}

/*
GnarkEval evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEval(api frontend.API, inputs []zk.WrappedVariable) zk.WrappedVariable {

	/*
		First, build a buffer to store the intermediate results
	*/
	intermediateRes := make([][]zk.WrappedVariable, len(b.Nodes))
	for level := range b.Nodes {
		intermediateRes[level] = make([]zk.WrappedVariable, len(b.Nodes[level]))
	}

	/*
		Then, store the initial values in the level entries of the vector
	*/
	inputCursor := 0

	for i := range b.Nodes[0] {
		switch op := b.Nodes[0][i].Operator.(type) {
		case Constant:
			tmp := op.Val.GetExt()
			intermediateRes[0][i] = zk.ValueOf(tmp) // @thomas ext or base ?
		case Variable:
			intermediateRes[0][i] = inputs[inputCursor]
			inputCursor++
		}
	}

	/*
		Then computes the levels one by one
	*/
	for level := 1; level < len(b.Nodes); level++ {
		for pos, node := range b.Nodes[level] {
			/*
				Collect the inputs of the current node from the intermediateRes
			*/
			nodeInputs := make([]zk.WrappedVariable, len(node.Children))
			for i, childID := range node.Children {
				nodeInputs[i] = intermediateRes[childID.level()][childID.posInLevel()]
			}

			/*
				Run the evaluation
			*/
			res := node.Operator.GnarkEval(api, nodeInputs)

			/*
				Registers the result in the intermediate results
			*/
			intermediateRes[level][pos] = res
		}
	}

	// Deep-copy the result from the last level (which assumedly contains only one node)
	if len(intermediateRes[len(intermediateRes)-1]) > 1 {
		panic("multiple heads")
	}
	return intermediateRes[len(b.Nodes)-1][0]

}

/*
GnarkEvalExt evaluates the expression in a gnark circuit
*/
func (b *ExpressionBoard) GnarkEvalExt(api frontend.API, inputs []gnarkfext.E4Gen) gnarkfext.E4Gen {

	/*
		First, build a buffer to store the intermediate results
	*/
	intermediateRes := make([][]gnarkfext.E4Gen, len(b.Nodes))
	for level := range b.Nodes {
		intermediateRes[level] = make([]gnarkfext.E4Gen, len(b.Nodes[level]))
	}

	/*
		Then, store the initial values in the level entries of the vector
	*/
	inputCursor := 0

	for i := range b.Nodes[0] {
		switch op := b.Nodes[0][i].Operator.(type) {
		case Constant:
			intermediateRes[0][i] = gnarkfext.NewE4Gen(op.Val.GetExt())
		case Variable:
			intermediateRes[0][i] = inputs[inputCursor]
			inputCursor++
		}
	}

	/*
		Then computes the levels one by one
	*/
	for level := 1; level < len(b.Nodes); level++ {
		for pos, node := range b.Nodes[level] {
			/*
				Collect the inputs of the current node from the intermediateRes
			*/
			nodeInputs := make([]gnarkfext.E4Gen, len(node.Children))
			for i, childID := range node.Children {
				nodeInputs[i] = intermediateRes[childID.level()][childID.posInLevel()]
			}

			/*
				Run the evaluation
			*/
			res := node.Operator.GnarkEvalExt(api, nodeInputs)

			/*
				Registers the result in the intermediate results
			*/
			intermediateRes[level][pos] = res
		}
	}

	// Deep-copy the result from the last level (which assumedly contains only one node)
	if len(intermediateRes[len(intermediateRes)-1]) > 1 {
		panic("multiple heads")
	}
	return intermediateRes[len(b.Nodes)-1][0]

}

// DumpToString is a debug utility which print out the expression in a readable
// format.
func (b *ExpressionBoard) DumpToString() string {
	res := ""
	for i := range b.Nodes {
		res += fmt.Sprintf("LEVEL %v\n", i)
		for _, node := range b.Nodes[i] {
			res += fmt.Sprintf("\t(%T) %++v\n", node.Operator, node)
		}
	}
	return res
}

// CountNodes returns the node count of the expression
func (b *ExpressionBoard) CountNodes() int {
	res := 0
	for i := 0; i < len(b.Nodes); i++ {
		res += len(b.Nodes[i])
	}
	return res
}
