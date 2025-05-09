package zkevm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
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

func TestZKEVM(t *testing.T) {
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}

	// Run existing subtests for attributes
	t.Run("Arithmetization", TestArithmetization)
	t.Run("Sha2", TestSha2)
	t.Run("StateManager", TestStateManager)
	t.Run("CompiledIOP", TestCompiledIOP)

	// Failing tests due to not supporting serialization of `func()`

	t.Run("Keccak", TestKeccak)
	t.Run("Modexp", TestModexp)
	t.Run("Ecadd", TestEcadd)
	t.Run("Ecmul", TestEcmul)
	t.Run("Ecpair", TestEcpair)
}

func TestModexp(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Modexp, deserializedModexp) {
		t.Fatalf("Mis-matched fields after serde Modexp (ignoring unexported fields)")
	}
}

func TestEcadd(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Ecadd, deserializedEcadd) {
		t.Fatalf("Mis-matched fields after serde Ecadd (ignoring unexported fields)")
	}
}

func TestEcmul(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Ecmul, deserializedEcmul) {
		t.Fatalf("Mis-matched fields after serde Ecmul (ignoring unexported fields)")
	}
}

func TestEcpair(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Ecpair, deserializedEcpair) {
		t.Fatalf("Mis-matched fields after serde Ecpair (ignoring unexported fields)")
	}
}

// TestArithmetization tests serialization and deserialization of the Arithmetization field.
func TestArithmetization(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Arithmetization, deserialized) {
		t.Fatalf("Mis-matched fields after serde Arithmetization (ignoring unexported fields)")
	}
}

// TestKeccak tests serialization and deserialization of the Keccak field.
func TestKeccak(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Keccak, deserialized) {
		t.Fatalf("Mis-matched fields after serde Keccak (ignoring unexported fields)")
	}
}

// TestSha2 tests serialization and deserialization of the Sha2 field.
func TestSha2(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.Sha2, deserialized) {
		t.Fatalf("Mis-matched fields after serde Sha2 (ignoring unexported fields)")
	}
}

// TestCompiledIOP tests serialization and deserialization of the WizardIOP field.
func TestCompiledIOP(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.WizardIOP.Columns, deserializedIOP.Columns) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Columns (ignoring unexported fields)")
	}

	if !compareExportedFields(z.WizardIOP.Coins, deserializedIOP.Coins) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Coins (ignoring unexported fields)")
	}

	if !compareExportedFields(z.WizardIOP.QueriesParams, deserializedIOP.QueriesParams) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: QueriesParams (ignoring unexported fields)")
	}

	if !compareExportedFields(z.WizardIOP.QueriesNoParams, deserializedIOP.QueriesNoParams) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: QueriesNoParams (ignoring unexported fields)")
	}

	fmt.Printf("Original Prover action:%+v \n", z.WizardIOP.SubProvers)
	fmt.Printf("Deserialized Prover action:%+v \n", deserializedIOP.SubProvers)
	if !compareExportedFields(z.WizardIOP.SubProvers, deserializedIOP.SubProvers) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: SubProvers (ignoring unexported fields)")
	}

	fmt.Printf("Original Verifier action:%+v \n", z.WizardIOP.SubVerifiers)
	fmt.Printf("Deserialized Verifier action:%+v \n", deserializedIOP.SubVerifiers)
	if !compareExportedFields(z.WizardIOP.SubVerifiers, deserializedIOP.SubVerifiers) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: SubVerifiers (ignoring unexported fields)")
	}

	fmt.Printf("Original FSHookPreSampling:%+v \n", z.WizardIOP.FiatShamirHooksPreSampling)
	fmt.Printf("Deserialized FSHookPreSampling:%+v \n", deserializedIOP.FiatShamirHooksPreSampling)
	if !compareExportedFields(z.WizardIOP.FiatShamirHooksPreSampling, deserializedIOP.FiatShamirHooksPreSampling) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: FiatShamirHookPreSampling (ignoring unexported fields)")
	}

	fmt.Println("Original precomputed map:", z.WizardIOP.Precomputed)
	fmt.Println("Deserialized precomputed map:", deserializedIOP.Precomputed)
	if !compareExportedFields(z.WizardIOP.Precomputed, deserializedIOP.Precomputed) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: Precomputed (ignoring unexported fields)")
	}

	fmt.Println("Original PcsCtxs:", z.WizardIOP.PcsCtxs)
	fmt.Println("Deserialized PcsCtxs:", deserializedIOP.PcsCtxs)
	if !compareExportedFields(z.WizardIOP.PcsCtxs, deserializedIOP.PcsCtxs) {
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
	if !compareExportedFields(z.WizardIOP.PublicInputs, deserializedIOP.PublicInputs) {
		t.Fatalf("Mis-matched fields after serde CompiledIOP: PublicInputs (ignoring unexported fields)")
	}

}

// TestStateManager tests serialization and deserialization of the StateManager field.
func TestStateManager(t *testing.T) {
	z := GetZkEVM()
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
	if !compareExportedFields(z.StateManager, deserialized) {
		t.Fatalf("Mis-matched fields after serde StateManager (ignoring unexported fields)")
	}

}

// compareExportedFields checks if two values are equal, ignoring unexported fields, including in nested structs.
// It logs mismatched fields with their paths and values.
func compareExportedFields(a, b interface{}) bool {
	return compareExportedFieldsWithPath(a, b, "")
}

// compareExportedFieldsWithPath is a helper that tracks the field path for logging.
func compareExportedFieldsWithPath(a, b interface{}, path string) bool {
	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(b)

	// Ensure both values are valid
	if !v1.IsValid() || !v2.IsValid() {
		if !v1.IsValid() && !v2.IsValid() {
			return true
		}
		fmt.Printf("Mismatch at %s: one value is invalid (v1: %v, v2: %v, types: %v, %v)\n", path, a, b, reflect.TypeOf(a), reflect.TypeOf(b))
		return false
	}

	// Ensure same type
	if v1.Type() != v2.Type() {
		fmt.Printf("Mismatch at %s: types differ (v1: %v, v2: %v, types: %v, %v)\n", path, a, b, v1.Type(), v2.Type())
		return false
	}

	// Handle maps
	if v1.Kind() == reflect.Map {
		// fmt.Printf("Map comparision for v1:%v v2:%v\n", v1, v2)
		if v1.Len() != v2.Len() {
			fmt.Printf("Mismatch at %s: map lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1.Len(), v2.Len(), v1.Type())
			return false
		}
		for _, key := range v1.MapKeys() {
			value1 := v1.MapIndex(key)
			value2 := v2.MapIndex(key)
			if !value2.IsValid() {
				fmt.Printf("Mismatch at %s: key %v is missing in second map\n", path, key)
				return false
			}
			keyPath := fmt.Sprintf("%s[%v]", path, key)
			if !compareExportedFieldsWithPath(value1.Interface(), value2.Interface(), keyPath) {
				return false
			}
		}
		return true
	}

	// Handle pointers by dereferencing
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v1.IsNil() != v2.IsNil() {
			fmt.Printf("Mismatch at %s: nil status differs (v1: %v, v2: %v, type: %v)\n", path, a, b, v1.Type())
			return false
		}
		return compareExportedFieldsWithPath(v1.Elem().Interface(), v2.Elem().Interface(), path)
	}

	// Handle structs
	if v1.Kind() == reflect.Struct {
		// fmt.Printf("Handling struct comparision\n")
		equal := true
		for i := 0; i < v1.NumField(); i++ {
			// Skip unexported fields
			if !v1.Type().Field(i).IsExported() {
				continue
			}
			f1 := v1.Field(i)
			f2 := v2.Field(i)
			fieldName := v1.Type().Field(i).Name
			// Construct field path
			fieldPath := fieldName
			if path != "" {
				fieldPath = path + "." + fieldName
			}

			// Recursively compare field values
			if !compareExportedFieldsWithPath(f1.Interface(), f2.Interface(), fieldPath) {
				equal = false
			}
		}
		return equal
	}

	// Handle slices
	if v1.Kind() == reflect.Slice {
		if v1.Len() != v2.Len() {
			fmt.Printf("Mismatch at %s: slice lengths differ (v1: %v, v2: %v, type: %v)\n", path, v1, v2, v1.Type())
			return false
		}
		equal := true
		for i := 0; i < v1.Len(); i++ {
			// Construct element path
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			if !compareExportedFieldsWithPath(v1.Index(i).Interface(), v2.Index(i).Interface(), elemPath) {
				equal = false
			}
		}
		return equal
	}

	// For other types, use DeepEqual and log if mismatched
	if !reflect.DeepEqual(a, b) {
		fmt.Printf("Mismatch at %s: values differ (v1: %v, v2: %v, type: %v)\n", path, a, b, v1.Type())
		return false
	}
	return true
}

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := config.TracesLimits{
		Add:                                  1 << 23,
		Bin:                                  1 << 22,
		Blake2Fmodexpdata:                    1 << 18,
		Blockdata:                            1 << 16,
		Blockhash:                            1 << 16,
		Ecdata:                               1 << 22,
		Euc:                                  1 << 20,
		Exp:                                  1 << 18,
		Ext:                                  1 << 24,
		Gas:                                  1 << 20,
		Hub:                                  1 << 25,
		Logdata:                              1 << 20,
		Loginfo:                              1 << 16,
		Mmio:                                 1 << 25,
		Mmu:                                  1 << 25,
		Mod:                                  1 << 21,
		Mul:                                  1 << 20,
		Mxp:                                  1 << 23,
		Oob:                                  1 << 22,
		Rlpaddr:                              1 << 16,
		Rlptxn:                               1 << 21,
		Rlptxrcpt:                            1 << 21,
		Rom:                                  1 << 26,
		Romlex:                               1 << 16,
		Shakiradata:                          1 << 19,
		Shf:                                  1 << 20,
		Stp:                                  1 << 18,
		Trm:                                  1 << 19,
		Txndata:                              1 << 18,
		Wcp:                                  1 << 22,
		Binreftable:                          1 << 24,
		Shfreftable:                          1 << 16,
		Instdecoder:                          1 << 13,
		PrecompileEcrecoverEffectiveCalls:    1 << 13,
		PrecompileSha2Blocks:                 1 << 13,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       1 << 10,
		PrecompileModexpEffectiveCalls4096:   1 << 4,
		PrecompileEcaddEffectiveCalls:        1 << 12,
		PrecompileEcmulEffectiveCalls:        1 << 9,
		PrecompileEcpairingEffectiveCalls:    1 << 9,
		PrecompileEcpairingMillerLoops:       1 << 10,
		PrecompileEcpairingG2MembershipCalls: 1 << 10,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 17,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    1 << 12,
		ShomeiMerkleProofs:                   1 << 18,
	}

	return FullZKEVMWithSuite(&traceLimits, CompilationSuite{}, &config.Config{})
}

// GetAffinities returns a list of affinities for the following modules. This
// affinities regroup how the modules are grouped.
//
//	ecadd / ecmul / ecpairing
//	hub / hub.scp / hub.acp
//	everything related to keccak
func GetAffinities(z *ZkEvm) [][]column.Natural {

	return [][]column.Natural{
		{
			z.Ecmul.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecadd.AlignedGnarkData.IsActive.(column.Natural),
			z.Ecpair.AlignedFinalExpCircuit.IsActive.(column.Natural),
			z.Ecpair.AlignedG2MembershipData.IsActive.(column.Natural),
			z.Ecpair.AlignedMillerLoopCircuit.IsActive.(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("hub.HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.scp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.acp_ADDRESS_HI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.ccp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.envcp_HUB_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("hub.stkcp_PEEK_AT_STACK_POW_4").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("KECCAK_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("CLEANING_KECCAK_CleanLimb").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_KECCAK_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_KECCAK_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_IS_ACTIVE_").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAKF_BLOCK_BASE_2_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("KECCAK_OVER_BLOCKS_TAGS_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("HASH_OUTPUT_Hash_Lo").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("SHA2_IMPORT_PAD_HASH_NUM").(column.Natural),
			z.WizardIOP.Columns.GetHandle("DECOMPOSITION_SHA2_Decomposed_Len_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LENGTH_CONSISTENCY_SHA2_BYTE_LEN_0_0").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_FILTERS_SPAGHETTI").(column.Natural),
			z.WizardIOP.Columns.GetHandle("LANE_SHA2_Lane").(column.Natural),
			z.WizardIOP.Columns.GetHandle("Coefficient_SHA2").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_IS_ACTIVE").(column.Natural),
			z.WizardIOP.Columns.GetHandle("SHA2_OVER_BLOCK_SHA2_COMPRESSION_CIRCUIT_IS_ACTIVE").(column.Natural),
		},
		{
			z.WizardIOP.Columns.GetHandle("mmio.CN_ABC").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmio.MMIO_STAMP").(column.Natural),
			z.WizardIOP.Columns.GetHandle("mmu.STAMP").(column.Natural),
		},
	}
}
