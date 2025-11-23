package gnarkutil

import (
	"bytes"
	"testing"

	fr377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

func TestPartialChecksumBatchesPacked(t *testing.T) {
	sizes := []int{30, 50, 70, 90, 16}
	b := make([]byte, 256)
	expectedHashPrefixesHex := []string{
		"0x00000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d00",
		"0x1210253ea482", "0x129c0059d630", "0x1188d373d5f54f4485ee23eb",
		"0x00f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff000000000000000000000000000000"}
	for i := range b {
		b[i] = uint8(i)
	}
	for i := range sizes {
		res := PartialMiMCChecksumLooselyPackedBytes(b[:sizes[i]])
		b = b[sizes[i]:]
		expectedHashPrefix, err := hexutil.Decode(expectedHashPrefixesHex[i])
		require.NoError(t, err)
		if !bytes.HasPrefix(res, expectedHashPrefix) {
			t.Fatalf("expected checksum %s..., got 0x%x", expectedHashPrefixesHex[i], res)
		}
	}
}

// python style
func _range(n int) []byte {
	res := make([]byte, n)
	for i := range n {
		res[i] = byte(i)
	}
	return res
}

func TestPackLoose(t *testing.T) {
	data := _range(32)
	var (
		bb       bytes.Buffer
		expected [64]byte
	)
	copy(expected[1:], data)
	expected[32] = 0
	expected[33] = 31
	require.NoError(t, PackLoose(&bb, data, fr377.Bytes, 1))
	require.Equal(t, expected[:], bb.Bytes())
}
