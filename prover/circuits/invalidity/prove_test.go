package invalidity_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	keccakDummy "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	invalidityPI "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// compileSetupProve compiles the circuit, runs PLONK setup with an unsafe SRS,
// then generates and verifies a proof from the given assignment. To make the setup generation efficient we use small tree-depth, mock zkevm and dummy compilation suite.
func compileSetupProve(t *testing.T, builder circuits.Builder, circuitID circuits.CircuitID, assignment *invalidity.CircuitInvalidity) {
	t.Helper()

	t.Logf("Compiling circuit %s...", circuitID)
	ccs, err := builder.Compile()
	require.NoError(t, err)
	t.Logf("Compiled: %d constraints", ccs.GetNbConstraints())

	t.Log("Running MakeSetup with unsafe SRS...")
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuitID,
		ccs,
		circuits.NewUnsafeSRSProvider(),
		map[string]any{},
	)
	require.NoError(t, err)

	t.Log("Generating and verifying proof...")
	proof, err := circuits.ProveCheck(
		&setup,
		assignment,
		emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
		emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
	)
	require.NoError(t, err)
	require.NotNil(t, proof)
	t.Logf("Proof generated and verified for %s", circuitID)
}

// TestProveFilteredAddress generates and verifies a proof for the
// invalidity-filtered-address circuit (FilteredAddressFrom case).
func TestProveFilteredAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping proof generation test in short mode")
	}

	const maxRlpByteSize = 1024
	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f2bD50")
	fromAddr := common.HexToAddress("0x00aed6")

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(59144),
		Nonce:     1,
		GasTipCap: big.NewInt(1000000000),
		GasFeeCap: big.NewInt(100000000000),
		Gas:       21000,
		To:        &toAddr,
		Value:     big.NewInt(1000),
	})

	encodedTx := ethereum.EncodeTxForSigning(tx)
	rlpEncodedTx := make([]byte, len(encodedTx))
	copy(rlpEncodedTx, encodedTx)

	txHash := ethereum.GetTxHash(tx)
	stateRoot := field.RandomOctuplet()

	// Keccak compiled IOP for the builder (circuit compilation)
	keccakCompForBuilder := invalidity.MakeKeccakCompiledIOP(maxRlpByteSize, keccakDummy.Compile)

	// Keccak compiled IOP + proof for the assignment (witness generation)
	kcomp, kproof := invalidity.MakeKeccakProofs(tx, maxRlpByteSize, keccakDummy.Compile)

	builder := invalidity.NewBuilder(
		invalidity.Config{
			KeccakCompiledIOP: keccakCompForBuilder,
			MaxRlpByteSize:    maxRlpByteSize,
		},
		&invalidity.FilteredAddressCircuit{},
	)

	assi := invalidity.AssigningInputs{
		Transaction:       tx,
		FromAddress:       fromAddr,
		MaxRlpByteSize:    maxRlpByteSize,
		InvalidityType:    invalidity.FilteredAddressFrom,
		RlpEncodedTx:      rlpEncodedTx,
		KeccakCompiledIOP: kcomp,
		KeccakProof:       kproof,
		StateRootHash:     stateRoot,
		FuncInputs: public_input.Invalidity{
			TxHash:         txHash,
			FromAddress:    linTypes.EthAddress(fromAddr),
			StateRootHash:  linTypes.KoalaOctuplet(stateRoot),
			FromIsFiltered: true,
			ToIsFiltered:   false,
			ToAddress:      linTypes.EthAddress(*tx.To()),
		},
	}

	assignment := &invalidity.CircuitInvalidity{
		SubCircuit: &invalidity.FilteredAddressCircuit{},
	}
	assignment.Assign(assi)

	compileSetupProve(t, builder, circuits.InvalidityFilteredAddressCircuitID, assignment)
}

// TestProveNonceBalance generates and verifies a proof for the
// invalidity-nonce-balance circuit (BadBalance case).
func TestProveNonceBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping proof generation test in short mode")
	}

	const (
		maxRlpByteSize = 1024
		treeDepth      = 10
	)

	tree, _, _ := genShomei(t, tcases, treeDepth)
	root := tree.Root

	// Use the first test case (BadBalance: account has zero balance, tx has value=1)
	tcase := tcases[0]
	proof, _ := tree.Prove(0)
	leaf := tcase.Leaf.Hash()

	assi := invalidity.AssigningInputs{
		AccountTrieInputs: invalidity.AccountTrieInputs{
			Proof:       proof,
			Leaf:        leaf,
			Root:        root,
			Account:     tcase.Account,
			LeafOpening: tcase.Leaf,
		},
		Transaction:    types.NewTx(&tcase.Tx),
		FromAddress:    tcase.FromAddress,
		MaxRlpByteSize: maxRlpByteSize,
		InvalidityType: tcase.InvalidityType,
		FuncInputs: public_input.Invalidity{
			ToAddress:     linTypes.EthAddress(*tcase.Tx.To),
			StateRootHash: linTypes.KoalaOctuplet(root),
			FromAddress:   linTypes.EthAddress(tcase.FromAddress),
		},
	}

	b := ethereum.EncodeTxForSigning(assi.Transaction)
	assi.RlpEncodedTx = make([]byte, len(b))
	copy(assi.RlpEncodedTx, b)

	assi.FuncInputs.TxHash = common.Hash(crypto.Keccak256(assi.RlpEncodedTx))

	// Keccak compiled IOP for the builder
	keccakCompForBuilder := invalidity.MakeKeccakCompiledIOP(maxRlpByteSize, keccakDummy.Compile)

	// Keccak compiled IOP + proof for the assignment
	kcomp, kproof := invalidity.MakeKeccakProofs(assi.Transaction, maxRlpByteSize, keccakDummy.Compile)
	assi.KeccakCompiledIOP = kcomp
	assi.KeccakProof = kproof

	// Use the same treeDepth for the builder so the circuit's Merkle proof
	// allocation matches the test tree's proof size.
	builder := invalidity.NewBuilder(
		invalidity.Config{
			Depth:             treeDepth,
			KeccakCompiledIOP: keccakCompForBuilder,
			MaxRlpByteSize:    maxRlpByteSize,
		},
		&invalidity.BadNonceBalanceCircuit{},
	)

	assignment := &invalidity.CircuitInvalidity{
		SubCircuit: &invalidity.BadNonceBalanceCircuit{},
	}
	assignment.Assign(assi)

	compileSetupProve(t, builder, circuits.InvalidityNonceBalanceCircuitID, assignment)
}

// TestProveBadPrecompile generates and verifies a proof for the
// invalidity-precompile-logs circuit (BadPrecompile case).
func TestProveBadPrecompile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping proof generation test in short mode")
	}

	// Use mockZkevmPI (from bad_precompile_test.go) which directly registers
	// limb-based public input columns. This is the same mock that the existing
	// TestBadPrecompileCircuit uses and is proven to produce wizard values
	// consistent with the circuit's checkPublicInputs extraction.
	inputs := invalidityPI.Inputs{
		FixedInputs: fixedInputs,
		CaseInputs: invalidityPI.CaseInputs{
			HasBadPrecompile: true,
			NumL2Logs:        5,
		},
	}

	comp, wizProof := mockZkevmPI(inputs)
	testZkevm := &zkevm.ZkEvm{InitialCompiledIOP: comp}

	txToAddr := common.HexToAddress("0xdeadbeefdeadb")
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID: big.NewInt(59144),
		To:      &txToAddr,
	})

	builder := invalidity.NewBuilder(
		invalidity.Config{
			Zkevm: testZkevm,
		},
		&invalidity.BadPrecompileCircuit{},
	)

	assi := invalidity.AssigningInputs{
		InvalidityType:   invalidity.BadPrecompile,
		Zkevm:            testZkevm,
		ZkevmWizardProof: wizProof,
		Transaction:      tx,
		FuncInputs: public_input.Invalidity{
			StateRootHash: piStateRootHash,
			TxHash:        common.Hash(piTxHashBytes),
			FromAddress:   linTypes.EthAddress(piFromAddressBytes),
			ToAddress:     linTypes.EthAddress(*tx.To()),
		},
	}

	assignment := &invalidity.CircuitInvalidity{
		SubCircuit: &invalidity.BadPrecompileCircuit{
			InvalidityType: invalidity.BadPrecompile,
		},
	}
	assignment.Assign(assi)

	compileSetupProve(t, builder, circuits.InvalidityPrecompileLogsCircuitID, assignment)
}
