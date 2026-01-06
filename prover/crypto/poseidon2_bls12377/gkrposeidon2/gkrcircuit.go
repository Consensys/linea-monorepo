package gkrposeidon2

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/constraint/solver/gkrgates"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/gkrapi"
	"github.com/consensys/gnark/std/gkrapi/gkr"
	_ "github.com/consensys/gnark/std/hash/all" // to ensure all hashes are registered

	"github.com/consensys/gnark-crypto/ecc"
	frBls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	poseidon2Bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
)

// extKeyGate applies the external matrix mul, then adds the round key.
// Because of its symmetry, we don't need to define distinct x1 and x2 versions of it.
func extKeyGate(roundKey frontend.Variable) gkr.GateFunction {
	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		if len(x) != 2 {
			panic("expected 2 inputs")
		}
		return api.Add(api.Mul(x[0], 2), x[1], roundKey)
	}
}

// pow4Gate computes a -> a^4
func pow4Gate(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 1 {
		panic("expected 1 input")
	}
	y := api.Mul(x[0], x[0])
	y = api.Mul(y, y)
	return y
}

// pow4TimesGate computes a, b -> a^4 * b
func pow4TimesGate(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	y := api.Mul(x[0], x[0])
	y = api.Mul(y, y)
	return api.Mul(y, x[1])
}

// pow2Gate computes a -> a^2
func pow2Gate(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 1 {
		panic("expected 1 input")
	}
	return api.Mul(x[0], x[0])
}

// pow2TimesGate computes a, b -> a^2 * b
func pow2TimesGate(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	return api.Mul(x[0], x[0], x[1])
}

// extGate2 applies the external matrix mul, outputting the second element of the result.
func extGate2(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	return api.Add(api.Mul(x[1], 2), x[0])
}

// intKeyGate2 applies the internal matrix mul, then adds the round key.
func intKeyGate2(roundKey frontend.Variable) gkr.GateFunction {
	return func(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
		if len(x) != 2 {
			panic("expected 2 inputs")
		}
		return api.Add(api.Mul(x[1], 3), x[0], roundKey)
	}
}

// intGate2 applies the internal matrix mul. The round key is zero.
func intGate2(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	return api.Add(api.Mul(x[1], 3), x[0])
}

// extAddGate applies the first row of the external matrix to the first two elements and adds the third.
func extAddGate(api gkr.GateAPI, x ...frontend.Variable) frontend.Variable {
	if len(x) != 3 {
		panic("expected 3 inputs")
	}
	return api.Add(api.Mul(x[0], 2), x[1], x[2])
}

// defineCircuit defines the GKR circuit for the Poseidon2 compression over BLS12-377.
// insLeft and insRight are the inputs to the compression permutation (t=2).
// They must be padded to a power of 2.
func defineCircuit(insLeft, insRight []frontend.Variable) (*gkrapi.API, gkr.Variable, error) {
	const (
		xI = iota
		yI
	)

	p := poseidon2Bls12377.GetDefaultParameters()
	gateNamer := newRoundGateNamer(p)
	rF := p.NbFullRounds
	rP := p.NbPartialRounds
	halfRf := rF / 2

	gkrApi := gkrapi.New()

	x, err := gkrApi.Import(insLeft)
	if err != nil {
		return nil, -1, err
	}
	y, err := gkrApi.Import(insRight)
	if err != nil {
		return nil, -1, err
	}
	y0 := y // save to feed forward at the end

	// apply the s-box to u. For BLS12-377 Poseidon2 the s-box degree is 17:
	// u^17 = (u^4)^4 * u
	sBox := func(u gkr.Variable) gkr.Variable {
		v := gkrApi.Gate(pow4Gate, u)           // u^4
		return gkrApi.Gate(pow4TimesGate, v, u) // u^17
	}

	extKeySBox := func(round, varI int, a, b gkr.Variable) gkr.Variable {
		return sBox(gkrApi.NamedGate(gateNamer.linear(varI, round), a, b))
	}

	intKeySBox2 := func(round int, a, b gkr.Variable) gkr.Variable {
		return sBox(gkrApi.NamedGate(gateNamer.linear(yI, round), a, b))
	}

	fullRound := func(i int) {
		x1 := extKeySBox(i, xI, x, y)
		x, y = x1, extKeySBox(i, yI, y, x) // external matrix is symmetric
	}

	for i := range halfRf {
		fullRound(i)
	}

	{
		// i = halfRf: first partial round
		x1 := extKeySBox(halfRf, xI, x, y)
		x, y = x1, gkrApi.Gate(extGate2, x, y)
	}

	for i := halfRf + 1; i < halfRf+rP; i++ {
		x1 := extKeySBox(i, xI, x, y) // first row of internal matrix matches external
		x, y = x1, gkrApi.Gate(intGate2, x, y)
	}

	{
		i := halfRf + rP
		// first iteration of the final batch of full rounds
		x1 := extKeySBox(i, xI, x, y)
		x, y = x1, intKeySBox2(i, x, y)
	}

	for i := halfRf + rP + 1; i < rP+rF; i++ {
		fullRound(i)
	}

	// Apply external matrix one last time and add y0 to perform the feed-forward
	// (this matches gnark/std/permutation/poseidon2.(*Permutation).Compress).
	y = gkrApi.NamedGate(gateNamer.linear(yI, rP+rF), y, x, y0)

	return gkrApi, y, nil
}

// registerGkrSolverOptions is a wrapper for RegisterGkrSolverOptions
// that performs the registration for the curve associated with api.
func registerGkrSolverOptions(api frontend.API) {
	_ = api
	RegisterGkrSolverOptions(ecc.BLS12_377)
}

func permuteHint(m *big.Int, ins, outs []*big.Int) error {
	if m.Cmp(ecc.BLS12_377.ScalarField()) != 0 {
		return errors.New("only bls12-377 supported")
	}
	if len(ins) != 2 || len(outs) != 1 {
		return errors.New("expected 2 inputs and 1 output")
	}
	var x [2]frBls12377.Element
	x[0].SetBigInt(ins[0])
	x[1].SetBigInt(ins[1])
	y0 := x[1]

	err := bls12377Permutation().Permutation(x[:])
	x[1].Add(&x[1], &y0) // feed forward
	x[1].BigInt(outs[0])
	return err
}

var bls12377Permutation = sync.OnceValue(func() *poseidon2Bls12377.Permutation {
	params := poseidon2Bls12377.GetDefaultParameters()
	return poseidon2Bls12377.NewPermutation(2, params.NbFullRounds, params.NbPartialRounds)
})

// RegisterGkrSolverOptions registers the GKR gates corresponding to the given curves for the solver.
func RegisterGkrSolverOptions(curves ...ecc.ID) {
	if len(curves) == 0 {
		panic("expected at least one curve")
	}
	solver.RegisterHint(permuteHint)
	for _, curve := range curves {
		switch curve {
		case ecc.BLS12_377:
			if err := registerGkrGatesBls12377(); err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("curve %s not currently supported", curve))
		}
	}
}

func registerGkrGatesBls12377() error {
	const (
		x = iota
		y
	)

	p := poseidon2Bls12377.GetDefaultParameters()
	halfRf := p.NbFullRounds / 2
	gateNames := newRoundGateNamer(p)

	if err := gkrgates.Register(pow2Gate, 1, gkrgates.WithUnverifiedDegree(2), gkrgates.WithNoSolvableVar()); err != nil {
		return err
	}
	if err := gkrgates.Register(pow4Gate, 1, gkrgates.WithUnverifiedDegree(4), gkrgates.WithNoSolvableVar()); err != nil {
		return err
	}
	if err := gkrgates.Register(pow2TimesGate, 2, gkrgates.WithUnverifiedDegree(3), gkrgates.WithNoSolvableVar()); err != nil {
		return err
	}
	if err := gkrgates.Register(pow4TimesGate, 2, gkrgates.WithUnverifiedDegree(5), gkrgates.WithNoSolvableVar()); err != nil {
		return err
	}

	if err := gkrgates.Register(intGate2, 2, gkrgates.WithUnverifiedDegree(1), gkrgates.WithUnverifiedSolvableVar(0)); err != nil {
		return err
	}

	extKeySBox := func(round int, varIndex int) error {
		return gkrgates.Register(
			extKeyGate(&p.RoundKeys[round][varIndex]),
			2,
			gkrgates.WithUnverifiedDegree(1),
			gkrgates.WithUnverifiedSolvableVar(0),
			gkrgates.WithName(gateNames.linear(varIndex, round)),
		)
	}

	intKeySBox2 := func(round int) error {
		return gkrgates.Register(
			intKeyGate2(&p.RoundKeys[round][1]),
			2,
			gkrgates.WithUnverifiedDegree(1),
			gkrgates.WithUnverifiedSolvableVar(0),
			gkrgates.WithName(gateNames.linear(y, round)),
		)
	}

	fullRound := func(i int) error {
		if err := extKeySBox(i, x); err != nil {
			return err
		}
		return extKeySBox(i, y)
	}

	for round := range halfRf {
		if err := fullRound(round); err != nil {
			return err
		}
	}

	{ // round = halfRf: first partial one
		if err := extKeySBox(halfRf, x); err != nil {
			return err
		}
	}

	for round := halfRf + 1; round < halfRf+p.NbPartialRounds; round++ {
		if err := extKeySBox(round, x); err != nil {
			return err
		}
	}

	{
		round := halfRf + p.NbPartialRounds
		if err := extKeySBox(round, x); err != nil {
			return err
		}
		if err := intKeySBox2(round); err != nil {
			return err
		}
	}

	for round := halfRf + p.NbPartialRounds + 1; round < p.NbPartialRounds+p.NbFullRounds; round++ {
		if err := fullRound(round); err != nil {
			return err
		}
	}

	return gkrgates.Register(
		extAddGate,
		3,
		gkrgates.WithUnverifiedDegree(1),
		gkrgates.WithUnverifiedSolvableVar(0),
		gkrgates.WithName(gateNames.linear(y, p.NbPartialRounds+p.NbFullRounds)),
	)
}

type roundGateNamer string

func newRoundGateNamer(p fmt.Stringer) roundGateNamer {
	return roundGateNamer(p.String())
}

func (n roundGateNamer) linear(varIndex, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("x%d-l-op-round=%d;%s", varIndex, round, n))
}

// Included for parity with gnark's reference implementation.
func (n roundGateNamer) integrated(varIndex, round int) gkr.GateName {
	return gkr.GateName(fmt.Sprintf("x%d-i-op-round=%d;%s", varIndex, round, n))
}

func init() {
	// Ensure the hint is registered even if callers don't explicitly call RegisterGkrSolverOptions.
	// Gate registration is done lazily in finalize via RegisterGkrSolverOptions(curve).
	solver.RegisterHint(permuteHint)
}
