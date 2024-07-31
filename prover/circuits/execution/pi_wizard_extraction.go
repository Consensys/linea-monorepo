package execution

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput"
)

// checkPublicInputs checks that the values in fi are consistent with the
// wizard.VerifierCircuit
func checkPublicInputs(
	api frontend.API,
	wvc *wizard.WizardVerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark,
	wizardFuncInp publicInput.FunctionalInputExtractor,
) {

	var (
		finalRollingHash   = internal.CombineBytesIntoElements(api, gnarkFuncInp.FinalRollingHash)
		initialRollingHash = internal.CombineBytesIntoElements(api, gnarkFuncInp.InitialRollingHash)
	)

	hsh, err := mimc.NewMiMC(api)
	if err != nil {
		panic(err)
	}

	hsh.Write(wvc.GetLocalPointEvalParams(wizardFuncInp.DataNbBytes.ID).Y, wvc.GetLocalPointEvalParams(wizardFuncInp.DataChecksum.ID).Y)
	api.AssertIsEqual(hsh.Sum(), gnarkFuncInp.DataChecksum)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.L2MessageHash.ID).Y,
		// TODO: this operation is done a second time when computing the final
		// public input which is wasteful although not dramatic (~8000 unused
		// constraints)
		gnarkFuncInp.L2MessageHashes.Checksum(api),
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialStateRootHash.ID).Y,
		gnarkFuncInp.InitialStateRootHash,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialBlockNumber.ID).Y,
		gnarkFuncInp.InitialBlockNumber,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialBlockTimestamp.ID).Y,
		gnarkFuncInp.InitialBlockTimestamp,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialRollingHash[0].ID).Y,
		initialRollingHash[0],
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialRollingHash[1].ID).Y,
		initialRollingHash[1],
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.InitialRollingHashNumber.ID).Y,
		gnarkFuncInp.InitialRollingHashNumber,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalStateRootHash.ID).Y,
		gnarkFuncInp.FinalStateRootHash,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalBlockNumber.ID).Y,
		gnarkFuncInp.FinalBlockNumber,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalBlockTimestamp.ID).Y,
		gnarkFuncInp.FinalBlockTimestamp,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalRollingHash[0].ID).Y,
		finalRollingHash[0],
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalRollingHash[1].ID).Y,
		finalRollingHash[1],
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.FinalRollingHashNumber.ID).Y,
		gnarkFuncInp.FinalRollingHashNumber,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.ChainID.ID).Y,
		gnarkFuncInp.ChainID,
	)

	api.AssertIsEqual(
		wvc.GetLocalPointEvalParams(wizardFuncInp.L2MessageServiceAddr.ID).Y,
		gnarkFuncInp.L2MessageServiceAddr,
	)

}
