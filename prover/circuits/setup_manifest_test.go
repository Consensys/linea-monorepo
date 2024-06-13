package circuits

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const exampleJson = `{
	"circuitName": "aggregation-10",
	"timestamp": "2024-04-24T21:41:20.104431-05:00",
	"checksums": {
	  "verifyingKey": "0x78022b8d128bc34096eafcecc1d5a36f95affdb4264f52cbeeb043dc73948f3d",
	  "verifierContract": "",
	  "circuit": "0x59d950e0fb10c5aa20f8c0082ffc550283a4aa2b5325be03623c331e651801eb"
	},
	"nbConstraints": 3116379,
	"curveID": "bw6_761",
	"extraFlags": {
	  "allowedVkForAggregationDigests": [
		"0xd1624b8e9e5987f7bbf85cb32bb7b9787144aceb4527864f31ba1957e300f7eb",
		"0xc4a868954d361bf8c18d4b3699c4fa973a6b2e4543ddea0ce7970d6941f55758"
	  ]
	}
  }`

func TestParseManifestAndFlags(t *testing.T) {
	assert := require.New(t)

	var m SetupManifest
	err := json.NewDecoder(bytes.NewReader([]byte(exampleJson))).Decode(&m)
	assert.NoError(err)

	keys, err := m.GetStringArray("allowedVkForAggregationDigests")
	assert.NoError(err)

	assert.Equal(2, len(keys))
	assert.Equal("0xd1624b8e9e5987f7bbf85cb32bb7b9787144aceb4527864f31ba1957e300f7eb", keys[0])
	assert.Equal("0xc4a868954d361bf8c18d4b3699c4fa973a6b2e4543ddea0ce7970d6941f55758", keys[1])
}
