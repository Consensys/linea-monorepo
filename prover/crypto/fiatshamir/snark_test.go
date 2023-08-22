package fiatshamir_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/fiatshamir"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc/gkrmimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/require"
)

type GnarkFSTestCircuit struct {
	X, YFe, YInt frontend.Variable
}

func (c *GnarkFSTestCircuit) Define(api frontend.API) error {
	// Run what would be otherwise in the Define
	factory := gkrmimc.NewHasherFactory(api)
	fs := fiatshamir.NewGnarkFiatShamir(api, factory)
	fs.Update(c.X)
	actualYFe := fs.RandomField()
	actualInts := fs.RandomManyIntegers(2, 16)
	api.AssertIsEqual(c.YFe, actualYFe)
	api.Println(c.YFe, actualYFe)
	api.AssertIsEqual(c.YInt, actualInts[0])
	api.Println(c.YInt, actualInts[0])
	return nil
}

type GnarkFSCircuitEmptyHash struct {
	Y frontend.Variable
}

func (c *GnarkFSCircuitEmptyHash) Define(api frontend.API) error {
	factory := gkrmimc.NewHasherFactory(api)
	fs := fiatshamir.NewGnarkFiatShamir(api, factory)
	y := fs.RandomField()
	api.AssertIsEqual(c.Y, y)
	return nil
}

func TestGnarkFiatShamirEmpty(t *testing.T) {

	ccs, err := frontend.Compile(
		ecc.BN254.ScalarField(),
		scs.NewBuilder,
		&GnarkFSCircuitEmptyHash{},
	)
	require.NoError(t, err)

	fs := fiatshamir.NewMiMCFiatShamir()
	y := fs.RandomField()

	assignment := GnarkFSCircuitEmptyHash{
		Y: y,
	}

	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	require.NoError(t, err)

	err = ccs.IsSolved(witness, gkrmimc.SolverOpts(ccs)...)
	require.NoError(t, err)
}

func TestGnarkFiatShamir(t *testing.T) {

	ccs, err := frontend.Compile(
		ecc.BN254.ScalarField(),
		scs.NewBuilder,
		&GnarkFSTestCircuit{},
	)
	require.NoError(t, err)

	x := field.NewElement(2)
	fs := fiatshamir.NewMiMCFiatShamir()
	fs.Update(x)
	yFe := fs.RandomField()
	yInt := field.NewElement(uint64(fs.RandomManyIntegers(2, 16)[0]))

	assignment := GnarkFSTestCircuit{
		X:    x,
		YFe:  yFe,
		YInt: yInt,
	}
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	require.NoError(t, err)

	err = ccs.IsSolved(witness, gkrmimc.SolverOpts(ccs)...)
	require.NoError(t, err)
}
