package symbolic

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempoolext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"sync"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Evaluate the board for a batch  of inputs in parallel
func (b *ExpressionBoard) EvaluateExt(inputs []sv.SmartVector, p ...mempoolext.MemPool) sv.SmartVector {

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

	if len(p) > 0 && p[0].Size() != MaxChunkSize {
		utils.Panic("the pool should be a pool of vectors of size=%v but it is %v", MaxChunkSize, p[0].Size())
	}

	/*
		The is no vector input iff totalSize is 0
		Thus the condition below catch the two cases where:
			- There is no input vectors (scalars only)
			- The vectors are smaller than the min chunk size
	*/

	if totalSize < MaxChunkSize {
		// never pass the pool here as the pool assumes that all vectors have a
		// size of MaxChunkSize. Thus, it would not work here.
		return b.evaluateSingleThread(inputs).DeepCopy()
	}

	// This is the code-path that is used for benchmarking when the size of the
	// vectors is exactly MaxChunkSize. In production, it will rather use the
	// multi-threaded option. The above condition cannot use the pool because we
	// assume here that the pool has a vector size of exactly MaxChunkSize.
	if totalSize == MaxChunkSize {
		return b.evaluateSingleThreadExt(inputs, p...).DeepCopy()
	}

	if totalSize%MaxChunkSize != 0 {
		utils.Panic("Unsupported : totalSize %v is not divided by the chunk size %v", totalSize, MaxChunkSize)
	}

	numChunks := totalSize / MaxChunkSize
	res := make([]fext.Element, totalSize)

	parallel.ExecuteFromChan(numChunks, func(wg *sync.WaitGroup, id *parallel.AtomicCounter) {

		var pool []mempoolext.MemPool
		if len(p) > 0 {
			if _, ok := p[0].(*mempoolext.DebuggeableCall); !ok {
				pool = append(pool, mempoolext.WrapsWithMemCache(p[0]))
			}
		}

		chunkInputs := make([]sv.SmartVector, len(inputs))

		for {
			chunkID, ok := id.Next()
			if !ok {
				break
			}

			var (
				chunkStart = chunkID * MaxChunkSize
				chunkStop  = (chunkID + 1) * MaxChunkSize
			)

			for i, inp := range inputs {
				chunkInputs[i] = inp.SubVector(chunkStart, chunkStop)
				// Sanity-check : the output of SubVector must have the correct length
				if chunkInputs[i].Len() != chunkStop-chunkStart {
					utils.Panic("Subvector failed, subvector should have size %v but size is %v", chunkStop-chunkStart, chunkInputs[i].Len())
				}
			}

			// We don't parallelize evaluations where the inputs are all scalars
			// Therefore the cast is safe.
			chunkRes := b.evaluateSingleThreadExt(chunkInputs, pool...)

			// No race condition here as each call write to different places
			// of vec.
			chunkRes.WriteInSliceExt(res[chunkStart:chunkStop])

			wg.Done()
		}

		if len(p) > 0 {
			if sa, ok := pool[0].(*mempoolext.SliceArena); ok {
				sa.TearDown()
			}
		}
	})

	return smartvectorsext.NewRegularExt(res)
}

// evaluateSingleThread evaluates a boarded expression. The inputs can be either
// vector or scalars. The vector's input length should be smaller than a chunk.
func (b *ExpressionBoard) evaluateSingleThreadExt(inputs []sv.SmartVector, p ...mempoolext.MemPool) sv.SmartVector {

	var (
		length         = inputs[0].Len()
		pool, hasPool  = mempoolext.ExtractCheckOptionalSoft(length, p...)
		nodeAssignment = b.prepareNodeAssignments(inputs)
	)

	// Then computes the levels one by one
	for level := 1; level < len(nodeAssignment); level++ {
		for pil := range nodeAssignment[level] {

			node := &nodeAssignment[level][pil]
			nodeAssignment.evalExt(node, pool)
		}
	}

	// Assertion, the last level contains only one node (the final result)
	if len(nodeAssignment[len(nodeAssignment)-1]) > 1 {
		panic("multiple heads")
	}

	resBuf := nodeAssignment[len(b.Nodes)-1][0].Value

	if resBuf == nil {
		panic("resbuf is nil")
	}

	// Deep-copy the last node and put resBuf back in the pool. It's cleanier
	// that way.
	if hasPool {
		if reg, ok := resBuf.(*smartvectorsext.PooledExt); ok {
			resGC := reg.DeepCopy()
			reg.Free(pool)
			resBuf = resGC
		}
	}

	return resBuf
}

func (b *ExpressionBoard) EvaluateMixed(inputs []sv.SmartVector, p ...mempool.GenericMemPool) sv.SmartVector {
	basePools := []mempool.MemPool{}
	extPools := []mempoolext.MemPool{}
	// separate the pools into base and extension pools
	for _, pool := range p {
		if extPool, ok := pool.(mempoolext.MemPool); ok {
			// we have an extension pool
			extPools = append(extPools, extPool)
		} else {
			// we have a base pool
			basePools = append(basePools, pool.(mempool.MemPool))
		}
	}
	vectorsBase := make([]sv.SmartVector, 0, len(inputs))
	vectorsExt := make([]sv.SmartVector, 0, len(inputs))
	for _, input := range inputs {
		if _, isBaseError := input.GetBase(0); isBaseError == nil {
			// we have a base vector
			vectorsBase = append(vectorsBase, input)
		} else {
			// we have an extension vector
			vectorsExt = append(vectorsExt, input)
		}
	}
	// perform a size check between base and extension vectors
	// compatibility of the  sizes inside the base set and extension set are
	// checked inside the Evaluate function
	if len(vectorsBase) > 0 && len(vectorsExt) > 0 &&
		vectorsBase[0].Len() != vectorsExt[0].Len() {
		utils.Panic("base and extension vectors have different sizes")
	}
	var resBase sv.SmartVector
	var resExt sv.SmartVector
	// evaluate the base vectors
	if len(vectorsBase) > 0 {
		// we have base vectors
		// evaluate the base vectors
		// use the base pools
		resBase := b.Evaluate(vectorsBase, basePools...)
	}
	// evaluate the extension vectors
	if len(vectorsExt) > 0 {
		// we have extension vectors
		// evaluate the extension vectors
		// use the extension pools
		resExt := b.EvaluateExt(vectorsExt, extPools...)
	}
	var res sv.SmartVector = sv.NewConstant(field.Zero(), inputs[0].Len())
	if resBase != nil {
		res = sv.Add(res, resBase)
	}
	if resExt != nil {
		res = sv.Add(res, resExt)
	}
	return res
}
