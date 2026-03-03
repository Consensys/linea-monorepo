package execution

import (
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/config"
	mimc "github.com/consensys/linea-monorepo/prover/crypto/mimc_bls12377"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

type limitlessBuilder struct {
	congWIOP     *wizard.CompiledIOP
	traceLimits  *config.TracesLimits
	WizardAssets *zkevm.LimitlessZkEVM
}

func NewBuilderLimitless(congWIOP *wizard.CompiledIOP, traceLimits *config.TracesLimits) *limitlessBuilder {
	return &limitlessBuilder{congWIOP: congWIOP, traceLimits: traceLimits}
}

func (b *limitlessBuilder) Compile() (constraint.ConstraintSystem, error) {
	return makeCSLimitless(b), nil
}

// Makes the constraint system for the execution-limitless circuit
func makeCSLimitless(b *limitlessBuilder) constraint.ConstraintSystem {
	circuit := AllocateLimitless(b.congWIOP, b.traceLimits)

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}

// checkLimitlessConglomerationCompletion checks that the conglomeration proof
// is complete:
//
// - all the segments have been aggregated
// - the horner/log-derivative/grand-product cancelled out
// - the shared randomness is correctly computed
// - the general multiset cancels out
func (c *CircuitExecution) checkLimitlessConglomerationCompletion(api frontend.API) {

	wvc := c.WizardVerifier

	// This loops counts the number of modules by counting the number of public
	// inputs starting with [TargetNbSegmentPublicInputBase]
	numModule := 0
	for _, pub := range wvc.Spec.PublicInputs {
		// In principle, IsPrefix should be the right one but Contains has more
		// chances to stand the test of time if we decide to play with prefixes.
		if strings.Contains(pub.Name, distributed.TargetNbSegmentPublicInputBase) {
			numModule++
		}
	}

	// This checks that the GL and LPP segment counts match the target number of
	// segment for each module.
	var (
		target           = distributed.GetPublicInputListGnark(api, &wvc, distributed.TargetNbSegmentPublicInputBase, numModule)
		countGL          = distributed.GetPublicInputListGnark(api, &wvc, distributed.SegmentCountGLPublicInputBase, numModule)
		countLPP         = distributed.GetPublicInputListGnark(api, &wvc, distributed.SegmentCountLPPPublicInputBase, numModule)
		generalMSet      = distributed.GetPublicInputListGnark(api, &wvc, distributed.GeneralMultiSetPublicInputBase, mimc.MSetHashSize)
		sharedRandMSet   = distributed.GetPublicInputListGnark(api, &wvc, distributed.SharedRandomnessMultiSetPublicInputBase, mimc.MSetHashSize)
		initRandomness   = wvc.GetPublicInput(api, distributed.InitialRandomnessPublicInput)
		logDerivativeSum = wvc.GetPublicInput(api, distributed.LogDerivativeSumPublicInput)
		gdProduct        = wvc.GetPublicInput(api, distributed.GrandProductPublicInput)
		hornerSum        = wvc.GetPublicInput(api, distributed.HornerPublicInput)
		vk0              = wvc.GetPublicInput(api, distributed.VerifyingKeyPublicInput)
		vk1              = wvc.GetPublicInput(api, distributed.VerifyingKey2PublicInput)
		vkMerkleRoot     = wvc.GetPublicInput(api, distributed.VerifyingKeyMerkleRootPublicInput)
	)

	for module := 0; module < numModule; module++ {
		api.AssertIsEqual(target[module], countGL[module])
		api.AssertIsEqual(target[module], countLPP[module])
	}

	for k := range generalMSet {
		api.AssertIsEqual(generalMSet[k], 0)
	}

	api.AssertIsEqual(logDerivativeSum, 0)
	api.AssertIsEqual(hornerSum, 0)
	api.AssertIsEqual(gdProduct, 1)
	newRand := mimc.GnarkHashVec(api, sharedRandMSet[:])
	api.AssertIsEqual(initRandomness, newRand)
	api.AssertIsEqual(vk0, c.CongloVK[0])
	api.AssertIsEqual(vk1, c.CongloVK[1])
	api.AssertIsEqual(vkMerkleRoot, c.VKMerkleRoot)

}
