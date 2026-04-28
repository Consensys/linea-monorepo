package plonk2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/stretchr/testify/require"
)

type e2eMulCircuit struct {
	X frontend.Variable
	Y frontend.Variable
	Z frontend.Variable `gnark:",public"`
}

func (c *e2eMulCircuit) Define(api frontend.API) error {
	api.AssertIsEqual(api.Mul(c.X, c.Y), c.Z)
	return nil
}

func TestPlonkE2E_AllTargetCurves(t *testing.T) {
	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			ccs, err := frontend.Compile(
				curveID.ScalarField(),
				scs.NewBuilder,
				&e2eMulCircuit{},
			)
			require.NoError(t, err, "compiling circuit should succeed")

			srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
			require.NoError(t, err, "creating unsafe test SRS should succeed")

			pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
			require.NoError(t, err, "PlonK setup should succeed")

			validWitness, err := frontend.NewWitness(
				&e2eMulCircuit{X: 3, Y: 11, Z: 33},
				curveID.ScalarField(),
			)
			require.NoError(t, err, "creating valid witness should succeed")
			publicWitness, err := validWitness.Public()
			require.NoError(t, err, "extracting public witness should succeed")

			proof, err := gnarkplonk.Prove(ccs, pk, validWitness)
			require.NoError(t, err, "PlonK prove should succeed")
			require.NoError(
				t,
				gnarkplonk.Verify(proof, vk, publicWitness),
				"PlonK verify should succeed",
			)

			invalidWitness, err := frontend.NewWitness(
				&e2eMulCircuit{X: 3, Y: 11, Z: 34},
				curveID.ScalarField(),
			)
			require.NoError(t, err, "creating invalid witness should succeed")
			_, err = gnarkplonk.Prove(ccs, pk, invalidWitness)
			require.Error(t, err, "invalid witness should not prove")
		})
	}
}
