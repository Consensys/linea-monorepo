package messagebus_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/wiop"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/logderivativesum"
	"github.com/consensys/linea-monorepo/prover-ray/wiop/compilers/messagebus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeVec builds a base-field ConcreteVector from uint64 literals.
func makeVec(vals ...uint64) *wiop.ConcreteVector {
	elems := make([]field.Element, len(vals))
	for i, v := range vals {
		elems[i].SetUint64(v)
	}
	return &wiop.ConcreteVector{Plain: field.VecFromBase(elems)}
}

// runRound executes every prover action registered on the runtime's current
// round.
func runRound(rt *wiop.Runtime) {
	for _, a := range rt.CurrentRound().ProverActions {
		a.Run(*rt)
	}
}

// checkAllVerifierActions evaluates every verifier action across every round
// of the runtime and returns the first non-nil error.
func checkAllVerifierActions(rt *wiop.Runtime) error {
	for _, r := range rt.System.Rounds {
		for _, va := range r.VerifierActions {
			if err := va.Check(*rt); err != nil {
				return err
			}
		}
	}
	return nil
}

// setupMessageBusCoins allocates a fresh round, registers the shared
// (α, β) pair on it, and stores both on the System. The fresh round
// stands in for the round on which the production Fiat–Shamir mechanism
// would sample the coins (always strictly after the round-0 columns it
// derives them from, so the runtime can sample them once those columns
// have been absorbed into the FS state).
func setupMessageBusCoins(sys *wiop.System) *wiop.Round {
	r := sys.NewRound()
	sys.MessageBusAlpha = r.NewCoinField(sys.Context.Childf("mb-alpha"))
	sys.MessageBusBeta = r.NewCoinField(sys.Context.Childf("mb-beta"))
	return r
}

// drive runs the canonical "assign-witness → advance to coin round → advance
// to result round → run prover" loop for a message-bus pipeline. After this
// returns, every prover action has executed and the verifier actions are
// ready to be checked.
//
// Round structure produced by setupMessageBusCoins + messagebus.Compile +
// logderivativesum.Compile:
//
//   - Round 0: user-witness columns (selectors, value columns, multiplicities).
//   - Round 1: the externally-supplied (α, β) coins; no prover actions.
//   - Round 2: LogDerivativeSum Result cells + Z columns; one prover action
//     per LogDerivativeSum query that assigns Z and the result.
func drive(rt *wiop.Runtime) {
	rt.AdvanceRound() // → coin round, samples α and β
	rt.AdvanceRound() // → result round
	runRound(rt)      // assigns Z columns + result cells
}

// TestCompile_TwoSegmentsBalanced is the canonical completeness case: one
// handle, one Send segment, one Receive segment, multiplicities all 1, every
// sent row matches a received row. The verifier must accept.
func TestCompile_TwoSegmentsBalanced(t *testing.T) {
	sys := wiop.NewSystemf("mb-balanced")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 4, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 4, wiop.PaddingDirectionNone)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "ping",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "ping",
		wiop.NewTable(colB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, makeVec(10, 20, 30, 40))
	rt.AssignColumn(colB, makeVec(10, 20, 30, 40))
	rt.AssignColumn(mulB, makeVec(1, 1, 1, 1))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_TamperedMultiplicity exercises the soundness path: with a
// receiver multiplicity that no longer counts the senders correctly, the
// per-handle cells should not sum to zero and the verifier must reject.
func TestCompile_TamperedMultiplicity(t *testing.T) {
	sys := wiop.NewSystemf("mb-tampered")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 4, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 4, wiop.PaddingDirectionNone)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "ping",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "ping",
		wiop.NewTable(colB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, makeVec(10, 20, 30, 40))
	rt.AssignColumn(colB, makeVec(10, 20, 30, 40))
	rt.AssignColumn(mulB, makeVec(2, 1, 1, 1)) // wrong: row 0 is counted twice

	drive(&rt)
	assert.Error(t, checkAllVerifierActions(&rt),
		"verifier must reject a multiplicity that miscounts the senders")
}

// TestCompile_MultipleSendersOneReceiver verifies that several Send segments
// can balance against a single Receive segment if multiplicities tally them
// correctly. Sender 1 emits values [10, 20]; sender 2 also emits [10, 20];
// the receiver holds [10, 20] with multiplicity [2, 2].
func TestCompile_MultipleSendersOneReceiver(t *testing.T) {
	sys := wiop.NewSystemf("mb-multi-senders")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modS1 := sys.NewSizedModule(sys.Context.Childf("modS1"), 2, wiop.PaddingDirectionNone)
	modS2 := sys.NewSizedModule(sys.Context.Childf("modS2"), 2, wiop.PaddingDirectionNone)
	modR := sys.NewSizedModule(sys.Context.Childf("modR"), 2, wiop.PaddingDirectionNone)
	colS1 := modS1.NewColumn(sys.Context.Childf("S1"), wiop.VisibilityOracle, r0)
	colS2 := modS2.NewColumn(sys.Context.Childf("S2"), wiop.VisibilityOracle, r0)
	colR := modR.NewColumn(sys.Context.Childf("R"), wiop.VisibilityOracle, r0)
	mulR := modR.NewColumn(sys.Context.Childf("mR"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-S1"),
		"segS1", "bus",
		wiop.NewTable(colS1.View()),
	)
	sys.NewMessageBusSend(
		sys.Context.Childf("send-S2"),
		"segS2", "bus",
		wiop.NewTable(colS2.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-R"),
		"segR", "bus",
		wiop.NewTable(colR.View()),
		mulR.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colS1, makeVec(10, 20))
	rt.AssignColumn(colS2, makeVec(10, 20))
	rt.AssignColumn(colR, makeVec(10, 20))
	rt.AssignColumn(mulR, makeVec(2, 2))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_MultiColumnTuples covers width > 1 tables: senders push
// (key, value) pairs; receivers consume them with matching multiplicity. The
// compiler must allocate an α coin (in addition to β) and fold each tuple
// via Horner before hashing.
func TestCompile_MultiColumnTuples(t *testing.T) {
	sys := wiop.NewSystemf("mb-tuples")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 4, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 4, wiop.PaddingDirectionNone)
	keyA := modA.NewColumn(sys.Context.Childf("kA"), wiop.VisibilityOracle, r0)
	valA := modA.NewColumn(sys.Context.Childf("vA"), wiop.VisibilityOracle, r0)
	keyB := modB.NewColumn(sys.Context.Childf("kB"), wiop.VisibilityOracle, r0)
	valB := modB.NewColumn(sys.Context.Childf("vB"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "kv",
		wiop.NewTable(keyA.View(), valA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "kv",
		wiop.NewTable(keyB.View(), valB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(keyA, makeVec(1, 2, 3, 4))
	rt.AssignColumn(valA, makeVec(10, 20, 30, 40))
	rt.AssignColumn(keyB, makeVec(1, 2, 3, 4))
	rt.AssignColumn(valB, makeVec(10, 20, 30, 40))
	rt.AssignColumn(mulB, makeVec(1, 1, 1, 1))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_FilteredSelectors covers Tables with a selector column on
// either side: only rows where the selector is 1 contribute to the
// accumulator.
func TestCompile_FilteredSelectors(t *testing.T) {
	sys := wiop.NewSystemf("mb-filter")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 4, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 4, wiop.PaddingDirectionNone)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	selA := modA.NewColumn(sys.Context.Childf("selA"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	selB := modB.NewColumn(sys.Context.Childf("selB"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "filtered",
		wiop.NewFilteredTable(selA.View(), colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "filtered",
		wiop.NewFilteredTable(selB.View(), colB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	// Active sends: A[0]=10, A[2]=20 (selA = [1,0,1,0]).
	// Active receives at same values, multiplicities counted by selB-respecting
	// filter (selB = [1,1,0,0], mulB = [1,1,0,0]) → receives 10 and 20 once each.
	rt.AssignColumn(colA, makeVec(10, 99, 20, 99))
	rt.AssignColumn(selA, makeVec(1, 0, 1, 0))
	rt.AssignColumn(colB, makeVec(10, 20, 77, 77))
	rt.AssignColumn(selB, makeVec(1, 1, 0, 0))
	rt.AssignColumn(mulB, makeVec(1, 1, 0, 0))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_TwoHandlesIndependent verifies that two unrelated handles in
// the same system are checked independently: they share the global (α, β)
// coins but each handle still gets its own LogDerivativeSum cells and its
// own verifier action, so tampering one handle cannot mask the other.
func TestCompile_TwoHandlesIndependent(t *testing.T) {
	sys := wiop.NewSystemf("mb-two-handles")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 2, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 2, wiop.PaddingDirectionNone)
	modC := sys.NewSizedModule(sys.Context.Childf("modC"), 2, wiop.PaddingDirectionNone)
	modD := sys.NewSizedModule(sys.Context.Childf("modD"), 2, wiop.PaddingDirectionNone)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)
	colC := modC.NewColumn(sys.Context.Childf("C"), wiop.VisibilityOracle, r0)
	colD := modD.NewColumn(sys.Context.Childf("D"), wiop.VisibilityOracle, r0)
	mulD := modD.NewColumn(sys.Context.Childf("mD"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"), "segA", "alpha",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"), "segB", "alpha",
		wiop.NewTable(colB.View()), mulB.View(),
	)
	sys.NewMessageBusSend(
		sys.Context.Childf("send-C"), "segC", "beta",
		wiop.NewTable(colC.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-D"), "segD", "beta",
		wiop.NewTable(colD.View()), mulD.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, makeVec(7, 8))
	rt.AssignColumn(colB, makeVec(7, 8))
	rt.AssignColumn(mulB, makeVec(1, 1))
	rt.AssignColumn(colC, makeVec(100, 200))
	rt.AssignColumn(colD, makeVec(100, 200))
	rt.AssignColumn(mulD, makeVec(1, 1))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_ReceiveWithoutMultiplicity covers the "implicit-1 multiplicity"
// path on the Receive side. Pass nil and the compiler must treat it as the
// constant 1.
func TestCompile_ReceiveWithoutMultiplicity(t *testing.T) {
	sys := wiop.NewSystemf("mb-no-mul")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewSizedModule(sys.Context.Childf("modA"), 2, wiop.PaddingDirectionNone)
	modB := sys.NewSizedModule(sys.Context.Childf("modB"), 2, wiop.PaddingDirectionNone)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"), "segA", "impl-one",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"), "segB", "impl-one",
		wiop.NewTable(colB.View()),
		nil, // explicit nil → constant-1 multiplicity
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, makeVec(5, 6))
	rt.AssignColumn(colB, makeVec(5, 6))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_NoMessageBusesIsNoOp asserts that running the compiler against
// a system with no message-bus queries is a no-op: no rounds, coins, or
// verifier actions are appended.
func TestCompile_NoMessageBusesIsNoOp(t *testing.T) {
	sys := wiop.NewSystemf("mb-empty")
	r0 := sys.NewRound()
	_ = r0

	roundsBefore := len(sys.Rounds)
	messagebus.Compile(sys)
	assert.Equal(t, roundsBefore, len(sys.Rounds),
		"Compile must not append rounds when there are no MessageBus queries")
}

// TestCompile_DynamicModule_TwoSegmentsBalanced is the completeness counterpart
// of TestCompile_TwoSegmentsBalanced for dynamic modules: the participating
// modules' sizes are not declared statically but established at runtime by the
// first column assignment. The same compiled System is then re-driven across
// two different runtime sizes to confirm size-agnostic compilation.
//
// TODO: unskip once [logderivativesum] supports dynamic modules. The
// messagebus pass itself is already size-agnostic — it never inspects
// Module.Size() — but logderivativesum.buildZ currently panics on a dynamic
// module because it needs the static size to allocate the Z column. Until
// that lands, the end-to-end path stops at logderivativesum.Compile.
func TestCompile_DynamicModule_TwoSegmentsBalanced(t *testing.T) {
	t.Skip("dynamic modules not yet supported by logderivativesum.Compile (panics on m.Size()==0)")

	sys := wiop.NewSystemf("mb-dyn-balanced")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewDynamicModule(sys.Context.Childf("modA"), wiop.PaddingDirectionRight)
	modB := sys.NewDynamicModule(sys.Context.Childf("modB"), wiop.PaddingDirectionRight)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "ping",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "ping",
		wiop.NewTable(colB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	cases := []struct {
		name string
		vals []uint64
	}{
		{"size-4", []uint64{10, 20, 30, 40}},
		{"size-8", []uint64{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rt := wiop.NewRuntime(sys)
			rt.AssignColumn(colA, makeVec(tc.vals...))
			rt.AssignColumn(colB, makeVec(tc.vals...))
			ones := make([]uint64, len(tc.vals))
			for i := range ones {
				ones[i] = 1
			}
			rt.AssignColumn(mulB, makeVec(ones...))

			drive(&rt)
			require.NoError(t, checkAllVerifierActions(&rt))
		})
	}
}

// TestCompile_DynamicModule_TamperedFails is the soundness counterpart on a
// dynamic module: a multiplicity that miscounts the senders must be rejected
// regardless of the runtime size.
//
// TODO: unskip once [logderivativesum] supports dynamic modules — see
// [TestCompile_DynamicModule_TwoSegmentsBalanced] for the rationale.
func TestCompile_DynamicModule_TamperedFails(t *testing.T) {
	t.Skip("dynamic modules not yet supported by logderivativesum.Compile (panics on m.Size()==0)")

	sys := wiop.NewSystemf("mb-dyn-tampered")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewDynamicModule(sys.Context.Childf("modA"), wiop.PaddingDirectionRight)
	modB := sys.NewDynamicModule(sys.Context.Childf("modB"), wiop.PaddingDirectionRight)
	colA := modA.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colB := modB.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "ping",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "ping",
		wiop.NewTable(colB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(colA, makeVec(10, 20, 30, 40))
	rt.AssignColumn(colB, makeVec(10, 20, 30, 40))
	rt.AssignColumn(mulB, makeVec(2, 1, 1, 1)) // wrong: row 0 counted twice

	drive(&rt)
	assert.Error(t, checkAllVerifierActions(&rt),
		"verifier must reject a multiplicity that miscounts senders on a dynamic module")
}

// TestCompile_DynamicModule_MultiColumnTuples covers the width-2 α-fold over
// dynamic modules: (key, value) tuples sent and received with matching
// multiplicities. Exercises both the Horner fold and the runtime-determined
// row count simultaneously.
//
// TODO: unskip once [logderivativesum] supports dynamic modules — see
// [TestCompile_DynamicModule_TwoSegmentsBalanced] for the rationale.
func TestCompile_DynamicModule_MultiColumnTuples(t *testing.T) {
	t.Skip("dynamic modules not yet supported by logderivativesum.Compile (panics on m.Size()==0)")

	sys := wiop.NewSystemf("mb-dyn-tuples")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	modA := sys.NewDynamicModule(sys.Context.Childf("modA"), wiop.PaddingDirectionRight)
	modB := sys.NewDynamicModule(sys.Context.Childf("modB"), wiop.PaddingDirectionRight)
	keyA := modA.NewColumn(sys.Context.Childf("kA"), wiop.VisibilityOracle, r0)
	valA := modA.NewColumn(sys.Context.Childf("vA"), wiop.VisibilityOracle, r0)
	keyB := modB.NewColumn(sys.Context.Childf("kB"), wiop.VisibilityOracle, r0)
	valB := modB.NewColumn(sys.Context.Childf("vB"), wiop.VisibilityOracle, r0)
	mulB := modB.NewColumn(sys.Context.Childf("mB"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-A"),
		"segA", "kv",
		wiop.NewTable(keyA.View(), valA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-B"),
		"segB", "kv",
		wiop.NewTable(keyB.View(), valB.View()),
		mulB.View(),
	)

	messagebus.Compile(sys)
	logderivativesum.Compile(sys)

	rt := wiop.NewRuntime(sys)
	rt.AssignColumn(keyA, makeVec(1, 2, 3, 4))
	rt.AssignColumn(valA, makeVec(10, 20, 30, 40))
	rt.AssignColumn(keyB, makeVec(1, 2, 3, 4))
	rt.AssignColumn(valB, makeVec(10, 20, 30, 40))
	rt.AssignColumn(mulB, makeVec(1, 1, 1, 1))

	drive(&rt)
	require.NoError(t, checkAllVerifierActions(&rt))
}

// TestCompile_WidthMismatchPanics asserts that the compiler rejects two
// participants on the same handle with different column widths — the
// alpha-fold is only meaningful when every participant has the same width.
func TestCompile_WidthMismatchPanics(t *testing.T) {
	sys := wiop.NewSystemf("mb-width-mismatch")
	r0 := sys.NewRound()
	setupMessageBusCoins(sys)

	mod := sys.NewSizedModule(sys.Context.Childf("mod"), 2, wiop.PaddingDirectionNone)
	colA := mod.NewColumn(sys.Context.Childf("A"), wiop.VisibilityOracle, r0)
	colA2 := mod.NewColumn(sys.Context.Childf("A2"), wiop.VisibilityOracle, r0)
	colB := mod.NewColumn(sys.Context.Childf("B"), wiop.VisibilityOracle, r0)

	sys.NewMessageBusSend(
		sys.Context.Childf("send-1col"), "segA", "h",
		wiop.NewTable(colA.View()),
	)
	sys.NewMessageBusReceive(
		sys.Context.Childf("recv-2col"), "segB", "h",
		wiop.NewTable(colB.View(), colA2.View()),
		nil,
	)

	assert.Panics(t, func() { messagebus.Compile(sys) })
}

// TestCompile_MissingCoinsPanics asserts that running the compiler with
// unreduced MessageBus entries but no externally-supplied (α, β) coins
// panics. The pass intentionally does not allocate its own coins.
func TestCompile_MissingCoinsPanics(t *testing.T) {
	sys := wiop.NewSystemf("mb-missing-coins")
	r0 := sys.NewRound()

	mod := sys.NewSizedModule(sys.Context.Childf("m"), 2, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)
	sys.NewMessageBusSend(
		sys.Context.Childf("send"), "seg", "h",
		wiop.NewTable(col.View()),
	)

	assert.Panics(t, func() { messagebus.Compile(sys) },
		"Compile must panic when MessageBusAlpha/MessageBusBeta are unset")
}

// TestNewMessageBusReceive_WrongMultiplicityModulePanics asserts that the
// constructor rejects a vector-valued multiplicity bound to a different
// module than the receiving table.
func TestNewMessageBusReceive_WrongMultiplicityModulePanics(t *testing.T) {
	sys := wiop.NewSystemf("mb-bad-mul-mod")
	r0 := sys.NewRound()
	_ = r0
	mod := sys.NewSizedModule(sys.Context.Childf("m"), 2, wiop.PaddingDirectionNone)
	foreign := sys.NewSizedModule(sys.Context.Childf("foreign"), 2, wiop.PaddingDirectionNone)
	col := mod.NewColumn(sys.Context.Childf("c"), wiop.VisibilityOracle, r0)

	assert.Panics(t, func() {
		sys.NewMessageBusReceive(
			sys.Context.Childf("recv-bad-mod"), "segA", "h",
			wiop.NewTable(col.View()),
			wiop.NewConstantVector(foreign, field.NewFromString("1")),
		)
	})
}
