package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnMarhsalAddress(t *testing.T) {

	tcases := []struct {
		jsonString string
		hexAddress string
	}{
		{
			jsonString: `{"address": "0x"}`,
			hexAddress: "0x0000000000000000000000000000000000000000",
		},
		{
			jsonString: `{"address": "0xabcd"}`,
			hexAddress: "0x000000000000000000000000000000000000abcd",
		},
		{
			jsonString: `{"address": "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},
		{
			jsonString: `{"address": "0x1234125737deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}`,
			hexAddress: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},
	}

	for _, c := range tcases {

		var deserialized struct {
			Address EthAddress `json:"address"`
		}

		if err := json.Unmarshal([]byte(c.jsonString), &deserialized); err != nil {
			t.Fatalf("could not deserialize")
		}

		assert.Equal(t, c.hexAddress, deserialized.Address.Hex())
	}

}
