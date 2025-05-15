package serdetests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecpair"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/modexp"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager"
)

// serializeValue serializes a value using the serialization package.
func serializeValue(value interface{}) ([]byte, error) {
	serializedData, err := serialization.SerializeValue(reflect.ValueOf(value), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize value: %w", err)
	}
	return serializedData, nil
}

func TestSerdeZKEVM(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}

	// Run existing subtests for attributes
	t.Run("Arithmetization", TestSerdeArithmetization)
	t.Run("Sha2", TestSerdeSha2)
	t.Run("StateManager", TestSerdeStateManager)
	t.Run("CompiledIOP", TestSerdeCompiledIOP)
	t.Run("Keccak", TestSerdeKeccak)
	t.Run("Ecadd", TestSerdeEcadd)
	t.Run("Ecmul", TestSerdeEcmul)
	t.Run("ECDSA", TestSerdeECDSA)
	t.Run("Modexp", TestSerdeModexp)
	t.Run("Ecpair", TestSerdeEcpair)
}

func TestSerdeModexp(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Modexp == nil {
		t.Fatal("Modexp field is nil")
	}

	// Serialize the original Modexp
	modexpSer, err := serializeValue(z.Modexp)
	if err != nil {
		t.Fatalf("Failed to serialize Modexp: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Modexp
	deserializedModexpVal, err := serialization.DeserializeValue(modexpSer, serialization.DeclarationMode, reflect.TypeOf(&modexp.Module{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Modexp: %v", err)
	}
	deserializedModexp, ok := deserializedModexpVal.Interface().(*modexp.Module)
	if !ok {
		t.Fatalf("Deserialized value is not *modexp.Module: got %T", deserializedModexpVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Modexp, deserializedModexp) {
		t.Fatalf("Mis-matched fields after serde Modexp (ignoring unexported fields)")
	}
}

func TestSerdeECDSA(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Ecdsa == nil {
		t.Fatal("Ecadd field is nil")
	}

	// Serialize the original Ecdsa
	ecdsaSer, err := serializeValue(z.Ecdsa)
	if err != nil {
		t.Fatalf("Failed to serialize Ecdsa: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Ecadd
	deserializedEcdsaVal, err := serialization.DeserializeValue(ecdsaSer, serialization.DeclarationMode, reflect.TypeOf(&ecdsa.EcdsaZkEvm{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Ecdsa: %v", err)
	}
	deserializedEcdsa, ok := deserializedEcdsaVal.Interface().(*ecdsa.EcdsaZkEvm)
	if !ok {
		t.Fatalf("Deserialized value is not *ecarith.EcAdd: got %T", deserializedEcdsaVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Ecdsa, deserializedEcdsa) {
		t.Fatalf("Mis-matched fields after serde Ecadd (ignoring unexported fields)")
	}
}

func TestSerdeEcadd(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Ecadd == nil {
		t.Fatal("Ecadd field is nil")
	}

	// Serialize the original Ecadd
	ecaddSer, err := serializeValue(z.Ecadd)
	if err != nil {
		t.Fatalf("Failed to serialize Ecadd: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Ecadd
	deserializedEcaddVal, err := serialization.DeserializeValue(ecaddSer, serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcAdd{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Ecadd: %v", err)
	}
	deserializedEcadd, ok := deserializedEcaddVal.Interface().(*ecarith.EcAdd)
	if !ok {
		t.Fatalf("Deserialized value is not *ecarith.EcAdd: got %T", deserializedEcaddVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Ecadd, deserializedEcadd) {
		t.Fatalf("Mis-matched fields after serde Ecadd (ignoring unexported fields)")
	}
}

func TestSerdeEcmul(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Ecmul == nil {
		t.Fatal("Ecmul field is nil")
	}

	// Serialize the original Ecmul
	ecmulSer, err := serializeValue(z.Ecmul)
	if err != nil {
		t.Fatalf("Failed to serialize Ecmul: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Ecmul
	deserializedEcmulVal, err := serialization.DeserializeValue(ecmulSer, serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcMul{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Ecmul: %v", err)
	}
	deserializedEcmul, ok := deserializedEcmulVal.Interface().(*ecarith.EcMul)
	if !ok {
		t.Fatalf("Deserialized value is not *ecarith.EcMul: got %T", deserializedEcmulVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Ecmul, deserializedEcmul) {
		t.Fatalf("Mis-matched fields after serde Ecmul (ignoring unexported fields)")
	}
}

func TestSerdeEcpair(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Ecpair == nil {
		t.Fatal("Ecpair field is nil")
	}

	// Serialize the original Ecpair
	ecpairSer, err := serializeValue(z.Ecpair)
	if err != nil {
		t.Fatalf("Failed to serialize Ecpair: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Ecpair
	deserializedEcpairVal, err := serialization.DeserializeValue(ecpairSer, serialization.DeclarationMode, reflect.TypeOf(&ecpair.ECPair{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Ecpair: %v", err)
	}
	deserializedEcpair, ok := deserializedEcpairVal.Interface().(*ecpair.ECPair)
	if !ok {
		t.Fatalf("Deserialized value is not *ecpair.ECPair: got %T", deserializedEcpairVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Ecpair, deserializedEcpair) {
		t.Fatalf("Mis-matched fields after serde Ecpair (ignoring unexported fields)")
	}
}

// TestArithmetization tests serialization and deserialization of the Arithmetization field.
func TestSerdeArithmetization(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Arithmetization == nil {
		t.Fatal("Arithmetization field is nil")
	}

	// Serialize the original Arithmetization
	arithmetizationSer, err := serializeValue(z.Arithmetization)
	if err != nil {
		t.Fatalf("Failed to serialize Arithmetization: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Arithmetization
	deserializedVal, err := serialization.DeserializeValue(arithmetizationSer, serialization.DeclarationMode, reflect.TypeOf(&arithmetization.Arithmetization{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Arithmetization: %v", err)
	}
	deserialized, ok := deserializedVal.Interface().(*arithmetization.Arithmetization)
	if !ok {
		t.Fatalf("Deserialized value is not *arithmetization.Arithmetization: got %T", deserializedVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Arithmetization, deserialized) {
		t.Fatalf("Mis-matched fields after serde Arithmetization (ignoring unexported fields)")
	}
}

// TestKeccak tests serialization and deserialization of the Keccak field.
func TestSerdeKeccak(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Keccak == nil {
		t.Fatal("Keccak field is nil")
	}

	// Serialize the original Keccak
	keccakSer, err := serializeValue(z.Keccak)
	if err != nil {
		t.Fatalf("Failed to serialize Keccak: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Keccak
	deserializedVal, err := serialization.DeserializeValue(keccakSer, serialization.DeclarationMode, reflect.TypeOf(&keccak.KeccakZkEVM{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Keccak: %v", err)
	}
	deserialized, ok := deserializedVal.Interface().(*keccak.KeccakZkEVM)
	if !ok {
		t.Fatalf("Deserialized value is not *keccak.KeccakZkEVM: got %T", deserializedVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Keccak, deserialized) {
		t.Fatalf("Mis-matched fields after serde Keccak (ignoring unexported fields)")
	}
}

// TestSha2 tests serialization and deserialization of the Sha2 field.
func TestSerdeSha2(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Sha2 == nil {
		t.Fatal("Sha2 field is nil")
	}

	// Serialize the original Sha2
	sha2Ser, err := serializeValue(z.Sha2)
	if err != nil {
		t.Fatalf("Failed to serialize Sha2: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new Sha2
	deserializedVal, err := serialization.DeserializeValue(sha2Ser, serialization.DeclarationMode, reflect.TypeOf(&sha2.Sha2SingleProvider{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize Sha2: %v", err)
	}
	deserialized, ok := deserializedVal.Interface().(*sha2.Sha2SingleProvider)
	if !ok {
		t.Fatalf("Deserialized value is not *sha2.Sha2SingleProvider: got %T", deserializedVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.Sha2, deserialized) {
		t.Fatalf("Mis-matched fields after serde Sha2 (ignoring unexported fields)")
	}
}

// TestCompiledIOP tests serialization and deserialization of the WizardIOP field.
func TestSerdeCompiledIOP(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.WizardIOP == nil {
		t.Fatal("WizardIOP field is nil")
	}

	// Serialize the original CompiledIOP
	compiledIOPSer, err := serialization.SerializeCompiledIOP(z.WizardIOP)
	if err != nil {
		t.Fatalf("Failed to serialize CompiledIOP: %v", err)
	}

	// Deserialize into a new CompiledIOP
	deserializedIOP, err := serialization.DeserializeCompiledIOP(compiledIOPSer)
	if err != nil {
		t.Fatalf("Failed to deserialize CompiledIOP: %v", err)
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.WizardIOP.Columns, deserializedIOP.Columns) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Columns (ignoring unexported fields)")
	}

	if !test_utils.CompareExportedFields(z.WizardIOP.Coins, deserializedIOP.Coins) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Coins (ignoring unexported fields)")
	}

	if !test_utils.CompareExportedFields(z.WizardIOP.QueriesParams, deserializedIOP.QueriesParams) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: QueriesParams (ignoring unexported fields)")
	}

	if !test_utils.CompareExportedFields(z.WizardIOP.QueriesNoParams, deserializedIOP.QueriesNoParams) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: QueriesNoParams (ignoring unexported fields)")
	}

	fmt.Printf("Original Prover action:%+v \n", z.WizardIOP.SubProvers)
	fmt.Printf("Deserialized Prover action:%+v \n", deserializedIOP.SubProvers)
	if !test_utils.CompareExportedFields(z.WizardIOP.SubProvers, deserializedIOP.SubProvers) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: SubProvers (ignoring unexported fields)")
	}

	fmt.Printf("Original Verifier action:%+v \n", z.WizardIOP.SubVerifiers)
	fmt.Printf("Deserialized Verifier action:%+v \n", deserializedIOP.SubVerifiers)
	if !test_utils.CompareExportedFields(z.WizardIOP.SubVerifiers, deserializedIOP.SubVerifiers) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: SubVerifiers (ignoring unexported fields)")
	}

	fmt.Printf("Original FSHookPreSampling:%+v \n", z.WizardIOP.FiatShamirHooksPreSampling)
	fmt.Printf("Deserialized FSHookPreSampling:%+v \n", deserializedIOP.FiatShamirHooksPreSampling)
	if !test_utils.CompareExportedFields(z.WizardIOP.FiatShamirHooksPreSampling, deserializedIOP.FiatShamirHooksPreSampling) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: FiatShamirHookPreSampling (ignoring unexported fields)")
	}

	fmt.Println("Original precomputed map:", z.WizardIOP.Precomputed)
	fmt.Println("Deserialized precomputed map:", deserializedIOP.Precomputed)
	if !test_utils.CompareExportedFields(z.WizardIOP.Precomputed, deserializedIOP.Precomputed) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Precomputed (ignoring unexported fields)")
	}

	fmt.Println("Original PcsCtxs:", z.WizardIOP.PcsCtxs)
	fmt.Println("Deserialized PcsCtxs:", deserializedIOP.PcsCtxs)
	if !test_utils.CompareExportedFields(z.WizardIOP.PcsCtxs, deserializedIOP.PcsCtxs) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: PcsCtxs (ignoring unexported fields)")
	}

	if z.WizardIOP.DummyCompiled != deserializedIOP.DummyCompiled {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: DummyCompiled")
	}

	fmt.Println("Original FiatShamirSetup:", z.WizardIOP.FiatShamirSetup)
	fmt.Println("Deserialized FiatShamirSetup:", deserializedIOP.FiatShamirSetup)
	if z.WizardIOP.FiatShamirSetup != deserializedIOP.FiatShamirSetup {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: FiatShamirSetup")
	}

	fmt.Println("Original PublicInputs:", z.WizardIOP.PublicInputs)
	fmt.Println("Deserialized PublicInputs:", deserializedIOP.PublicInputs)
	if !test_utils.CompareExportedFields(z.WizardIOP.PublicInputs, deserializedIOP.PublicInputs) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: PublicInputs (ignoring unexported fields)")
	}

	if z.WizardIOP.DummyCompiled != deserializedIOP.DummyCompiled {
		t.Fatalf("Mismatch in DummyCompiled: Original=%v, Deserialized=%v", z.WizardIOP.DummyCompiled, deserializedIOP.DummyCompiled)
	}

	if z.WizardIOP.SelfRecursionCount != deserializedIOP.SelfRecursionCount {
		t.Fatalf("Mismatch in SelfRecursionCount: Original=%v, Deserialized=%v", z.WizardIOP.SelfRecursionCount, deserializedIOP.SelfRecursionCount)
	}

}

// TestStateManager tests serialization and deserialization of the StateManager field.
func TestSerdeStateManager(t *testing.T) {
	z := distributed.GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.StateManager == nil {
		t.Fatal("StateManager field is nil")
	}

	// Serialize the original StateManager
	stateManagerSer, err := serializeValue(z.StateManager)
	if err != nil {
		t.Fatalf("Failed to serialize StateManager: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new StateManager
	deserializedVal, err := serialization.DeserializeValue(stateManagerSer, serialization.DeclarationMode, reflect.TypeOf(&statemanager.StateManager{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize StateManager: %v", err)
	}
	deserialized, ok := deserializedVal.Interface().(*statemanager.StateManager)
	if !ok {
		t.Fatalf("Deserialized value is not *statemanager.StateManager: got %T", deserializedVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(z.StateManager, deserialized) {
		t.Fatalf("Mis-matched fields after serde StateManager (ignoring unexported fields)")
	}

}
