package execution

import (
	"math/big"
	"reflect"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	poseidon2permutation "github.com/consensys/gnark/std/permutation/poseidon2/gkr-poseidon2"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput"
)

// checkPublicInputs checks that the values in fi are consistent with the
// wizard.VerifierCircuit
func checkPublicInputs(
	koalaAPI *koalagnark.API,
	wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark,
	execData [1 << 17]frontend.Variable,
) {

	// Checking the state root hash concomittance
	checkStateRootHash(koalaAPI, wvc, gnarkFuncInp)

	// Checking the block number concomittance
	checkBlockNumber(koalaAPI, wvc, gnarkFuncInp)

	// Checking the block timestamp concomittance
	checkBlockTimestamp(koalaAPI, wvc, gnarkFuncInp)

	// Checking the rolling hash concomittance
	checkRollingHash(koalaAPI, wvc, gnarkFuncInp)

	// Checking the rolling hash number concomittance
	checkRollingHashNumber(koalaAPI, wvc, gnarkFuncInp)

	// Checking the concomittance of the dynamic chain config (L2MsgService,
	// BaseFee, CoinBase, ChainID)
	checkDynamicChainConfig(koalaAPI, wvc, gnarkFuncInp)

	// Checking the execution data
	checkExecutionData(koalaAPI, wvc, gnarkFuncInp, execData)

	// Checking the L2 Msg hash
	checkL2MSgHashes(koalaAPI, wvc, gnarkFuncInp)
}

// checkStateRootHash checks the concomittance of the state root hashes between
// the functional inputs and the public inputs extracted from the wizard circuit.
func checkStateRootHash(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialStateRootHashWords = getPublicInputArr(koalaAPI, wvc, pie.InitialStateRootHash[:])
		extrFinalStateRootHashWords   = getPublicInputArr(koalaAPI, wvc, pie.FinalStateRootHash[:])
	)

	combineKoala := func(api frontend.API, vs []frontend.Variable) frontend.Variable {
		p32 := big.NewInt(1)
		p32.Lsh(p32, 32)
		return compress.ReadNum(api, vs, p32)
	}

	api.AssertIsEqual(
		gnarkFuncInp.InitialStateRootHash[0],
		combineKoala(api, extrInitialStateRootHashWords[:4]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.InitialStateRootHash[1],
		combineKoala(api, extrInitialStateRootHashWords[4:]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalStateRootHash[0],
		combineKoala(api, extrFinalStateRootHashWords[:4]),
	)

	api.AssertIsEqual(
		gnarkFuncInp.FinalStateRootHash[1],
		combineKoala(api, extrFinalStateRootHashWords[4:]),
	)
}

// checkBlockNumber checks the concomittance of the block number between the
// functional inputs and the public inputs extracted from the wizard circuit.
func checkBlockNumber(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialBlockNumberWords = getPublicInputArr(koalaAPI, wvc, pie.InitialBlockNumber[:])
		extrFinalBlockNumberWords   = getPublicInputArr(koalaAPI, wvc, pie.FinalBlockNumber[:])
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
func checkBlockTimestamp(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialTimestampWords = getPublicInputArr(koalaAPI, wvc, pie.InitialBlockTimestamp[:])
		extrFinalTimestampWords   = getPublicInputArr(koalaAPI, wvc, pie.FinalBlockTimestamp[:])
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
func checkRollingHash(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialRollingHashWordsHi = getPublicInputArr(koalaAPI, wvc, pie.FirstRollingHashUpdate[:8])
		extrInitialRollingHashWordsLo = getPublicInputArr(koalaAPI, wvc, pie.FirstRollingHashUpdate[8:])
		extrFinalRollingHashWordsHi   = getPublicInputArr(koalaAPI, wvc, pie.LastRollingHashUpdate[:8])
		extrFinalRollingHashWordsLo   = getPublicInputArr(koalaAPI, wvc, pie.LastRollingHashUpdate[8:])

		funcInpInitialRollingHashWordsHi = internal.CombineByteIntoWords(api, gnarkFuncInp.InitialRollingHashUpdate[:16])
		funcInpInitialRollingHashWordsLo = internal.CombineByteIntoWords(api, gnarkFuncInp.InitialRollingHashUpdate[16:])
		funcInpFinalRollingHashWordsHi   = internal.CombineByteIntoWords(api, gnarkFuncInp.FinalRollingHashUpdate[:16])
		funcInpFinalRollingHashWordsLo   = internal.CombineByteIntoWords(api, gnarkFuncInp.FinalRollingHashUpdate[16:])
	)

	for i := range extrInitialRollingHashWordsHi {
		api.AssertIsEqual(funcInpInitialRollingHashWordsHi[i], extrInitialRollingHashWordsHi[i])
		api.AssertIsEqual(funcInpInitialRollingHashWordsLo[i], extrInitialRollingHashWordsLo[i])
		api.AssertIsEqual(funcInpFinalRollingHashWordsHi[i], extrFinalRollingHashWordsHi[i])
		api.AssertIsEqual(funcInpFinalRollingHashWordsLo[i], extrFinalRollingHashWordsLo[i])
	}
}

// checkRollingHashNumber checks the concomittance of the rolling hash number
// between the functional inputs and the public inputs extracted from the wizard
// circuit.
func checkRollingHashNumber(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrInitialRollingHashNumberWords = getPublicInputArr(koalaAPI, wvc, pie.FirstRollingHashUpdateNumber[:])
		extrFinalRollingHashNumberWords   = getPublicInputArr(koalaAPI, wvc, pie.LastRollingHashUpdateNumber[:])
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
func checkDynamicChainConfig(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	var (
		pie = getPublicInputExtractor(wvc)

		extrChainIDLimbs    = getPublicInputArr(koalaAPI, wvc, pie.ChainID[:])
		extrBaseFeeLimbs    = getPublicInputArr(koalaAPI, wvc, pie.BaseFee[:])
		extrCoinBaseLimbs   = getPublicInputArr(koalaAPI, wvc, pie.CoinBase[:])
		extrMsgServiceLimbs = getPublicInputArr(koalaAPI, wvc, pie.L2MessageServiceAddr[:])

		chainID    = internal.CombineWordsIntoElements(api, extrChainIDLimbs)
		baseFee    = internal.CombineWordsIntoElements(api, extrBaseFeeLimbs)
		coinBase   = internal.CombineWordsIntoElements(api, extrCoinBaseLimbs)
		msgService = internal.CombineWordsIntoElements(api, extrMsgServiceLimbs)
	)

	mustBeEqualIfExtractedNonZero := func(fn, ex frontend.Variable) {
		api.AssertIsEqual(api.Mul(ex, api.Sub(ex, fn)), 0)
	}

	mustBeEqualIfExtractedNonZero(gnarkFuncInp.ChainID, chainID)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.BaseFee, baseFee)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.CoinBase, coinBase)
	mustBeEqualIfExtractedNonZero(gnarkFuncInp.L2MessageServiceAddr, msgService)
}

// checkExecutionData computes the BLS execution data hash and checks it is
// consistent with the public input extracted from the wizard circuit using the
// multilateral commitment.
func checkExecutionData(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit,
	gnarkFuncInp FunctionalPublicInputSnark, execData [1 << 17]frontend.Variable,
) {
	api := koalaAPI.Frontend()

	hsh, err := poseidon2permutation.NewCompressor(api)
	if err != nil {
		panic(err)
	}

	var (
		pie = getPublicInputExtractor(wvc)

		extrSZX       = getPublicInputExt(koalaAPI, wvc, pie.DataSZX)
		extrSZY       = getPublicInputExt(koalaAPI, wvc, pie.DataSZY)
		extrKoalaHash = getPublicInputArr(koalaAPI, wvc, pie.DataChecksum[:])
		extrDataNByte = getPublicInput(koalaAPI, wvc, pie.DataNbBytes)
	)

	// @alex: in theory we could simplify a little bit the code by just not
	// asking the user to provider execDataNByte, but doing it this way allows
	// easily diagnosing if there is a mismatching between what is extracted
	// from the inner-proof and what is provided by the user.
	api.AssertIsEqual(extrDataNByte, gnarkFuncInp.DataChecksum.Length)

	recoveredX, recoveredY, hashBLS := public_input.CheckExecDataMultiCommitmentOpeningGnark(
		api, execData, extrDataNByte, [8]frontend.Variable(extrKoalaHash), hsh,
	)

	for i := range extrSZX {
		api.AssertIsEqual(extrSZX[i], recoveredX[i])
		api.AssertIsEqual(extrSZY[i], recoveredY[i])
	}

	api.AssertIsEqual(gnarkFuncInp.DataChecksum.PartialHash, hashBLS)

	if err := gnarkFuncInp.DataChecksum.Check(api); err != nil {
		panic(err)
	}
}

// checkL2MSgHashes checks the concomittance of the L2 message hashes extracted
// from the wvc to their purported BLS hash held in the gnarkFuncInp.
func checkL2MSgHashes(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, gnarkFuncInp FunctionalPublicInputSnark) {
	api := koalaAPI.Frontend()

	pie := getPublicInputExtractor(wvc)

	if len(pie.L2Messages) != len(gnarkFuncInp.L2MessageHashes.Values) {
		utils.Panic("L2MessageHashes length mismatch: %d != %d", len(pie.L2Messages), len(gnarkFuncInp.L2MessageHashes.Values))
	}

	// This converts the provided L2MsgHash (in 8-bits words) into 16-bytes and
	// then directly compare with the public input extracted from the circuit.

	for i := range gnarkFuncInp.L2MessageHashes.Values {

		var (
			funcL2MessageHashBytes = gnarkFuncInp.L2MessageHashes.Values[i]
			funcL2MessageHashWords = internal.CombineByteIntoWords(api, funcL2MessageHashBytes[:])
			extrL2MessageHashWords = getPublicInputArr(koalaAPI, wvc, pie.L2Messages[i][:])
		)

		if len(funcL2MessageHashWords) != len(extrL2MessageHashWords) {
			utils.Panic("L2MessageHashes[%d] length mismatch: %d != %d", i, len(funcL2MessageHashWords), len(extrL2MessageHashWords))
		}

		for j := range funcL2MessageHashWords {
			api.AssertIsEqual(funcL2MessageHashWords[j], extrL2MessageHashWords[j])
		}
	}
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
func getPublicInputExt(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, pi wizard.PublicInput) [4]frontend.Variable {
	// this prefixing is needed because the full-prover circuit is being recursed
	res := wvc.GetPublicInputExt(koalaAPI, pi.Name)
	return [4]frontend.Variable{
		res.B0.A0.Native(),
		res.B0.A1.Native(),
		res.B1.A0.Native(),
		res.B1.A1.Native(),
	}
}

// getPublicInputArr returns a slice of values from the public input
func getPublicInputArr(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, pis []wizard.PublicInput) []frontend.Variable {
	res := make([]frontend.Variable, len(pis))
	for i := range pis {
		r := wvc.GetPublicInput(koalaAPI, pis[i].Name)
		res[i] = r.Native()
	}
	return res
}

// getPublicInput returns a value from the public input
func getPublicInput(koalaAPI *koalagnark.API, wvc *wizard.VerifierCircuit, pi wizard.PublicInput) frontend.Variable {
	r := wvc.GetPublicInput(koalaAPI, pi.Name)
	return r.Native()
}
