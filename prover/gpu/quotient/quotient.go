// GPU-accelerated quotient computation with batch NTT + pinned H2D.
//
//go:build cuda

package quotient

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/gpu"
	gpusym "github.com/consensys/linea-monorepo/prover/gpu/symbolic"
	gpuvortex "github.com/consensys/linea-monorepo/prover/gpu/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
)

const maxGPUSlots = 8192

func RunGPU(
	dev *gpu.Device, run *wizard.ProverRuntime,
	domainSize int, ratios []int,
	boards []symbolic.ExpressionBoard,
	rootsForRatio [][]ifaces.Column,
	shiftedForRatio [][]ifaces.Column,
	quotientShares [][]ifaces.Column,
	constraintsByRatio map[int][]int,
) error {
	stopTimer := profiling.LogTimer("GPU quotient (domain size %d)", domainSize)
	defer stopTimer()

	maxRatio := 0
	for _, r := range ratios {
		if r > maxRatio {
			maxRatio = r
		}
	}

	// ── Compile boards ───────────────────────────────────────────────────
	t0 := time.Now()
	type cb struct {
		pgm  *gpusym.GPUSymProgram
		meta []symbolic.Metadata
	}
	compiled := make([]cb, len(boards))
	var (
		gpuBoardCount   int
		cpuBoardCount   int
		maxBoardSlots   int
		fallbackBySlots int
	)
	for k := range boards {
		ops := gpusym.BoardToNodeOps(&boards[k])
		if len(ops) == 0 {
			cpuBoardCount++
			continue
		}
		p := gpusym.CompileGPU(ops)
		if p.NumSlots > maxBoardSlots {
			maxBoardSlots = p.NumSlots
		}
		if len(p.Bytecode) == 0 || p.NumSlots > maxGPUSlots {
			cpuBoardCount++
			if p.NumSlots > maxGPUSlots {
				fallbackBySlots++
			}
			continue
		}
		dp, err := gpusym.CompileSymGPU(dev, p)
		if err != nil {
			panic(fmt.Sprintf("gpu/quotient: compile[%d]: %v", k, err))
		}
		defer dp.Free()
		compiled[k] = cb{pgm: dp, meta: boards[k].ListVariableMetadata()}
		gpuBoardCount++
	}
	tCompile := time.Since(t0)

	// ── GPU NTT domain ───────────────────────────────────────────────────
	nttDomain, err := gpuvortex.NewGPUFFTDomain(dev, domainSize)
	if err != nil {
		return fmt.Errorf("gpu/quotient: NTT domain init (size %d): %w", domainSize, err)
	}
	defer nttDomain.Free()
	var nInv field.Element
	nInv.SetUint64(uint64(domainSize))
	nInv.Inverse(&nInv)

	cpuDomain0 := fft.NewDomain(uint64(domainSize), fft.WithCache())

	var annBase []field.Element
	var annExt []fext.Element
	annBaseDone, annExtDone := false, false

	// ── Collect ALL roots, split base / ext ───────────────────────────────
	t0 = time.Now()
	allRoots := make(map[ifaces.ColID]ifaces.Column)
	for _, roots := range rootsForRatio {
		for _, r := range roots {
			allRoots[r.GetColID()] = r
		}
	}

	isExtRoot := make(map[ifaces.ColID]bool)
	var baseRootIDs []ifaces.ColID
	var extRootIDs []ifaces.ColID
	for id, root := range allRoots {
		w, ok := run.Columns.TryGet(id)
		if !ok {
			w = root.GetColAssignment(run)
		}
		if !smartvectors.IsBase(w) {
			isExtRoot[id] = true
			extRootIDs = append(extRootIDs, id)
		} else {
			baseRootIDs = append(baseRootIDs, id)
		}
	}
	nBaseRoots := len(baseRootIDs)
	nExtRoots := len(extRootIDs)
	tRootSplit := time.Since(t0)

	ratioRootStats := make(map[int][2]int, len(constraintsByRatio))
	for ratio, constraintsIndices := range constraintsByRatio {
		seen := make(map[ifaces.ColID]struct{})
		var baseCount, extCount int
		for _, j := range constraintsIndices {
			for _, root := range rootsForRatio[j] {
				id := root.GetColID()
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				if isExtRoot[id] {
					extCount++
				} else {
					baseCount++
				}
			}
		}
		ratioRootStats[ratio] = [2]int{baseCount, extCount}
	}

	t0 = time.Now()

	// ── Ext root witness data (read once, cache coefficients in CPU memory) ──
	extRootIdx := make(map[ifaces.ColID]int, nExtRoots)
	for i, id := range extRootIDs {
		extRootIdx[id] = i
	}
	extCoeffs := make([][]fext.Element, nExtRoots)
	t0 = time.Now()
	if nExtRoots > 0 {
		parallel.Execute(nExtRoots, func(start, stop int) {
			for k := start; k < stop; k++ {
				id := extRootIDs[k]
				w, ok := run.Columns.TryGet(id)
				if !ok {
					w = allRoots[id].GetColAssignment(run)
				}
				r := make([]fext.Element, domainSize)
				w.WriteInSliceExt(r)
				cpuDomain0.FFTInverseExt(r, fft.DIF, fft.WithNbTasks(1))
				extCoeffs[k] = r
			}
		})
	}
	tExtIFFT := time.Since(t0)

	type ratioData struct {
		baseIDs     []ifaces.ColID
		baseIdx     map[ifaces.ColID]int
		dPacked     *gpuvortex.KBVector
		dEvals      *gpuvortex.KBVector
		extIDs      []ifaces.ColID
		extIdx      map[ifaces.ColID]int
		dExtCoeffs  *gpuvortex.KBVector // SoA, natural-order coefficients
		dExtEvals   *gpuvortex.KBVector // SoA, per-coset evaluations
		extEvalPtrs []unsafe.Pointer    // ptr to one root block in dExtEvals
	}

	ratioPrepared := make(map[int]*ratioData, len(constraintsByRatio))
	var tPack, tH2D, tIFFT time.Duration
	var tExtPrepGPU time.Duration

	for ratio, constraintsIndices := range constraintsByRatio {
		rd := &ratioData{
			baseIdx: make(map[ifaces.ColID]int),
			extIdx:  make(map[ifaces.ColID]int),
		}
		seen := make(map[ifaces.ColID]struct{})
		for _, j := range constraintsIndices {
			for _, root := range rootsForRatio[j] {
				id := root.GetColID()
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				if isExtRoot[id] {
					rd.extIdx[id] = len(rd.extIDs)
					rd.extIDs = append(rd.extIDs, id)
				} else {
					rd.baseIdx[id] = len(rd.baseIDs)
					rd.baseIDs = append(rd.baseIDs, id)
				}
			}
		}
		ratioRootStats[ratio] = [2]int{len(rd.baseIDs), len(rd.extIDs)}

		if len(rd.baseIDs) > 0 {
			t0 = time.Now()
			// Cached pinned buffer keyed on (deviceID, capacity). The
			// first call on a given (device, ratio-shape) pays the
			// cudaMallocHost; subsequent calls reuse it. This is the
			// single biggest pre-optimization improvement for the
			// quotient hot path — see gpu/vortex/pinned_cache.go.
			deviceID := 0
			if dev != nil {
				deviceID = dev.DeviceID()
			}
			capacity := len(rd.baseIDs) * domainSize
			pinnedBuf := gpuvortex.GetPinned(deviceID, capacity)[:capacity]
			parallel.Execute(len(rd.baseIDs), func(start, stop int) {
				for k := start; k < stop; k++ {
					id := rd.baseIDs[k]
					w, ok := run.Columns.TryGet(id)
					if !ok {
						w = allRoots[id].GetColAssignment(run)
					}
					dst := pinnedBuf[k*domainSize : (k+1)*domainSize]
					if reg, ok := w.(*sv.Regular); ok {
						copy(dst, *reg)
					} else {
						w.WriteInSlice(dst)
					}
				}
			})
			tPack += time.Since(t0)

			t0 = time.Now()
			rd.dPacked, err = gpuvortex.NewKBVector(dev, len(rd.baseIDs)*domainSize)
			if err != nil {
				return fmt.Errorf("gpu/quotient: alloc dPacked (ratio %d, %d elems): %w", ratio, len(rd.baseIDs)*domainSize, err)
			}
			defer rd.dPacked.Free()
			rd.dPacked.CopyFromHostPinned(pinnedBuf)
			tH2D += time.Since(t0)

			t0 = time.Now()
			nttDomain.BatchIFFTScale(rd.dPacked, len(rd.baseIDs), nInv)
			gpuvortex.Sync(dev)
			tIFFT += time.Since(t0)
			// Pinned buffer stays in the cache; reused on next call.

			rd.dEvals, err = gpuvortex.NewKBVector(dev, len(rd.baseIDs)*domainSize)
			if err != nil {
				return fmt.Errorf("gpu/quotient: alloc dEvals (ratio %d, %d elems): %w", ratio, len(rd.baseIDs)*domainSize, err)
			}
			defer rd.dEvals.Free()
		}

		if len(rd.extIDs) > 0 {
			t0 = time.Now()
			rd.dExtCoeffs, err = gpuvortex.NewKBVector(dev, len(rd.extIDs)*domainSize*4)
			if err != nil {
				return fmt.Errorf("gpu/quotient: alloc dExtCoeffs (ratio %d, %d elems): %w", ratio, len(rd.extIDs)*domainSize*4, err)
			}
			defer rd.dExtCoeffs.Free()
			rd.dExtEvals, err = gpuvortex.NewKBVector(dev, len(rd.extIDs)*domainSize*4)
			if err != nil {
				return fmt.Errorf("gpu/quotient: alloc dExtEvals (ratio %d, %d elems): %w", ratio, len(rd.extIDs)*domainSize*4, err)
			}
			defer rd.dExtEvals.Free()
			rd.extEvalPtrs = make([]unsafe.Pointer, len(rd.extIDs))
			for k := range rd.extIDs {
				rd.extEvalPtrs[k] = unsafe.Add(rd.dExtEvals.DevicePtr(), k*domainSize*4*4)
			}

			extSOA := make([]field.Element, len(rd.extIDs)*domainSize*4)
			for k, id := range rd.extIDs {
				globalIdx := extRootIdx[id]
				coeffs := extCoeffs[globalIdx]
				base := k * domainSize * 4
				dst0 := extSOA[base : base+domainSize]
				dst1 := extSOA[base+domainSize : base+2*domainSize]
				dst2 := extSOA[base+2*domainSize : base+3*domainSize]
				dst3 := extSOA[base+3*domainSize : base+4*domainSize]
				for j := range coeffs {
					dst0[j] = coeffs[j].B0.A0
					dst1[j] = coeffs[j].B0.A1
					dst2[j] = coeffs[j].B1.A0
					dst3[j] = coeffs[j].B1.A1
				}
			}
			rd.dExtCoeffs.CopyFromHost(extSOA)
			for vec := 0; vec < len(rd.extIDs)*4; vec++ {
				ptr := unsafe.Add(rd.dExtCoeffs.DevicePtr(), vec*domainSize*4)
				gpuvortex.BitRevRaw(dev, ptr, domainSize)
			}
			gpuvortex.Sync(dev)
			tExtPrepGPU += time.Since(t0)
		}

		ratioPrepared[ratio] = rd
	}

	var tCosetPrep, tCosetNTT, tExtFFT, tSymEval time.Duration
	var tSymInputs, tSymKernel, tSymPost, tSymAssign, tSymAuxVec, tSymAuxFree time.Duration
	var symAuxVecCount int

	for i := 0; i < maxRatio; i++ {
		for ratio, constraintsIndices := range constraintsByRatio {
			if i%(maxRatio/ratio) != 0 {
				continue
			}
			share := i * ratio / maxRatio
			shift := computeShift(uint64(domainSize), ratio, share)
			rd := ratioPrepared[ratio]

			t0 = time.Now()
			cosetDomain := fft.NewDomain(uint64(domainSize), fft.WithShift(shift), fft.WithCache())
			tCosetPrep += time.Since(t0)

			// ── Batch coset FFT for ratio-specific base roots ─
			t0 = time.Now()
			if rd.dPacked != nil {
				rd.dEvals.CopyFromDevice(rd.dPacked)
				nttDomain.BatchCosetFFTBitRev(rd.dEvals, len(rd.baseIDs), shift)
				gpuvortex.Sync(dev)
			}
			tCosetNTT += time.Since(t0)

			evalPtr := func(id ifaces.ColID) unsafe.Pointer {
				return unsafe.Add(rd.dEvals.DevicePtr(), rd.baseIdx[id]*domainSize*4)
			}

			// ── Ratio-specific extension roots: GPU batch coset FFT on SoA blocks ─
			t0 = time.Now()
			if rd.dExtCoeffs != nil {
				rd.dExtEvals.CopyFromDevice(rd.dExtCoeffs)
				nttDomain.BatchCosetFFTBitRev(rd.dExtEvals, len(rd.extIDs)*4, shift)
				gpuvortex.Sync(dev)
			}
			tExtFFT += time.Since(t0)

			miscVecs := make(map[string]*gpuvortex.KBVector)

			// ── GPU symbolic eval ────────────────────────────────────
			t0 = time.Now()
			for _, j := range constraintsIndices {
				c := &compiled[j]
				if c.pgm == nil {
					cpuFallback(run, &boards[j], rootsForRatio[j], shiftedForRatio[j],
						cpuDomain0, cosetDomain, quotientShares[j][share].GetColID(),
						domainSize, i, maxRatio, &annBase, &annExt, &annBaseDone, &annExtDone)
					continue
				}

				tIn := time.Now()
				inputs := make([]gpusym.SymInput, len(c.meta))
				for k, mi := range c.meta {
					switch m := mi.(type) {
					case ifaces.Column:
						root := column.RootParents(m)
						rid := root.GetColID()
						isExt := isExtRoot[rid]
						if shifted, isShifted := m.(column.Shifted); isShifted {
							if isExt {
								inputs[k] = gpusym.SymInput{Tag: gpusym.SymInputRotE4SOA, DPtr: rd.extEvalPtrs[rd.extIdx[rid]], Offset: shifted.Offset}
							} else {
								inputs[k] = gpusym.SymInput{Tag: gpusym.SymInputRotKB, DPtr: evalPtr(rid), Offset: shifted.Offset}
							}
						} else {
							if isExt {
								inputs[k] = gpusym.SymInput{Tag: gpusym.SymInputE4VecSOA, DPtr: rd.extEvalPtrs[rd.extIdx[rid]]}
							} else {
								inputs[k] = gpusym.SymInput{Tag: gpusym.SymInputKB, DPtr: evalPtr(rid)}
							}
						}
					case coin.Info:
						inputs[k] = gpusym.SymInputFromConst(run.GetRandomCoinFieldExt(m.Name))
					case variables.X:
						key := fmt.Sprintf("X_%d", i)
						if dV, ok := miscVecs[key]; ok {
							inputs[k] = gpusym.SymInputFromVec(dV)
						} else {
							tAux := time.Now()
							h := make([]field.Element, domainSize)
							m.EvalCoset(domainSize, i, maxRatio, true).WriteInSlice(h)
							dV, auxErr := gpuvortex.NewKBVector(dev, domainSize)
							if auxErr != nil {
								return fmt.Errorf("gpu/quotient: alloc X vec: %w", auxErr)
							}
							dV.CopyFromHost(h)
							miscVecs[key] = dV
							tSymAuxVec += time.Since(tAux)
							symAuxVecCount++
							inputs[k] = gpusym.SymInputFromVec(dV)
						}
					case variables.PeriodicSample:
						key := fmt.Sprintf("%s_%d", m.String(), i)
						if dV, ok := miscVecs[key]; ok {
							inputs[k] = gpusym.SymInputFromVec(dV)
						} else {
							tAux := time.Now()
							h := make([]field.Element, domainSize)
							m.EvalCoset(domainSize, i, maxRatio, true).WriteInSlice(h)
							dV, auxErr := gpuvortex.NewKBVector(dev, domainSize)
							if auxErr != nil {
								return fmt.Errorf("gpu/quotient: alloc PeriodicSample vec: %w", auxErr)
							}
							dV.CopyFromHost(h)
							miscVecs[key] = dV
							tSymAuxVec += time.Since(tAux)
							symAuxVecCount++
							inputs[k] = gpusym.SymInputFromVec(dV)
						}
					case ifaces.Accessor:
						if m.IsBase() {
							v := m.GetVal(run)
							var e4 fext.Element
							fext.SetFromBase(&e4, &v)
							inputs[k] = gpusym.SymInputFromConst(e4)
						} else {
							inputs[k] = gpusym.SymInputFromConst(m.GetValExt(run))
						}
					default:
						utils.Panic("unknown metadata type %v", reflect.TypeOf(mi))
					}
				}
				tSymInputs += time.Since(tIn)

				tK := time.Now()
				result := gpusym.EvalSymGPU(dev, c.pgm, inputs, domainSize)
				tSymKernel += time.Since(tK)

				tP := time.Now()
				var assigned sv.SmartVector
				allBaseField := true
				for k := range result {
					if !fext.IsBase(&result[k]) {
						allBaseField = false
						break
					}
				}
				if allBaseField {
					if !annBaseDone {
						annBase = fastpoly.EvalXnMinusOneOnACoset(domainSize, domainSize*maxRatio)
						annBase = field.ParBatchInvert(annBase, runtime.GOMAXPROCS(0))
						annBaseDone = true
					}
					br := make([]field.Element, domainSize)
					for k := range result {
						br[k] = result[k].B0.A0
					}
					vq := field.Vector(br)
					vq.ScalarMul(vq, &annBase[i])
					assigned = sv.NewRegular(br)
				} else {
					if !annExtDone {
						annExt = fastpolyext.EvalXnMinusOneOnACoset(domainSize, domainSize*maxRatio)
						annExt = fext.ParBatchInvert(annExt, runtime.GOMAXPROCS(0))
						annExtDone = true
					}
					extensions.Vector(result).ScalarMul(extensions.Vector(result), &annExt[i])
					assigned = sv.NewRegularExt(result)
				}
				tSymPost += time.Since(tP)

				tA := time.Now()
				run.AssignColumn(quotientShares[j][share].GetColID(), assigned)
				tSymAssign += time.Since(tA)
			}
			tSymEval += time.Since(t0)

			tF := time.Now()
			for _, v := range miscVecs {
				v.Free()
			}
			tSymAuxFree += time.Since(tF)
		}
	}

	fmt.Printf("gpu/quotient TIMING: compile=%v rootSplit=%v extIFFT=%v pack=%v h2d=%v ifft=%v cosetPrep=%v cosetNTT=%v extFFT=%v symEval=%v (inputs=%v kernel=%v post=%v assign=%v auxBuild=%v auxFree=%v auxVecs=%d) | %d base, %d ext roots | boards gpu=%d cpu=%d maxSlots=%d slotFallback=%d\n",
		tCompile, tRootSplit, tExtIFFT, tPack, tH2D, tIFFT, tCosetPrep, tCosetNTT, tExtFFT, tSymEval,
		tSymInputs, tSymKernel, tSymPost, tSymAssign, tSymAuxVec, tSymAuxFree, symAuxVecCount,
		nBaseRoots, nExtRoots, gpuBoardCount, cpuBoardCount, maxBoardSlots, fallbackBySlots)
	fmt.Printf("gpu/quotient ROOTS per ratio: %v\n", ratioRootStats)
	return nil
}

func computeShift(n uint64, cosetRatio int, cosetID int) field.Element {
	var s field.Element
	g := fft.GeneratorFullMultiplicativeGroup()
	omega, _ := fft.Generator(n * uint64(cosetRatio))
	omega.Exp(omega, new(big.Int).SetInt64(int64(cosetID)))
	s.Mul(&g, &omega)
	return s
}

func cpuFallback(
	run *wizard.ProverRuntime, board *symbolic.ExpressionBoard,
	roots, shifted []ifaces.Column, domain0, cosetDomain *fft.Domain,
	colID ifaces.ColID, domainSize, cosetIdx, maxRatio int,
	annBase *[]field.Element, annExt *[]fext.Element, annBaseDone, annExtDone *bool,
) {
	reeval := make(map[ifaces.ColID]sv.SmartVector)
	for _, root := range roots {
		id := root.GetColID()
		if _, ok := reeval[id]; ok {
			continue
		}
		w, ok := run.Columns.TryGet(id)
		if !ok {
			w = root.GetColAssignment(run)
		}
		if smartvectors.IsBase(w) {
			r := make([]field.Element, domainSize)
			w.WriteInSlice(r)
			domain0.FFTInverse(r, fft.DIF, fft.WithNbTasks(2))
			cosetDomain.FFT(r, fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))
			reeval[id] = sv.NewRegular(r)
		} else {
			r := make([]fext.Element, domainSize)
			w.WriteInSliceExt(r)
			domain0.FFTInverseExt(r, fft.DIF, fft.WithNbTasks(2))
			cosetDomain.FFTExt(r, fft.DIT, fft.OnCoset(), fft.WithNbTasks(2))
			reeval[id] = sv.NewRegularExt(r)
		}
	}
	for _, pol := range shifted {
		pid := pol.GetColID()
		if _, ok := reeval[pid]; ok {
			continue
		}
		rt := column.RootParents(pol)
		if s, ok := pol.(column.Shifted); ok {
			switch v := reeval[rt.GetColID()].(type) {
			case *sv.Regular:
				reeval[pid] = sv.SoftRotate(v, s.Offset)
			case *sv.RegularExt:
				reeval[pid] = sv.SoftRotateExt(v, s.Offset)
			}
		}
	}
	metas := board.ListVariableMetadata()
	ins := make([]sv.SmartVector, len(metas))
	for k, mi := range metas {
		switch m := mi.(type) {
		case ifaces.Column:
			ins[k] = reeval[m.GetColID()]
		case coin.Info:
			ins[k] = sv.NewConstantExt(run.GetRandomCoinFieldExt(m.Name), domainSize)
		case variables.X:
			ins[k] = m.EvalCoset(domainSize, cosetIdx, maxRatio, true)
		case variables.PeriodicSample:
			ins[k] = m.EvalCoset(domainSize, cosetIdx, maxRatio, true)
		case ifaces.Accessor:
			if m.IsBase() {
				ins[k] = sv.NewConstant(m.GetVal(run), domainSize)
			} else {
				ins[k] = sv.NewConstantExt(m.GetValExt(run), domainSize)
			}
		}
	}
	qs := board.Evaluate(ins)
	switch q := qs.(type) {
	case *sv.Regular:
		if !*annBaseDone {
			*annBase = fastpoly.EvalXnMinusOneOnACoset(domainSize, domainSize*maxRatio)
			*annBase = field.ParBatchInvert(*annBase, runtime.GOMAXPROCS(0))
			*annBaseDone = true
		}
		vq := field.Vector(*q)
		vq.ScalarMul(vq, &(*annBase)[cosetIdx])
	case *sv.RegularExt:
		if !*annExtDone {
			*annExt = fastpolyext.EvalXnMinusOneOnACoset(domainSize, domainSize*maxRatio)
			*annExt = fext.ParBatchInvert(*annExt, runtime.GOMAXPROCS(0))
			*annExtDone = true
		}
		extensions.Vector(*q).ScalarMul(extensions.Vector(*q), &(*annExt)[cosetIdx])
	}
	run.AssignColumn(colID, qs)
}
