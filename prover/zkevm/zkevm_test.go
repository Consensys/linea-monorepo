package zkevm

import (
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

// serializeAndHash serializes a value and computes its SHA-256 hash.
func serializeAndHash(value interface{}) ([]byte, []byte, error) {
	serializedData, err := serialization.SerializeValue(reflect.ValueOf(value), serialization.DeclarationMode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize value: %w", err)
	}
	hash := sha256.Sum256(serializedData)
	return serializedData, hash[:], nil
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

	// Serialize and compute hash for original Arithmetization
	arithmetizationSer, originalHash, err := serializeAndHash(z.Arithmetization)
	if err != nil {
		t.Fatalf("Failed to serialize and hash Arithmetization: %v", err)
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

	// Serialize and compute hash for deserialized Arithmetization
	_, deserializedHash, err := serializeAndHash(deserialized)
	if err != nil {
		t.Fatalf("Failed to serialize and hash deserialized Arithmetization: %v", err)
	}

	// Compare hashes
	if !reflect.DeepEqual(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
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

	// Serialize and compute hash for original Keccak
	keccakSer, originalHash, err := serializeAndHash(z.Keccak)
	if err != nil {
		t.Fatalf("Failed to serialize and hash Keccak: %v", err)
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

	// Serialize and compute hash for deserialized Keccak
	_, deserializedHash, err := serializeAndHash(deserialized)
	if err != nil {
		t.Fatalf("Failed to serialize and hash deserialized Keccak: %v", err)
	}

	// Compare hashes
	if !reflect.DeepEqual(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
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

	// Serialize and compute hash for original Sha2
	sha2Ser, originalHash, err := serializeAndHash(z.Sha2)
	if err != nil {
		t.Fatalf("Failed to serialize and hash Sha2: %v", err)
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

	// Serialize and compute hash for deserialized Sha2
	_, deserializedHash, err := serializeAndHash(deserialized)
	if err != nil {
		t.Fatalf("Failed to serialize and hash deserialized Sha2: %v", err)
	}

	// Compare hashes
	if !reflect.DeepEqual(originalHash, deserializedHash) {
		t.Errorf("Hashes do not match:\nOriginal: %x\nDeserialized: %x", originalHash, deserializedHash)
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

	// Serialize and compute hash for original CompiledIOP
	compiledIOPSer, err := serialization.SerializeCompiledIOP(z.WizardIOP)
	if err != nil {
		t.Fatalf("Failed to serialize and hash CompiledIOP: %v", err)
	}

	// Deserialize into a new CompiledIOP
	deserializedIOP, err := serialization.DeserializeCompiledIOP(compiledIOPSer)
	if err != nil {
		t.Fatalf("Failed to deserialize CompiledIOP: %v", err)
	}

	// Compare hashes
	if !reflect.DeepEqual(z.WizardIOP, deserializedIOP) {

	}
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
