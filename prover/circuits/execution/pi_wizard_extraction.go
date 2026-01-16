package execution

import (
	"github.com/consensys/gnark/frontend"
	gkrposeidon2 "github.com/consensys/gnark/std/hash/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
)

// checkPublicInputs checks that the values in fi are consistent with the
// wizard.VerifierCircuit
func checkPublicInputs(
	api frontend.API,
	wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark,
) {

	var (
		lastRollingHash  = internal.CombineBytesIntoElements(api, gnarkFuncInp.FinalRollingHashUpdate)
		firstRollingHash = internal.CombineBytesIntoElements(api, gnarkFuncInp.InitialRollingHashUpdate)
		execDataHash     = execDataHash(api, wvc)

		_ = firstRollingHash // to make the compiler happy
		_ = lastRollingHash  // to make the compiler happy
	)

	// As we have this issue, the execDataHash will not match what we have in the
	// functional input (the txnrlp is incorrect). It should be converted into
	// an [api.AssertIsEqual] once this is resolved.
	//
	api.AssertIsEqual(execDataHash, gnarkFuncInp.DataChecksum)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.L2MessageHash),
		// TODO: this operation is done a second time when computing the final
		// public input which is wasteful although not dramatic (~8000 unused
		// constraints)
		gnarkFuncInp.L2MessageHashes.CheckSumPoseidon2(api),
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.InitialStateRootHash),
		gnarkFuncInp.InitialStateRootHash,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.InitialBlockNumber),
		gnarkFuncInp.InitialBlockNumber,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.InitialBlockTimestamp),
		gnarkFuncInp.InitialBlockTimestamp,
	)

	panic("fix the exposition of the rolling hash updates. It should be accessible as an array of 16 limbs elements")

	// api.AssertIsEqual(
	// 	wvc.GetPublicInput(api, publicInput.FirstRollingHashUpdate_0),
	// 	firstRollingHash[0],
	// )

	// api.AssertIsEqual(
	// 	wvc.GetPublicInput(api, publicInput.FirstRollingHashUpdate_1),
	// 	firstRollingHash[1],
	// )

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.FirstRollingHashUpdateNumber),
		gnarkFuncInp.FirstRollingHashUpdateNumber,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.FinalStateRootHash),
		gnarkFuncInp.FinalStateRootHash,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.FinalBlockNumber),
		gnarkFuncInp.FinalBlockNumber,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.FinalBlockTimestamp),
		gnarkFuncInp.FinalBlockTimestamp,
	)

	panic("uncomment the code")

	// api.AssertIsEqual(
	// 	wvc.GetPublicInput(api, publicInput.LastRollingHashUpdate_0),
	// 	lastRollingHash[0],
	// )

	// api.AssertIsEqual(
	// 	wvc.GetPublicInput(api, publicInput.LastRollingHashUpdate_1),
	// 	lastRollingHash[1],
	// )

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.LastRollingHashNumberUpdate),
		gnarkFuncInp.LastRollingHashUpdateNumber,
	)

	panic("limb split the L2MessageServiceAddr")

	// var (
	// 	twoPow128     = new(big.Int).SetInt64(1)
	// 	twoPow112     = new(big.Int).SetInt64(1)
	// 	_             = twoPow128.Lsh(twoPow128, 128)
	// 	_             = twoPow112.Lsh(twoPow112, 112)
	// 	bridgeAddress = api.Add(
	// 		api.Mul(
	// 			twoPow128,
	// 			wvc.GetPublicInput(api, publicInput.L2MessageServiceAddrHi),
	// 		),
	// 		wvc.GetPublicInput(api, publicInput.L2MessageServiceAddrLo),
	// 	)
	// )

	// // In principle, we should enforce a strict equality between the purported
	// // chainID and the one extracted from the traces. But in case, the executed
	// // block has only legacy transactions (e.g. transactions without a specified
	// // chainID) then the traces will return a chainID of zero.
	// api.AssertIsEqual(
	// 	api.Mul(
	// 		wvc.GetPublicInput(api, publicInput.ChainID),
	// 		api.Sub(
	// 			api.Div(
	// 				wvc.GetPublicInput(api, publicInput.ChainID),
	// 				twoPow112,
	// 			),
	// 			gnarkFuncInp.ChainID,
	// 		),
	// 	),
	// 	0,
	// )

	// api.AssertIsEqual(bridgeAddress, gnarkFuncInp.L2MessageServiceAddr)

	// To do: @gusiri
	// This will need an update (as for the whole file as the inputs are broken down in limbs now)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.BaseFee),
		gnarkFuncInp.BaseFee,
	)

	api.AssertIsEqual(
		wvc.GetPublicInput(api, publicInput.CoinBase),
		gnarkFuncInp.CoinBase,
	)

}

// execDataHash hash the execution-data with its length so that we can guard
// against padding attack (although the padding attacks are not possible to
// being with due to the encoding of the plaintext)
func execDataHash(
	api frontend.API,
	wvc *wizard.VerifierCircuit,
) frontend.Variable {

	hsh, err := gkrposeidon2.New(api)
	if err != nil {
		panic(err)
	}

	hsh.Write(
		wvc.GetPublicInput(api, publicInput.DataNbBytes),
		wvc.GetPublicInput(api, publicInput.DataChecksum),
	)

	return hsh.Sum()
}
