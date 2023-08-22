package json_test

import (
	"bytes"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/json"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestJsonExample(t *testing.T) {

	filenames := []string{"./testjson2.json", "./testjson3.json"}
	for _, fname := range filenames {

		t.Logf("testing file %v", fname)

		buf := bytes.Buffer{}
		f := files.MustRead(fname)
		n, err := buf.ReadFrom(f)
		require.NotZero(t, n, "read zero bytes")
		require.NoError(t, err, "error parsing the json")

		v, err := fastjson.ParseBytes(buf.Bytes())
		require.NoErrorf(t, err, "the test json is broken")
		require.NotNilf(t, v, "parser output was nil")

		// Get the field result.zkStateMerkleProof
		blocks := v.Get("result", "zkStateMerkleProof")
		require.NotNilf(t, blocks, "could not get `result.zkStateMerkleProof`")

		// Gets the parent state root hash
		parentRootHash, err := jsonutil.TryGetDigest(*v, "result", "zkParentStateRootHash")
		require.NoErrorf(t, err, "failed parsing the parent root hash")
		require.NotNilf(t, parentRootHash, "could not get `result.parentStateRootHash`")

		parsed, err := json.ParseStateManagerTraces(*blocks)
		require.NoError(t, err, "ParseBlockTraces returned an error")
		require.NotNil(t, parsed, "ParseBlockTraces returned nil")

		// We keep track of the oldest root hash
		parent := parentRootHash
		for i := range parsed {
			old, new, err := eth.CheckTraces(parsed[i])
			require.NoError(t, err, "inspection found an error in the traces")
			require.Equal(t, parent.Hex(), old.Hex(), "expected parent and recovered parent root hash mismatch")
			parent = new
		}

		t.Logf("testing file %v - PASS", fname)
	}

}
