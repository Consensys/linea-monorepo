package execution

import (
	"reflect"

	"github.com/consensys/gnark/frontend"
	poseidon2permutation "github.com/consensys/gnark/std/permutation/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
)

// checkPublicInputs checks that the values in fi are consistent with the
// wizard.VerifierCircuit
func checkPublicInputs(
	api frontend.API,
	wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark,
	execData [1 << 17]frontend.Variable,
	execDataNByte frontend.Variable,
) {

	// Checking the state root hash concomittance
	checkStateRootHash(api, wvc, gnarkFuncInp)

	// Checking the block number concomittance
	checkBlockNumber(api, wvc, gnarkFuncInp)

	// Checking the block timestamp concomittance
	checkBlockTimestamp(api, wvc, gnarkFuncInp)

	// Checking the rolling hash concomittance
	checkRollingHash(api, wvc, gnarkFuncInp)

	// Checking the rolling hash number concomittance
	checkRollingHashNumber(api, wvc, gnarkFuncInp)

	// Checking the concomittance of the dynamic chain config (L2MsgService,
	// BaseFee, CoinBase, ChainID)
	checkDynamicChainConfig(api, wvc, gnarkFuncInp)

	// Checking the execution data
	checkExecutionData(api, wvc, gnarkFuncInp, execData, execDataNByte)
}

// checkStateRootHash checks the concomittance of the state root hashes between
// the functional inputs and the public inputs extracted from the wizard circuit.
func checkStateRootHash(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialStateRootHashWords = getPublicInputArr(api, wvc, pie.InitialStateRootHash[:])
		extrFinalStateRootHashWords   = getPublicInputArr(api, wvc, pie.FinalStateRootHash[:])
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialStateRootHash[0],
		internal.CombineWordsIntoElements(api, extrInitialStateRootHashWords[:8]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialStateRootHash[1],
		internal.CombineWordsIntoElements(api, extrInitialStateRootHashWords[8:]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalStateRootHash[0],
		internal.CombineWordsIntoElements(api, extrFinalStateRootHashWords[:8]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalStateRootHash[1],
		internal.CombineWordsIntoElements(api, extrFinalStateRootHashWords[8:]),
	)
}

// checkBlockNumber checks the concomittance of the block number between the
// functional inputs and the public inputs extracted from the wizard circuit.
func checkBlockNumber(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialBlockNumberWords = getPublicInputArr(api, wvc, pie.InitialBlockNumber[:])
		extrFinalBlockNumberWords   = getPublicInputArr(api, wvc, pie.FinalBlockNumber[:])
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialBlockNumber,
		internal.CombineWordsIntoElements(api, extrInitialBlockNumberWords),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalBlockNumber,
		internal.CombineWordsIntoElements(api, extrFinalBlockNumberWords),
	)
}

// checkBlockTimestamp checks the concomittance of the block timestamp between the
// functional inputs and the public inputs extracted from the wizard circuit.
func checkBlockTimestamp(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialTimestampWords = getPublicInputArr(api, wvc, pie.InitialBlockTimestamp[:])
		extrFinalTimestampWords   = getPublicInputArr(api, wvc, pie.FinalBlockTimestamp[:])
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialBlockTimestamp,
		internal.CombineWordsIntoElements(api, extrInitialTimestampWords),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalBlockTimestamp,
		internal.CombineWordsIntoElements(api, extrFinalTimestampWords),
	)
}

// checkRollingHash checks the concomittance of the rolling hash between the
// functional inputs and the public inputs extracted from the wizard circuit.
func checkRollingHash(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialRollingHashWordsHi = getPublicInputArr(api, wvc, pie.FirstRollingHashUpdate[:8])
		extrInitialRollingHashWordsLo = getPublicInputArr(api, wvc, pie.FirstRollingHashUpdate[8:])
		extrFinalRollingHashWordsHi   = getPublicInputArr(api, wvc, pie.LastRollingHashUpdate[:8])
		extrFinalRollingHashWordsLo   = getPublicInputArr(api, wvc, pie.LastRollingHashUpdate[8:])
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialRollingHashUpdate[0],
		internal.CombineWordsIntoElements(api, extrInitialRollingHashWordsHi),
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialRollingHashUpdate[1],
		internal.CombineWordsIntoElements(api, extrInitialRollingHashWordsLo),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalRollingHashUpdate[0],
		internal.CombineWordsIntoElements(api, extrFinalRollingHashWordsHi),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalRollingHashUpdate[1],
		internal.CombineWordsIntoElements(api, extrFinalRollingHashWordsLo),
	)
}

// checkRollingHashNumber checks the concomittance of the rolling hash number
// between the functional inputs and the public inputs extracted from the wizard
// circuit.
func checkRollingHashNumber(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialRollingHashNumberWords = getPublicInputArr(api, wvc, pie.FirstRollingHashUpdateNumber[:])
		extrFinalRollingHashNumberWords   = getPublicInputArr(api, wvc, pie.LastRollingHashUpdateNumber[:])
	)

	api.AssertIsEqual(
		gnarkFuncInp.FirstRollingHashUpdateNumber,
		internal.CombineWordsIntoElements(api, extrInitialRollingHashNumberWords),
	)

	api.AssertIsEqual(
		gnarkFuncInp.LastRollingHashUpdateNumber,
		internal.CombineWordsIntoElements(api, extrFinalRollingHashNumberWords),
	)
}

// checkDynamicChainConfig checks the concomittance of the dynamic chain config
// between the functional inputs and the public inputs extracted from the wizard
// circuit.
func checkDynamicChainConfig(api frontend.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {

	var (
		pie = getPublicInputExtractor(wvc)

		extrChainIDLimbs    = getPublicInputArr(api, wvc, pie.ChainID[:])
		extrBaseFeeLimbs    = getPublicInputArr(api, wvc, pie.BaseFee[:])
		extrCoinBaseLimbs   = getPublicInputArr(api, wvc, pie.CoinBase[:])
		extrMsgServiceLimbs = getPublicInputArr(api, wvc, pie.L2MessageServiceAddr[:])

		chainID    = internal.CombineWordsIntoElements(api, extrChainIDLimbs)
		baseFee    = internal.CombineWordsIntoElements(api, extrBaseFeeLimbs)
		coinBase   = internal.CombineWordsIntoElements(api, extrCoinBaseLimbs)
		msgService = internal.CombineWordsIntoElements(api, extrMsgServiceLimbs)
	)

	mustBeEqualIfExtractedNonZero := func(fn, ex frontend.Variable) {
		api.AssertIsEqual(api.Mul(fn, api.Sub(ex, fn)), 0)
	}

	mustBeEqualIfExtractedNonZero(gnarkFuncInp.ChainID, chainID)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.BaseFee, baseFee)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.CoinBase, coinBase)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.L2MessageServiceAddr, msgService)
}

// checkExecutionData computes the BLS execution data hash and checks it is
// consistent with the public input extracted from the wizard circuit using the
// multilateral commitment.
func checkExecutionData(api frontend.API, wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark, execData [1 << 17]frontend.Variable,
	execDataNByte frontend.Variable,
) {

	hsh, err := poseidon2permutation.NewCompressor(api)
	if err != nil {
		panic(err)
	}

	var (
		pie = getPublicInputExtractor(wvc)

		extrSZX       = getPublicInputExt(api, wvc, pie.DataSZX)
		extrSZY       = getPublicInputExt(api, wvc, pie.DataSZY)
		extrKoalaHash = getPublicInputArr(api, wvc, pie.DataChecksum[:])
		extrDataNByte = getPublicInput(api, wvc, pie.DataNbBytes)
	)

	// @alex: in theory we could simplify a little bit the code by just not
	// asking the user to provider execDataNByte, but doing it this way allows
	// easily diagnosing if there is a mismatching between what is extracted
	// from the inner-proof and what is provided by the user.
	api.AssertIsEqual(extrDataNByte, execDataNByte)

	recoveredX, recoveredY, hashBLS := public_input.CheckExecDataMultiCommitmentOpeningGnark(
		api, execData, execDataNByte, [8]frontend.Variable(extrKoalaHash), hsh,
	)

	for i := range extrSZX {
		api.AssertIsEqual(extrSZX[i], recoveredX[i])
		api.AssertIsEqual(extrSZY[i], recoveredY[i])
	}

	api.AssertIsEqual(gnarkFuncInp.DataChecksum, hashBLS)
}

// getPublicInputExtractor extracts the public input from the wizard circuit
func getPublicInputExtractor(wvc *wizard.VerifierCircuit) *publicInput.FunctionalInputExtractor {
	extraData, extraDataFound := wvc.Spec.ExtraData[publicInput.PublicInputExtractorMetadata]
	if !extraDataFound {
		panic("public input extractor not found")
	}
	pie, ok := extraData.(*publicInput.FunctionalInputExtractor)
	if !ok {
		panic("public input extractor not of the right type: " + reflect.TypeOf(extraData).String())
	}
	return pie
}

// getPublicInputExt returns a field extension public input coordinates in array
// form
func getPublicInputExt(api frontend.API, wvc *wizard.VerifierCircuit, pi wizard.PublicInput) [4]frontend.Variable {
	// this prefixing is needed because the full-prover circuit is being recursed
	res := wvc.GetPublicInputExt(api, pi.Name)
	return [4]frontend.Variable{
		res.B0.A0.Native(),
		res.B0.A1.Native(),
		res.B1.A0.Native(),
		res.B1.A1.Native(),
	}
}

// getPublicInputArr returns a slice of values from the public input
func getPublicInputArr(api frontend.API, wvc *wizard.VerifierCircuit, pis []wizard.PublicInput) []frontend.Variable {
	res := make([]frontend.Variable, len(pis))
	for i := range pis {
		r := pis[i].Acc.GetFrontendVariable(api, wvc)
		res[i] = r.Native()
	}
	return res
}

// getPublicInput returns a value from the public input
func getPublicInput(api frontend.API, wvc *wizard.VerifierCircuit, pi wizard.PublicInput) frontend.Variable {
	r := pi.Acc.GetFrontendVariable(api, wvc)
	return r.Native()
}

func assertAreEqual(api frontend.API, wvc *wizard.VerifierCircuit, funcValue []frontend.Variable, extractedValue []wizard.PublicInput) {
	if len(funcValue) != len(extractedValue) {
		panic("mismatched lengths")
	}
	for i := range funcValue {
		api.AssertIsEqual(funcValue[i], extractedValue[i].Acc.GetFrontendVariable(api, wvc))
	}
}
