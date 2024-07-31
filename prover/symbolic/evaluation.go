package symbolic

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

const (
	// MaxChunkSize is a fine-tuned by hand number. Smaller is slower, larger
	// comsumes for more memory. For testing, it can be overriden even though
	// it will become flakky if we have several tests.
	MaxChunkSize int = 1 << 9
)

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

// Evaluate the board for a batch  of inputs in parallel
func (b *ExpressionBoard) Evaluate(inputs []sv.SmartVector, p ...*mempool.Pool) sv.SmartVector {

	/*
		Find the size of the vector
	*/
	totalSize := 0
	for i, inp := range inputs {

		if totalSize > 0 && totalSize != inp.Len() {
			// Expects that all vector inputs have the same size
			utils.Panic("Mismatch in the size: len(v) %v, totalsize %v, pos %v", inp.Len(), totalSize, i)
		}
		if totalSize == 0 {
			totalSize = inp.Len()
		}
	}

	if totalSize == 0 {
		// Either, there is no input or the inputs are empty. Both
		// cases are panic
		utils.Panic("Either there is no input or the inputs all have size 0")
	}

	/*
		The is no vector input iff totalSize is 0
		Thus the condition below catch the two cases where:
			- There is no input vectors (scalars only)
			- The vectors are smaller than the min chunk size
	*/

	if totalSize <= MaxChunkSize {
		// never pass the pool here
		return b.evaluateSingleThread(inputs)
	}

	if totalSize%MaxChunkSize != 0 {
		utils.Panic("Unsupported : totalSize %v is not divided by the chunk size %v", totalSize, MaxChunkSize)
	}

	numChunks := totalSize / MaxChunkSize
	res := make([]field.Element, totalSize)

	parallel.ExecuteChunky(numChunks, func(start, stop int) {
		for chunkID := start; chunkID < stop; chunkID++ {

			chunkStart := chunkID * MaxChunkSize
			chunkStop := (chunkID + 1) * MaxChunkSize
			chunkInputs := make([]sv.SmartVector, len(inputs))

			for i, inp := range inputs {
				chunkInputs[i] = inp.SubVector(chunkStart, chunkStop)
				// Sanity-check : the output of SubVector must have the correct length
				if chunkInputs[i].Len() != chunkStop-chunkStart {
					utils.Panic("Subvector failed, subvector should have size %v but size is %v", chunkStop-chunkStart, chunkInputs[i].Len())
				}
			}

			// We don't parallelize evaluations where the inputs are all scalars
			// Therefore the cast is safe.
			chunkRes := b.evaluateSingleThread(chunkInputs, p...)

			// No race condition here as each call write to different places
			// of vec.
			chunkRes.WriteInSlice(res[chunkStart:chunkStop])
		}
	})

	return sv.NewRegular(res)
}

// evaluateSingleThread evaluates a boarded expression. The inputs can be either
// vector or scalars. The vector's input length should be smaller than a chunk.
func (b *ExpressionBoard) evaluateSingleThread(inputs []sv.SmartVector, p ...*mempool.Pool) sv.SmartVector {

	length := inputs[0].Len()
	pool, hasPool := mempool.ExtractCheckOptionalSoft(length, p...)

	/*
		First, build a buffer to store the intermediate results
	*/
	intermediateRes := make([][]sv.SmartVector, len(b.Nodes))
	parentCount := make([][]int, len(b.Nodes))
	for level := range b.Nodes {
		intermediateRes[level] = make([]sv.SmartVector, len(b.Nodes[level]))
		parentCount[level] = make([]int, len(b.Nodes[level]))
	}

	/*
		Then, store the initial values in the level entries of the vector
	*/
	inputCursor := 0

	for i := range b.Nodes[0] {
		switch op := b.Nodes[0][i].Operator.(type) {
		case Constant:
			// The constants are identified to constant vectors
			intermediateRes[0][i] = sv.NewConstant(op.Val, length)
			parentCount[0][i] = -1 // that way it never reaches zero
		case Variable:
			// Sanity-check the input should have the correct length
			if inputs[inputCursor].Len() != length {
				utils.Panic("Subvector failed, subvector should have size %v but size is %v", length, inputs[inputCursor].Len())
			}
			intermediateRes[0][i] = inputs[inputCursor]
			// track the number of time the node will be used
			parentCount[0][i] = len(b.Nodes[0][i].Parents)
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
			nodeInputs := make([]sv.SmartVector, len(node.Children))
			for i, childID := range node.Children {
				lvl, pil := childID.level(), childID.posInLevel()
				nodeInputs[i] = intermediateRes[lvl][pil]
			}

			/*
				Run the evaluation
			*/
			res := node.Operator.Evaluate(nodeInputs, pool)
			if res.Len() != length {
				utils.Panic("Subvector failed, subvector should have size %v but size is %v", length, inputs[inputCursor].Len())
			}

			// If the pool is used, free the childre
			for i, childID := range node.Children {
				lvl, pil := childID.level(), childID.posInLevel()

				// bar the node input to used again. We won't need it.
				nodeInputs[i] = nil

				// beside, if all its parents have been computed. We can
				// remove it from the intermediate result (and possibly
				// pool it).
				parentCount[lvl][pil]--
				if parentCount[lvl][pil] == 0 {

					reg, isRegular := intermediateRes[lvl][pil].(*sv.Regular)
					// remove the pooled slice from the structure to ensure
					// it is not used anymore.
					intermediateRes[lvl][pil] = nil

					// And pool it if relevant
					if hasPool && isRegular && lvl > 0 {
						v := []field.Element(*reg)
						pool.Free(&v)
					}
				}
			}

			/*
				Registers the result in the intermediate results
			*/
			intermediateRes[level][pos] = res
			parentCount[level][pos] = len(node.Parents)
		}
	}

	// Assertion, the last level contains only one node (the final result)
	if len(intermediateRes[len(intermediateRes)-1]) > 1 {
		panic("multiple heads")
	}

	resBuf := intermediateRes[len(b.Nodes)-1][0]

	// Deep-copy the last node and put resBuf back in the pool. It's cleanier
	// that way.
	if hasPool {
		if reg, ok := resBuf.(*sv.Regular); ok {
			resGC := reg.DeepCopy()
			v := []field.Element(*reg)
			pool.Free(&v)
			resBuf = resGC
		}
	}

	return resBuf
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
func (b *ExpressionBoard) GnarkEval(api frontend.API, inputs []frontend.Variable) frontend.Variable {

	/*
		First, build a buffer to store the intermediate results
	*/
	intermediateRes := make([][]frontend.Variable, len(b.Nodes))
	for level := range b.Nodes {
		intermediateRes[level] = make([]frontend.Variable, len(b.Nodes[level]))
	}

	/*
		Then, store the initial values in the level entries of the vector
	*/
	inputCursor := 0

	for i := range b.Nodes[0] {
		switch op := b.Nodes[0][i].Operator.(type) {
		case Constant:
			intermediateRes[0][i] = op.Val
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
			nodeInputs := make([]frontend.Variable, len(node.Children))
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

// CountNodes returns the node count of the expression, without accounting for
// the level zero.
func (b *ExpressionBoard) CountNodes() int {
	res := 0
	for i := 1; i < len(b.Nodes); i++ {
		res += len(b.Nodes[i])
	}
	return res
}
