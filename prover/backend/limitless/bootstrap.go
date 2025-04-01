package limitless

import (
	"errors"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

type Bootstrapper struct {
	disc       *distributed.StandardModuleDiscoverer
	zkevm      *zkevm.ZkEvm
	distWizard *distributed.DistributedWizard
}

func NewBootstrapper(cfg *config.Config, targetWeight int) (*Bootstrapper, error) {
	// Initialize module discoverer
	disc := &distributed.StandardModuleDiscoverer{
		TargetWeight: targetWeight,
	}

	// Get zkEVM instance
	zkevmInstance := GetZkEVM()

	// Distribute the wizard protocol
	distWizard := distributed.DistributeWizard(zkevmInstance.WizardIOP, disc)

	// Apply production compilers to the bootstrapper (precompilation is already done in DistributeWizard)
	distWizard.CompileModules(
		mimc.CompileMiMC,
		logderivativesum.LookupIntoLogDerivativeSumWithSegmenter(disc),
		permutation.CompileIntoGdProduct,
		horner.ProjectionToHorner,
	)

	return &Bootstrapper{
		disc:       disc,
		zkevm:      zkevmInstance,
		distWizard: &distWizard,
	}, nil
}

func (b *Bootstrapper) GenerateBootstrapperProof(req *execution.Request) (*wizard.Proof, error) {
	// Generate zkEVM witness from request and config
	witness := GetZkevmWitness(req, nil)
	if witness == nil {
		return nil, errors.New("failed to generate zkEVM witness")
	}

	// Run the prover on the bootstrapper
	runtimeBoot := wizard.RunProver(b.distWizard.Bootstrapper, b.zkevm.GetMainProverStep(witness))
	if runtimeBoot == nil {
		return nil, errors.New("bootstrapper prover failed")
	}

	// Extract and return the proof
	proof := runtimeBoot.ExtractProof()
	return &proof, nil
}

func (b *Bootstrapper) VerifyBootstrapperProof(proof wizard.Proof) error {
	if err := wizard.Verify(b.distWizard.Bootstrapper, proof); err != nil {
		return errors.New("proof verification failed:" + err.Error())
	}
	return nil
}

// GetZkevmWitness returns a [zkevm.Witness]
func GetZkevmWitness(req *execution.Request, cfg *config.Config) *zkevm.Witness {
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	return witness.ZkEVM
}

// GetZKEVM returns a [zkevm.ZkEvm] with its trace limits inflated so that it
// can be used as input for the package functions. The zkevm is returned
// without any compilation.
func GetZkEVM() *zkevm.ZkEvm {

	// This are the config trace-limits from sepolia. All multiplied by 16.
	traceLimits := config.TracesLimits{
		Add:                                  1 << 19,
		Bin:                                  1 << 18,
		Blake2Fmodexpdata:                    1 << 14,
		Blockdata:                            1 << 12,
		Blockhash:                            1 << 12,
		Ecdata:                               1 << 18,
		Euc:                                  1 << 16,
		Exp:                                  1 << 14,
		Ext:                                  1 << 20,
		Gas:                                  1 << 16,
		Hub:                                  1 << 21,
		Logdata:                              1 << 16,
		Loginfo:                              1 << 12,
		Mmio:                                 1 << 21,
		Mmu:                                  1 << 21,
		Mod:                                  1 << 17,
		Mul:                                  1 << 16,
		Mxp:                                  1 << 19,
		Oob:                                  1 << 18,
		Rlpaddr:                              1 << 12,
		Rlptxn:                               1 << 17,
		Rlptxrcpt:                            1 << 17,
		Rom:                                  1 << 22,
		Romlex:                               1 << 12,
		Shakiradata:                          1 << 15,
		Shf:                                  1 << 16,
		Stp:                                  1 << 14,
		Trm:                                  1 << 15,
		Txndata:                              1 << 14,
		Wcp:                                  1 << 18,
		Binreftable:                          1 << 20,
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
	return zkevm.FullZKEVMWithSuite(&traceLimits, []func(*wizard.CompiledIOP){}, &config.Config{})
}
