package invalidity_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type invalidityHashCircuit struct {
	Inputs invalidity.FunctionalPublicInputsGnark `gnark:",secret"`
	Hash   frontend.Variable                      `gnark:",public"`
}

func (c *invalidityHashCircuit) Define(api frontend.API) error {

	api.AssertIsEqual(c.Hash, c.Inputs.Sum(api))
	return nil
}

func TestInvalidityPublicInputHashConsistency(t *testing.T) {
	gnarkutil.RegisterHints()

	var ftxRollingHash [32]byte
	for i := range ftxRollingHash {
		ftxRollingHash[i] = byte(i + 1)
	}

	pi := public_input.Invalidity{
		TxHash:              common.HexToHash("0x11223344556677889900aabbccddeeff00112233445566778899aabbccddeeff"),
		TxNumber:            7,
		FromAddress:         linTypes.EthAddress(common.HexToAddress("0x1111111111111111111111111111111111111111")),
		ExpectedBlockHeight: 42,
		StateRootHash:       linTypes.MustHexToKoalabearOctuplet("0x0b1dfeef3db4956540da8a5f785917ef1ba432e521368da60a0a1ce430425666"),
		FtxRollingHash:      ftxRollingHash,
	}

	// Expected hash from native byte-stream implementation.
	var expected fr.Element
	nativeSum := pi.Sum(nil)
	expected.SetBytes(nativeSum)

	var gnarkInputs invalidity.FunctionalPublicInputsGnark
	gnarkInputs.Assign(pi)

	circuit := &invalidityHashCircuit{}
	assignment := &invalidityHashCircuit{
		Inputs: gnarkInputs,
		Hash:   expected,
	}

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
	require.NoError(t, err)

	witness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	require.NoError(t, ccs.IsSolved(witness))
}
