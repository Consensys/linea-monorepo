package gkrverifier

import (
	"math/big"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/gkr"
	gkrkoalabear "github.com/consensys/gnark/gkr/koalabear"
	gkrposeidon2 "github.com/consensys/gnark/gkr/permutation/gkr-poseidon2/koalabear"
)

const gateDriver = "../../src/test/zkc/gkr/gate.zkc"

// gateCheck is an evaluation of a gate, as the driver takes it.
type gateCheck struct {
	gate     int
	inputs   []extensions.E6
	expected extensions.E6
}

// gateChecksInputs serializes the checks, returning the input of the driver and the index in
// "checks" of the first word of each check's expected value.
func gateChecksInputs(checks []gateCheck) (inputs map[string]zkcInput, expectedAt []int) {
	var words zkcInput
	for _, c := range checks {
		words = append(words, uint32(c.gate))
		for i := range c.inputs {
			words = append(words, e6Words(&c.inputs[i])...)
		}
		expectedAt = append(expectedAt, len(words))
		words = append(words, e6Words(&c.expected)...)
	}
	return map[string]zkcInput{"nb_checks": {uint32(len(checks))}, "checks": words}, expectedAt
}

func e6FromBase(x koalabear.Element) (res extensions.E6) {
	res.B0.A0 = x
	return
}

// evaluateGate is a reference interpreter of gkr.GateBytecode over E₆: it returns the result of the gate on the inputs.
func evaluateGate(gate gkr.GateBytecode, inputs []extensions.E6) extensions.E6 {
	vars := make([]extensions.E6, 0, gate.EvaluatorSize(len(inputs)))
	for _, c := range gate.Constants {
		var e koalabear.Element
		e.SetBigInt(c)
		vars = append(vars, e6FromBase(e))
	}
	vars = append(vars, inputs...)
	for _, inst := range gate.Instructions {
		arg := func(i int) *extensions.E6 { return &vars[inst.Inputs[i]] }
		var res extensions.E6
		switch inst.Op {
		case gkr.OpAdd:
			res = *arg(0)
			for i := 1; i < len(inst.Inputs); i++ {
				res.Add(&res, arg(i))
			}
		case gkr.OpSub:
			res = *arg(0)
			for i := 1; i < len(inst.Inputs); i++ {
				res.Sub(&res, arg(i))
			}
		case gkr.OpMul:
			res = *arg(0)
			for i := 1; i < len(inst.Inputs); i++ {
				res.Mul(&res, arg(i))
			}
		case gkr.OpNeg:
			res.Neg(arg(0))
		case gkr.OpMulAcc:
			res.Mul(arg(0), arg(1))
			res.Add(&res, arg(2))
		case gkr.OpSumExp17:
			res.Add(arg(0), arg(1))
			res.Add(&res, arg(2))
			res.ExpInt64(res, 17)
		default:
			panic("unknown operation")
		}
		vars = append(vars, res)
	}
	return vars[len(vars)-1]
}

func randomE6(rng *rand.Rand) (x extensions.E6) {
	coords := make([]koalabear.Element, 6)
	for i := range coords {
		coords[i].SetUint64(uint64(rng.Uint32()))
	}
	return *x.SetBase(coords)
}

// randomGateChecks returns a check of every gate of the circuit on random inputs.
func randomGateChecks(rng *rand.Rand, circuit gkr.SerializableCircuit, gateOfWire []int) []gateCheck {
	var checks []gateCheck
	done := make(map[int]bool)
	for wireI, gateI := range gateOfWire {
		if gateI < 0 || done[gateI] {
			continue
		}
		done[gateI] = true
		inputs := make([]extensions.E6, len(circuit[wireI].Inputs))
		for i := range inputs {
			inputs[i] = randomE6(rng)
		}
		checks = append(checks, gateCheck{gateI, inputs, evaluateGate(circuit[wireI].Gate.Evaluate, inputs)})
	}
	return checks
}

// runGateChecks runs the driver on the checks, with the gate tables in the file tables. If corrupt
// is not nil, it is applied to the driver's inputs first, given the index in inputs["checks"] of
// the first word of each check's expected value. It reports whether the driver accepted.
func runGateChecks(t *testing.T, tables string, checks []gateCheck, corrupt func(inputs map[string]zkcInput, expectedAt []int)) bool {
	t.Helper()
	inputs, expectedAt := gateChecksInputs(checks)
	if corrupt != nil {
		corrupt(inputs, expectedAt)
	}
	return runZkc(t, inputs, gateDriver, tables)
}

func writeTables(t *testing.T, circuit gkr.SerializableCircuit) (path string, gateOfWire []int) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "gates.zkc")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if gateOfWire, err = WriteGateTables(f, circuit); err != nil {
		t.Fatal(err)
	}
	return
}

// testGateChecks checks that the driver accepts the checks and rejects each corrupted one.
func testGateChecks(t *testing.T, tables string, checks []gateCheck) {
	t.Helper()
	if !runGateChecks(t, tables, checks, nil) {
		t.Fatal("correct evaluations rejected")
	}
	for _, i := range []int{0, len(checks) / 2, len(checks) - 1} {
		corruptedCheck := func(coordinate int) func(map[string]zkcInput, []int) {
			return func(inputs map[string]zkcInput, expectedAt []int) {
				w := &inputs["checks"][expectedAt[i]+coordinate]
				*w = (*w + 1) % modulusU32
			}
		}
		for _, coordinate := range []int{0, 5} {
			if runGateChecks(t, tables, checks, corruptedCheck(coordinate)) {
				t.Errorf("corrupted check %d, coordinate %d, accepted", i, coordinate)
			}
		}
	}
}

// TestGates checks the gate interpreter against the gates of the Poseidon2 compression circuit,
// on the values gnark computes for its wires and on random elements of E₆.
func TestGates(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))

	compressor, err := gkrposeidon2.NewCompressor(poseidon2.GetDefaultParameters())
	if err != nil {
		t.Fatal(err)
	}
	circuit := compressor.Circuit

	tables, gateOfWire := writeTables(t, circuit)
	nbGates := 0
	for _, g := range gateOfWire {
		nbGates = max(nbGates, g+1)
	}
	t.Logf("%d wires, %d distinct gates", len(circuit), nbGates)

	t.Run("wire values", func(t *testing.T) {
		const nbInstances = 2
		assignment := make(gkrkoalabear.WireAssignment, len(circuit))
		for _, in := range circuit.Inputs() {
			assignment[in] = make([]koalabear.Element, nbInstances)
			koalabear.Vector(assignment[in]).MustSetRandom()
		}
		assignment.Complete(circuit)

		var checks []gateCheck
		for wireI, wire := range circuit {
			if wire.IsInput() {
				continue
			}
			for instanceI := range nbInstances {
				c := gateCheck{gate: gateOfWire[wireI], expected: e6FromBase(assignment[wireI][instanceI])}
				for _, in := range wire.Inputs {
					c.inputs = append(c.inputs, e6FromBase(assignment[in][instanceI]))
				}
				checks = append(checks, c)
			}
		}
		testGateChecks(t, tables, checks)

		// zkc execute compiles and runs; a run of no checks is the compilation alone.
		start := time.Now()
		if !runGateChecks(t, tables, nil, nil) {
			t.Fatal("no checks rejected")
		}
		compilation := time.Since(start)
		start = time.Now()
		if !runGateChecks(t, tables, checks, nil) {
			t.Fatal("correct evaluations rejected")
		}
		t.Logf("compilation of the tables of %d gates: %v; execution of %d checks: %v", nbGates, compilation, len(checks), time.Since(start)-compilation)
	})

	t.Run("E6 inputs", func(t *testing.T) {
		testGateChecks(t, tables, randomGateChecks(rng, circuit, gateOfWire))
	})
}

// TestGateOperations checks every operation of gkr.GateOp, most of which the compression circuit
// does not use.
func TestGateOperations(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 10))

	gate := func(nbConsts int, instructions ...gkr.GateInstruction) gkr.GateBytecode {
		consts := []*big.Int{big.NewInt(5), big.NewInt(-3), new(big.Int).Lsh(big.NewInt(1), 40)}[:nbConsts]
		return gkr.GateBytecode{Constants: consts, Instructions: instructions}
	}
	wire := func(g gkr.GateBytecode, inputs ...int) gkr.SerializableWire {
		return gkr.SerializableWire{Gate: gkr.SerializableGate{Evaluate: g, NbIn: len(inputs)}, Inputs: inputs}
	}

	// Wires 0, 1, 2 are inputs. The slots of a gate are its constants, then its inputs, then the
	// results of its instructions.
	circuit := gkr.SerializableCircuit{
		{}, {}, {},
		wire(gkr.IdentityBytecode(), 0),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpAdd, Inputs: []uint16{0, 1, 2, 1}}), 0, 1, 2),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpSub, Inputs: []uint16{0, 1, 2}}), 0, 1, 2),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpMul, Inputs: []uint16{0, 1, 2, 0}}), 0, 1, 2),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpNeg, Inputs: []uint16{1}}), 0, 1),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpMulAcc, Inputs: []uint16{0, 1, 2}}), 0, 1, 2),
		wire(gate(0, gkr.GateInstruction{Op: gkr.OpSumExp17, Inputs: []uint16{0, 1, 2}}), 0, 1, 2),
		// the constants are slots 0, 1, 2, the inputs x, y are slots 3, 4, and the results are slots 5, ...
		wire(gate(3,
			gkr.GateInstruction{Op: gkr.OpMul, Inputs: []uint16{0, 3}},         // slot 5: 5x
			gkr.GateInstruction{Op: gkr.OpMulAcc, Inputs: []uint16{1, 4, 5}},   // slot 6: -3y + 5x
			gkr.GateInstruction{Op: gkr.OpSub, Inputs: []uint16{6, 5, 3}},      // slot 7
			gkr.GateInstruction{Op: gkr.OpSumExp17, Inputs: []uint16{7, 3, 4}}, // slot 8
			gkr.GateInstruction{Op: gkr.OpAdd, Inputs: []uint16{8, 7, 2}},      // slot 9: includes the constant 2⁴⁰
			gkr.GateInstruction{Op: gkr.OpNeg, Inputs: []uint16{9}},            // slot 10
		), 1, 2),
	}

	tables, gateOfWire := writeTables(t, circuit)
	testGateChecks(t, tables, randomGateChecks(rng, circuit, gateOfWire))
}
