package symbolic

import (
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Program is a compiled expression board ready for evaluation.

type opCode uint8

const (
	opLoadConst opCode = iota
	opLoadInput
	opMul
	opLinComb
	opPolyEval
)

func areAllConstants(inp []smartvectors.SmartVector) bool {
	for _, v := range inp {
		// must be sv.Constant or sv.ConstantExt
		switch v.(type) {
		case *smartvectors.Constant:
		case *smartvectors.ConstantExt:
		default:
			return false
		}
	}
	return true
}

func (b *ExpressionBoard) Evaluate(inputs []smartvectors.SmartVector) smartvectors.SmartVector {
	if b.ProgramNodesCount != len(b.Nodes) {
		b.Compile()
	}

	if len(inputs) == 0 {
		panic("no input provided")
	}
	totalSize := inputs[0].Len()
	if totalSize == 0 {
		panic("inputs all have size 0")
	}
	for i := 1; i < len(inputs); i++ {
		if inputs[i].Len() != totalSize {
			utils.Panic("mismatch in the size: len(v0) %v, len(v%v) %v", totalSize, i, inputs[i].Len())
		}
	}
	// defaultChunkSize; we need at least 16 to leverage
	// avx512 instructions. Larger values will allocate more memory.
	const defaultChunkSize = 1 << 5
	chunkSize := min(defaultChunkSize, totalSize)
	if totalSize%chunkSize != 0 {
		panic("chunk size should divide total size")
	}
	numChunks := totalSize / chunkSize

	isBase := smartvectors.AreAllBase(inputs)

	if isBase {
		res := make([]field.Element, totalSize)
		parallel.Execute(numChunks, func(start, stop int) {
			vm := newVMBase(b, chunkSize)
			for chunkID := start; chunkID < stop; chunkID++ {
				chunkStart := chunkID * chunkSize
				chunkStop := (chunkID + 1) * chunkSize
				vm.execute(inputs, chunkStart, chunkStop)

				lastDst := b.ResultSlot
				copy(res[chunkStart:chunkStop], vm.memory[lastDst*chunkSize:(lastDst+1)*chunkSize])
			}
		})

		if areAllConstants(inputs) {
			return smartvectors.NewConstant(res[0], totalSize)
		}
		return smartvectors.NewRegular(res)
	} else {
		res := make([]fext.Element, totalSize)
		parallel.Execute(numChunks, func(start, stop int) {
			vm := newVMExt(b, chunkSize)
			for chunkID := start; chunkID < stop; chunkID++ {
				chunkStart := chunkID * chunkSize
				chunkStop := (chunkID + 1) * chunkSize
				vm.execute(inputs, chunkStart, chunkStop)

				lastDst := b.ResultSlot
				copy(res[chunkStart:chunkStop], vm.memory[lastDst*chunkSize:(lastDst+1)*chunkSize])
			}
		})

		if areAllConstants(inputs) {
			return smartvectors.NewConstantExt(res[0], totalSize)
		}
		return smartvectors.NewRegularExt(res)
	}
}

// Compile compiles the expression board into a program.
func (b *ExpressionBoard) Compile() {
	if len(b.Nodes) == 0 {
		b.Bytecode = nil
		b.Constants = nil
		b.NumSlots = 0
		b.ProgramNodesCount = 0
		return
	}

	// 1. Liveness Analysis
	lastUse := make([]int, len(b.Nodes))
	for i := range lastUse {
		lastUse[i] = -1
	}
	// The last node is the output, so it is implicitly used at the end
	lastUse[len(b.Nodes)-1] = len(b.Nodes)

	for i, node := range b.Nodes {
		for _, childID := range node.Children {
			if i > lastUse[childID] {
				lastUse[childID] = i
			}
		}
	}

	// 2. Instruction Generation & Register Allocation
	bytecode := make([]int, 0, len(b.Nodes)*4)
	constants := make([]fext.Element, 0)
	slots := make([]int, len(b.Nodes))
	freeSlots := make([]int, 0)
	nextSlot := 0

	// Helper to get a slot
	getSlot := func() int {
		if len(freeSlots) > 0 {
			s := freeSlots[len(freeSlots)-1]
			freeSlots = freeSlots[:len(freeSlots)-1]
			return s
		}
		s := nextSlot
		nextSlot++
		return s
	}

	inputCursor := 0

	for i, node := range b.Nodes {
		dstSlot := getSlot()
		slots[i] = dstSlot

		switch op := node.Operator.(type) {
		case Constant:
			bytecode = append(bytecode, int(opLoadConst), dstSlot, len(constants))
			constants = append(constants, op.Val.GetExt())
		case Variable:
			bytecode = append(bytecode, int(opLoadInput), dstSlot, inputCursor)
			inputCursor++
		case Product:
			bytecode = append(bytecode, int(opMul), dstSlot, len(node.Children))
			for k, childID := range node.Children {
				bytecode = append(bytecode, slots[childID], op.Exponents[k])
			}
		case LinComb:
			bytecode = append(bytecode, int(opLinComb), dstSlot, len(node.Children))
			for k, childID := range node.Children {
				bytecode = append(bytecode, slots[childID], op.Coeffs[k])
			}
		case PolyEval:
			bytecode = append(bytecode, int(opPolyEval), dstSlot, len(node.Children))
			for _, childID := range node.Children {
				bytecode = append(bytecode, slots[childID])
			}
		default:
			utils.Panic("unknown op %T", op)
		}

		// Free slots of children that are dead
		freed := make(map[int]bool)
		for _, childID := range node.Children {
			cid := int(childID)
			if lastUse[cid] == i {
				if !freed[slots[cid]] {
					freeSlots = append(freeSlots, slots[cid])
					freed[slots[cid]] = true
				}
			}
		}
	}

	b.Bytecode = bytecode
	b.Constants = constants
	b.NumSlots = nextSlot
	b.ResultSlot = slots[len(b.Nodes)-1]
	b.ProgramNodesCount = len(b.Nodes)
}

// VM for Base elements
type vmBase struct {
	memory    []field.Element
	scratch   []field.Element
	chunkSize int
	board     *ExpressionBoard
}

func newVMBase(b *ExpressionBoard, chunkSize int) *vmBase {
	return &vmBase{
		memory:    make([]field.Element, b.NumSlots*chunkSize),
		scratch:   make([]field.Element, chunkSize),
		chunkSize: chunkSize,
		board:     b,
	}
}

func (vm *vmBase) execute(inputs []smartvectors.SmartVector, chunkStart, chunkStop int) {
	chunkLen := chunkStop - chunkStart
	bytecode := vm.board.Bytecode
	constants := vm.board.Constants
	pc := 0

	for pc < len(bytecode) {
		op := opCode(bytecode[pc])
		pc++
		dstSlot := bytecode[pc]
		pc++
		dstOffset := dstSlot * vm.chunkSize
		dst := vm.memory[dstOffset : dstOffset+chunkLen]

		switch op {
		case opLoadConst:
			constIdx := bytecode[pc]
			pc++
			val := constants[constIdx].B0.A0
			for i := 0; i < chunkLen; i++ {
				dst[i] = val
			}
		case opLoadInput:
			inputID := bytecode[pc]
			pc++
			input := inputs[inputID]
			switch rv := input.(type) {
			case *smartvectors.Regular:
				copy(dst, (*rv)[chunkStart:chunkStop])
			default:
				sb := input.SubVector(chunkStart, chunkStop)
				sb.WriteInSlice(dst)
			}
		case opMul: // Handles Product logic
			numSrc := bytecode[pc]
			pc++
			vRes := field.Vector(dst)
			vTmp := field.Vector(vm.scratch[:chunkLen])
			for k := 0; k < numSrc; k++ {
				srcSlot := bytecode[pc]
				pc++
				exp := bytecode[pc]
				pc++
				srcOffset := srcSlot * vm.chunkSize
				vInput := field.Vector(vm.memory[srcOffset : srcOffset+chunkLen])
				if k == 0 {
					vRes.Exp(vInput, int64(exp))
				} else {
					vTmp.Exp(vInput, int64(exp))
					vRes.Mul(vRes, vTmp)
				}
			}
		case opLinComb:
			numSrc := bytecode[pc]
			pc++
			vRes := field.Vector(dst)
			vTmp := field.Vector(vm.scratch[:chunkLen])
			var t0 field.Element
			for k := 0; k < numSrc; k++ {
				srcSlot := bytecode[pc]
				pc++
				coeff := bytecode[pc]
				pc++
				srcOffset := srcSlot * vm.chunkSize
				vInput := field.Vector(vm.memory[srcOffset : srcOffset+chunkLen])

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
					t0.SetInt64(int64(coeff))
					vTmp.ScalarMul(vInput, &t0)
					vRes.Add(vRes, vTmp)
				}
			}
		case opPolyEval:
			numSrc := bytecode[pc]
			pc++
			srcStart := pc
			pc += numSrc

			// input[0] is x (constant)
			xSlot := bytecode[srcStart]
			xOffset := xSlot * vm.chunkSize
			x := vm.memory[xOffset]

			vRes := field.Vector(dst)

			lastSrcSlot := bytecode[srcStart+numSrc-1]
			lastSrcOffset := lastSrcSlot * vm.chunkSize
			copy(vRes, field.Vector(vm.memory[lastSrcOffset:lastSrcOffset+chunkLen]))

			for k := numSrc - 2; k >= 1; k-- {
				srcSlot := bytecode[srcStart+k]
				srcOffset := srcSlot * vm.chunkSize
				vTmp := field.Vector(vm.memory[srcOffset : srcOffset+chunkLen])
				vRes.ScalarMul(vRes, &x)
				vRes.Add(vRes, vTmp)
			}
		}
	}
}

// VM for Ext elements
type vmExt struct {
	memory    []fext.Element
	scratch   []fext.Element
	chunkSize int
	board     *ExpressionBoard
}

func newVMExt(b *ExpressionBoard, chunkSize int) *vmExt {
	return &vmExt{
		memory:    make([]fext.Element, b.NumSlots*chunkSize),
		scratch:   make([]fext.Element, chunkSize),
		chunkSize: chunkSize,
		board:     b,
	}
}

func (vm *vmExt) execute(inputs []smartvectors.SmartVector, chunkStart, chunkStop int) {
	chunkLen := chunkStop - chunkStart
	bytecode := vm.board.Bytecode
	constants := vm.board.Constants
	pc := 0

	for pc < len(bytecode) {
		op := opCode(bytecode[pc])
		pc++
		dstSlot := bytecode[pc]
		pc++
		dstOffset := dstSlot * vm.chunkSize
		dst := vm.memory[dstOffset : dstOffset+chunkLen]

		switch op {
		case opLoadConst:
			constIdx := bytecode[pc]
			pc++
			val := constants[constIdx]
			for i := 0; i < chunkLen; i++ {
				dst[i] = val
			}
		case opLoadInput:
			inputID := bytecode[pc]
			pc++
			input := inputs[inputID]
			switch rv := input.(type) {
			case *smartvectors.RotatedExt:
				rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, dst)
			case *smartvectors.Rotated:
				rv.WriteSubVectorInSliceExt(chunkStart, chunkStop, dst)
			case *smartvectors.Regular:
				for i := 0; i < chunkLen; i++ {
					dst[i].B0.A0 = (*rv)[chunkStart+i]
					dst[i].B0.A1.SetZero()
					dst[i].B1.A0.SetZero()
					dst[i].B1.A1.SetZero()
				}
			case *smartvectors.RegularExt:
				copy(dst, (*rv)[chunkStart:chunkStop])
			case *smartvectors.ConstantExt:
				for i := 0; i < chunkLen; i++ {
					dst[i].Set(&rv.Value)
				}
			default:
				sb := input.SubVector(chunkStart, chunkStop)
				sb.WriteInSliceExt(dst)
			}
		case opMul:
			numSrc := bytecode[pc]
			pc++
			vRes := extensions.Vector(dst)
			vTmp := extensions.Vector(vm.scratch[:chunkLen])

			srcSlot0 := bytecode[pc]
			pc++
			exp0 := bytecode[pc]
			pc++
			srcOffset0 := srcSlot0 * vm.chunkSize
			vInput0 := extensions.Vector(vm.memory[srcOffset0 : srcOffset0+chunkLen])
			vRes.Exp(vInput0, int64(exp0))

			for k := 1; k < numSrc; k++ {
				srcSlot := bytecode[pc]
				pc++
				exp := bytecode[pc]
				pc++
				srcOffset := srcSlot * vm.chunkSize
				vInput := extensions.Vector(vm.memory[srcOffset : srcOffset+chunkLen])
				if exp == 1 {
					vRes.Mul(vRes, vInput)
					continue
				}
				vTmp.Exp(vInput, int64(exp))
				vRes.Mul(vRes, vTmp)
			}
		case opLinComb:
			numSrc := bytecode[pc]
			pc++
			vRes := extensions.Vector(dst)
			vTmp := extensions.Vector(vm.scratch[:chunkLen])
			var t0 field.Element
			for k := 0; k < numSrc; k++ {
				srcSlot := bytecode[pc]
				pc++
				coeff := bytecode[pc]
				pc++
				srcOffset := srcSlot * vm.chunkSize
				vInput := extensions.Vector(vm.memory[srcOffset : srcOffset+chunkLen])

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
						vRes.ScalarMulByElement(vInput, &t0)
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
					t0.SetInt64(int64(coeff))
					vTmp.ScalarMulByElement(vInput, &t0)
					vRes.Add(vRes, vTmp)
				}
			}
		case opPolyEval:
			numSrc := bytecode[pc]
			pc++
			srcStart := pc
			pc += numSrc

			xSlot := bytecode[srcStart]
			xOffset := xSlot * vm.chunkSize
			x := vm.memory[xOffset]

			vRes := extensions.Vector(dst)

			lastSrcSlot := bytecode[srcStart+numSrc-1]
			lastSrcOffset := lastSrcSlot * vm.chunkSize
			copy(vRes, extensions.Vector(vm.memory[lastSrcOffset:lastSrcOffset+chunkLen]))

			for k := numSrc - 2; k >= 1; k-- {
				srcSlot := bytecode[srcStart+k]
				srcOffset := srcSlot * vm.chunkSize
				vTmp := extensions.Vector(vm.memory[srcOffset : srcOffset+chunkLen])
				vRes.ScalarMul(vRes, &x)
				vRes.Add(vRes, vTmp)
			}
		}
	}
}
