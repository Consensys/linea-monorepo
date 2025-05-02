package zkevm

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
)

// hashArithmetization computes a SHA-256 hash of the serialized Arithmetization instance.
func hashArithmetization(a *arithmetization.Arithmetization) ([]byte, error) {
	data, err := serialization.SerializeValue(reflect.ValueOf(a), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("serialize Arithmetization: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// TestArithmetization tests serialization and deserialization of the Arithmetization field.
func TestArithmetization(t *testing.T) {
	// Get a valid ZkEvm instance with inflated trace limits
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Arithmetization == nil {
		t.Fatal("Arithmetization field is nil")
	}

	// Compute hash of original Arithmetization
	originalHash, err := hashArithmetization(z.Arithmetization)
	if err != nil {
		t.Fatalf("Failed to hash original Arithmetization: %v", err)
	}

	// Serialize Arithmetization
	arithmetizationSer, err := serialization.SerializeValue(reflect.ValueOf(z.Arithmetization), serialization.DeclarationMode)
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

	// Compute hash of deserialized Arithmetization
	deserializedHash, err := hashArithmetization(deserialized)
	if err != nil {
		t.Fatalf("Failed to hash deserialized Arithmetization: %v", err)
	}

	// Compare hashes
	if !bytes.Equal(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
	}
}

// hashKeccak computes a SHA-256 hash of the serialized Keccak instance.
func hashKeccak(k *keccak.KeccakZkEVM) ([]byte, error) {
	data, err := serialization.SerializeValue(reflect.ValueOf(k), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("serialize Keccak: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// TestKeccak tests serialization and deserialization of the Keccak field.
func TestKeccak(t *testing.T) {
	// Get a valid ZkEvm instance with inflated trace limits
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Keccak == nil {
		t.Fatal("Keccak field is nil")
	}

	// Compute hash of original Keccak
	originalHash, err := hashKeccak(z.Keccak)
	if err != nil {
		t.Fatalf("Failed to hash original Keccak: %v", err)
	}

	// Serialize Keccak
	keccakSer, err := serialization.SerializeValue(reflect.ValueOf(z.Keccak), serialization.DeclarationMode)
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

	// Compute hash of deserialized Keccak
	deserializedHash, err := hashKeccak(deserialized)
	if err != nil {
		t.Fatalf("Failed to hash deserialized Keccak: %v", err)
	}

	// Compare hashes
	if !bytes.Equal(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
	}
}

// hashSha2 computes a SHA-256 hash of the serialized Sha2 instance.
func hashSha2(s *sha2.Sha2SingleProvider) ([]byte, error) {
	data, err := serialization.SerializeValue(reflect.ValueOf(s), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("serialize Sha2: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// TestSha2 tests serialization and deserialization of the Sha2 field.
func TestSha2(t *testing.T) {
	// Get a valid ZkEvm instance with inflated trace limits
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Sha2 == nil {
		t.Fatal("Sha2 field is nil")
	}

	// Compute hash of original Sha2
	originalHash, err := hashSha2(z.Sha2)
	if err != nil {
		t.Fatalf("Failed to hash original Sha2: %v", err)
	}

	// Serialize Sha2
	sha2Ser, err := serialization.SerializeValue(reflect.ValueOf(z.Sha2), serialization.DeclarationMode)
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

	// Compute hash of deserialized Sha2
	deserializedHash, err := hashSha2(deserialized)
	if err != nil {
		t.Fatalf("Failed to hash deserialized Sha2: %v", err)
	}

	// Compare hashes
	if !bytes.Equal(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
	}
}

func TestDiscoverMissingTypesForArithmetization(t *testing.T) {
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}
	if z.Arithmetization == nil {
		t.Fatal("Arithmetization field is nil")
	}

	data, missingTypes, err := serialization.SerializeValueWithDiscovery(reflect.ValueOf(z.Arithmetization), serialization.DeclarationMode)
	if err != nil {
		t.Logf("Serialization encountered issues: %v", err)
	}
	if len(missingTypes) > 0 {
		t.Logf("Missing types for Arithmetization:")
		for _, typ := range missingTypes {
			t.Logf(" - %s", typ)
		}
	} else {
		t.Log("No missing types found for Arithmetization.")
	}
	if data == nil {
		t.Error("Serialized data is nil")
	}
}

/*
// hashZkEvm computes a SHA-256 hash of the serialized ZkEvm instance.
func hashZkEvm(z *ZkEvm) ([]byte, error) {
	data, err := serialization.SerializeValue(reflect.ValueOf(z), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("serialize ZkEvm: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// TestSeDeserZkEVM tests serialization and deserialization of a ZkEvm instance field by field.
func TestSeDeserZkEVM(t *testing.T) {
	// Get a valid ZkEvm instance with inflated trace limits
	z := GetZkEVM()
	if z == nil {
		t.Fatal("GetZkEVM returned nil")
	}

	// Compute hash of original ZkEvm
	originalHash, err := hashZkEvm(z)
	if err != nil {
		t.Fatalf("Failed to hash original ZkEvm: %v", err)
	}

	// Serialize fields individually
	arithmetizationSer, err := serialization.SerializeValue(reflect.ValueOf(z.Arithmetization), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize arithmetization: %v", err)
	}

	keccakSer, err := serialization.SerializeValue(reflect.ValueOf(z.Keccak), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize keccak: %v", err)
	}

	stateManagerSer, err := serialization.SerializeValue(reflect.ValueOf(z.StateManager), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize stateManager: %v", err)
	}

	publicInputSer, err := serialization.SerializeValue(reflect.ValueOf(z.PublicInput), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize publicInput: %v", err)
	}

	ecdsaSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecdsa), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize ecdsa: %v", err)
	}

	modexpSer, err := serialization.SerializeValue(reflect.ValueOf(z.Modexp), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize modexp: %v", err)
	}

	ecaddSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecadd), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize ecadd: %v", err)
	}

	ecmulSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecmul), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize ecmul: %v", err)
	}

	ecpairSer, err := serialization.SerializeValue(reflect.ValueOf(z.Ecpair), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize ecpair: %v", err)
	}

	sha2Ser, err := serialization.SerializeValue(reflect.ValueOf(z.Sha2), serialization.DeclarationMode)
	if err != nil {
		t.Fatalf("Failed to serialize sha2: %v", err)
	}

	wizardIOPSer, err := serialization.SerializeCompiledIOP(z.WizardIOP)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Store serialized data in a slice (order matters)
	serializedData := [][]byte{
		arithmetizationSer,
		keccakSer,
		stateManagerSer,
		publicInputSer,
		ecdsaSer,
		modexpSer,
		ecaddSer,
		ecmulSer,
		ecpairSer,
		sha2Ser,
		wizardIOPSer,
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new ZkEvm
	deserialized := &ZkEvm{}

	arithmetizationVal, err := serialization.DeserializeValue(serializedData[0], serialization.DeclarationMode, reflect.TypeOf(&arithmetization.Arithmetization{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize arithmetization: %v", err)
	}
	deserialized.Arithmetization = arithmetizationVal.Interface().(*arithmetization.Arithmetization)

	keccakVal, err := serialization.DeserializeValue(serializedData[1], serialization.DeclarationMode, reflect.TypeOf(&keccak.KeccakZkEVM{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize keccak: %v", err)
	}
	deserialized.Keccak = keccakVal.Interface().(*keccak.KeccakZkEVM)

	stateManagerVal, err := serialization.DeserializeValue(serializedData[2], serialization.DeclarationMode, reflect.TypeOf(&statemanager.StateManager{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize stateManager: %v", err)
	}
	deserialized.StateManager = stateManagerVal.Interface().(*statemanager.StateManager)

	publicInputVal, err := serialization.DeserializeValue(serializedData[3], serialization.DeclarationMode, reflect.TypeOf(&publicInput.PublicInput{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize publicInput: %v", err)
	}
	deserialized.PublicInput = publicInputVal.Interface().(*publicInput.PublicInput)

	ecdsaVal, err := serialization.DeserializeValue(serializedData[4], serialization.DeclarationMode, reflect.TypeOf(&ecdsa.EcdsaZkEvm{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize ecdsa: %v", err)
	}
	deserialized.Ecdsa = ecdsaVal.Interface().(*ecdsa.EcdsaZkEvm)

	modexpVal, err := serialization.DeserializeValue(serializedData[5], serialization.DeclarationMode, reflect.TypeOf(&modexp.Module{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize modexp: %v", err)
	}
	deserialized.Modexp = modexpVal.Interface().(*modexp.Module)

	ecaddVal, err := serialization.DeserializeValue(serializedData[6], serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcAdd{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize ecadd: %v", err)
	}
	deserialized.Ecadd = ecaddVal.Interface().(*ecarith.EcAdd)

	ecmulVal, err := serialization.DeserializeValue(serializedData[7], serialization.DeclarationMode, reflect.TypeOf(&ecarith.EcMul{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize ecmul: %v", err)
	}
	deserialized.Ecmul = ecmulVal.Interface().(*ecarith.EcMul)

	ecpairVal, err := serialization.DeserializeValue(serializedData[8], serialization.DeclarationMode, reflect.TypeOf(&ecpair.ECPair{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize ecpair: %v", err)
	}
	deserialized.Ecpair = ecpairVal.Interface().(*ecpair.ECPair)

	sha2Val, err := serialization.DeserializeValue(serializedData[9], serialization.DeclarationMode, reflect.TypeOf(&sha2.Sha2SingleProvider{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize sha2: %v", err)
	}
	deserialized.Sha2 = sha2Val.Interface().(*sha2.Sha2SingleProvider)

	deserialized.WizardIOP, err = serialization.DeserializeCompiledIOP(serializedData[10])
	if err != nil {
		t.Fatalf("Failed to deserialize wizardIOP: %v", err)
	}

	// Compute hash of deserialized ZkEvm
	deserializedHash, err := hashZkEvm(deserialized)
	if err != nil {
		t.Fatalf("Failed to hash deserialized ZkEvm: %v", err)
	}

	// Compare hashes
	if !bytes.Equal(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
	}
} */

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
