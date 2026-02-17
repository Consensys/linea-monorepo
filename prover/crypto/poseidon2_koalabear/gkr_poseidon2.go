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

// gkrInitOnce ensures one-time initialization of GKR parameters.
var gkrInitOnce sync.Once

func gkrInit() {
	gkrInitOnce.Do(func() {
		gkrParams = poseidon2.NewParameters(gkrWidth, gkrNbFullRounds, gkrNbPartRounds)
		computeIntDiagPlusOne()
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

	for round := 0; round < gkrNbTotalRounds; round++ {
		for pos := 0; pos < gkrWidth; pos++ {
			var (
				name   gkr.GateName
				f      gkr.GateFunction
				degree int
			)

			if shouldApplySBox(round, pos) {
				// Combined gate: linear + cube (degree 3)
				degree = 3
				if isExtRound(round) {
					name = combinedExtGateName(pos, round)
					f = newCombinedExtGate(pos, round)
				} else {
					name = combinedIntGateName(pos, round)
					f = newCombinedIntGate(pos, round)
				}
			} else {
				// Linear-only gate (degree 1) for non-S-boxed positions
				degree = 1
				if isExtRound(round) {
					name = extGateName(pos, round)
					f = newExtGate(pos, round)
				} else {
					name = intGateName(pos, round)
					f = newIntGate(pos, round)
				}
			}

			if err := gkrgates.Register(f, gkrWidth,
				gkrgates.WithUnverifiedDegree(degree),
				gkrgates.WithUnverifiedSolvableVar(0),
				gkrgates.WithName(name),
			); err != nil {
				return fmt.Errorf("failed to register gate %s: %w", name, err)
			}
		}
	}

	// Register final ext+feed-forward gates for output positions 8..15
	for i := 0; i < gkrBlockSize; i++ {
		pos := gkrBlockSize + i
		name := finalExtGateName(pos)
		if err := gkrgates.Register(newFinalExtGate(pos), gkrWidth+1,
			gkrgates.WithUnverifiedDegree(1),
			gkrgates.WithUnverifiedSolvableVar(0),
			gkrgates.WithName(name),
		); err != nil {
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

	// Create 16 input variables
	var state [gkrWidth]gkr.Variable
	for i := 0; i < gkrWidth; i++ {
		state[i] = gkrApi.NewInput()
		ins[i] = state[i]
	}

	// Process all rounds — each round is a single GKR layer.
	// S-boxed positions use combined gates (linear+cube, degree 3).
	// Non-S-boxed positions use linear-only gates (degree 1).
	for round := 0; round < gkrNbTotalRounds; round++ {
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

	// Final step: apply ext_matrix (no round key) to the last sBox outputs
	// and add feed-forward from the original data inputs.
	// Output[i] = extMatrix_row_{8+i}(state[0..15]) + originalData[i]
	// where originalData[i] = ins[8+i]
	for i := 0; i < gkrBlockSize; i++ {
		outputPos := gkrBlockSize + i
		// Build input list: 16 state variables + 1 feed-forward variable
		gateInputs := make([]gkr.Variable, gkrWidth+1)
		copy(gateInputs[:gkrWidth], state[:])
		gateInputs[gkrWidth] = ins[outputPos] // original data[i]
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
