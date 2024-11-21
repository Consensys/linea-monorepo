package circuits

import (
	"bytes"
	"context"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/kzg"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRSStore(t *testing.T) {
	assert := require.New(t)

	srsStore, err := NewSRSStore("../prover-assets/kzgsrs")
	assert.NoError(err)

	assert.Greater(len(srsStore.entries), 0)

	// log the entries
	for curveID, entries := range srsStore.entries {
		t.Logf("curveID %s\n", curveID)
		for _, entry := range entries {
			t.Logf("entry %v\n", entry)
		}
	}
}

func TestLagrange(t *testing.T) {

	cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit{Input: make([]frontend.Variable, 10)})
	require.NoError(t, err)

	canonicalSize, lagrangeSize := plonk.SRSSize(cs)
	canonical, lagrange, err := NewUnsafeSRSProvider().GetSRS(context.TODO(), cs)
	require.NoError(t, err)

	var bb bytes.Buffer
	require.NoError(t, canonical.WriteDump(&bb))

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("kzg_srs_canonical_%d_bls12377_aleo.memdump", canonicalSize)), bb.Bytes(), 0600))

	srsStore, err := NewSRSStore(dir)
	require.NoError(t, err)

	canonicalBack, lagrangeBack, err := srsStore.GetSRS(context.TODO(), cs)
	require.NoError(t, err)

	assertSerializablesEqual(t, canonical, canonicalBack)
	assertSerializablesEqual(t, lagrange, lagrangeBack)

	// now read the Lagrange file
	lagrangeFromFile, err := os.ReadFile(filepath.Join(dir, fmt.Sprintf("kzg_srs_lagrange_%d_bls12377_aleo.memdump", lagrangeSize)))
	require.NoError(t, err)

	bb.Reset()
	require.NoError(t, lagrange.WriteDump(&bb))
	require.NoError(t, test_utils.BytesEqual(lagrangeFromFile, bb.Bytes()))
}

func assertSerializablesEqual(t *testing.T, a, b kzg.Serializable) {
	var ab, bb bytes.Buffer
	require.NoError(t, a.WriteDump(&ab))
	require.NoError(t, b.WriteDump(&bb))
	require.NoError(t, test_utils.BytesEqual(ab.Bytes(), bb.Bytes()))
}

type circuit struct {
	// this is a dummy circuit that does nothing
	// it is used to generate the SRS
	Input []frontend.Variable `gnark:",public"`
}

func (circuit *circuit) Define(api frontend.API) error {
	c, err := api.(frontend.Committer).Commit(circuit.Input...)
	if err != nil {
		return err
	}
	prod := frontend.Variable(1)
	for _, x := range circuit.Input {
		prod = api.Mul(prod, x)
	}
	api.AssertIsDifferent(c, prod)
	return nil
}
