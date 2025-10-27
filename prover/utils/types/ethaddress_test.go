package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnMarhsalAddress(t *testing.T) {

	tcases := []struct {
		jsonString string
		hexAddress string
		Expected   string
	}{
		{
			jsonString: `{"address": "0x"}`,
			hexAddress: "0x0000000000000000000000000000000000000000",
			Expected:   "00000000000000000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			jsonString: `{"address": "0xabcd"}`,
			hexAddress: "0x000000000000000000000000000000000000abcd",
			Expected:   "0000000000000000000000000000000000000000000000000000000000000000000000000000abcd",
		},
		{
			jsonString: `{"address": "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Expected:   "0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef",
		},
		{
			jsonString: `{"address": "0x1234125737deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Expected:   "0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef0000dead0000beef",
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
		assert.Equal(
			t,
			c.Expected,
			hex,
		)
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
