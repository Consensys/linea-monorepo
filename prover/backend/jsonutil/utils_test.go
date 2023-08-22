package jsonutil

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestHexString(t *testing.T) {

	testJsonTmpl := `{"key": "%v"}`

	cases := [][]byte{
		make([]byte, 32),
		make([]byte, 31),
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		hexutil.MustDecode("0x1234546421542124"),
	}

	for _, cas := range cases {
		// format the JSON with the intended bytes
		testString := fmt.Sprintf(testJsonTmpl, hexutil.Encode(cas))

		// it should parse
		v, err := fastjson.Parse(testString)
		require.NoErrorf(t, err, "the test json is broken")
		require.NotNilf(t, v, "parser output was nil")

		// and we should recover the original bytes
		parsedBytes, err := TryGetHexBytes(*v, "key")
		require.NoError(t, err, "ParseBlockTraces returned an error")
		require.NotNil(t, parsedBytes, "ParseBlockTraces returned nil")
		require.Equal(t, cas, parsedBytes)
	}
}

func TestString(t *testing.T) {

	testJsonTmpl := `{"key": "%v"}`

	cases := []string{
		"",
		"Hello, World",
		"%%è$&@¤£µ*§",
		"AaA455",
	}

	for _, cas := range cases {
		// format the JSON with the intended bytes
		testString := fmt.Sprintf(testJsonTmpl, cas)

		// it should parse
		v, err := fastjson.Parse(testString)
		require.NoErrorf(t, err, "the test json is broken, %q", cas)
		require.NotNilf(t, v, "parser output was nil, %q", cas)

		// and we should recover the original bytes
		parsedBytes, err := TryGetString(*v, "key")
		require.NoErrorf(t, err, "ParseBlockTraces returned an error : %q", cas)
		require.NotNilf(t, parsedBytes, "ParseBlockTraces returned nil : %q", cas)
		require.Equalf(t, cas, parsedBytes, "case : %q", cas)
	}
}
