package symbolic

import (
	"fmt"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/gnark/frontend"
)

const (
	// It's a fine-tuned by hand number. Smaller is slower, larger comsumes for more memory.
	// For testing, it can be overriden even though it will become flakky if we have several
	// tests.
	MAX_CHUNK_SIZE int = 1 << 9
)

type GetDegree = func(m interface{}) int

/*
ExpressionBoard is a shared space for defining expressions. Several expressions
can use the same board. Ideally, all expressions uses the same common board.
*/
type ExpressionBoard struct {
	// Topologically sorted list of (deduplicated) nodes
	// The nodes are sorted by level : e.g the first entry corresponds to
	// the leaves and the last one (which should be unique). To the root.
	Nodes [][]Node
	// Maps nodes to their level in the DAG structure. The 32 MSB bits
	// of the ID indicates the level and the LSB bits indicates the position
	// in the level.
	ESHashesToPos map[field.Element]NodeID
}

// Empty board
func EmptyBoard() ExpressionBoard {
	return ExpressionBoard{
		Nodes:         [][]Node{},
		ESHashesToPos: map[field.Element]NodeID{},
	}
}

// Returns a node from an ID
func (b *ExpressionBoard) GetNode(id NodeID) *Node {
	return &b.Nodes[id.Level()][id.PosInLevel()]
}

// Returns a node from an ESH and nil if not found
func (b *ExpressionBoard) NodeFromESH(esh field.Element) *Node {
	nodeid, ok := b.ESHashesToPos[esh]
	if !ok {
		return nil
	}
	return b.GetNode(nodeid)
}

// List the metadata of the variables into the board
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
func (b *ExpressionBoard) Evaluate(inputs []sv.SmartVector) sv.SmartVector {

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

	if totalSize <= MAX_CHUNK_SIZE {
		return b.evaluateSingleThread(inputs)
	}

	if totalSize%MAX_CHUNK_SIZE != 0 {
		utils.Panic("Unsupported : totalSize %v is not divided by the chunk size %v", totalSize, MAX_CHUNK_SIZE)
	}

	numChunks := totalSize / MAX_CHUNK_SIZE
	res := make([]field.Element, totalSize)

	parallel.ExecuteChunky(numChunks, func(start, stop int) {
		for chunkID := start; chunkID < stop; chunkID++ {

			chunkStart := chunkID * MAX_CHUNK_SIZE
			chunkStop := (chunkID + 1) * MAX_CHUNK_SIZE
			chunkInputs := make([]sv.SmartVector, len(inputs))

			for i, inp := range inputs {
				chunkInputs[i] = inp.SubVector(chunkStart, chunkStop)
				// Sanity-check : the output of SubVector must have the correct length
				if chunkInputs[i].Len() != chunkStop-chunkStart {
					utils.Panic("subvector failed, subvector should have size %v but size is %v", chunkStop-chunkStart, chunkInputs[i].Len())
				}
			}

			// We don't parallelize evaluations where the inputs are all scalars
			// Therefore the cast is safe.
			chunkRes := b.evaluateSingleThread(chunkInputs)

			// No race condition here as each call write to different places
			// of vec.
			chunkRes.WriteInSlice(res[chunkStart:chunkStop])
		}
	})

	return sv.NewRegular(res)
}

// Evaluates a boarded expression. The inputs can be either vector or scalars
// The vector's input length should be smaller than a chunk
func (b *ExpressionBoard) evaluateSingleThread(inputs []sv.SmartVector) sv.SmartVector {

	// At this point, we are already guaranteed that
	length := inputs[0].Len()

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
				utils.Panic("subvector failed, subvector should have size %v but size is %v", length, inputs[inputCursor].Len())
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
				lvl, pil := childID.Level(), childID.PosInLevel()
				nodeInputs[i] = intermediateRes[lvl][pil]
				parentCount[lvl][pil]--
				if parentCount[lvl][pil] == 0 {
					intermediateRes[lvl][pil] = nil // free the children
				}
			}

			/*
				Run the evaluation
			*/
			res := node.Operator.Evaluate(nodeInputs)
			if res.Len() != length {
				utils.Panic("subvector failed, subvector should have size %v but size is %v", length, inputs[inputCursor].Len())
			}

			/*
				Registers the result in the intermediate results
			*/
			intermediateRes[level][pos] = res
			parentCount[level][pos] = len(node.Parents)
		}
	}

	// Deep-copy the result from the last level (which assumedly contains only one node)
	if len(intermediateRes[len(intermediateRes)-1]) > 1 {
		panic("multiple heads")
	}

	resBuf := intermediateRes[len(b.Nodes)-1][0]
	return resBuf
}

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
				childrenDeg[i] = intermediateRes[childID.Level()][childID.PosInLevel()]
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
Evaluates the expression in a gnark circuit
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
				nodeInputs[i] = intermediateRes[childID.Level()][childID.PosInLevel()]
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
