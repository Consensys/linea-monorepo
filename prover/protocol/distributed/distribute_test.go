package experiment

import (
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

const (
	SizeXs = 1 << (14 + 2*iota)
	SizeS
	SizeM
	SizeL
	SizeXL
	SizeXXL
)

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *zkevm.ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := config.TracesLimits{
		Add:                                  SizeXL,
		Bin:                                  SizeXL,
		Blake2Fmodexpdata:                    SizeM,
		Blockdata:                            SizeXs,
		Blockhash:                            SizeXs,
		Ecdata:                               SizeL,
		Euc:                                  SizeM,
		Exp:                                  SizeM,
		Ext:                                  SizeXXL,
		Gas:                                  SizeM,
		Hub:                                  SizeXXL,
		Logdata:                              SizeL,
		Loginfo:                              SizeS,
		Mmio:                                 SizeXXL,
		Mmu:                                  SizeXXL,
		Mod:                                  SizeL,
		Mul:                                  SizeM,
		Mxp:                                  SizeL,
		Oob:                                  SizeL,
		Rlpaddr:                              SizeS,
		Rlptxn:                               SizeL,
		Rlptxrcpt:                            SizeM,
		Rom:                                  SizeL,
		Romlex:                               SizeXs,
		Shakiradata:                          SizeM,
		Shf:                                  SizeL,
		Stp:                                  SizeM,
		Trm:                                  SizeM,
		Txndata:                              SizeS,
		Wcp:                                  SizeL,
		Binreftable:                          262144,
		Shfreftable:                          4096,
		Instdecoder:                          512,
		PrecompileEcrecoverEffectiveCalls:    SizeXs,
		PrecompileSha2Blocks:                 SizeS,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       SizeXs,
		PrecompileEcaddEffectiveCalls:        SizeS,
		PrecompileEcmulEffectiveCalls:        SizeXs,
		PrecompileEcpairingEffectiveCalls:    SizeXs,
		PrecompileEcpairingMillerLoops:       SizeXs,
		PrecompileEcpairingG2MembershipCalls: SizeXs,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          SizeM,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    SizeXs,
		ShomeiMerkleProofs:                   SizeS,
	}

	return zkevm.FullZKEVMWithSuite(&traceLimits, []func(*wizard.CompiledIOP){})
}
