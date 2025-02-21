package experiment

import (
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// GetZkEVM returns the zkEVM compiled-IOP without any compilation. It
// uses the sepolia-config.
func GetZkEVM() *zkevm.ZkEvm {

	traceLimits := &config.TracesLimits{
		Add:                                  524288,
		Bin:                                  262144,
		Blake2Fmodexpdata:                    16384,
		Blockdata:                            4096,
		Blockhash:                            512,
		Ecdata:                               262144,
		Euc:                                  65536,
		Exp:                                  8192,
		Ext:                                  1048576,
		Gas:                                  65536,
		Hub:                                  1048576,
		Logdata:                              65536,
		Loginfo:                              4096,
		Mmio:                                 1048576,
		Mmu:                                  1048576,
		Mod:                                  131072,
		Mul:                                  65536,
		Mxp:                                  524288,
		Oob:                                  262144,
		Rlpaddr:                              4096,
		Rlptxn:                               131072,
		Rlptxrcpt:                            65536,
		Rom:                                  1048576,
		Romlex:                               1024,
		Shakiradata:                          32768,
		Shf:                                  65536,
		Stp:                                  16384,
		Trm:                                  32768,
		Txndata:                              8192,
		Wcp:                                  262144,
		PrecompileEcrecoverEffectiveCalls:    128,
		PrecompileSha2Blocks:                 671,
		PrecompileModexpEffectiveCalls:       4,
		PrecompileEcaddEffectiveCalls:        16384,
		PrecompileEcmulEffectiveCalls:        32,
		PrecompileEcpairingEffectiveCalls:    16,
		PrecompileEcpairingMillerLoops:       64,
		PrecompileEcpairingG2MembershipCalls: 64,
		BlockKeccak:                          8192,
		BlockL1Size:                          1000000,
		BlockL2L1Logs:                        16,
		BlockTransactions:                    200,
		Binreftable:                          262144,
		Shfreftable:                          4096,
		Instdecoder:                          512,
	}

	return zkevm.FullZKEVMWithSuite(traceLimits, zkevm.CompilationSuite{})
}
