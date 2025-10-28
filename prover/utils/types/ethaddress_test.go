package types_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/consensys/linea-monorepo/prover/utils/types"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnMarhsalAddress(t *testing.T) {

	tcases := []struct {
		jsonString  string
		hexAddress  string
		Expected    string
		AccountHash string
	}{
		{
			jsonString:  `{"address": "0x"}`,
			hexAddress:  "0x0000000000000000000000000000000000000000",
			Expected:    "00000000000000000000000000000000000000000000000000000000000000000000000000000000",
			AccountHash: "0x225471e76dca970375f18cc4222539bf1893f40b2c28b7d62119af207704c426",
		},
		{
			jsonString:  `{"address": "0xabcd"}`,
			hexAddress:  "0x000000000000000000000000000000000000abcd",
			Expected:    "0000000000000000000000000000000000000000000000000000000000000000000000000000abcd",
			AccountHash: "0x3e455e235bb3ee213d538e4b3696f11e2dd918bd79d502703f9f6c4c0a230b59",
		},
		{
			jsonString:  `{"address": "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress:  "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Expected:    "0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef",
			AccountHash: "0x4f465a0b4b9bdb1c3d379bb2200f6340153aaf124c320a3014f53550491b9446",
		},
		{
			jsonString:  `{"address": "0x1234125737deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress:  "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Expected:    "0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef",
			AccountHash: "0x4f465a0b4b9bdb1c3d379bb2200f6340153aaf124c320a3014f53550491b9446",
		},
	}

	for _, c := range tcases {

		var deserialized struct {
			Address EthAddress `json:"address"`
		}

		if err := json.Unmarshal([]byte(c.jsonString), &deserialized); err != nil {
			t.Fatalf("could not deserialize")
		}
		buffer := &bytes.Buffer{}
		deserialized.Address.WriteTo(buffer)

		assert.Equal(t, c.hexAddress, deserialized.Address.Hex())

		hex := hex.EncodeToString(buffer.Bytes())
		assert.Equal(t, c.Expected, hex)

		// Calculate the hash
		hasher := poseidon2.Poseidon2()
		deserialized.Address.WriteTo(hasher)
		hash := hasher.Sum(nil)

		// Format the hash as a "0x" prefixed hex string
		hashString := fmt.Sprintf("0x%x", hash)

		assert.Equal(t, c.AccountHash, hashString)

	}

}

func TestReadWriteAddress(t *testing.T) {
	hexAddress, _ := AddressFromHex("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	buf := &bytes.Buffer{}
	_, err := hexAddress.WriteTo(buf)
	require.NoError(t, err, "failed to write hexAddress to buffer")

	var add EthAddress
	_, err = add.ReadFrom(buf)

	require.NoError(t, err)
	assert.Equal(t, hexAddress, add)

}
