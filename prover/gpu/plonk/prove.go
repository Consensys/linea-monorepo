//go:build cuda

package plonk

// plonk_prove.go — GPU-accelerated PlonK prover.
//
// Architecture: canonical-only MSM. All polynomial commitments use a single
// canonical SRS MSM context. Wire polynomials and Z are obtained in Lagrange
// form, converted to canonical via GPU iFFT, then committed.
//
// Structural layers:
//
//   GPUProvingKey   — slim: VerifyingKey + lazy instance reference
//   gpuInstance     — persistent GPU resources + circuit data + GPU-resident polys
//   gpuProver       — per-proof mutable state; methods implement prove phases
//
// Instance is created lazily on first GPUProve call. GPU resources persist
// across proofs (MSM context, FFT domain, permutation table, selector polynomials).
// Release explicitly via Close().
//
// GPU-resident polynomials (dQl, dQr, dQm, dQo, dS1, dS2, dS3, dQkFixed, dQcp)
// are uploaded once during instance init and used via fast device-to-device copies
// in the hot quotient and linearized polynomial computation paths.

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/bits"
	"runtime"
	"sync"
	"time"
	"unsafe"

	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/hash_to_field"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/iop"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"

	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/constraint/solver"
	fcs "github.com/consensys/gnark/frontend/cs"
	"golang.org/x/sync/errgroup"

	plonk377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	"github.com/consensys/linea-monorepo/prover/gpu"
)

// Polynomial IDs matching gnark's internal ordering.
const (
	id_L int = iota
	id_R
	id_O
	id_Z
	id_ZS
	id_Ql
	id_Qr
	id_Qm
	id_Qo
	id_Qk
	id_S1
	id_S2
	id_S3
	id_Qci // Qcp[i] at id_Qci+2*i, Pi[i] at id_Qci+2*i+1
)

// Blinding polynomial orders.
const (
	orderBlindingL = 1
	orderBlindingR = 1
	orderBlindingO = 1
	orderBlindingZ = 2
)

// msmMaxPoints is the number of extra canonical SRS points beyond n needed
// for blinding. Z has blinding degree 2, so blinded Z needs n+3 coefficients.
// We use n+6 as a safe margin matching gnark's SRS convention.
const msmExtraPoints = 6

// ─────────────────────────────────────────────────────────────────────────────
// GPUProvingKey — slim wrapper: VerifyingKey + lazy instance
// ─────────────────────────────────────────────────────────────────────────────

// GPUProvingKey holds the data needed to generate GPU-accelerated PlonK proofs.
//
// Uses canonical-only MSM: a single canonical SRS MSM context handles all
// polynomial commitments. Wire/Z polynomials are converted from Lagrange to
// canonical via GPU iFFT before committing.
//
// Construct via NewGPUProvingKey or NewGPUProvingKeyFromPinned.
// Call Close() to release GPU resources.
type GPUProvingKey struct {
	mu sync.Mutex

	// Verifying key (needed for transcript binding and openings).
	Vk *plonk377.VerifyingKey

	// SRS data (consumed during instance initialization).
	Kzg             []G1TEPoint  // canonical SRS in TE format (nil'd after pinning)
	n               int          // domain size (power of 2)
	pinnedCanonical *G1MSMPoints // pinned SRS (nil'd after MSM takes ownership)

	// Lazily initialized instance (persists across proofs, freed on Close).
	inst *gpuInstance
}

// NewGPUProvingKey creates a GPUProvingKey from pre-converted TE SRS points
// and a verifying key. This is a lightweight constructor — circuit-specific data
// and GPU resources are initialized lazily on the first call to GPUProve.
//
// The domain size is derived from vk.Size. Only the canonical SRS is needed.
func NewGPUProvingKey(canonicalTE []G1TEPoint, vk *plonk377.VerifyingKey) *GPUProvingKey {
	n := 0
	if vk != nil {
		n = int(vk.Size)
	}
	gpk := &GPUProvingKey{
		Kzg: canonicalTE,
		Vk:  vk,
		n:   n,
	}
	runtime.SetFinalizer(gpk, (*GPUProvingKey).Close)
	return gpk
}

// NewGPUProvingKeyFromPinned creates a GPUProvingKey from pre-pinned SRS points.
// The canonical pinned memory is owned by the GPUProvingKey and freed on Close().
//
// This is the production path: SRSStore → pinned memory → GPUProvingKey.
func NewGPUProvingKeyFromPinned(canonicalPinned *G1MSMPoints, vk *plonk377.VerifyingKey) *GPUProvingKey {
	n := 0
	if vk != nil {
		n = int(vk.Size)
	}
	gpk := &GPUProvingKey{
		Vk:              vk,
		n:               n,
		pinnedCanonical: canonicalPinned,
	}
	runtime.SetFinalizer(gpk, (*GPUProvingKey).Close)
	return gpk
}

// Size returns the domain size n.
func (gpk *GPUProvingKey) Size() int {
	return gpk.n
}

// Prepare performs one-time circuit and GPU setup for proving on the given device.
// Callers can use this to move lazy initialization out of per-proof timing.
func (gpk *GPUProvingKey) Prepare(dev *gpu.Device, spr *cs.SparseR1CS) error {
	gpk.mu.Lock()
	defer gpk.mu.Unlock()

	if gpk.inst != nil && gpk.inst.dev == dev {
		return nil
	}
	if gpk.inst != nil {
		gpk.inst.close()
		gpk.inst = nil
	}
	inst, err := newGPUInstance(dev, gpk, spr)
	if err != nil {
		return err
	}
	gpk.inst = inst
	return nil
}

// Close releases all GPU resources held by this proving key.
func (gpk *GPUProvingKey) Close() {
	gpk.mu.Lock()
	defer gpk.mu.Unlock()

	if gpk.inst != nil {
		gpk.inst.close()
		gpk.inst = nil
	}
	if gpk.pinnedCanonical != nil {
		gpk.pinnedCanonical.Free()
		gpk.pinnedCanonical = nil
	}
	runtime.SetFinalizer(gpk, nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// gpuInstance — persistent GPU resources + circuit data + GPU-resident polys
//
// Created lazily on first GPUProve call. Persists across proofs — subsequent
// proofs skip init. Release with close().
//
//   ┌──────────────────────────────────────────────────────────────────────┐
//   │ gpuInstance                                                          │
//   │                                                                      │
//   │  GPU infrastructure:    MSM · FFT domain · permutation table        │
//   │                                                                      │
//   │  GPU-resident polys:    dQl dQr dQm dQo dS1 dS2 dS3 dQkFixed dQcp │
//   │  (uploaded once, d2d-copied per use — avoids hot-path H2D)         │
//   │                                                                      │
//   │  Host circuit data:     canonical selectors (for CPU eval at ζ)     │
//   └──────────────────────────────────────────────────────────────────────┘
// ─────────────────────────────────────────────────────────────────────────────

type gpuInstance struct {
	dev     *gpu.Device
	vk      *plonk377.VerifyingKey
	n       int
	log2n   uint
	domain0 *fft.Domain

	// GPU infrastructure
	msm    *G1MSM
	fftDom *GPUFFTDomain
	dPerm  unsafe.Pointer // device permutation table (3n int64s)

	// GPU-resident canonical polynomials (uploaded once during init)
	//
	// These are used in the hot quotient and linearized polynomial paths.
	// Each coset iteration copies from dXxx → working vector (d2d, ~100×
	// faster than host upload at large n) → CosetFFT in-place.
	dQl, dQr, dQm, dQo *FrVector
	dS1, dS2, dS3      *FrVector
	dQkFixed           *FrVector
	dQcp               []*FrVector

	// Host canonical data (for CPU polynomial evaluation at ζ, batch opening, etc.)
	qlCanonical, qrCanonical, qmCanonical, qoCanonical fr.Vector
	qkFixedCanonical                                   fr.Vector
	s1Canonical, s2Canonical, s3Canonical              fr.Vector
	qcpCanonical                                       []fr.Vector
	qkLagrange                                         fr.Vector
	permutation                                        []int64
	nbPublicVariables                                  int
	commitmentInfo                                     []uint64

	// Pre-allocated host buffers reused across proofs (single-proof-at-a-time).
	// Eliminates ~3.3 GiB of per-proof Go heap allocations at n=2^23.
	hBufs hostBufs
}

// hostBufs holds pre-allocated host memory buffers reused across proofs.
// Eliminates per-proof GC pressure from large transient allocations.
//
//	Buffer            Size        Used by
//	─────────────     ─────────   ───────────────────────────────
//	lCanonical        n           commitToLRO (iFFT output)
//	rCanonical        n           commitToLRO
//	oCanonical        n           commitToLRO
//	zLagrange         n           buildZGPU (Z Lagrange form)
//	qkCanonical       n           commitToLRO (Qk canonical)
//	lBlinded          n+2         commitToLRO (blinded L̃)
//	rBlinded          n+2         commitToLRO (blinded R̃)
//	oBlinded          n+2         commitToLRO (blinded Õ)
//	zBlinded          n+3         buildZAndCommit (blinded Z̃)
//	hFull             max(4n,     computeNumeratorGPU (quotient)
//	                  3(n+2))
//	openZBuf          n+3         openAndFinalize (Horner quotient)
type hostBufs struct {
	lCanonical, rCanonical, oCanonical fr.Vector
	zLagrange                          fr.Vector
	qkCanonical                        fr.Vector
	qkCoeffs                           fr.Vector // Qk Lagrange (patched per proof)
	lBlinded, rBlinded, oBlinded       []fr.Element
	zBlinded                           []fr.Element
	hFull                              []fr.Element
	openZBuf                           []fr.Element
}

func (inst *gpuInstance) initHostBufs() {
	n := inst.n
	inst.hBufs = hostBufs{
		lCanonical:  make(fr.Vector, n),
		rCanonical:  make(fr.Vector, n),
		oCanonical:  make(fr.Vector, n),
		zLagrange:   make(fr.Vector, n),
		qkCanonical: make(fr.Vector, n),
		qkCoeffs:    make(fr.Vector, n),
		lBlinded:    make([]fr.Element, n+1+orderBlindingL),
		rBlinded:    make([]fr.Element, n+1+orderBlindingR),
		oBlinded:    make([]fr.Element, n+1+orderBlindingO),
		zBlinded:    make([]fr.Element, n+1+orderBlindingZ),
		openZBuf:    make([]fr.Element, n+1+orderBlindingZ),
	}
	hSize := 4 * n
	if needed := 3 * (n + 2); needed > hSize {
		hSize = needed
	}
	inst.hBufs.hFull = make([]fr.Element, hSize)
}

// newGPUInstance creates a fully initialized instance. Consumes SRS data from gpk.
func newGPUInstance(dev *gpu.Device, gpk *GPUProvingKey, spr *cs.SparseR1CS) (*gpuInstance, error) {
	return newGPUInstanceWithBSB22Ready(dev, gpk, spr, nil)
}

// newGPUInstanceWithBSB22Ready creates a fully initialized GPU instance.
//
// If onBSB22Ready is non-nil, it is called exactly once when the instance is
// safe for BSB22 hint work. This intentionally waits for full initialization:
// setup and BSB22 both use the same CUDA context and shared transfer staging,
// so overlapping them would make host-side API calls race even though kernels
// are stream-ordered on the device.
//
// Initialization pipeline:
//  1. Parse SparseR1CS → Lagrange-form selectors + permutation
//  2. Pin SRS → create MSM context (transfers SRS ownership)
//  3. Create FFT domain
//  4. Upload permutation table
//  5. GPU iFFT: Lagrange → canonical for all selectors (~20x faster than CPU)
//  6. Upload canonical polys to GPU FrVectors (one-time, persists across proofs)
//  7. Initialize multi-stream
//  8. Pre-allocate host buffers  ← BSB22-ready
func newGPUInstanceWithBSB22Ready(
	dev *gpu.Device,
	gpk *GPUProvingKey,
	spr *cs.SparseR1CS,
	onBSB22Ready func(*gpuInstance),
) (*gpuInstance, error) {
	inst := &gpuInstance{
		dev: dev,
		vk:  gpk.Vk,
		n:   gpk.n,
	}

	fail := func(msg string, err error) (*gpuInstance, error) {
		inst.close()
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// Step 1: circuit data from SparseR1CS
	if err := inst.initCircuitData(spr); err != nil {
		return fail("init circuit data", err)
	}

	// Step 2: pin SRS and create MSM
	if gpk.pinnedCanonical == nil && gpk.Kzg != nil {
		var err error
		gpk.pinnedCanonical, err = PinG1TEPoints(gpk.Kzg)
		if err != nil {
			return fail("pin canonical SRS", err)
		}
	}
	gpk.Kzg = nil // free Go heap copy

	if gpk.pinnedCanonical == nil {
		inst.close()
		return nil, errors.New("no SRS data available (already consumed)")
	}

	msmSize := inst.n + msmExtraPoints
	if msmSize > gpk.pinnedCanonical.N {
		msmSize = gpk.pinnedCanonical.N
	}
	var err error
	inst.msm, err = NewG1MSMN(dev, gpk.pinnedCanonical, msmSize)
	if err != nil {
		return fail("create canonical MSM", err)
	}
	gpk.pinnedCanonical = nil // ownership transferred to MSM

	// Pin work buffers across the wave of commitments that runs before the
	// quotient phase. Released in computeQuotientGPU around OffloadPoints
	// (the quotient needs that VRAM back). Saves ~5–10 ms per MSM call vs
	// lazy alloc/free + cudaHostRegister/Unregister.
	if perr := inst.msm.PinWorkBuffers(); perr != nil {
		return fail("pin MSM work buffers", perr)
	}

	// Step 3: FFT domain (enough for BSB22 hint iFFT path)
	inst.fftDom, err = NewFFTDomain(dev, inst.n)
	if err != nil {
		return fail("create FFT domain", err)
	}

	// Step 4: upload permutation table
	inst.dPerm, err = DeviceAllocCopyInt64(dev, inst.permutation)
	if err != nil {
		return fail("upload permutation", err)
	}

	// Step 5: GPU iFFT Lagrange → canonical
	if err := inst.initCanonicalGPU(); err != nil {
		return fail("init canonical", err)
	}

	// Step 6: upload canonical polys to GPU
	if err := inst.uploadPolynomials(); err != nil {
		return fail("upload polynomials", err)
	}

	// Step 7: initialize multi-stream (gpu.StreamTransfer, gpu.StreamMSM)
	if err := dev.InitMultiStream(); err != nil {
		return fail("init multi-stream", err)
	}

	// Step 8: pre-allocate host buffers (eliminates ~3 GiB per-proof GC pressure).
	inst.initHostBufs()

	if onBSB22Ready != nil {
		onBSB22Ready(inst)
	}

	return inst, nil
}

// initCircuitData parses the constraint system and builds Lagrange-form selectors.
func (inst *gpuInstance) initCircuitData(spr *cs.SparseR1CS) error {
	nbConstraints := spr.GetNbConstraints()
	sizeSystem := uint64(nbConstraints + len(spr.Public))
	inst.domain0 = fft.NewDomain(sizeSystem, fft.WithoutPrecompute())
	n := int(inst.domain0.Cardinality)

	if n != inst.n {
		return fmt.Errorf("domain size mismatch: spr gives %d, SRS has %d points", n, inst.n)
	}

	inst.log2n = uint(bits.TrailingZeros(uint(n)))

	trace := plonk377.NewTrace(spr, inst.domain0)

	// Take ownership of Lagrange coefficients directly — no Clone, no CPU iFFT.
	// Converted to canonical form on GPU in initCanonicalGPU (~20x faster).
	inst.qlCanonical = fr.Vector(trace.Ql.Coefficients())
	inst.qrCanonical = fr.Vector(trace.Qr.Coefficients())
	inst.qmCanonical = fr.Vector(trace.Qm.Coefficients())
	inst.qoCanonical = fr.Vector(trace.Qo.Coefficients())
	inst.s1Canonical = fr.Vector(trace.S1.Coefficients())
	inst.s2Canonical = fr.Vector(trace.S2.Coefficients())
	inst.s3Canonical = fr.Vector(trace.S3.Coefficients())

	// Qk needs a separate Lagrange copy (per-proof public input patching).
	inst.qkLagrange = make(fr.Vector, n)
	copy(inst.qkLagrange, trace.Qk.Coefficients())
	inst.qkFixedCanonical = fr.Vector(trace.Qk.Coefficients())

	inst.qcpCanonical = make([]fr.Vector, len(trace.Qcp))
	for i, p := range trace.Qcp {
		inst.qcpCanonical[i] = fr.Vector(p.Coefficients())
	}
	inst.permutation = trace.S
	inst.nbPublicVariables = len(spr.Public)
	inst.commitmentInfo = inst.vk.CommitmentConstraintIndexes

	return nil
}

// initCanonicalGPU converts selector polynomials from Lagrange to canonical form
// using GPU iFFT. ~20x faster than CPU iFFT at large sizes (e.g. n=2^26: ~2s vs ~40s).
func (inst *gpuInstance) initCanonicalGPU() error {
	n := inst.n
	gpuWork, err := NewFrVector(inst.dev, n)
	if err != nil {
		return fmt.Errorf("alloc work vector: %w", err)
	}
	defer gpuWork.Free()

	iFFTSelector := func(v fr.Vector) {
		gpuWork.CopyFromHost(v)
		inst.fftDom.BitReverse(gpuWork)
		inst.fftDom.FFTInverse(gpuWork)
		gpuWork.CopyToHost(v)
	}

	iFFTSelector(inst.qlCanonical)
	iFFTSelector(inst.qrCanonical)
	iFFTSelector(inst.qmCanonical)
	iFFTSelector(inst.qoCanonical)
	iFFTSelector(inst.qkFixedCanonical)
	iFFTSelector(inst.s1Canonical)
	iFFTSelector(inst.s2Canonical)
	iFFTSelector(inst.s3Canonical)
	for _, v := range inst.qcpCanonical {
		iFFTSelector(v)
	}

	if err := inst.dev.Sync(); err != nil {
		return fmt.Errorf("sync after canonical conversion: %w", err)
	}
	return nil
}

// uploadPolynomials uploads canonical selector polynomials to GPU FrVectors.
// These persist across proofs and are used via fast d2d copies in hot paths.
func (inst *gpuInstance) uploadPolynomials() error {
	upload := func(hostData fr.Vector) (*FrVector, error) {
		v, err := NewFrVector(inst.dev, inst.n)
		if err != nil {
			return nil, err
		}
		v.CopyFromHost(hostData)
		return v, nil
	}

	var err error
	inst.dQl, err = upload(inst.qlCanonical)
	if err != nil {
		return fmt.Errorf("upload ql: %w", err)
	}
	inst.dQr, err = upload(inst.qrCanonical)
	if err != nil {
		return fmt.Errorf("upload qr: %w", err)
	}
	inst.dQm, err = upload(inst.qmCanonical)
	if err != nil {
		return fmt.Errorf("upload qm: %w", err)
	}
	inst.dQo, err = upload(inst.qoCanonical)
	if err != nil {
		return fmt.Errorf("upload qo: %w", err)
	}
	inst.dS1, err = upload(inst.s1Canonical)
	if err != nil {
		return fmt.Errorf("upload s1: %w", err)
	}
	inst.dS2, err = upload(inst.s2Canonical)
	if err != nil {
		return fmt.Errorf("upload s2: %w", err)
	}
	inst.dS3, err = upload(inst.s3Canonical)
	if err != nil {
		return fmt.Errorf("upload s3: %w", err)
	}
	inst.dQkFixed, err = upload(inst.qkFixedCanonical)
	if err != nil {
		return fmt.Errorf("upload qkFixed: %w", err)
	}

	inst.dQcp = make([]*FrVector, len(inst.qcpCanonical))
	for i, v := range inst.qcpCanonical {
		inst.dQcp[i], err = upload(v)
		if err != nil {
			return fmt.Errorf("upload qcp[%d]: %w", i, err)
		}
	}

	return nil
}

// close releases all GPU resources held by this instance.
func (inst *gpuInstance) close() {
	if inst.msm != nil {
		inst.msm.Close() // also frees owned pinned points
		inst.msm = nil
	}
	if inst.fftDom != nil {
		inst.fftDom.Close()
		inst.fftDom = nil
	}
	if inst.dPerm != nil {
		DeviceFreePtr(inst.dPerm)
		inst.dPerm = nil
	}
	// Free GPU-resident polynomials
	for _, v := range []*FrVector{inst.dQl, inst.dQr, inst.dQm, inst.dQo,
		inst.dS1, inst.dS2, inst.dS3, inst.dQkFixed} {
		if v != nil {
			v.Free()
		}
	}
	inst.dQl, inst.dQr, inst.dQm, inst.dQo = nil, nil, nil, nil
	inst.dS1, inst.dS2, inst.dS3, inst.dQkFixed = nil, nil, nil, nil
	for _, v := range inst.dQcp {
		if v != nil {
			v.Free()
		}
	}
	inst.dQcp = nil
}

// ─────────────────────────────────────────────────────────────────────────────
// gpuProver — per-proof mutable state; methods implement prove phases
//
// Created fresh for each GPUProve call. Holds transient data that lives only
// for the duration of a single proof. Methods are called from errgroup
// goroutines in GPUProve, matching gnark's instance method pattern.
// ─────────────────────────────────────────────────────────────────────────────

type gpuProver struct {
	inst *gpuInstance

	proof plonk377.Proof
	fs    *fiatshamir.Transcript

	// BSB22
	preSolved      *preSolvedData
	commitmentInfo constraint.PlonkCommitments
	commitmentVal  []fr.Element
	pi2Canonical   [][]fr.Element

	// Per-proof polynomial state
	evalL, evalR, evalO          fr.Vector
	wWitness                     fr.Vector
	bpL, bpR, bpO, bpZ           *iop.Polynomial
	qkCoeffs                     fr.Vector
	qkCanonicalHost              fr.Vector
	qkCanonicalGPU               *FrVector
	lBlinded, rBlinded, oBlinded []fr.Element
	zBlinded                     []fr.Element
	gpuWork                      *FrVector
	h1, h2, h3                   []fr.Element
	gamma, beta, alpha, zeta     fr.Element

	logTime func(string)
}

// cleanup frees per-proof GPU resources.
func (p *gpuProver) cleanup() {
	if p.gpuWork != nil {
		p.gpuWork.Free()
		p.gpuWork = nil
	}
	if p.qkCanonicalGPU != nil {
		p.qkCanonicalGPU.Free()
		p.qkCanonicalGPU = nil
	}
}

func runGPUProveStep(label string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s panic: %v", label, r)
		}
	}()
	return fn()
}

// ─── gpuProver methods (each prove phase) ────────────────────────────────────

// initBlindingPolynomials generates random blinding polynomials.
func (p *gpuProver) initBlindingPolynomials() {
	p.bpL = getRandomPolynomial(orderBlindingL)
	p.bpR = getRandomPolynomial(orderBlindingR)
	p.bpO = getRandomPolynomial(orderBlindingO)
	p.bpZ = getRandomPolynomial(orderBlindingZ)
}

// solvePreSolved copies pre-solved data into the prover.
func (p *gpuProver) solvePreSolved() error {
	ps := p.preSolved
	if len(ps.commitmentVal) != len(p.commitmentVal) {
		return fmt.Errorf("pre-solved commitment value mismatch: have=%d want=%d",
			len(ps.commitmentVal), len(p.commitmentVal))
	}
	if len(ps.pi2Canonical) != len(p.commitmentInfo) {
		return fmt.Errorf("pre-solved pi2 mismatch: have=%d want=%d",
			len(ps.pi2Canonical), len(p.commitmentInfo))
	}
	if len(ps.bsb22Commitments) != len(p.proof.Bsb22Commitments) {
		return fmt.Errorf("pre-solved BSB22 commitment mismatch: have=%d want=%d",
			len(ps.bsb22Commitments), len(p.proof.Bsb22Commitments))
	}
	p.evalL = ps.evalL
	p.evalR = ps.evalR
	p.evalO = ps.evalO
	p.wWitness = ps.wWitness
	p.pi2Canonical = ps.pi2Canonical
	copy(p.commitmentVal, ps.commitmentVal)
	copy(p.proof.Bsb22Commitments, ps.bsb22Commitments)
	return nil
}

// solve runs constraint solving with BSB22 commitment callbacks.
func (p *gpuProver) solve(
	spr *cs.SparseR1CS,
	fullWitness witness.Witness,
	waitBSB22Inst func() (*gpuInstance, error),
) error {
	var (
		bsb22Inst    *gpuInstance
		bsb22Err     error
		bsb22Once    sync.Once
		getBSB22Inst = func() (*gpuInstance, error) {
			bsb22Once.Do(func() {
				bsb22Inst, bsb22Err = waitBSB22Inst()
			})
			return bsb22Inst, bsb22Err
		}
	)

	var solverOpts []solver.Option
	if len(p.commitmentInfo) > 0 {
		htfFunc := hash_to_field.New([]byte("BSB22-Plonk"))
		bsb22ID := solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder)
		solverOpts = []solver.Option{solver.OverrideHint(bsb22ID, func(_ *big.Int, ins, outs []*big.Int) error {
			inst, err := getBSB22Inst()
			if err != nil {
				return err
			}

			commDepth := int(ins[0].Int64())
			ins = ins[1:]

			ci := p.commitmentInfo[commDepth]
			committedValues := make([]fr.Element, inst.domain0.Cardinality)
			offset := inst.nbPublicVariables
			for i := range ins {
				committedValues[offset+ci.Committed[i]].SetBigInt(ins[i])
			}
			committedValues[offset+ci.CommitmentIndex].SetRandom()
			committedValues[offset+spr.GetNbConstraints()-1].SetRandom()

			// iFFT to canonical, then canonical MSM (no blinding for BSB22).
			n := inst.n
			p.gpuWork.CopyFromHost(fr.Vector(committedValues[:n]))
			inst.fftDom.BitReverse(p.gpuWork)
			inst.fftDom.FFTInverse(p.gpuWork)
			canonicalBuf := make(fr.Vector, n)
			p.gpuWork.CopyToHost(canonicalBuf)
			p.pi2Canonical[commDepth] = canonicalBuf

			jacs := inst.msm.MultiExp(canonicalBuf)
			p.proof.Bsb22Commitments[commDepth].FromJacobian(&jacs[0])

			htfFunc.Write(p.proof.Bsb22Commitments[commDepth].Marshal())
			hashBts := htfFunc.Sum(nil)
			htfFunc.Reset()
			nbBuf := fr.Bytes
			if htfFunc.Size() < fr.Bytes {
				nbBuf = htfFunc.Size()
			}
			p.commitmentVal[commDepth].SetBytes(hashBts[:nbBuf])
			p.commitmentVal[commDepth].BigInt(outs[0])
			return nil
		})}
	}

	solution_, err := spr.Solve(fullWitness, solverOpts...)
	if err != nil {
		return fmt.Errorf("solve: %w", err)
	}
	solution := solution_.(*cs.SparseR1CSSolution)
	p.evalL = fr.Vector(solution.L)
	p.evalR = fr.Vector(solution.R)
	p.evalO = fr.Vector(solution.O)

	var ok bool
	p.wWitness, ok = fullWitness.Vector().(fr.Vector)
	if !ok {
		return errors.New("invalid witness type")
	}
	return nil
}

// completeQk patches Qk with public inputs + BSB22 commitment values.
func (p *gpuProver) completeQk() error {
	inst := p.inst
	n := inst.n

	p.qkCoeffs = inst.hBufs.qkCoeffs
	copy(p.qkCoeffs, inst.qkLagrange)
	copy(p.qkCoeffs, p.wWitness[:inst.nbPublicVariables])
	for i := range p.commitmentInfo {
		p.qkCoeffs[inst.nbPublicVariables+p.commitmentInfo[i].CommitmentIndex] = p.commitmentVal[i]
	}

	for i := range p.commitmentInfo {
		if len(p.pi2Canonical[i]) != n {
			return fmt.Errorf("missing BSB22 canonical polynomial %d", i)
		}
	}
	return nil
}

// commitToLRO performs GPU iFFT on L,R,O (overlaps with completeQk), then
// waits for Qk and blinding, blinds wire polynomials, and commits [L],[R],[O].
func (p *gpuProver) commitToLRO(waitQk, waitBlinding func() error) error {
	inst := p.inst
	hb := &inst.hBufs

	// GPU iFFT L,R,O into pre-allocated canonical buffers.
	gpuToCanonical := func(lagrange, dst fr.Vector) {
		p.gpuWork.CopyFromHost(lagrange)
		inst.fftDom.BitReverse(p.gpuWork)
		inst.fftDom.FFTInverse(p.gpuWork)
		p.gpuWork.CopyToHost(dst)
	}

	gpuToCanonical(p.evalL, hb.lCanonical)
	gpuToCanonical(p.evalR, hb.rCanonical)
	gpuToCanonical(p.evalO, hb.oCanonical)

	// Wait for Qk to be patched, then iFFT it.
	if err := waitQk(); err != nil {
		return err
	}
	p.gpuWork.CopyFromHost(p.qkCoeffs)
	inst.fftDom.BitReverse(p.gpuWork)
	inst.fftDom.FFTInverse(p.gpuWork)
	if qkGPU, err := NewFrVector(inst.dev, inst.n); err == nil {
		qkGPU.CopyFromDevice(p.gpuWork)
		p.qkCanonicalGPU = qkGPU
		p.qkCanonicalHost = nil
	} else {
		p.gpuWork.CopyToHost(hb.qkCanonical)
		p.qkCanonicalHost = hb.qkCanonical
	}
	p.qkCoeffs = nil

	// Wait for blinding polynomials.
	if err := waitBlinding(); err != nil {
		return err
	}

	// Blind: p(X) + b(X)*(X^n - 1) in canonical form.
	// Uses pre-allocated buffers to avoid ~1 GiB per-proof allocations.
	var blindWG sync.WaitGroup
	blindWG.Add(3)
	go func() { defer blindWG.Done(); p.lBlinded = blindInto(hb.lBlinded, hb.lCanonical, p.bpL) }()
	go func() { defer blindWG.Done(); p.rBlinded = blindInto(hb.rBlinded, hb.rCanonical, p.bpR) }()
	go func() { defer blindWG.Done(); p.oBlinded = blindInto(hb.oBlinded, hb.oCanonical, p.bpO) }()
	blindWG.Wait()

	p.logTime("iFFT L,R,O,Qk + blind")

	// Commit [L], [R], [O] (batched canonical MSM).
	lroCommits := gpuCommitN(inst.msm, p.lBlinded, p.rBlinded, p.oBlinded)
	p.proof.LRO[0] = lroCommits[0]
	p.proof.LRO[1] = lroCommits[1]
	p.proof.LRO[2] = lroCommits[2]

	p.logTime("MSM commit L,R,O")
	return nil
}

// deriveGammaBeta derives Fiat-Shamir challenges γ and β.
func (p *gpuProver) deriveGammaBeta() error {
	inst := p.inst

	if err := bindPublicData(p.fs, "gamma", inst.vk, p.wWitness[:inst.nbPublicVariables]); err != nil {
		return err
	}
	var err error
	p.gamma, err = deriveRandomness(p.fs, "gamma", &p.proof.LRO[0], &p.proof.LRO[1], &p.proof.LRO[2])
	if err != nil {
		return err
	}
	p.beta, err = deriveRandomness(p.fs, "beta")
	if err != nil {
		return err
	}
	p.wWitness = nil

	p.logTime("derive gamma,beta")
	return nil
}

// buildZAndCommit builds the Z permutation polynomial, converts to canonical,
// blinds, commits, and derives α.
func (p *gpuProver) buildZAndCommit() error {
	inst := p.inst

	zLagrange, err := buildZGPU(inst, p.gpuWork,
		p.evalL, p.evalR, p.evalO, p.beta, p.gamma)
	if err != nil {
		return fmt.Errorf("build Z: %w", err)
	}
	p.evalL = nil
	p.evalR = nil
	p.evalO = nil

	p.logTime("build Z")

	// iFFT Z → canonical, blind, commit (using pre-allocated buffers).
	hb := &inst.hBufs
	p.gpuWork.CopyFromHost(zLagrange)
	inst.fftDom.BitReverse(p.gpuWork)
	inst.fftDom.FFTInverse(p.gpuWork)
	p.gpuWork.CopyToHost(hb.zLagrange) // reuse zLagrange as canonical scratch
	p.zBlinded = blindInto(hb.zBlinded, hb.zLagrange, p.bpZ)

	p.proof.Z = gpuCommit(inst.msm, p.zBlinded)

	p.logTime("iFFT+commit Z")

	// Derive α.
	alphaDeps := make([]*curve.G1Affine, len(p.proof.Bsb22Commitments)+1)
	for i := range p.proof.Bsb22Commitments {
		alphaDeps[i] = &p.proof.Bsb22Commitments[i]
	}
	alphaDeps[len(alphaDeps)-1] = &p.proof.Z
	p.alpha, err = deriveRandomness(p.fs, "alpha", alphaDeps...)
	if err != nil {
		return err
	}

	p.logTime("derive alpha")
	return nil
}

// computeQuotientAndCommit computes the quotient polynomial, commits h1,h2,h3,
// and derives ζ.
func (p *gpuProver) computeQuotientAndCommit() error {
	inst := p.inst
	n := inst.n

	if n < 8 {
		return fmt.Errorf("domain size %d too small for GPU prover (minimum 8)", n)
	}

	// Offload MSM points and release work buffers during quotient compute
	// (saves ~6 GiB of points + multi-GB sort buffers at large n).
	inst.msm.OffloadPoints()
	if rerr := inst.msm.ReleaseWorkBuffers(); rerr != nil {
		return fmt.Errorf("release MSM work buffers: %w", rerr)
	}
	pointsOffloaded := true
	defer func() {
		if pointsOffloaded {
			inst.msm.ReloadPoints()
			// Re-pin for the post-quotient commit wave (h1/h2/h3, linPol, etc.).
			_ = inst.msm.PinWorkBuffers()
		}
	}()

	var qErr error
	p.h1, p.h2, p.h3, qErr = computeNumeratorGPU(
		inst, p.gpuWork,
		p.lBlinded, p.rBlinded, p.oBlinded, p.zBlinded,
		p.qkCanonicalHost, p.qkCanonicalGPU, p.pi2Canonical,
		p.alpha, p.beta, p.gamma,
	)
	if qErr != nil {
		return fmt.Errorf("compute quotient: %w", qErr)
	}
	p.qkCanonicalHost = nil
	if p.qkCanonicalGPU != nil {
		p.qkCanonicalGPU.Free()
		p.qkCanonicalGPU = nil
	}

	p.logTime("quotient GPU")

	// Reload MSM points and commit h1, h2, h3 (batched canonical MSM).
	inst.msm.ReloadPoints()
	if perr := inst.msm.PinWorkBuffers(); perr != nil {
		return fmt.Errorf("re-pin MSM work buffers: %w", perr)
	}
	pointsOffloaded = false
	hCommits := gpuCommitN(inst.msm, p.h1, p.h2, p.h3)
	p.proof.H[0] = hCommits[0]
	p.proof.H[1] = hCommits[1]
	p.proof.H[2] = hCommits[2]

	p.logTime("MSM commit h1,h2,h3")

	// Derive ζ.
	var zetaErr error
	p.zeta, zetaErr = deriveRandomness(p.fs, "zeta", &p.proof.H[0], &p.proof.H[1], &p.proof.H[2])
	if zetaErr != nil {
		return zetaErr
	}
	return nil
}

// openAndFinalize opens Z at ωζ, evaluates + linearizes, commits linearized
// polynomial, and performs the batch opening.
//
// Memory-aware scheduling: linearized poly (2 FrVectors) completes BEFORE
// Z opening MSM (transient sort buffers). This prevents overlapping GPU
// memory pressure that would exceed VRAM at n=2^27.
//
//	CPU goroutine:  [─── Horner quotient(Z) ───]→ bzuzeta
//	CPU main:       [── eval at ζ (5+ polys) ──]→ wait bzuzeta
//	GPU main:                                      [linearized poly]→ free 2 FrVectors
//	GPU main:                                                         [Z opening MSM]
func (p *gpuProver) openAndFinalize() error {
	inst := p.inst

	var zetaShifted fr.Element
	zetaShifted.Mul(&p.zeta, &inst.domain0.Generator)

	// Goroutine: Horner quotient for Z opening (CPU-only, no GPU).
	// Uses pre-allocated buffer — parallelHornerQuotient modifies in place.
	openZPoly := inst.hBufs.openZBuf[:len(p.zBlinded)]
	copy(openZPoly, p.zBlinded)
	bzuzetaCh := make(chan fr.Element, 1)
	go func() {
		parallelHornerQuotient(openZPoly, zetaShifted)
		bzuzetaCh <- openZPoly[0]
	}()

	// Evaluate polynomials at ζ:
	//   GPU-resident (s1, s2, qcp) → PolyEvalGPU (no upload, data on device)
	//   Host blinded (L̃, R̃, Õ)   → CPU concurrent Horner
	//
	// GPU and CPU work overlaps: GPU evals run on device while CPU goroutines
	// run Horner on host. This is faster than running all on CPU at large n.
	var blzeta, brzeta, bozeta, s1Zeta, s2Zeta fr.Element
	qcpzeta := make([]fr.Element, len(p.commitmentInfo))

	// Start CPU evals concurrently.
	var evalWG sync.WaitGroup
	evalWG.Add(3)
	go func() { defer evalWG.Done(); blzeta = parallelEvaluateCanonical(p.lBlinded, p.zeta) }()
	go func() { defer evalWG.Done(); brzeta = parallelEvaluateCanonical(p.rBlinded, p.zeta) }()
	go func() { defer evalWG.Done(); bozeta = parallelEvaluateCanonical(p.oBlinded, p.zeta) }()

	// GPU evals from resident FrVectors (overlaps with CPU work above).
	s1Zeta = PolyEvalGPU(inst.dev, inst.dS1, p.zeta)
	s2Zeta = PolyEvalGPU(inst.dev, inst.dS2, p.zeta)
	for i := range p.commitmentInfo {
		qcpzeta[i] = PolyEvalGPU(inst.dev, inst.dQcp[i], p.zeta)
	}

	evalWG.Wait()

	bzuzeta := <-bzuzetaCh
	p.proof.ZShiftedOpening.ClaimedValue.Set(&bzuzeta)

	// Linearized polynomial (GPU, 2 FrVectors — freed on return).
	linPol := gpuComputeLinearizedPoly(
		inst,
		blzeta, brzeta, bozeta, p.alpha, p.beta, p.gamma, p.zeta, bzuzeta,
		s1Zeta, s2Zeta,
		qcpzeta, p.zBlinded, p.pi2Canonical,
		p.h1, p.h2, p.h3,
	)
	p.h1 = nil
	p.h2 = nil
	p.h3 = nil
	p.pi2Canonical = nil

	// Z opening MSM (runs AFTER linearized poly frees its FrVectors).
	p.proof.ZShiftedOpening.H = gpuCommit(inst.msm, openZPoly[1:])

	p.logTime("eval+linearize+open Z")

	// Commit linearized polynomial.
	linPolZetaCh := make(chan fr.Element, 1)
	go func() {
		linPolZetaCh <- parallelEvaluateCanonical(linPol, p.zeta)
	}()

	linPolDigest := gpuCommit(inst.msm, linPol)

	p.logTime("MSM commit linPol")

	// Batch opening (GPU MSM).
	nPolysToOpen := 6 + len(inst.qcpCanonical)
	claimedValues := make([]fr.Element, nPolysToOpen)
	claimedValues[0] = <-linPolZetaCh
	claimedValues[1] = blzeta
	claimedValues[2] = brzeta
	claimedValues[3] = bozeta
	claimedValues[4] = s1Zeta
	claimedValues[5] = s2Zeta
	for i := range inst.qcpCanonical {
		claimedValues[6+i] = qcpzeta[i]
	}

	polysToOpen := make([][]fr.Element, nPolysToOpen)
	polysToOpen[0] = linPol
	polysToOpen[1] = p.lBlinded
	polysToOpen[2] = p.rBlinded
	polysToOpen[3] = p.oBlinded
	polysToOpen[4] = inst.s1Canonical
	polysToOpen[5] = inst.s2Canonical
	for i := range inst.qcpCanonical {
		polysToOpen[6+i] = inst.qcpCanonical[i]
	}

	digestsToOpen := make([]curve.G1Affine, nPolysToOpen)
	digestsToOpen[0] = linPolDigest
	digestsToOpen[1] = p.proof.LRO[0]
	digestsToOpen[2] = p.proof.LRO[1]
	digestsToOpen[3] = p.proof.LRO[2]
	digestsToOpen[4] = inst.vk.S[0]
	digestsToOpen[5] = inst.vk.S[1]
	copy(digestsToOpen[6:], inst.vk.Qcp)

	var err error
	p.proof.BatchedProof, err = gpuBatchOpen(
		inst.msm,
		polysToOpen,
		digestsToOpen,
		claimedValues,
		p.zeta,
		p.proof.ZShiftedOpening.ClaimedValue.Marshal(),
	)
	if err != nil {
		return fmt.Errorf("batch opening: %w", err)
	}

	p.logTime("batch opening (GPU)")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Pre-solved witness support
// ─────────────────────────────────────────────────────────────────────────────

type preSolvedData struct {
	evalL, evalR, evalO fr.Vector
	wWitness            fr.Vector
	commitmentVal       []fr.Element
	pi2Canonical        [][]fr.Element
	bsb22Commitments    []kzg.Digest
}

type preSolvedWitness struct {
	witness.Witness
	data *preSolvedData
}

func (w *preSolvedWitness) preSolvedData() *preSolvedData {
	return w.data
}

// GPUPreSolve runs witness solving and BSB22 hint commitments ahead of proving.
// The returned witness can be passed to GPUProve to skip the internal solve phase.
func GPUPreSolve(dev *gpu.Device, gpk *GPUProvingKey, spr *cs.SparseR1CS, fullWitness witness.Witness) (witness.Witness, error) {
	if fullWitness == nil {
		return nil, errors.New("nil witness")
	}
	if err := gpk.Prepare(dev, spr); err != nil {
		return nil, err
	}
	inst := gpk.inst
	n := inst.n

	var commitmentInfo constraint.PlonkCommitments
	if spr.CommitmentInfo != nil {
		commitmentInfo = spr.CommitmentInfo.(constraint.PlonkCommitments)
	}
	commitmentVal := make([]fr.Element, len(commitmentInfo))
	pi2Canonical := make([][]fr.Element, len(commitmentInfo))
	bsb22Commitments := make([]kzg.Digest, len(commitmentInfo))

	gpuWork, err := NewFrVector(dev, n)
	if err != nil {
		return nil, fmt.Errorf("alloc pre-solve work vector: %w", err)
	}
	defer gpuWork.Free()

	var solverOpts []solver.Option
	if len(commitmentInfo) > 0 {
		htfFunc := hash_to_field.New([]byte("BSB22-Plonk"))
		bsb22ID := solver.GetHintID(fcs.Bsb22CommitmentComputePlaceholder)
		solverOpts = []solver.Option{solver.OverrideHint(bsb22ID, func(_ *big.Int, ins, outs []*big.Int) error {
			commDepth := int(ins[0].Int64())
			ins = ins[1:]
			if commDepth < 0 || commDepth >= len(commitmentInfo) {
				return fmt.Errorf("invalid commitment depth %d", commDepth)
			}

			ci := commitmentInfo[commDepth]
			committedValues := make([]fr.Element, inst.domain0.Cardinality)
			offset := inst.nbPublicVariables
			for i := range ins {
				committedValues[offset+ci.Committed[i]].SetBigInt(ins[i])
			}
			committedValues[offset+ci.CommitmentIndex].SetRandom()
			committedValues[offset+spr.GetNbConstraints()-1].SetRandom()

			gpuWork.CopyFromHost(fr.Vector(committedValues[:n]))
			inst.fftDom.BitReverse(gpuWork)
			inst.fftDom.FFTInverse(gpuWork)
			canonicalBuf := make(fr.Vector, n)
			gpuWork.CopyToHost(canonicalBuf)
			pi2Canonical[commDepth] = canonicalBuf

			jacs := inst.msm.MultiExp(canonicalBuf)
			bsb22Commitments[commDepth].FromJacobian(&jacs[0])

			htfFunc.Write(bsb22Commitments[commDepth].Marshal())
			hashBts := htfFunc.Sum(nil)
			htfFunc.Reset()
			nbBuf := fr.Bytes
			if htfFunc.Size() < fr.Bytes {
				nbBuf = htfFunc.Size()
			}
			commitmentVal[commDepth].SetBytes(hashBts[:nbBuf])
			commitmentVal[commDepth].BigInt(outs[0])
			return nil
		})}
	}

	solution_, err := spr.Solve(fullWitness, solverOpts...)
	if err != nil {
		return nil, fmt.Errorf("solve: %w", err)
	}
	solution := solution_.(*cs.SparseR1CSSolution)

	wWitness, ok := fullWitness.Vector().(fr.Vector)
	if !ok {
		return nil, errors.New("invalid witness type")
	}
	for i := range commitmentInfo {
		if len(pi2Canonical[i]) != n {
			return nil, fmt.Errorf("missing BSB22 canonical polynomial %d", i)
		}
	}

	data := &preSolvedData{
		evalL:            fr.Vector(solution.L),
		evalR:            fr.Vector(solution.R),
		evalO:            fr.Vector(solution.O),
		wWitness:         wWitness,
		commitmentVal:    commitmentVal,
		pi2Canonical:     pi2Canonical,
		bsb22Commitments: bsb22Commitments,
	}

	return &preSolvedWitness{Witness: fullWitness, data: data}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GPUProve — top-level prove API
//
// Orchestrates gpuProver methods via errgroup + channels:
//
//   initInstance(newGPUInstance) ─────────┐
//        │                                │
//        ├─ chInstHintReady ──> solve()   │
//        │                     (BSB22 hint │
//        │                     waits once) │
//        │                                │
//        └─ chInstReady ──────────────────┤
//                                         │
//   solve ───────────────────────────────┐│
//   (spr.Solve, BSB22)                  ││
//        │                              ││
//        ├─ chSolved ───────────────────┤│
//        │                              ││
//   initBlinding ───────────────────────┤│
//   (random bpL,bpR,bpO,bpZ)            ││
//        │                              ││
//        ├─ chBlinding ─────────────────┤│
//        │                              ││
//   completeQk ─────────────────────────┤│
//   (waits chSolved + chInstReady)      ││
//        │                              ││
//        ├─ chQk ───────────────────────┤│
//        │                              ││
//   main pipeline ──────────────────────┘│
//   (waits chSolved + chInstReady)       │
//        │
//        ├─ commitToLRO (waits chQk, chBlinding mid-execution)
//        ├─ deriveGammaBeta
//        ├─ buildZAndCommit
//        ├─ computeQuotientAndCommit
//        └─ openAndFinalize
// ─────────────────────────────────────────────────────────────────────────────

func GPUProve(dev *gpu.Device, gpk *GPUProvingKey, spr *cs.SparseR1CS, fullWitness witness.Witness) (*plonk377.Proof, error) {
	gpk.mu.Lock()
	defer gpk.mu.Unlock()

	vk := gpk.Vk
	if vk == nil {
		return nil, errors.New("gpu proving key missing verifying key")
	}
	n := gpk.n

	proveStart := time.Now()
	logTime := func(label string) {
		log.Printf("  [GPUProve n=%d] %s: %v", n, label, time.Since(proveStart))
	}

	// BSB22 commitment support.
	var commitmentInfo constraint.PlonkCommitments
	if spr.CommitmentInfo != nil {
		commitmentInfo = spr.CommitmentInfo.(constraint.PlonkCommitments)
	}
	if len(commitmentInfo) != len(vk.CommitmentConstraintIndexes) {
		return nil, fmt.Errorf("commitment metadata mismatch: spr=%d gpk=%d",
			len(commitmentInfo), len(vk.CommitmentConstraintIndexes))
	}

	// Allocate per-proof work vector.
	gpuWork, err := NewFrVector(dev, n)
	if err != nil {
		return nil, fmt.Errorf("create work vector: %w", err)
	}
	logTime("setupGPU")

	// Detect pre-solved witness.
	var preSolved *preSolvedData
	if psw, ok := fullWitness.(interface{ preSolvedData() *preSolvedData }); ok {
		preSolved = psw.preSolvedData()
	}

	// Create per-proof prover.
	p := &gpuProver{
		fs:             fiatshamir.NewTranscript(sha256.New(), "gamma", "beta", "alpha", "zeta"),
		preSolved:      preSolved,
		commitmentInfo: commitmentInfo,
		commitmentVal:  make([]fr.Element, len(commitmentInfo)),
		pi2Canonical:   make([][]fr.Element, len(commitmentInfo)),
		gpuWork:        gpuWork,
		logTime:        logTime,
	}
	p.proof.Bsb22Commitments = make([]kzg.Digest, len(commitmentInfo))

	// ── Channels ──
	chSolved := make(chan struct{})
	chBlinding := make(chan struct{})
	chQk := make(chan struct{})
	chInstHintReady := make(chan struct{})
	chInstReady := make(chan struct{})

	var (
		initInst *gpuInstance
		initErr  error
	)
	var hintReadyOnce sync.Once
	signalInstHintReady := func(inst *gpuInstance) {
		if inst != nil && initInst == nil {
			initInst = inst
		}
		hintReadyOnce.Do(func() { close(chInstHintReady) })
	}

	g, ctx := errgroup.WithContext(context.Background())
	wait := func(ch <-chan struct{}) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			return nil
		}
	}
	waitClosed := func(ch <-chan struct{}) error {
		select {
		case <-ch:
			return nil
		default:
			return wait(ch)
		}
	}
	waitInstHintReady := func() (*gpuInstance, error) {
		if err := waitClosed(chInstHintReady); err != nil {
			return nil, err
		}
		if initInst != nil {
			return initInst, nil
		}
		if err := waitClosed(chInstReady); err != nil {
			return nil, err
		}
		if initErr != nil {
			return nil, initErr
		}
		return nil, errors.New("gpu instance init failed before BSB22 readiness")
	}
	waitInstReady := func() (*gpuInstance, error) {
		if err := waitClosed(chInstReady); err != nil {
			return nil, err
		}
		if initErr != nil {
			return nil, initErr
		}
		if initInst == nil {
			return nil, errors.New("gpu instance init produced no instance")
		}
		return initInst, nil
	}
	safeGo := func(label string, fn func() error) {
		g.Go(func() error {
			return runGPUProveStep(label, fn)
		})
	}

	// ── initInstance ──
	safeGo("init instance", func() error {
		defer close(chInstReady)
		defer signalInstHintReady(nil)

		if gpk.inst != nil && gpk.inst.dev == dev {
			signalInstHintReady(gpk.inst)
			return nil
		}
		if gpk.inst != nil {
			gpk.inst.close()
			gpk.inst = nil
		}

		inst, err := newGPUInstanceWithBSB22Ready(dev, gpk, spr, signalInstHintReady)
		if initInst == nil {
			initInst = inst
		}
		initErr = err
		if err != nil {
			return err
		}
		gpk.inst = inst
		return nil
	})

	// ── solve ──
	if preSolved != nil {
		safeGo("solve pre-solved witness", func() error {
			if err := p.solvePreSolved(); err != nil {
				return err
			}
			logTime("solve (pre-solved)")
			close(chSolved)
			return nil
		})
	} else {
		safeGo("solve witness", func() error {
			if err := p.solve(spr, fullWitness, waitInstHintReady); err != nil {
				return err
			}
			logTime("solve")
			close(chSolved)
			return nil
		})
	}

	// ── initBlinding ──
	safeGo("initialize blinding polynomials", func() error {
		p.initBlindingPolynomials()
		close(chBlinding)
		return nil
	})

	var setInstOnce sync.Once
	setProverInst := func(inst *gpuInstance) {
		setInstOnce.Do(func() { p.inst = inst })
	}

	// ── completeQk ──
	safeGo("complete qk", func() error {
		if err := wait(chSolved); err != nil {
			return err
		}
		inst, err := waitInstReady()
		if err != nil {
			return err
		}
		setProverInst(inst)
		if err := p.completeQk(); err != nil {
			return err
		}
		logTime("completeQk")
		close(chQk)
		return nil
	})

	// ── main prove pipeline ──
	// Starts iFFT L,R,O as soon as solve completes (overlaps with completeQk).
	// Then runs the sequential prove phases: commit → derive → build Z → quotient → open.
	safeGo("main prove pipeline", func() error {
		if err := wait(chSolved); err != nil {
			return err
		}
		inst, err := waitInstReady()
		if err != nil {
			return err
		}
		setProverInst(inst)

		if err := p.commitToLRO(
			func() error { return wait(chQk) },
			func() error { return wait(chBlinding) },
		); err != nil {
			return err
		}

		if err := p.deriveGammaBeta(); err != nil {
			return err
		}
		if err := p.buildZAndCommit(); err != nil {
			return err
		}
		if err := p.computeQuotientAndCommit(); err != nil {
			return err
		}

		// Free gpuWork before open phase — not needed after quotient,
		// and freeing it reduces peak VRAM during Z opening MSM.
		p.gpuWork.Free()
		p.gpuWork = nil

		return p.openAndFinalize()
	})

	// Wait for all goroutines.
	if err := g.Wait(); err != nil {
		p.cleanup()
		if initErr != nil && initInst != nil && gpk.inst != initInst {
			initInst.close()
		}
		return nil, err
	}
	p.cleanup()

	return &p.proof, nil
}

// ─────────────────────────────────────────────────────────────────────
// GPU Phase B: Quotient polynomial computation
// ─────────────────────────────────────────────────────────────────────

// computeNumeratorGPU computes the quotient polynomial h(X) on the GPU.
//
// Algorithm:
//
//	Given blinded canonical polynomials L̃(X), R̃(X), Õ(X), Z̃(X) and
//	circuit-fixed selectors Ql..Qk, S1..S3, for each coset gₖ·H:
//
//	  Step 1. Evaluate all 12 polynomials on the coset via CosetFFT:
//	          L̃(gₖ·ωⁱ), R̃(gₖ·ωⁱ), ..., Qk(gₖ·ωⁱ)  for i ∈ [0,n)
//
//	  Step 2. Compute L₁⁻¹ denominators: (gₖ·ωⁱ - 1)⁻¹ via batch invert
//
//	  Step 3. Fused quotient kernel (single pass per element):
//	          h̃ₖ[i] = (α·perm + α²·boundary + gate) / Z_H(gₖ)
//
//	  Step 4. BSB22: h̃ₖ[i] += Σⱼ Qcp[j]·π[j] / Z_H(gₖ)
//
//	Then recover h₁,h₂,h₃ via decomposed iFFT(4n):
//	  4 × CosetFFTInverse(n) + Butterfly4Inverse + scaling
//
// Selector polynomials (Ql,Qr,Qm,Qo,S1,S2,S3,Qk,Qcp) are loaded from
// GPU-resident FrVectors via fast device-to-device copies, avoiding
// host→device transfers in this hot loop.
func computeNumeratorGPU(
	inst *gpuInstance,
	gpuWork *FrVector,
	lBlinded, rBlinded, oBlinded, zBlinded []fr.Element,
	qkCanonicalHost fr.Vector,
	qkCanonicalGPU *FrVector,
	pi2Canonical [][]fr.Element,
	alpha, beta, gamma fr.Element,
) (h1, h2, h3 []fr.Element, err error) {

	n := inst.n
	dev := inst.dev
	fftDom := inst.fftDom
	domain0 := inst.domain0
	cosetShift := inst.vk.CosetShift

	// 8 mandatory working vectors.
	var allocErr error
	alloc := func() *FrVector {
		r, err := NewFrVector(dev, n)
		if err != nil {
			allocErr = err
		}
		return r
	}
	gpuL := alloc()
	gpuR := alloc()
	gpuO := alloc()
	gpuZ := alloc()
	gpuS1 := alloc()
	gpuS2 := alloc()
	gpuS3 := alloc()
	gpuResult := alloc()
	if allocErr != nil {
		return nil, nil, nil, fmt.Errorf("allocate gpu vectors: %w", allocErr)
	}

	// Try 4 extra source vectors for GPU-side coset reduction.
	// At n=2^27 these cost 16 GiB; if VRAM is tight we fall back to
	// CPU-side reduceBlindedForCoset + H2D upload per coset.
	gpuLSrc, gpuRSrc, gpuOSrc, gpuZSrc := alloc(), alloc(), alloc(), alloc()
	gpuCosetReduce := allocErr == nil
	if !gpuCosetReduce {
		allocErr = nil // reset — the 8 mandatory vectors are fine
		for _, v := range []*FrVector{gpuLSrc, gpuRSrc, gpuOSrc, gpuZSrc} {
			if v != nil {
				v.Free()
			}
		}
		gpuLSrc, gpuRSrc, gpuOSrc, gpuZSrc = nil, nil, nil, nil
	}

	// Keep per-proof quotient inputs in device layout across all cosets.
	// Allocation failure is non-fatal; the hot loop falls back to the previous
	// H2D upload path under tight VRAM.
	gpuQkSrc := qkCanonicalGPU
	freeGpuQkSrc := false
	if gpuQkSrc == nil {
		if v, e := NewFrVector(dev, n); e == nil {
			gpuQkSrc = v
			freeGpuQkSrc = true
			gpuQkSrc.CopyFromHost(qkCanonicalHost)
		}
	}
	gpuPi2Src := make([]*FrVector, len(pi2Canonical))
	for j := range pi2Canonical {
		if len(pi2Canonical[j]) != n {
			continue
		}
		if v, e := NewFrVector(dev, n); e == nil {
			gpuPi2Src[j] = v
			gpuPi2Src[j].CopyFromHost(fr.Vector(pi2Canonical[j]))
		}
	}

	// Preserve the first three coset results on device and keep the fourth in
	// gpuResult. This avoids four n-sized D2H transfers followed by four H2D
	// uploads before the decomposed iFFT(4n). If VRAM is tight, fall back to
	// the host staging path below.
	var gpuCosetBlocks [3]*FrVector
	gpuCosetResultsOnDevice := true
	for k := range gpuCosetBlocks {
		v, e := NewFrVector(dev, n)
		if e != nil {
			gpuCosetResultsOnDevice = false
			for _, block := range gpuCosetBlocks {
				if block != nil {
					block.Free()
				}
			}
			gpuCosetBlocks = [3]*FrVector{}
			break
		}
		gpuCosetBlocks[k] = v
	}

	defer func() {
		gpuL.Free()
		gpuR.Free()
		gpuO.Free()
		gpuZ.Free()
		gpuS1.Free()
		gpuS2.Free()
		gpuS3.Free()
		gpuResult.Free()
		if gpuLSrc != nil {
			gpuLSrc.Free()
		}
		if gpuRSrc != nil {
			gpuRSrc.Free()
		}
		if gpuOSrc != nil {
			gpuOSrc.Free()
		}
		if gpuZSrc != nil {
			gpuZSrc.Free()
		}
		if freeGpuQkSrc && gpuQkSrc != nil {
			gpuQkSrc.Free()
		}
		for _, v := range gpuPi2Src {
			if v != nil {
				v.Free()
			}
		}
		for _, v := range gpuCosetBlocks {
			if v != nil {
				v.Free()
			}
		}
	}()

	// Host scratch for CPU-side fallback path (reused across 4 cosets).
	var hostScratch fr.Vector
	if gpuCosetReduce {
		// GPU path: upload blinded[0:n] once to persistent src vectors.
		gpuLSrc.CopyFromHost(fr.Vector(lBlinded[:n]))
		gpuRSrc.CopyFromHost(fr.Vector(rBlinded[:n]))
		gpuOSrc.CopyFromHost(fr.Vector(oBlinded[:n]))
		gpuZSrc.CopyFromHost(fr.Vector(zBlinded[:n]))
	} else {
		hostScratch = make(fr.Vector, n)
	}

	if err := dev.Sync(); err != nil {
		return nil, nil, nil, fmt.Errorf("sync cached quotient inputs: %w", err)
	}

	lTail := lBlinded[n:]
	rTail := rBlinded[n:]
	oTail := oBlinded[n:]
	zTail := zBlinded[n:]

	// ── Host result buffer (pre-allocated) ──
	hFull := inst.hBufs.hFull
	hostChunks := [4]fr.Vector{
		fr.Vector(hFull[0:n]),
		fr.Vector(hFull[n : 2*n]),
		fr.Vector(hFull[2*n : 3*n]),
		fr.Vector(hFull[3*n : 4*n]),
	}

	// ── Domain constants ──
	domain1 := fft.NewDomain(4*uint64(n), fft.WithoutPrecompute())
	u := domain1.FrMultiplicativeGen
	g1 := domain1.Generator

	var cosetShiftSq fr.Element
	cosetShiftSq.Square(&cosetShift)

	bn := big.NewInt(int64(n))
	var one fr.Element
	one.SetOne()

	// ── Main coset loop ──
	//
	// 2-stream pipeline: D2D copies on gpu.StreamTransfer overlap with FFT on
	// gpu.StreamCompute. Each coset evaluates all 12 polynomials, computes
	// permutation + boundary + gate constraints, and stores the result.
	//
	const (
		evDataReady gpu.EventID = 0 // gpu.StreamTransfer → gpu.StreamCompute: data loaded
		evPermDone  gpu.EventID = 1 // gpu.StreamCompute → gpu.StreamTransfer: safe to overwrite
		evCosetDone gpu.EventID = 2 // gpu.StreamCompute → gpu.StreamTransfer: coset fully stored
	)

	var cosetGen fr.Element
	for k := 0; k < 4; k++ {
		if k > 0 && gpuCosetResultsOnDevice {
			dev.WaitEvent(gpu.StreamTransfer, evCosetDone)
		}

		if k == 0 {
			cosetGen.Set(&u)
		} else {
			cosetGen.Mul(&cosetGen, &g1)
		}

		var cosetPowN fr.Element
		cosetPowN.Exp(cosetGen, bn)

		// ── Phase 1a: D2D S1,S2,S3 on gpu.StreamTransfer while wire FFTs run ──
		gpuS1.CopyFromDevice(inst.dS1, gpu.StreamTransfer)
		gpuS2.CopyFromDevice(inst.dS2, gpu.StreamTransfer)
		gpuS3.CopyFromDevice(inst.dS3, gpu.StreamTransfer)
		dev.RecordEvent(gpu.StreamTransfer, evDataReady)

		// Wire polynomial coset reduction + FFT on gpu.StreamCompute.
		// GPU path: fused kernel reads from persistent src vectors.
		// CPU fallback: reduceBlindedForCoset on host, then H2D upload.
		if gpuCosetReduce {
			ReduceBlindedCoset(gpuL, gpuLSrc, lTail, cosetPowN)
			ReduceBlindedCoset(gpuR, gpuRSrc, rTail, cosetPowN)
			ReduceBlindedCoset(gpuO, gpuOSrc, oTail, cosetPowN)
			ReduceBlindedCoset(gpuZ, gpuZSrc, zTail, cosetPowN)
		} else {
			reduceBlindedForCoset(hostScratch, lBlinded, cosetPowN)
			gpuL.CopyFromHost(hostScratch)
			reduceBlindedForCoset(hostScratch, rBlinded, cosetPowN)
			gpuR.CopyFromHost(hostScratch)
			reduceBlindedForCoset(hostScratch, oBlinded, cosetPowN)
			gpuO.CopyFromHost(hostScratch)
			reduceBlindedForCoset(hostScratch, zBlinded, cosetPowN)
			gpuZ.CopyFromHost(hostScratch)
		}
		fftDom.CosetFFT(gpuL, cosetGen)
		fftDom.CosetFFT(gpuR, cosetGen)
		fftDom.CosetFFT(gpuO, cosetGen)
		fftDom.CosetFFT(gpuZ, cosetGen)

		// ── Phase 1b: S1,S2,S3 CosetFFT ──
		dev.WaitEvent(gpu.StreamCompute, evDataReady)
		fftDom.CosetFFT(gpuS1, cosetGen)
		fftDom.CosetFFT(gpuS2, cosetGen)
		fftDom.CosetFFT(gpuS3, cosetGen)

		// ── Phase 1c: L₁ denominator inverse ──
		ComputeL1Den(gpuWork, cosetGen, fftDom)
		gpuWork.BatchInvert(gpuResult)

		// ── Phase 1d: Permutation + boundary constraint ──
		var l1Scalar fr.Element
		l1Scalar.Sub(&cosetPowN, &one)
		l1Scalar.Mul(&l1Scalar, &domain0.CardinalityInv)

		PlonkPermBoundary(
			gpuResult, gpuL, gpuR, gpuO, gpuZ,
			gpuS1, gpuS2, gpuS3, gpuWork,
			alpha, beta, gamma, l1Scalar,
			cosetShift, cosetShiftSq, cosetGen,
			fftDom,
		)

		// ── Phase 2a: Pipelined D2D+FFT for Ql,Qr,Qm,Qo,Qk ──
		//   gpuZ → Ql, gpuS1 → Qr, gpuS2 → Qm, gpuS3 → Qo, gpuWork → Qk
		dev.RecordEvent(gpu.StreamCompute, evPermDone)

		// Ql → gpuZ on stream 0 (safe — same stream as PermBoundary)
		gpuZ.CopyFromDevice(inst.dQl)
		fftDom.CosetFFT(gpuZ, cosetGen)

		// Qr → gpuS1: gpu.StreamTransfer waits for PermBoundary, then d2d
		dev.WaitEvent(gpu.StreamTransfer, evPermDone)
		gpuS1.CopyFromDevice(inst.dQr, gpu.StreamTransfer)
		dev.RecordEvent(gpu.StreamTransfer, evDataReady)
		dev.WaitEvent(gpu.StreamCompute, evDataReady)
		fftDom.CosetFFT(gpuS1, cosetGen)

		// Qm → gpuS2
		gpuS2.CopyFromDevice(inst.dQm, gpu.StreamTransfer)
		dev.RecordEvent(gpu.StreamTransfer, evDataReady)
		dev.WaitEvent(gpu.StreamCompute, evDataReady)
		fftDom.CosetFFT(gpuS2, cosetGen)

		// Qo → gpuS3
		gpuS3.CopyFromDevice(inst.dQo, gpu.StreamTransfer)
		dev.RecordEvent(gpu.StreamTransfer, evDataReady)
		dev.WaitEvent(gpu.StreamCompute, evDataReady)
		fftDom.CosetFFT(gpuS3, cosetGen)

		// Qk → gpuWork (per-proof, depends on public inputs)
		if gpuQkSrc != nil {
			gpuWork.CopyFromDevice(gpuQkSrc, gpu.StreamTransfer)
		} else {
			gpuWork.CopyFromHost(qkCanonicalHost, gpu.StreamTransfer)
		}
		dev.RecordEvent(gpu.StreamTransfer, evDataReady)
		dev.WaitEvent(gpu.StreamCompute, evDataReady)
		fftDom.CosetFFT(gpuWork, cosetGen)

		// ── Phase 2b: Gate constraint accumulation ──
		var zhKInv fr.Element
		zhKInv.Sub(&cosetPowN, &one)
		zhKInv.Inverse(&zhKInv)

		PlonkGateAccum(
			gpuResult, gpuZ, gpuS1, gpuS2, gpuS3, gpuWork,
			gpuL, gpuR, gpuO,
			zhKInv,
		)

		// ── BSB22 commitment accumulation ──
		for j := range pi2Canonical {
			gpuZ.CopyFromDevice(inst.dQcp[j])
			fftDom.CosetFFT(gpuZ, cosetGen)
			if gpuPi2Src[j] != nil {
				gpuWork.CopyFromDevice(gpuPi2Src[j])
			} else {
				gpuWork.CopyFromHost(fr.Vector(pi2Canonical[j]))
			}
			fftDom.CosetFFT(gpuWork, cosetGen)
			gpuZ.Mul(gpuZ, gpuWork)
			gpuResult.AddScalarMul(gpuZ, zhKInv)
		}

		// ── Store coset result ──
		if gpuCosetResultsOnDevice {
			if k < len(gpuCosetBlocks) {
				gpuCosetBlocks[k].CopyFromDevice(gpuResult)
			}
			dev.RecordEvent(gpu.StreamCompute, evCosetDone)
		} else {
			gpuResult.CopyToHost(hostChunks[k])
		}
	}

	// ── Decomposed iFFT(4n) ──
	//
	// Recover h₁,h₂,h₃ from the 4 coset evaluations via:
	//   4 × CosetFFTInverse(n) + Butterfly4Inverse + u⁻ˡⁿ scaling
	var blocks [4]*FrVector
	if gpuCosetResultsOnDevice {
		blocks = [4]*FrVector{gpuCosetBlocks[0], gpuCosetBlocks[1], gpuCosetBlocks[2], gpuResult}
	} else {
		blocks = [4]*FrVector{gpuL, gpuR, gpuO, gpuZ}
	}

	cosetGen.Set(&u)
	for k := 0; k < 4; k++ {
		if k > 0 {
			cosetGen.Mul(&cosetGen, &g1)
		}
		if !gpuCosetResultsOnDevice {
			blocks[k].CopyFromHost(hostChunks[k])
		}
		var cosetGenInv fr.Element
		cosetGenInv.Inverse(&cosetGen)
		fftDom.CosetFFTInverse(blocks[k], cosetGenInv)
	}

	var omega4Inv, quarter fr.Element
	{
		var omega4 fr.Element
		omega4.Exp(g1, bn)
		omega4Inv.Inverse(&omega4)
	}
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)
	Butterfly4Inverse(blocks[0], blocks[1], blocks[2], blocks[3], omega4Inv, quarter)

	var uInvN fr.Element
	{
		var uN fr.Element
		uN.Exp(u, bn)
		uInvN.Inverse(&uN)
	}
	blocks[1].ScalarMul(uInvN)
	var uInv2N fr.Element
	uInv2N.Mul(&uInvN, &uInvN)
	blocks[2].ScalarMul(uInv2N)
	var uInv3N fr.Element
	uInv3N.Mul(&uInv2N, &uInvN)
	blocks[3].ScalarMul(uInv3N)

	if err := dev.Sync(); err != nil {
		return nil, nil, nil, fmt.Errorf("quotient GPU sync: %w", err)
	}

	for k := 0; k < 4; k++ {
		blocks[k].CopyToHost(fr.Vector(hFull[k*n : (k+1)*n]))
	}

	np2 := n + 2
	h1 = hFull[:np2]
	h2 = hFull[np2 : 2*np2]
	h3 = hFull[2*np2 : 3*np2]

	return h1, h2, h3, nil
}

// reduceBlindedForCoset reduces a blinded polynomial to n coefficients for a coset.
func reduceBlindedForCoset(dst fr.Vector, blinded []fr.Element, cosetPowN fr.Element) {
	n := len(dst)
	copy(dst, blinded[:n])
	for j := n; j < len(blinded); j++ {
		var tmp fr.Element
		tmp.Mul(&blinded[j], &cosetPowN)
		dst[j-n].Add(&dst[j-n], &tmp)
	}
}

// ─────────────────────────────────────────────────────────────────────
// GPU Z polynomial construction
// ─────────────────────────────────────────────────────────────────────

func buildZGPU(
	inst *gpuInstance, gpuWork *FrVector,
	evalL, evalR, evalO fr.Vector, beta, gamma fr.Element,
) (fr.Vector, error) {
	n := inst.n
	dev := inst.dev
	domain0 := inst.domain0

	gpuR, err := NewFrVector(dev, n)
	if err != nil {
		return nil, fmt.Errorf("alloc gpuR: %w", err)
	}
	defer gpuR.Free()
	gpuO, err := NewFrVector(dev, n)
	if err != nil {
		return nil, fmt.Errorf("alloc gpuO: %w", err)
	}
	defer gpuO.Free()

	gpuWork.CopyFromHost(evalL)
	gpuR.CopyFromHost(evalR)
	gpuO.CopyFromHost(evalO)

	gMul := domain0.FrMultiplicativeGen
	var gSq fr.Element
	gSq.Mul(&gMul, &gMul)

	PlonkZComputeFactors(gpuWork, gpuR, gpuO, inst.dPerm,
		beta, gamma, gMul, gSq, inst.log2n, inst.fftDom)

	gpuR.BatchInvert(gpuO)
	gpuWork.Mul(gpuWork, gpuR)

	ZPrefixProduct(dev, gpuR, gpuWork, gpuO)

	gpuR.CopyToHost(inst.hBufs.zLagrange)

	return inst.hBufs.zLagrange, nil
}

// ─────────────────────────────────────────────────────────────────────
// GPU MSM helpers
// ─────────────────────────────────────────────────────────────────────

// gpuCommit computes [P(τ)] = MSM(coeffs, canonical_SRS).
func gpuCommit(msm *G1MSM, coeffs []fr.Element) curve.G1Affine {
	jacs := msm.MultiExp(coeffs)
	var aff curve.G1Affine
	aff.FromJacobian(&jacs[0])
	return aff
}

// gpuCommitN computes multiple polynomial commitments in a single batched MSM call.
func gpuCommitN(msm *G1MSM, coeffSets ...[]fr.Element) []curve.G1Affine {
	jacs := msm.MultiExp(coeffSets...)
	affs := make([]curve.G1Affine, len(jacs))
	for i := range jacs {
		affs[i].FromJacobian(&jacs[i])
	}
	return affs
}

func gpuBatchOpen(
	msm *G1MSM,
	polys [][]fr.Element,
	digests []curve.G1Affine,
	claimedValues []fr.Element,
	point fr.Element,
	dataTranscript []byte,
) (kzg.BatchOpeningProof, error) {
	var res kzg.BatchOpeningProof
	nbPolys := len(polys)
	res.ClaimedValues = claimedValues

	fsGamma := fiatshamir.NewTranscript(sha256.New(), "gamma")
	if err := fsGamma.Bind("gamma", point.Marshal()); err != nil {
		return res, err
	}
	for i := range digests {
		if err := fsGamma.Bind("gamma", digests[i].Marshal()); err != nil {
			return res, err
		}
	}
	for i := range claimedValues {
		if err := fsGamma.Bind("gamma", claimedValues[i].Marshal()); err != nil {
			return res, err
		}
	}
	if len(dataTranscript) > 0 {
		if err := fsGamma.Bind("gamma", dataTranscript); err != nil {
			return res, err
		}
	}
	gammaByte, err := fsGamma.ComputeChallenge("gamma")
	if err != nil {
		return res, err
	}
	var gammaChallenge fr.Element
	gammaChallenge.SetBytes(gammaByte)

	largestPoly := 0
	for _, p := range polys {
		if len(p) > largestPoly {
			largestPoly = len(p)
		}
	}

	gammas := make([]fr.Element, nbPolys)
	gammas[0].SetOne()
	for i := 1; i < nbPolys; i++ {
		gammas[i].Mul(&gammas[i-1], &gammaChallenge)
	}

	// Fold polynomials: folded[j] = Σᵢ gammas[i]·polys[i][j]
	//
	// Uses fr.Vector.ScalarMul + fr.Vector.Add (AVX-512 IFMA when available).
	// Each goroutine processes a chunk with a reusable temp vector, so
	// total extra memory is O(n) rather than O(k·n).
	folded := make(fr.Vector, largestPoly)
	nCPU := runtime.NumCPU()
	chunkSize := (largestPoly + nCPU - 1) / nCPU
	var wg sync.WaitGroup
	for c := 0; c < largestPoly; c += chunkSize {
		start := c
		end := start + chunkSize
		if end > largestPoly {
			end = largestPoly
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			temp := make(fr.Vector, end-start)
			for i := range nbPolys {
				effEnd := end
				if effEnd > len(polys[i]) {
					effEnd = len(polys[i])
				}
				if start >= effEnd {
					continue
				}
				n := effEnd - start
				t := fr.Vector(temp[:n])
				t.ScalarMul(fr.Vector(polys[i][start:effEnd]), &gammas[i])
				f := fr.Vector(folded[start:effEnd])
				f.Add(f, t)
			}
		}()
	}
	wg.Wait()

	var foldedEval fr.Element
	for i := nbPolys - 1; i >= 0; i-- {
		foldedEval.Mul(&foldedEval, &gammaChallenge).Add(&foldedEval, &claimedValues[i])
	}

	folded[0].Sub(&folded[0], &foldedEval)
	parallelHornerQuotient(folded, point)
	h := folded[1:]

	res.H = gpuCommit(msm, h)

	return res, nil
}

// ─────────────────────────────────────────────────────────────────────
// Polynomial helpers
// ─────────────────────────────────────────────────────────────────────

// blindInto computes the blinded canonical coefficients into a pre-allocated buffer:
//
//	dst = [c0-b0, c1-b1, ..., c_{n-1}, b0, b1, ...]
//
// This is equivalent to p(X) + b(X)*(X^n - 1) in canonical form.
// dst must have length >= len(canonical)+len(bp.Coefficients()).
func blindInto(dst []fr.Element, canonical []fr.Element, bp *iop.Polynomial) []fr.Element {
	cbp := bp.Coefficients()
	result := dst[:len(canonical)+len(cbp)]
	copy(result, canonical)
	copy(result[len(canonical):], cbp)
	for i := 0; i < len(cbp); i++ {
		result[i].Sub(&result[i], &cbp[i])
	}
	return result
}

// getRandomPolynomial returns a random polynomial of the given degree.
func getRandomPolynomial(degree int) *iop.Polynomial {
	coeffs := make([]fr.Element, degree+1)
	for i := range coeffs {
		coeffs[i].SetRandom()
	}
	return iop.NewPolynomial(&coeffs, iop.Form{Basis: iop.Canonical, Layout: iop.Regular})
}

// ─────────────────────────────────────────────────────────────────────
// Polynomial evaluation helpers
// ─────────────────────────────────────────────────────────────────────

func evaluateCanonical(canonical []fr.Element, z fr.Element) fr.Element {
	var r fr.Element
	for i := len(canonical) - 1; i >= 0; i-- {
		r.Mul(&r, &z).Add(&r, &canonical[i])
	}
	return r
}

// parallelEvaluateCanonical evaluates p(z) = Σ p[i]·z^i using multi-core
// parallelism. Splits the polynomial into chunks, evaluates each via Horner,
// and combines: p(z) = p₀(z) + z^k·p₁(z) + z^{2k}·p₂(z) + ...
func parallelEvaluateCanonical(canonical []fr.Element, z fr.Element) fr.Element {
	n := len(canonical)
	nCPU := runtime.NumCPU()
	if n < 4096 || nCPU < 2 {
		return evaluateCanonical(canonical, z)
	}

	chunkSize := (n + nCPU - 1) / nCPU
	partials := make([]fr.Element, nCPU)

	var wg sync.WaitGroup
	for c := range nCPU {
		start := c * chunkSize
		if start >= n {
			break
		}
		end := start + chunkSize
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(idx, s, e int) {
			defer wg.Done()
			var r fr.Element
			for i := e - 1; i >= s; i-- {
				r.Mul(&r, &z).Add(&r, &canonical[i])
			}
			partials[idx] = r
		}(c, start, end)
	}
	wg.Wait()

	// z^chunkSize via repeated squaring.
	var zChunk fr.Element
	{
		zChunk.Set(&z)
		exp := chunkSize
		var acc fr.Element
		acc.SetOne()
		for exp > 0 {
			if exp&1 != 0 {
				acc.Mul(&acc, &zChunk)
			}
			zChunk.Square(&zChunk)
			exp >>= 1
		}
		zChunk.Set(&acc)
	}

	// Combine: p(z) = Σ z^{c·chunkSize} · partials[c]
	var result, zPow fr.Element
	zPow.SetOne()
	for c := range nCPU {
		if c*chunkSize >= n {
			break
		}
		var t fr.Element
		t.Mul(&partials[c], &zPow)
		result.Add(&result, &t)
		zPow.Mul(&zPow, &zChunk)
	}
	return result
}

// parallelHornerQuotient computes h(X) = (p(X) - p(z)) / (X - z) in place.
//
// After return: poly[0] = p(z) and poly[1:] contains the quotient coefficients.
//
// The recurrence q[i] = p[i] + z·q[i+1] (with q[n]=0) is parallelized via
// chunked prefix-scan:
//
//	Pass 1 (parallel): Local Horner within each chunk, assuming carry=0.
//	Pass 2 (sequential): Propagate true carries across chunk boundaries.
//	Pass 3 (parallel): Apply z^{offset}·carry correction to each element.
func parallelHornerQuotient(poly []fr.Element, z fr.Element) {
	n := len(poly)
	nCPU := runtime.NumCPU()
	if n < 4096 || nCPU < 2 {
		for i := n - 2; i >= 0; i-- {
			var tmp fr.Element
			tmp.Mul(&poly[i+1], &z)
			poly[i].Add(&poly[i], &tmp)
		}
		return
	}

	chunkSize := (n + nCPU - 1) / nCPU
	numChunks := (n + chunkSize - 1) / chunkSize

	// ── Pass 1: parallel local Horner ──
	var wg sync.WaitGroup
	for c := range numChunks {
		lo := c * chunkSize
		hi := lo + chunkSize
		if hi > n {
			hi = n
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for i := hi - 2; i >= lo; i-- {
				var tmp fr.Element
				tmp.Mul(&poly[i+1], &z)
				poly[i].Add(&poly[i], &tmp)
			}
		}(lo, hi)
	}
	wg.Wait()

	// ── Pass 2: sequential carry propagation (right → left) ──
	zk := expElement(z, chunkSize)

	carries := make([]fr.Element, numChunks) // carries[last] = 0
	for c := numChunks - 2; c >= 0; c-- {
		nextLo := (c + 1) * chunkSize
		nextLen := chunkSize
		if nextLo+nextLen > n {
			nextLen = n - nextLo
		}
		zkc := zk
		if nextLen != chunkSize {
			zkc = expElement(z, nextLen)
		}
		var tmp fr.Element
		tmp.Mul(&carries[c+1], &zkc)
		carries[c].Add(&poly[nextLo], &tmp)
	}

	// ── Pass 3: parallel correction ──
	for c := range numChunks {
		lo := c * chunkSize
		hi := lo + chunkSize
		if hi > n {
			hi = n
		}
		if carries[c].IsZero() {
			continue
		}
		wg.Add(1)
		go func(lo, hi, c int) {
			defer wg.Done()
			var zPow fr.Element
			zPow.Set(&z) // z^1 for i = hi-1
			for i := hi - 1; i >= lo; i-- {
				var corr fr.Element
				corr.Mul(&zPow, &carries[c])
				poly[i].Add(&poly[i], &corr)
				zPow.Mul(&zPow, &z)
			}
		}(lo, hi, c)
	}
	wg.Wait()
}

// expElement computes z^exp via repeated squaring.
func expElement(z fr.Element, exp int) fr.Element {
	var base, acc fr.Element
	base.Set(&z)
	acc.SetOne()
	for exp > 0 {
		if exp&1 != 0 {
			acc.Mul(&acc, &base)
		}
		base.Square(&base)
		exp >>= 1
	}
	return acc
}

// ─────────────────────────────────────────────────────────────────────
// Fiat-Shamir helpers
// ─────────────────────────────────────────────────────────────────────

func bindPublicData(fs *fiatshamir.Transcript, challenge string, vk *plonk377.VerifyingKey, publicInputs []fr.Element) error {
	if err := fs.Bind(challenge, vk.S[0].Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.S[1].Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.S[2].Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.Ql.Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.Qr.Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.Qm.Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.Qo.Marshal()); err != nil {
		return err
	}
	if err := fs.Bind(challenge, vk.Qk.Marshal()); err != nil {
		return err
	}
	for i := range vk.Qcp {
		if err := fs.Bind(challenge, vk.Qcp[i].Marshal()); err != nil {
			return err
		}
	}
	for i := range publicInputs {
		bPub := publicInputs[i].Marshal()
		if err := fs.Bind(challenge, bPub); err != nil {
			return err
		}
	}
	return nil
}

func deriveRandomness(fs *fiatshamir.Transcript, challenge string, points ...*curve.G1Affine) (fr.Element, error) {
	var buf [curve.SizeOfG1AffineUncompressed]byte
	var r fr.Element
	for _, p := range points {
		buf = p.RawBytes()
		if err := fs.Bind(challenge, buf[:]); err != nil {
			return r, err
		}
	}
	b, err := fs.ComputeChallenge(challenge)
	if err != nil {
		return r, err
	}
	r.SetBytes(b)
	return r, nil
}

// ─────────────────────────────────────────────────────────────────────
// Linearized polynomial
// ─────────────────────────────────────────────────────────────────────

func gpuComputeLinearizedPoly(
	inst *gpuInstance,
	lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu fr.Element,
	s1Zeta, s2Zeta fr.Element,
	qcpZeta []fr.Element, blindedZCanonical []fr.Element, pi2Canonical [][]fr.Element,
	h1, h2, h3 []fr.Element,
) []fr.Element {
	n := inst.n
	domain0 := inst.domain0
	cosetShift := inst.vk.CosetShift

	var rl fr.Element
	rl.Mul(&rZeta, &lZeta)

	var s1, tmp fr.Element
	s1.Mul(&s1Zeta, &beta).Add(&s1, &lZeta).Add(&s1, &gamma)
	tmp.Mul(&s2Zeta, &beta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s1.Mul(&s1, &tmp).Mul(&s1, &zu).Mul(&s1, &beta).Mul(&s1, &alpha)

	var s2 fr.Element
	var uzeta, uuzeta fr.Element
	uzeta.Mul(&zeta, &cosetShift)
	uuzeta.Mul(&uzeta, &cosetShift)
	s2.Mul(&beta, &zeta).Add(&s2, &lZeta).Add(&s2, &gamma)
	tmp.Mul(&beta, &uzeta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp)
	tmp.Mul(&beta, &uuzeta).Add(&tmp, &oZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp).Neg(&s2).Mul(&s2, &alpha)

	var zhZeta, zetaNPlusTwo, alphaSquareLagrangeZero, den fr.Element
	nbElmt := int64(domain0.Cardinality)
	alphaSquareLagrangeZero.Set(&zeta).Exp(alphaSquareLagrangeZero, big.NewInt(nbElmt))
	zetaNPlusTwo.Mul(&alphaSquareLagrangeZero, &zeta).Mul(&zetaNPlusTwo, &zeta)
	one := fr.One()
	alphaSquareLagrangeZero.Sub(&alphaSquareLagrangeZero, &one)
	zhZeta.Set(&alphaSquareLagrangeZero)
	den.Sub(&zeta, &one).Inverse(&den)
	alphaSquareLagrangeZero.Mul(&alphaSquareLagrangeZero, &den).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &domain0.CardinalityInv)

	// Try GPU path — use d2d copies from GPU-resident polynomials.
	dev := inst.dev
	gpuResult, err := NewFrVector(dev, n)
	if err != nil {
		return innerComputeLinearizedPoly(lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu,
			s1Zeta, s2Zeta, qcpZeta, blindedZCanonical, pi2Canonical, inst, h1, h2, h3)
	}
	defer gpuResult.Free()
	gpuWork, err := NewFrVector(dev, n)
	if err != nil {
		return innerComputeLinearizedPoly(lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu,
			s1Zeta, s2Zeta, qcpZeta, blindedZCanonical, pi2Canonical, inst, h1, h2, h3)
	}
	defer gpuWork.Free()

	var combinedZCoeff fr.Element
	combinedZCoeff.Add(&s2, &alphaSquareLagrangeZero)
	gpuResult.CopyFromHost(fr.Vector(blindedZCanonical[:n]))
	gpuResult.ScalarMul(combinedZCoeff)

	// d2d copies from GPU-resident polynomials (fast path).
	gpuWork.CopyFromDevice(inst.dS3)
	gpuResult.AddScalarMul(gpuWork, s1)

	gpuWork.CopyFromDevice(inst.dQl)
	gpuResult.AddScalarMul(gpuWork, lZeta)

	gpuWork.CopyFromDevice(inst.dQr)
	gpuResult.AddScalarMul(gpuWork, rZeta)

	gpuWork.CopyFromDevice(inst.dQm)
	gpuResult.AddScalarMul(gpuWork, rl)

	gpuWork.CopyFromDevice(inst.dQo)
	gpuResult.AddScalarMul(gpuWork, oZeta)

	gpuWork.CopyFromDevice(inst.dQkFixed)
	gpuResult.Add(gpuResult, gpuWork)

	for j := range qcpZeta {
		gpuWork.CopyFromHost(fr.Vector(pi2Canonical[j]))
		gpuResult.AddScalarMul(gpuWork, qcpZeta[j])
	}

	// Subtract h-terms directly: result -= zhZeta·ζ^{2(n+2)}·h3 + zhZeta·ζ^{n+2}·h2 + zhZeta·h1
	// Uses negated coefficients to avoid a third FrVector (gpuHterm).
	var negCoeff fr.Element
	negCoeff.Mul(&zhZeta, &zetaNPlusTwo).Mul(&negCoeff, &zetaNPlusTwo).Neg(&negCoeff)
	gpuWork.CopyFromHost(fr.Vector(h3[:n]))
	gpuResult.AddScalarMul(gpuWork, negCoeff)

	negCoeff.Mul(&zhZeta, &zetaNPlusTwo).Neg(&negCoeff)
	gpuWork.CopyFromHost(fr.Vector(h2[:n]))
	gpuResult.AddScalarMul(gpuWork, negCoeff)

	negCoeff.Neg(&zhZeta)
	gpuWork.CopyFromHost(fr.Vector(h1[:n]))
	gpuResult.AddScalarMul(gpuWork, negCoeff)

	gpuResult.CopyToHost(fr.Vector(blindedZCanonical[:n]))

	for i := n; i < len(blindedZCanonical); i++ {
		var t fr.Element
		t.Mul(&blindedZCanonical[i], &combinedZCoeff)
		if i < len(h3) {
			var hv fr.Element
			hv.Mul(&h3[i], &zetaNPlusTwo).
				Add(&hv, &h2[i]).
				Mul(&hv, &zetaNPlusTwo).
				Add(&hv, &h1[i]).
				Mul(&hv, &zhZeta)
			t.Sub(&t, &hv)
		}
		blindedZCanonical[i] = t
	}

	return blindedZCanonical
}

func innerComputeLinearizedPoly(
	lZeta, rZeta, oZeta, alpha, beta, gamma, zeta, zu fr.Element,
	s1Zeta, s2Zeta fr.Element,
	qcpZeta []fr.Element, blindedZCanonical []fr.Element, pi2Canonical [][]fr.Element,
	inst *gpuInstance,
	h1, h2, h3 []fr.Element,
) []fr.Element {
	domain0 := inst.domain0
	cosetShift := inst.vk.CosetShift

	var rl fr.Element
	rl.Mul(&rZeta, &lZeta)

	var s1, tmp fr.Element
	s1.Mul(&s1Zeta, &beta).Add(&s1, &lZeta).Add(&s1, &gamma)
	tmp.Mul(&s2Zeta, &beta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s1.Mul(&s1, &tmp).Mul(&s1, &zu).Mul(&s1, &beta).Mul(&s1, &alpha)

	var s2 fr.Element
	var uzeta, uuzeta fr.Element
	uzeta.Mul(&zeta, &cosetShift)
	uuzeta.Mul(&uzeta, &cosetShift)

	s2.Mul(&beta, &zeta).Add(&s2, &lZeta).Add(&s2, &gamma)
	tmp.Mul(&beta, &uzeta).Add(&tmp, &rZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp)
	tmp.Mul(&beta, &uuzeta).Add(&tmp, &oZeta).Add(&tmp, &gamma)
	s2.Mul(&s2, &tmp).Neg(&s2).Mul(&s2, &alpha)

	var zhZeta, zetaNPlusTwo, alphaSquareLagrangeZero, den fr.Element
	nbElmt := int64(domain0.Cardinality)
	alphaSquareLagrangeZero.Set(&zeta).Exp(alphaSquareLagrangeZero, big.NewInt(nbElmt))
	zetaNPlusTwo.Mul(&alphaSquareLagrangeZero, &zeta).Mul(&zetaNPlusTwo, &zeta)
	one := fr.One()
	alphaSquareLagrangeZero.Sub(&alphaSquareLagrangeZero, &one)
	zhZeta.Set(&alphaSquareLagrangeZero)
	den.Sub(&zeta, &one).Inverse(&den)
	alphaSquareLagrangeZero.Mul(&alphaSquareLagrangeZero, &den).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &alpha).
		Mul(&alphaSquareLagrangeZero, &domain0.CardinalityInv)

	s3can := []fr.Element(inst.s3Canonical)
	cql := []fr.Element(inst.qlCanonical)
	cqr := []fr.Element(inst.qrCanonical)
	cqm := []fr.Element(inst.qmCanonical)
	cqo := []fr.Element(inst.qoCanonical)
	cqk := []fr.Element(inst.qkFixedCanonical)

	total := len(blindedZCanonical)
	nCPU := runtime.NumCPU()
	chunkSize := (total + nCPU - 1) / nCPU
	var wg sync.WaitGroup
	for c := 0; c < total; c += chunkSize {
		start := c
		end := start + chunkSize
		if end > total {
			end = total
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			var t, t0, t1 fr.Element
			for i := start; i < end; i++ {
				t.Mul(&blindedZCanonical[i], &s2)
				if i < len(s3can) {
					t0.Mul(&s3can[i], &s1)
					t.Add(&t, &t0)
				}
				if i < len(cqm) {
					t1.Mul(&cqm[i], &rl)
					t.Add(&t, &t1)
					t0.Mul(&cql[i], &lZeta)
					t.Add(&t, &t0)
					t0.Mul(&cqr[i], &rZeta)
					t.Add(&t, &t0)
					t0.Mul(&cqo[i], &oZeta)
					t.Add(&t, &t0)
					t.Add(&t, &cqk[i])
					for j := range qcpZeta {
						t0.Mul(&pi2Canonical[j][i], &qcpZeta[j])
						t.Add(&t, &t0)
					}
				}
				t0.Mul(&blindedZCanonical[i], &alphaSquareLagrangeZero)
				blindedZCanonical[i].Add(&t, &t0)

				if i < len(h3) {
					t.Mul(&h3[i], &zetaNPlusTwo).
						Add(&t, &h2[i]).
						Mul(&t, &zetaNPlusTwo).
						Add(&t, &h1[i]).
						Mul(&t, &zhZeta)
					blindedZCanonical[i].Sub(&blindedZCanonical[i], &t)
				}
			}
		}()
	}
	wg.Wait()

	return blindedZCanonical
}
