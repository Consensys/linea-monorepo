package json

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestParseAccount(t *testing.T) {

	testJson := `{"key": "0x0000000000000000000000000000000000000000000000000000000000000041000000000000000000000000000000000000000000000000000000000000034328aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d112c7298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6eec5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4700000000000000000000000000000000000000000000000000000000000000000"}`

	// it should parse
	v, err := fastjson.Parse(testJson)
	require.NoErrorf(t, err, "the test json is broken, %q", testJson)
	require.NotNilf(t, v, "parser output was nil, %q", testJson)

	// We should recover the original account
	account, err := tryParseAccount(*v, "key")
	require.NoErrorf(t, err, "could not parse the json")
	require.Equal(t, account.Nonce, int64(65))
	require.Equal(t, account.Balance, big.NewInt(835))
	require.Equal(t, account.StorageRoot.Hex(), "0x28aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d11")
	require.Equal(t, account.CodeHash.Hex(), "0x2c7298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee")
	require.Equal(t, account.KeccakCodeHash.Hex(), "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	require.Equal(t, account.CodeSize, int64(0))

}
