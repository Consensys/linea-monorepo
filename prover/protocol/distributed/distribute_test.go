package experiment

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// TestDistribute attempts to run and compile the distributed protocol.
func TestDistribute(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
		}
		_ = Distribute(zkevm.WizardIOP, disc)
	)

}

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *zkevm.ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := config.TracesLimits{
		Add:                                  1 << 17,
		Bin:                                  1 << 16,
		Blake2Fmodexpdata:                    1 << 12,
		Blockdata:                            1 << 10,
		Blockhash:                            1 << 10,
		Ecdata:                               1 << 16,
		Euc:                                  1 << 14,
		Exp:                                  1 << 12,
		Ext:                                  1 << 18,
		Gas:                                  1 << 14,
		Hub:                                  1 << 19,
		Logdata:                              1 << 14,
		Loginfo:                              1 << 10,
		Mmio:                                 1 << 19,
		Mmu:                                  1 << 19,
		Mod:                                  1 << 15,
		Mul:                                  1 << 14,
		Mxp:                                  1 << 17,
		Oob:                                  1 << 16,
		Rlpaddr:                              1 << 10,
		Rlptxn:                               1 << 15,
		Rlptxrcpt:                            1 << 15,
		Rom:                                  1 << 20,
		Romlex:                               1 << 10,
		Shakiradata:                          1 << 13,
		Shf:                                  1 << 14,
		Stp:                                  1 << 12,
		Trm:                                  1 << 13,
		Txndata:                              1 << 12,
		Wcp:                                  1 << 16,
		Binreftable:                          1 << 18,
		Shfreftable:                          4096,
		Instdecoder:                          512,
		PrecompileEcrecoverEffectiveCalls:    500,
		PrecompileSha2Blocks:                 600,
		PrecompileRipemdBlocks:               0,
		PrecompileModexpEffectiveCalls:       64,
		PrecompileEcaddEffectiveCalls:        1 << 14,
		PrecompileEcmulEffectiveCalls:        32,
		PrecompileEcpairingEffectiveCalls:    32,
		PrecompileEcpairingMillerLoops:       64,
		PrecompileEcpairingG2MembershipCalls: 64,
		PrecompileBlakeEffectiveCalls:        0,
		PrecompileBlakeRounds:                0,
		BlockKeccak:                          1 << 13,
		BlockL1Size:                          100_000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    400,
		ShomeiMerkleProofs:                   1 << 14,
	}

	return zkevm.FullZKEVMWithSuite(&traceLimits, []func(*wizard.CompiledIOP){})
}
