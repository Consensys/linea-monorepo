package symbolic

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Evaluate the board for a batch  of inputs in parallel
func (b *ExpressionBoard) EvaluateExt(inputs []sv.SmartVector, p ...mempool.MemPool) sv.SmartVector {

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

		var pool []mempool.MemPool
		if len(p) > 0 {
			if _, ok := p[0].(*mempool.DebuggeableCall); !ok {
				pool = append(pool, mempool.WrapsWithMemCache(p[0]))
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
			if sa, ok := pool[0].(*mempool.SliceArena); ok {
				sa.TearDown()
			}
		}
	})

	return sv.NewRegularExt(res)
}

// evaluateSingleThread evaluates a boarded expression. The inputs can be either
// vector or scalars. The vector's input length should be smaller than a chunk.
func (b *ExpressionBoard) evaluateSingleThreadExt(inputs []sv.SmartVector, p ...mempool.MemPool) sv.SmartVector {

	var (
		length         = inputs[0].Len()
		pool, hasPool  = mempool.ExtractCheckOptionalSoft(length, p...)
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
		if reg, ok := resBuf.(*sv.PooledExt); ok {
			resGC := reg.DeepCopy()
			reg.Free(pool)
			resBuf = resGC
		}
	}

	return resBuf
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
		return b.evaluateSingleThreadMixed(inputs).DeepCopy()
	}

	// This is the code-path that is used for benchmarking when the size of the
	// vectors is exactly MaxChunkSize. In production, it will rather use the
	// multi-threaded option. The above condition cannot use the pool because we
	// assume here that the pool has a vector size of exactly MaxChunkSize.
	if totalSize == MaxChunkSize {
		fmt.Printf("doing single threaded mixed\n")
		return b.evaluateSingleThreadMixed(inputs, p...).DeepCopy()
	}

	if totalSize%MaxChunkSize != 0 {
		utils.Panic("Unsupported : totalSize %v is not divided by the chunk size %v", totalSize, MaxChunkSize)
	}

	numChunks := totalSize / MaxChunkSize

	// resIsBase := sv.AreAllBase(inputs)
	// if resIsBase {
	// 	fmt.Printf("doing base vectors, totalSize=%v, numChunks=%v\n", totalSize, numChunks)
	// 	// we are dealing with base vectors
	// 	res := make([]field.Element, totalSize)
	// 	parallel.ExecuteFromChan(numChunks, func(wg *sync.WaitGroup, id *parallel.AtomicCounter) {

	// 		var pool []mempool.MemPool
	// 		if len(p) > 0 {
	// 			if _, ok := p[0].(*mempool.DebuggeableCall); !ok {
	// 				pool = append(pool, mempool.WrapsWithMemCache(p[0]))
	// 			}
	// 		}

	// 		chunkInputs := make([]sv.SmartVector, len(inputs))

	// 		for {
	// 			chunkID, ok := id.Next()
	// 			if !ok {
	// 				break
	// 			}

	// 			var (
	// 				chunkStart = chunkID * MaxChunkSize
	// 				chunkStop  = (chunkID + 1) * MaxChunkSize
	// 			)

	// 			for i, inp := range inputs {
	// 				chunkInputs[i] = inp.SubVector(chunkStart, chunkStop)
	// 				// Sanity-check : the output of SubVector must have the correct length
	// 				if chunkInputs[i].Len() != chunkStop-chunkStart {
	// 					utils.Panic("Subvector failed, subvector should have size %v but size is %v", chunkStop-chunkStart, chunkInputs[i].Len())
	// 				}
	// 			}

	// 			// We don't parallelize evaluations where the inputs are all scalars
	// 			// Therefore the cast is safe.
	// 			chunkRes := b.evaluateSingleThreadMixed(chunkInputs, pool...)

	// 			// No race condition here as each call write to different places
	// 			// of vec.
	// 			chunkRes.WriteInSlice(res[chunkStart:chunkStop])

	// 			wg.Done()
	// 		}

	// 		if len(p) > 0 {
	// 			if sa, ok := pool[0].(*mempool.SliceArena); ok {
	// 				sa.TearDown()
	// 			}
	// 		}
	// 	})
	// 	return sv.NewRegular(res)
	// } else {
	nbNodes := 0
	for _, level := range b.Nodes {
		nbNodes += len(level)
	}
	fmt.Printf("doing ext vectors, totalSize=%v, numChunks=%v, nbNodes=%v\n", totalSize, numChunks, nbNodes)

	// The result is an extension vector
	res := make([]fext.Element, totalSize)
	parallel.Execute(numChunks, func(start, stop int) {

		// chunkInputs := make([]sv.SmartVector, len(inputs))
		ba := b.initNodeAssignmentsGautam(inputs)
		for chunkID := start; chunkID < stop; chunkID++ {
			// chunkID, ok := id.Next()
			// if !ok {
			// 	break
			// }

			var (
				chunkStart = chunkID * MaxChunkSize
				chunkStop  = (chunkID + 1) * MaxChunkSize
			)

			// for i, inp := range inputs {
			// 	chunkInputs[i] = inp.SubVector(chunkStart, chunkStop)
			// 	// Sanity-check : the output of SubVector must have the correct length
			// 	if chunkInputs[i].Len() != chunkStop-chunkStart {
			// 		utils.Panic("Subvector failed, subvector should have size %v but size is %v", chunkStop-chunkStart, chunkInputs[i].Len())
			// 	}
			// }

			b.resetNodeAssignmentsGautam(inputs, chunkStart, chunkStop, &ba)
			// We don't parallelize evaluations where the inputs are all scalars
			// Therefore the cast is safe.
			b.evaluateSingleThreadMixedGautam(inputs, chunkStart, chunkStop, &ba)

			// No race condition here as each call write to different places
			// of vec.
			// chunkRes.WriteInSliceExt(res[chunkStart:chunkStop])

			copy(res[chunkStart:chunkStop], ba.assignment[len(b.Nodes)-1][0].ConcreteValue[:])

			// wg.Done()
		}

	})
	return sv.NewRegularExt(res)
	// }
}

// evaluateSingleThread evaluates a boarded expression. The inputs can be either
// vector or scalars. The vector's input length should be smaller than a chunk.
func (b *ExpressionBoard) evaluateSingleThreadMixed(inputs []sv.SmartVector, p ...mempool.MemPool) sv.SmartVector {

	var (
		length         = inputs[0].Len()
		pool, hasPool  = mempool.ExtractCheckOptionalSoft(length, p...)
		nodeAssignment = b.prepareNodeAssignments(inputs)
	)

	// Then computes the levels one by one
	for level := 1; level < len(nodeAssignment); level++ {
		for pil := range nodeAssignment[level] {

			node := &nodeAssignment[level][pil]
			nodeAssignment.evalMixed(node, pool)

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

	// Deep-copy the last node and put resBuf back in the pool. It's cleaner
	// that way.
	if hasPool {
		if reg, ok := resBuf.(*sv.Pooled); ok {
			// we have a base Pooled smart vector
			resGC := reg.DeepCopy()
			reg.Free(pool)
			resBuf = resGC
		} else {
			// check for an extension
			if regExt, isPooledExtension := resBuf.(*sv.PooledExt); isPooledExtension {
				resGC := regExt.DeepCopy()
				regExt.Free(pool)
				resBuf = resGC
			}
		}
	}
	return resBuf
}

// evaluateSingleThread evaluates a boarded expression. The inputs can be either
// vector or scalars. The vector's input length should be smaller than a chunk.
func (b *ExpressionBoard) evaluateSingleThreadMixedGautam(inputs []sv.SmartVector, chunkStart, chunkStop int, ba *boardAssignmentGautam) {

	// var (
	// 	nodeAssignment = b.prepareNodeAssignments(inputs)
	// )

	// Then computes the levels one by one
	nodeAssignment := ba.assignment
	for level := 1; level < len(nodeAssignment); level++ {
		for pil := range nodeAssignment[level] {

			node := &nodeAssignment[level][pil]
			ba.evalMixedGautam(node)

		}
	}

	// // Assertion, the last level contains only one node (the final result)
	// if len(nodeAssignment[len(nodeAssignment)-1]) > 1 {
	// 	panic("multiple heads")
	// }

	// // create a new regular smart vector ext and write the concrete value in it.
	// res := make([]fext.Element, chunkStop-chunkStart) // TODO &gbotrel use pool ?
	// copy(res, nodeAssignment[len(b.Nodes)-1][0].ConcreteValue[:])
	// resExt := sv.NewRegularExt(res)

	// // resBuf := nodeAssignment[len(b.Nodes)-1][0].Value

	// // if resBuf == nil {
	// // 	panic("resbuf is nil")
	// // }

	// return resExt
}

func (b *ExpressionBoard) initNodeAssignmentsGautam(inputs []sv.SmartVector) boardAssignmentGautam {
	boardAssignmentsignment := boardAssignmentGautam{b: big.NewInt(0)}
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

func (b *ExpressionBoard) resetNodeAssignmentsGautam(inputs []sv.SmartVector, chunkStart, chunkStop int, nodeAssignments *boardAssignmentGautam) {
	inputCursor := 0
	// Init the constants and inputs.
	for i := range b.Nodes {
		for j := range b.Nodes[i] {
			na := &nodeAssignments.assignment[i][j]
			if !na.IsConstant {
				na.HasValue = false
			}

			if i == 0 {
				switch na.Node.Operator.(type) {
				case Variable:
					// leaves, we set the input.
					// Sanity-check the input should have the correct length
					sb := inputs[inputCursor].SubVector(chunkStart, chunkStop)
					sb.WriteInSliceExt(na.ConcreteValue[:])
					// nodeAssignments[0][i].Value = inputs[inputCursor]
					inputCursor++
					na.HasValue = true
				}
			}
		}
	}

}

func (b *boardAssignmentGautam) evalMixedGautam(na *nodeAssignmentGautam) {

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
			vTmp := extensions.Vector(b.tmp[:])
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
		vTmp := extensions.Vector(b.tmp[:])
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
				res.Add(res, vInput)
			}
		}
	default:
		eval := na.Node.Operator.EvaluateExt(na.Inputs)
		eval.WriteInSliceExt(na.ConcreteValue[:])
	}

	na.HasValue = true

	// for i := range val {
	// 	b.incParentKnownCountOfMixed(val[i], pool, false)
	// }
}

func exp(z *extensions.E4, x extensions.E4, k int64) {
	if k == 0 {
		z.SetOne()
		return
	}

	if k < 0 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q⁴) == (x⁻¹)ᵏ (mod q⁴)
		x.Inverse(&x)
		k = -k
	}

	z.SetOne()
	for i := 63; i >= 0; i-- {
		z.Square(z)
		if (k & (1 << uint(i))) != 0 {
			z.Mul(z, &x)
		}
	}

}
