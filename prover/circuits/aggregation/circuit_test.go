package aggregation

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal/test_utils"
	"github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection/keccak"
	public_input "github.com/consensys/zkevm-monorepo/prover/public-input"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublicInput(t *testing.T) {

	// test case taken from backend/aggregation
	testCases := []public_input.Aggregation{
		{
			FinalShnarf:                             "0x3f01b1a726e6317eb05d8fe8b370b1712dc16a7fde51dd38420d9a474401291c",
			ParentAggregationFinalShnarf:            "0x0f20c85d35a21767e81d5d2396169137a3ef03f58391767a17c7016cc82edf2e",
			ParentAggregationLastBlockTimestamp:     1711742796,
			FinalTimestamp:                          1711745271,
			LastFinalizedBlockNumber:                3237969,
			FinalBlockNumber:                        3238794,
			LastFinalizedL1RollingHash:              "0xe578e270cc6ee7164d4348ac7ca9a7cfc0c8c19b94954fc85669e75c1db46178",
			L1RollingHash:                           "0x0578f8009189d67ce0378619313b946f096ca20dde9cad0af12a245500054908",
			LastFinalizedL1RollingHashMessageNumber: 549238,
			L1RollingHashMessageNumber:              549263,
			L2MsgRootHashes:                         []string{"0xfb7ce9c89be905d39bfa2f6ecdf312f127f8984cf313cbea91bca882fca340cd"},
			L2MsgMerkleTreeDepth:                    5,
		},
	}

	for i := range testCases {

		fpi, err := public_input.NewAggregationFPI(&testCases[i])
		assert.NoError(t, err)

		sfpi := fpi.ToSnarkType()
		// TODO incorporate into public input hash or decide not to
		sfpi.NbDecompression = -1
		sfpi.InitialStateRootHash = -2
		sfpi.ChainID = -3
		sfpi.L2MessageServiceAddr = -4
		sfpi.NbL2Messages = -5

		var res [32]frontend.Variable
		assert.NoError(t, internal.CopyHexEncodedBytes(res[:], testCases[i].GetPublicInputHex()))

		test_utils.SnarkFunctionTest(func(api frontend.API) []frontend.Variable {
			sum := sfpi.Sum(api, keccak.NewHasher(api, 500))
			return sum[:]
		}, res[:]...)(t)
	}
}
