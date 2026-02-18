package poseidon2_koalabear

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/constraint/solver/gkrgates"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/gkrapi"
	"github.com/consensys/gnark/std/gkrapi/gkr"
	"github.com/consensys/gnark/std/hash"
)

// This file implements a GKR-based Poseidon2 compression function for KoalaBear
// field elements (width=16, full rounds=6, partial rounds=21, S-box degree=3).
//
// The GKR (sumcheck-based) approach batches all hash invocations in a circuit
// into a single GKR proof, replacing O(n * constraintsPerHash) with
// O(circuitSize * log(n)) constraints, where n is the number of hash calls.
//
// Architecture:
//   - 16 GKR input variables (8 state + 8 data)
//   - 27 rounds (3 full + 21 partial + 3 full)
//   - 8 GKR output variables (with feed-forward)
//   - Each round uses merged gates: combined linear+sbox (degree 3) for
//     S-boxed positions, linear-only (degree 1) for non-S-boxed positions.
//     This halves the number of GKR layers from ~55 to ~28.

const (
	gkrWidth          = 16
	gkrBlockSize      = 8
	gkrNbFullRounds   = 6
	gkrNbPartRounds   = 21
	gkrNbTotalRounds  = gkrNbFullRounds + gkrNbPartRounds
	gkrNbGroups       = gkrWidth / 4
	gkrHalfFullRounds = gkrNbFullRounds / 2

	// gkrIntBatchSize is the number of internal partial rounds optimized
	// with the partial S-box + batched anchoring approach.
	// Covers rounds gkrHalfFullRounds+1 through gkrHalfFullRounds+gkrNbPartRounds-1.
	gkrIntBatchSize = gkrNbPartRounds - 1 // = 20
)

// m4 is the circulant matrix circ(2,3,1,1) used in the external MDS layer.
var m4 = [4][4]int{
	{2, 3, 1, 1},
	{1, 2, 3, 1},
	{1, 1, 2, 3},
	{3, 1, 1, 2},
}

// gkrParams caches the KoalaBear Poseidon2 parameters.
var gkrParams *poseidon2.Parameters

// intDiagPlusOne[p] = (1 + diag[p]) as a KoalaBear field element (big.Int).
// For the internal matrix, output[p] = (1+diag[p]) * x[p] + sum_{q!=p} x[q].
var intDiagPlusOne [gkrWidth]*big.Int

// diagVal[p] = diag[p] as a KoalaBear field element (*big.Int).
var diagVal [gkrWidth]*big.Int

// diagPow[p][j] = diag[p]^j mod prime, for p=0..15 and j=0..gkrIntBatchSize.
var diagPow [gkrWidth][gkrIntBatchSize + 1]*big.Int

// eVals[m] = sum_{p=1}^{15} diag[p]^m mod prime, for m=0..gkrIntBatchSize-1.
// E_0 = 15 (the count of positions 1-15).
var eVals [gkrIntBatchSize]*big.Int

// gkrInitOnce ensures one-time initialization of GKR parameters.
var gkrInitOnce sync.Once

func gkrInit() {
	gkrInitOnce.Do(func() {
		gkrParams = poseidon2.NewParameters(gkrWidth, gkrNbFullRounds, gkrNbPartRounds)
		computeIntDiagPlusOne()
		computeDiagPrecomputations()
	})
}

// computeIntDiagPlusOne computes (1 + diag[p]) mod koalabear_prime for each position.
// diag = [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
func computeIntDiagPlusOne() {
	p := koalabear.Modulus()

	// fracVal computes (1 + num/den) mod p = (den + num) * inv(den) mod p
	fracVal := func(num, den int64) *big.Int {
		n := big.NewInt(den + num)
		d := big.NewInt(den)
		dInv := new(big.Int).ModInverse(d, p)
		v := new(big.Int).Mul(n, dInv)
		v.Mod(v, p)
		return v
	}

	// intVal computes (1 + diag) mod p for integer diagonal values
	intVal := func(diag int64) *big.Int {
		return fracVal(diag, 1)
	}

	intDiagPlusOne[0] = intVal(-2)              // 1 + (-2) = -1
	intDiagPlusOne[1] = intVal(1)               // 1 + 1 = 2
	intDiagPlusOne[2] = intVal(2)               // 1 + 2 = 3
	intDiagPlusOne[3] = fracVal(1, 2)           // 1 + 1/2 = 3/2
	intDiagPlusOne[4] = intVal(3)               // 1 + 3 = 4
	intDiagPlusOne[5] = intVal(4)               // 1 + 4 = 5
	intDiagPlusOne[6] = fracVal(-1, 2)          // 1 + (-1/2) = 1/2
	intDiagPlusOne[7] = intVal(-3)              // 1 + (-3) = -2
	intDiagPlusOne[8] = intVal(-4)              // 1 + (-4) = -3
	intDiagPlusOne[9] = fracVal(1, 1<<8)        // 1 + 1/256 = 257/256
	intDiagPlusOne[10] = fracVal(1, 1<<3)       // 1 + 1/8 = 9/8
	intDiagPlusOne[11] = fracVal(1, 1<<24)      // 1 + 1/2^24
	intDiagPlusOne[12] = fracVal(-1, 1<<8)      // 1 + (-1/256) = 255/256
	intDiagPlusOne[13] = fracVal(-1, 1<<3)      // 1 + (-1/8) = 7/8
	intDiagPlusOne[14] = fracVal(-1, 1<<4)      // 1 + (-1/16) = 15/16
	intDiagPlusOne[15] = fracVal(-1, 1<<24)     // 1 + (-1/2^24)
}

// computeDiagPrecomputations computes:
//   - diagVal[p] = diag[p] = intDiagPlusOne[p] - 1
//   - diagPow[p][j] = diag[p]^j for j=0..gkrIntBatchSize
//   - eVals[m] = sum_{p=1}^{15} diag[p]^m for m=0..gkrIntBatchSize-1
func computeDiagPrecomputations() {
	p := koalabear.Modulus()
	one := big.NewInt(1)

	for i := 0; i < gkrWidth; i++ {
		diagVal[i] = new(big.Int).Sub(intDiagPlusOne[i], one)
		diagVal[i].Mod(diagVal[i], p)
	}

	for i := 0; i < gkrWidth; i++ {
		diagPow[i][0] = new(big.Int).SetInt64(1)
		for j := 1; j <= gkrIntBatchSize; j++ {
			diagPow[i][j] = new(big.Int).Mul(diagPow[i][j-1], diagVal[i])
			diagPow[i][j].Mod(diagPow[i][j], p)
		}
	}

	for m := 0; m < gkrIntBatchSize; m++ {
		eVals[m] = new(big.Int)
		for i := 1; i < gkrWidth; i++ {
			eVals[m].Add(eVals[m], diagPow[i][m])
		}
		eVals[m].Mod(eVals[m], p)
	}
}

// --- Gate naming ---

func extGateName(pos, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-ext-p%d-r%d", pos, round))
}

func intGateName(pos, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-int-p%d-r%d", pos, round))
}

func finalExtGateName(pos int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-final-ext-p%d", pos))
}

func combinedExtGateName(pos, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-cext-p%d-r%d", pos, round))
}

func combinedIntGateName(pos, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-cint-p%d-r%d", pos, round))
}

// Gate names for the optimized batched partial rounds
func sumGateName() gkr.GateName {
	return "KB-P2-sum"
}

func momentGateName(j int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-moment-j%d", j))
}

func partialSBoxGateName(step int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-psbox-s%d", step))
}

func sUpdateGateName(step int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-supd-s%d", step))
}

func reconGateName(pos int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("KB-P2-recon-p%d", pos))
}

// --- Gate functions ---

// extCoeff returns the coefficient of x[q] in the external matrix output at position p.
// For p = 4g + j and q = 4g' + c:
//   - Same group (g == g'): 2 * M4[j][c]
//   - Different group: M4[j][c]
func extCoeff(p, q int) int {
	g := p / 4
	j := p % 4
	gPrime := q / 4
	c := q % 4
	if g == gPrime {
		return 2 * m4[j][c]
	}
	return m4[j][c]
}

// roundKeyBigInt extracts the round key at (round, pos) as a *big.Int.
// For partial rounds, only position 0 has a round key; others are zero.
func roundKeyBigInt(round, pos int) *big.Int {
	rk := gkrParams.RoundKeys[round]
	if pos >= len(rk) {
		return new(big.Int)
	}
	var b big.Int
	rk[pos].BigInt(&b)
	return &b
}

// newExtGate creates a gate function that computes the external matrix multiplication
// at position pos with the round key for the given round:
//
//	f(x0..x15) = sum_{q=0}^{15} extCoeff[pos][q] * x[q] + roundKey[round][pos]
func newExtGate(pos, round int) gkr.GateFunction {
	rk := roundKeyBigInt(round, pos)
	coeffs := [gkrWidth]int{}
	for q := 0; q < gkrWidth; q++ {
		coeffs[q] = extCoeff(pos, q)
	}

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// Start with the round key
		result := frontend.Variable(rk)
		for q := 0; q < gkrWidth; q++ {
			if coeffs[q] == 1 {
				result = api.Add(result, x[q])
			} else {
				result = api.Add(result, api.Mul(x[q], coeffs[q]))
			}
		}
		return result
	}
}

// newIntGate creates a gate function that computes the internal matrix multiplication
// at position pos with the round key for the given round:
//
//	f(x0..x15) = (1 + diag[pos]) * x[pos] + sum_{q != pos} x[q] + roundKey[round][pos]
func newIntGate(pos, round int) gkr.GateFunction {
	rk := roundKeyBigInt(round, pos)
	diagPlusOne := intDiagPlusOne[pos]

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// sum of all inputs except x[pos]
		result := frontend.Variable(rk)
		for q := 0; q < gkrWidth; q++ {
			if q == pos {
				result = api.Add(result, api.Mul(x[q], diagPlusOne))
			} else {
				result = api.Add(result, x[q])
			}
		}
		return result
	}
}

// newFinalExtGate creates a gate that computes the final external matrix multiplication
// (no round key) at position pos, plus feed-forward from the original data input.
// This gate takes 17 inputs: x[0..15] from the final sBox outputs, x[16] = original data[pos-8].
//
//	f(x0..x16) = sum_{q=0}^{15} extCoeff[pos][q] * x[q] + x[16]
func newFinalExtGate(pos int) gkr.GateFunction {
	coeffs := [gkrWidth]int{}
	for q := 0; q < gkrWidth; q++ {
		coeffs[q] = extCoeff(pos, q)
	}

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// x[16] is the feed-forward input (original data)
		result := x[gkrWidth] // start with feed-forward
		for q := 0; q < gkrWidth; q++ {
			if coeffs[q] == 1 {
				result = api.Add(result, x[q])
			} else {
				result = api.Add(result, api.Mul(x[q], coeffs[q]))
			}
		}
		return result
	}
}

// newCombinedExtGate creates a gate that merges external linear + sbox into one:
//
//	f(x0..x15) = (sum_{q=0}^{15} extCoeff[pos][q] * x[q] + roundKey[round][pos])³
func newCombinedExtGate(pos, round int) gkr.GateFunction {
	rk := roundKeyBigInt(round, pos)
	coeffs := [gkrWidth]int{}
	for q := 0; q < gkrWidth; q++ {
		coeffs[q] = extCoeff(pos, q)
	}

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		result := frontend.Variable(rk)
		for q := 0; q < gkrWidth; q++ {
			if coeffs[q] == 1 {
				result = api.Add(result, x[q])
			} else {
				result = api.Add(result, api.Mul(x[q], coeffs[q]))
			}
		}
		return api.Mul(result, result, result)
	}
}

// newCombinedIntGate creates a gate that merges internal linear + sbox into one:
//
//	f(x0..x15) = ((1 + diag[pos]) * x[pos] + sum_{q != pos} x[q] + roundKey[round][pos])³
func newCombinedIntGate(pos, round int) gkr.GateFunction {
	rk := roundKeyBigInt(round, pos)
	diagPlusOne := intDiagPlusOne[pos]

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		result := frontend.Variable(rk)
		for q := 0; q < gkrWidth; q++ {
			if q == pos {
				result = api.Add(result, api.Mul(x[q], diagPlusOne))
			} else {
				result = api.Add(result, x[q])
			}
		}
		return api.Mul(result, result, result)
	}
}

// --- Optimized partial-round gate constructors ---
//
// These gates implement the batched partial S-box + anchoring optimization.
// Instead of 16 gates per internal partial round, we track:
//   - x[0] evolution via partial S-box gates (degree 3)
//   - S = sum(state) via linear S-update gates (degree 1)
//   - Reconstruct positions 1-15 at the end from anchor values and S history
//
// Math: For internal MDS matrix M with output[p] = diag[p]*x[p] + S:
//   After k steps from anchor: x[p]_k = diag[p]^k * anchor[p] + sum_{j=0}^{k-1} diag[p]^{k-1-j} * S_j
//   S_k = x[0]_k + D_k + sum_{i=0}^{k-1} E_{k-1-i} * S_i
//   where D_j = sum_{p>0} diag[p]^j * anchor[p], E_m = sum_{p=1}^{15} diag[p]^m

// newSumGate: S_0 = sum(x[0..15])
// Takes 16 inputs (the anchor state), returns their sum.
func newSumGate() gkr.GateFunction {
	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		result := x[0]
		for i := 1; i < gkrWidth; i++ {
			result = api.Add(result, x[i])
		}
		return result
	}
}

// newMomentGate: D_j = sum_{p=1}^{15} diag[p]^j * x[p]
// Takes 15 inputs (anchor positions 1..15), returns weighted sum.
func newMomentGate(j int) gkr.GateFunction {
	coeffs := make([]*big.Int, gkrWidth-1)
	for p := 1; p < gkrWidth; p++ {
		coeffs[p-1] = diagPow[p][j]
	}
	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		result := api.Mul(x[0], coeffs[0])
		for i := 1; i < gkrWidth-1; i++ {
			result = api.Add(result, api.Mul(x[i], coeffs[i]))
		}
		return result
	}
}

// newPartialSBoxGate: x[0]_{step} = (diag[0]*x[0]_{step-1} + S_{step-1} + rk)^3
// Takes 2 inputs: x[0]_{step-1}, S_{step-1}
// The round index for this step is: gkrHalfFullRounds + step (1-indexed step)
func newPartialSBoxGate(step int) gkr.GateFunction {
	round := gkrHalfFullRounds + step
	rk := roundKeyBigInt(round, 0)
	diag0 := diagVal[0]

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// linear = diag[0]*x[0] + S + rk
		linear := api.Add(api.Mul(x[0], diag0), x[1], rk)
		return api.Mul(linear, linear, linear) // cube
	}
}

// newSUpdateGate: S_{step} = x[0]_{step} + D_{step} + sum_{i=0}^{step-1} E_{step-1-i} * S_i
// Takes (step+1) inputs: x[0]_{step}, S_0, S_1, ..., S_{step-1}
// Note: D_{step} is a constant baked into the gate.
func newSUpdateGate(step int) gkr.GateFunction {
	// Precompute E coefficients for this step
	eCoeffs := make([]*big.Int, step)
	for i := 0; i < step; i++ {
		eCoeffs[i] = eVals[step-1-i]
	}

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// x[0] = x[0]_{step}, x[1] = S_0, x[2] = S_1, ..., x[step] = S_{step-1}
		// Also x[step+1] = D_{step} (passed as input from moment gate)
		result := api.Add(x[0], x[step+1]) // x[0]_{step} + D_{step}
		for i := 0; i < step; i++ {
			result = api.Add(result, api.Mul(x[i+1], eCoeffs[i]))
		}
		return result
	}
}

// newReconGate: x[pos]_k = diag[pos]^k * anchor[pos] + sum_{j=0}^{k-1} diag[pos]^{k-1-j} * S_j
// Takes (k+1) inputs: anchor[pos], S_0, S_1, ..., S_{k-1}
// k = gkrIntBatchSize
func newReconGate(pos int) gkr.GateFunction {
	k := gkrIntBatchSize
	anchorCoeff := diagPow[pos][k]
	sCoeffs := make([]*big.Int, k)
	for j := 0; j < k; j++ {
		sCoeffs[j] = diagPow[pos][k-1-j]
	}

	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		// x[0] = anchor[pos], x[1] = S_0, ..., x[k] = S_{k-1}
		result := api.Mul(x[0], anchorCoeff)
		for j := 0; j < k; j++ {
			result = api.Add(result, api.Mul(x[j+1], sCoeffs[j]))
		}
		return result
	}
}

// shouldApplySBox returns true if the given position gets S-boxed in this round.
func shouldApplySBox(round, pos int) bool {
	if isFullSBoxRound(round) {
		return true
	}
	return pos == 0
}

// --- Fiat-Shamir hash for GKR verification ---

// gkrFiatShamirHash is the hash name passed to gkrApi.Compile.
// The blueprint appends "_KOALABEAR" to resolve "POSEIDON2_KOALABEAR" in gnark-crypto.
// The in-circuit verifier resolves "POSEIDON2" via the gnark std/hash registry.
const gkrFiatShamirHash = "POSEIDON2"

// kbHasher implements hash.FieldHasher using plain (non-GKR) CompressPoseidon2.
// Elements are buffered and compressed in 8-element blocks during Sum().
// The last partial block is left-padded with zeros.
// This matches the solver-side kbSolverHash in gnark/internal/gkr/koalabear.
type kbHasher struct {
	api    frontend.API
	state  GnarkOctuplet
	buffer []frontend.Variable
}

func (h *kbHasher) Write(data ...frontend.Variable) {
	h.buffer = append(h.buffer, data...)
}

func (h *kbHasher) Reset() {
	h.buffer = h.buffer[:0]
	for i := 0; i < 8; i++ {
		h.state[i] = 0
	}
}

func (h *kbHasher) Sum() frontend.Variable {
	for len(h.buffer) > 0 {
		var buf [BlockSize]frontend.Variable
		for i := range buf {
			buf[i] = 0
		}
		if len(h.buffer) < BlockSize {
			copy(buf[BlockSize-len(h.buffer):], h.buffer)
			h.buffer = h.buffer[:0]
		} else {
			copy(buf[:], h.buffer)
			h.buffer = h.buffer[BlockSize:]
		}
		h.state = CompressPoseidon2(h.api, h.state, buf)
	}
	return h.state[0]
}

// registerKBPoseidon2Hash overrides the POSEIDON2 hash registration to support
// KoalaBear. For non-KoalaBear fields, it falls back to the original behavior.
func registerKBPoseidon2Hash() {
	hash.Register(hash.POSEIDON2, func(api frontend.API) (hash.FieldHasher, error) {
		if api.Compiler().Field().Cmp(koalabear.Modulus()) == 0 {
			h := &kbHasher{api: api}
			for i := 0; i < 8; i++ {
				h.state[i] = 0
			}
			return h, nil
		}
		// For other fields, delegate to gnark's standard Poseidon2.
		// This import is safe since hash/all re-registers POSEIDON2 anyway.
		return nil, fmt.Errorf("POSEIDON2 not supported for field %s from KoalaBear override", api.Compiler().Field())
	})
}

// --- Gate registration ---

// RegisterGates registers all GKR gates for the KoalaBear Poseidon2 circuit.
// Must be called before circuit compilation that uses the GKR compressor.
//
// For each round and position, we register either:
//   - A combined gate (linear+sbox, degree 3) for positions that get S-boxed
//   - A linear-only gate (degree 1) for positions that don't
//
// This merges two GKR layers into one per round, halving the total layer count.
func RegisterGates() error {
	gkrInit()
	registerKBPoseidon2Hash()

	regGate := func(name gkr.GateName, f gkr.GateFunction, nInputs, degree int) error {
		return gkrgates.Register(f, nInputs,
			gkrgates.WithUnverifiedDegree(degree),
			gkrgates.WithUnverifiedSolvableVar(0),
			gkrgates.WithName(name),
		)
	}

	halfRf := gkrHalfFullRounds
	rp := gkrNbPartRounds
	firstIntPartial := halfRf + 1
	lastIntPartial := halfRf + rp - 1

	// Register gates for non-optimized rounds (full rounds + first/last partial)
	for round := 0; round < gkrNbTotalRounds; round++ {
		// Skip internal partial rounds — these are replaced by batched gates
		if round >= firstIntPartial && round <= lastIntPartial {
			continue
		}
		for pos := 0; pos < gkrWidth; pos++ {
			var (
				name   gkr.GateName
				f      gkr.GateFunction
				degree int
			)
			if shouldApplySBox(round, pos) {
				degree = 3
				if isExtRound(round) {
					name = combinedExtGateName(pos, round)
					f = newCombinedExtGate(pos, round)
				} else {
					name = combinedIntGateName(pos, round)
					f = newCombinedIntGate(pos, round)
				}
			} else {
				degree = 1
				if isExtRound(round) {
					name = extGateName(pos, round)
					f = newExtGate(pos, round)
				} else {
					name = intGateName(pos, round)
					f = newIntGate(pos, round)
				}
			}
			if err := regGate(name, f, gkrWidth, degree); err != nil {
				return fmt.Errorf("failed to register gate %s: %w", name, err)
			}
		}
	}

	// Register optimized batched partial-round gates
	// 1. Sum gate: S_0 = sum(anchor[0..15]), fan-in 16
	if err := regGate(sumGateName(), newSumGate(), gkrWidth, 1); err != nil {
		return fmt.Errorf("failed to register sum gate: %w", err)
	}

	// 2. Moment gates: D_j for j=1..gkrIntBatchSize, fan-in 15
	for j := 1; j <= gkrIntBatchSize; j++ {
		if err := regGate(momentGateName(j), newMomentGate(j), gkrWidth-1, 1); err != nil {
			return fmt.Errorf("failed to register moment gate j=%d: %w", j, err)
		}
	}

	// 3. Partial S-box gates: x[0]_{step}, fan-in 2, degree 3
	for step := 1; step <= gkrIntBatchSize; step++ {
		if err := regGate(partialSBoxGateName(step), newPartialSBoxGate(step), 2, 3); err != nil {
			return fmt.Errorf("failed to register partial sbox gate step=%d: %w", step, err)
		}
	}

	// 4. S-update gates: S_{step}, fan-in step+2 (x0_step, S_0..S_{step-1}, D_step)
	for step := 1; step <= gkrIntBatchSize; step++ {
		nInputs := step + 2 // x[0]_step + step S values + D_step
		if err := regGate(sUpdateGateName(step), newSUpdateGate(step), nInputs, 1); err != nil {
			return fmt.Errorf("failed to register S-update gate step=%d: %w", step, err)
		}
	}

	// 5. Reconstruction gates: x[pos]_k for pos=1..15, fan-in k+1
	for pos := 1; pos < gkrWidth; pos++ {
		nInputs := gkrIntBatchSize + 1 // anchor[pos] + k S values
		if err := regGate(reconGateName(pos), newReconGate(pos), nInputs, 1); err != nil {
			return fmt.Errorf("failed to register recon gate pos=%d: %w", pos, err)
		}
	}

	// Register final ext+feed-forward gates for output positions 8..15
	for i := 0; i < gkrBlockSize; i++ {
		pos := gkrBlockSize + i
		name := finalExtGateName(pos)
		if err := regGate(name, newFinalExtGate(pos), gkrWidth+1, 1); err != nil {
			return fmt.Errorf("failed to register gate %s: %w", name, err)
		}
	}

	return nil
}

// isExtRound returns true if round r uses the external matrix.
//
// The GKR regrouping shifts matrix operations: each GKR round applies the
// matrix from the PREVIOUS canonical round, plus the current round's key.
//
//   - Rounds 0..halfRf-1: initial ext + first full rounds → ext
//   - Round halfRf (first partial): ext from last full round → ext
//   - Rounds halfRf+1..halfRf+rp-1: int from partial rounds → int
//   - Round halfRf+rp (first of final full): int from last partial → int
//   - Rounds halfRf+rp+1..totalRounds-1: ext from full rounds → ext
func isExtRound(round int) bool {
	halfRf := gkrHalfFullRounds
	rp := gkrNbPartRounds
	// Rounds 0..halfRf use ext (initial ext + first full rounds)
	if round <= halfRf {
		return true
	}
	// Rounds halfRf+1..halfRf+rp use int (partial rounds)
	if round <= halfRf+rp {
		return false
	}
	// Remaining rounds use ext (final full rounds)
	return true
}

// isFullSBoxRound returns true if all 16 positions get S-boxed in this round.
func isFullSBoxRound(round int) bool {
	halfRf := gkrHalfFullRounds
	rp := gkrNbPartRounds
	return round < halfRf || round >= halfRf+rp
}

// --- GKR circuit definition ---

func defineGKRCircuit(api frontend.API) (gkrCircuit *gkrapi.Circuit, ins [gkrWidth]gkr.Variable, outs [gkrBlockSize]gkr.Variable, err error) {
	gkrInit()

	gkrApi, err := gkrapi.New(api)
	if err != nil {
		return
	}

	halfRf := gkrHalfFullRounds
	rp := gkrNbPartRounds

	// Create 16 input variables
	var state [gkrWidth]gkr.Variable
	for i := 0; i < gkrWidth; i++ {
		state[i] = gkrApi.NewInput()
		ins[i] = state[i]
	}

	// Helper: apply a full-width round (all 16 positions get gates)
	applyFullWidthRound := func(round int) {
		var newState [gkrWidth]gkr.Variable
		ext := isExtRound(round)
		for pos := 0; pos < gkrWidth; pos++ {
			var name gkr.GateName
			if shouldApplySBox(round, pos) {
				if ext {
					name = combinedExtGateName(pos, round)
				} else {
					name = combinedIntGateName(pos, round)
				}
			} else {
				if ext {
					name = extGateName(pos, round)
				} else {
					name = intGateName(pos, round)
				}
			}
			newState[pos] = gkrApi.NamedGate(name, state[:]...)
		}
		state = newState
	}

	// --- Phase 1: Full rounds + first partial round (rounds 0..halfRf) ---
	for round := 0; round <= halfRf; round++ {
		applyFullWidthRound(round)
	}
	// After round halfRf, state holds the "anchor" for the batched optimization.

	// --- Phase 2: Batched internal partial rounds (rounds halfRf+1..halfRf+rp-1) ---
	// Instead of 16 gates per round, we track x[0] and S = sum(state).

	// Save anchor state for reconstruction
	anchor := state

	// Compute S_0 = sum(anchor[0..15])
	sVars := make([]gkr.Variable, gkrIntBatchSize+1)
	sVars[0] = gkrApi.NamedGate(sumGateName(), anchor[:]...)

	// Compute D_j = sum_{p=1}^{15} diag[p]^j * anchor[p] for j=1..gkrIntBatchSize
	dVars := make([]gkr.Variable, gkrIntBatchSize+1)
	anchorTail := make([]gkr.Variable, gkrWidth-1)
	for p := 1; p < gkrWidth; p++ {
		anchorTail[p-1] = anchor[p]
	}
	for j := 1; j <= gkrIntBatchSize; j++ {
		dVars[j] = gkrApi.NamedGate(momentGateName(j), anchorTail...)
	}

	// Track x[0] evolution through each internal partial round
	x0 := anchor[0] // x[0]_0 = anchor[0]

	for step := 1; step <= gkrIntBatchSize; step++ {
		// x[0]_{step} = (diag[0]*x[0]_{step-1} + S_{step-1} + rk)^3
		x0 = gkrApi.NamedGate(partialSBoxGateName(step), x0, sVars[step-1])

		// S_{step} = x[0]_{step} + D_{step} + sum_{i=0}^{step-1} E_{step-1-i} * S_i
		sInputs := make([]gkr.Variable, step+2)
		sInputs[0] = x0            // x[0]_{step}
		for i := 0; i < step; i++ {
			sInputs[i+1] = sVars[i] // S_0, S_1, ..., S_{step-1}
		}
		sInputs[step+1] = dVars[step] // D_{step}
		sVars[step] = gkrApi.NamedGate(sUpdateGateName(step), sInputs...)
	}

	// Reconstruct positions 1-15:
	// x[pos]_k = diag[pos]^k * anchor[pos] + sum_{j=0}^{k-1} diag[pos]^{k-1-j} * S_j
	state[0] = x0
	for pos := 1; pos < gkrWidth; pos++ {
		reconInputs := make([]gkr.Variable, gkrIntBatchSize+1)
		reconInputs[0] = anchor[pos]
		for j := 0; j < gkrIntBatchSize; j++ {
			reconInputs[j+1] = sVars[j]
		}
		state[pos] = gkrApi.NamedGate(reconGateName(pos), reconInputs...)
	}

	// --- Phase 3: Last partial + final full rounds (rounds halfRf+rp..totalRounds-1) ---
	for round := halfRf + rp; round < gkrNbTotalRounds; round++ {
		applyFullWidthRound(round)
	}

	// Final step: apply ext_matrix (no round key) to the last sBox outputs
	// and add feed-forward from the original data inputs.
	for i := 0; i < gkrBlockSize; i++ {
		outputPos := gkrBlockSize + i
		gateInputs := make([]gkr.Variable, gkrWidth+1)
		copy(gateInputs[:gkrWidth], state[:])
		gateInputs[gkrWidth] = ins[outputPos]
		outs[i] = gkrApi.NamedGate(finalExtGateName(outputPos), gateInputs...)
	}

	gkrCircuit, err = gkrApi.Compile(gkrFiatShamirHash)
	return
}

// --- GKR Compressor ---

// GKRCompressor batches Poseidon2 compression calls using GKR.
// All Compress calls within a circuit compilation are batched into a single
// GKR proof, dramatically reducing constraint count.
type GKRCompressor struct {
	api        frontend.API
	gkrCircuit *gkrapi.Circuit
	ins        [gkrWidth]gkr.Variable
	outs       [gkrBlockSize]gkr.Variable
}

// keyValueStore is a local interface matching gnark's kvstore.Store.
// Used for caching the compressor across multiple NewGKRCompressor calls.
type keyValueStore interface {
	GetKeyValue(key any) any
	SetKeyValue(key, value any)
}

type gkrCompressorKey struct{}
// NewGKRCompressor creates (or retrieves a cached) GKR compressor for KoalaBear
// Poseidon2. All Compress calls share the same underlying GKR circuit.
func NewGKRCompressor(api frontend.API) (*GKRCompressor, error) {
	// Try to retrieve cached compressor
	if store, ok := api.Compiler().(keyValueStore); ok {
		if cached := store.GetKeyValue(gkrCompressorKey{}); cached != nil {
			if comp, ok := cached.(*GKRCompressor); ok {
				return comp, nil
			}
		}
	}

	gkrCircuit, ins, outs, err := defineGKRCircuit(api)
	if err != nil {
		return nil, fmt.Errorf("failed to define GKR circuit: %w", err)
	}

	comp := &GKRCompressor{
		api:        api,
		gkrCircuit: gkrCircuit,
		ins:        ins,
		outs:       outs,
	}

	// Cache the compressor
	if store, ok := api.Compiler().(keyValueStore); ok {
		store.SetKeyValue(gkrCompressorKey{}, comp)
	}

	return comp, nil
}

// Compress performs a Poseidon2 compression: newState = Compress(state, data).
// Each call adds an instance to the batched GKR proof.
func (c *GKRCompressor) Compress(state, data GnarkOctuplet) GnarkOctuplet {
	inputMap := make(map[gkr.Variable]frontend.Variable, gkrWidth)
	for i := 0; i < gkrBlockSize; i++ {
		inputMap[c.ins[i]] = state[i]
		inputMap[c.ins[gkrBlockSize+i]] = data[i]
	}

	outputMap, err := c.gkrCircuit.AddInstance(inputMap)
	if err != nil {
		panic(fmt.Sprintf("GKR AddInstance failed: %v", err))
	}

	var result GnarkOctuplet
	for i := 0; i < gkrBlockSize; i++ {
		result[i] = outputMap[c.outs[i]]
	}
	return result
}

// CompressPoseidon2GKR is a drop-in replacement for CompressPoseidon2 that uses
// GKR batching. The compressor must have been created with NewGKRCompressor.
func CompressPoseidon2GKR(comp *GKRCompressor, state, data GnarkOctuplet) GnarkOctuplet {
	return comp.Compress(state, data)
}
