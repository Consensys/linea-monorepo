//go:build !fuzzlight

package statemanager_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/stretchr/testify/require"
)

func TestJsonExample(t *testing.T) {

	dirEntries, err := os.ReadDir("./testdata")
	if err != nil {
		t.Fatalf("could not read testdata dir, %v. Did it change location?", err)
	}

	filenames := []string{}
	for i := range dirEntries {
		filenames = append(filenames, "./testdata/"+dirEntries[i].Name())
	}

	expectErr := []bool{false, false, false, false, false, false, false}

	for tcID, fname := range filenames {

		t.Run(fmt.Sprintf("file-%v", fname), func(t *testing.T) {
			f := files.MustRead(fname)
			var parsed statemanager.ShomeiOutput
			err := json.NewDecoder(f).Decode(&parsed)

			require.NoErrorf(t, err, "failed to decode the JSON file (%v)", fname)
			f.Close()

			// It's a bad test case. Expected to yield an error
			if expectErr[tcID] {

				gotErr := false

				parent := parsed.Result.ZkParentStateRootHash
				for i := range parsed.Result.ZkStateMerkleProof {
					old, new, err := statemanager.CheckTraces(parsed.Result.ZkStateMerkleProof[i])
					if err != nil {
						gotErr = true
					}
					if parent.Hex() != old.Hex() {
						gotErr = true
					}
					parent = new
				}

				require.Truef(t, gotErr, "did not get any error for %v", fname)
				return
			}

			// We keep track of the oldest root hash
			parent := parsed.Result.ZkParentStateRootHash

			t.Logf("file has %v traces", len(parsed.Result.ZkStateMerkleProof))
			for _, blockTraces := range parsed.Result.ZkStateMerkleProof {
				old, new, err := statemanager.CheckTraces(blockTraces)
				t.Logf("(fname= %v) old: %v, parent: %v", fname, old.Hex(), new.Hex())
				require.NoError(t, err, "inspection found an error in the traces (%v)", fname)
				require.Equal(t, parent.Hex(), old.Hex(), "expected parent and recovered parent root hash mismatch (%v)", fname)
				parent = new
			}

			t.Logf("testing file %v - PASS", fname)
		})

	}

}
