package fiatshamir_koalabear

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/stretchr/testify/require"
)

// TestFSConsistencyBN254 tests that GnarkFSKoalagnark in BN254 emulated mode
// produces the same RandomFieldExt and RandomManyIntegers as the native FS.
type fsConsistencyCircuit struct {
	// Inputs to feed to the FS
	Inputs [16]koalagnark.Element
	// Expected FieldExt coin (4 koalagnark elements)
	ExpectedCoin koalagnark.Ext
	// Expected integer vec (4 entries, upper bound 128)
	ExpectedInts [4]frontend.Variable
}

func (c *fsConsistencyCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	koalaAPI := koalagnark.NewAPI(api)

	// Update FS with inputs
	fs.Update(c.Inputs[:]...)

	// Sample a FieldExt coin
	coin := fs.RandomFieldExt()

	// Sample integer vec
	ints := fs.RandomManyIntegers(4, 128)

	// Check FieldExt coin
	koalaAPI.AssertIsEqualExt(coin, c.ExpectedCoin)

	// Check integer vec
	for i := 0; i < 4; i++ {
		api.AssertIsEqual(ints[i], c.ExpectedInts[i])
	}

	return nil
}

func TestFSConsistencyBN254(t *testing.T) {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		t.Fatal(err)
	}

	// Native computation
	nativeFS := NewFS()

	inputs := [16]field.Element{}
	for i := range inputs {
		inputs[i].SetUint64(uint64(i*12345 + 67890))
	}
	nativeFS.Update(inputs[:]...)

	// Sample FieldExt coin
	nativeCoin := nativeFS.RandomFext()
	t.Logf("Native coin: %v", nativeCoin.String())

	// Sample integer vec
	nativeInts := nativeFS.RandomManyIntegers(4, 128)
	t.Logf("Native ints: %v", nativeInts)

	// Build circuit witness
	var circuit fsConsistencyCircuit
	var assignment fsConsistencyCircuit

	for i := 0; i < 16; i++ {
		assignment.Inputs[i] = koalagnark.NewElementFromKoala(inputs[i])
	}
	assignment.ExpectedCoin = koalagnark.NewExt(nativeCoin)
	for i := 0; i < 4; i++ {
		assignment.ExpectedInts[i] = nativeInts[i]
	}

	err := test.IsSolved(&circuit, &assignment, ecc.BN254.ScalarField())
	require.NoError(t, err, "FS consistency check failed on BN254")
	t.Log("GnarkFSKoalagnark on BN254 is consistent with native FS")
}

// TestFSRandomFieldFullOctuplet tests all 8 elements of RandomField output
type fsFullOctupletCircuit struct {
	Inputs   [16]koalagnark.Element
	Expected [8]koalagnark.Element // ALL 8 elements from RandomField
}

func (c *fsFullOctupletCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	koalaAPI := koalagnark.NewAPI(api)

	fs.Update(c.Inputs[:]...)
	oct := fs.RandomField()

	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(oct[i], c.Expected[i])
	}
	return nil
}

func TestFSRandomFieldFullOctupletBN254(t *testing.T) {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		t.Fatal(err)
	}

	nativeFS := NewFS()
	inputs := [16]field.Element{}
	for i := range inputs {
		inputs[i].SetUint64(uint64(i*12345 + 67890))
	}
	nativeFS.Update(inputs[:]...)
	nativeOct := nativeFS.RandomField()

	t.Logf("Native full octuplet: %v", nativeOct)

	var circuit fsFullOctupletCircuit
	var assignment fsFullOctupletCircuit

	for i := 0; i < 16; i++ {
		assignment.Inputs[i] = koalagnark.NewElementFromKoala(inputs[i])
	}
	for i := 0; i < 8; i++ {
		assignment.Expected[i] = koalagnark.NewElementFromKoala(nativeOct[i])
	}

	err := test.IsSolved(&circuit, &assignment, ecc.BN254.ScalarField())
	require.NoError(t, err, "Full octuplet consistency check failed")
	t.Log("All 8 RandomField elements match on BN254")
}

// TestFSSecondRandomFieldAfterSafeguard tests that the second RandomField
// call (after safeguardUpdate) produces the correct result.
type fsSecondCallCircuit struct {
	Inputs          [16]koalagnark.Element
	ExpectedFirst   [8]koalagnark.Element // first RandomField
	ExpectedSecond  [8]koalagnark.Element // second RandomField (after safeguard)
}

func (c *fsSecondCallCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	koalaAPI := koalagnark.NewAPI(api)

	fs.Update(c.Inputs[:]...)

	// First RandomField (includes safeguardUpdate)
	first := fs.RandomField()
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(first[i], c.ExpectedFirst[i])
	}

	// Second RandomField (processes safeguard zero, then safeguardUpdate again)
	second := fs.RandomField()
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(second[i], c.ExpectedSecond[i])
	}

	return nil
}

func TestFSSecondRandomFieldBN254(t *testing.T) {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		t.Fatal(err)
	}

	nativeFS := NewFS()
	inputs := [16]field.Element{}
	for i := range inputs {
		inputs[i].SetUint64(uint64(i*12345 + 67890))
	}
	nativeFS.Update(inputs[:]...)

	nativeFirst := nativeFS.RandomField()
	t.Logf("Native first: %v", nativeFirst)

	nativeSecond := nativeFS.RandomField()
	t.Logf("Native second: %v", nativeSecond)

	var circuit fsSecondCallCircuit
	var assignment fsSecondCallCircuit

	for i := 0; i < 16; i++ {
		assignment.Inputs[i] = koalagnark.NewElementFromKoala(inputs[i])
	}
	for i := 0; i < 8; i++ {
		assignment.ExpectedFirst[i] = koalagnark.NewElementFromKoala(nativeFirst[i])
		assignment.ExpectedSecond[i] = koalagnark.NewElementFromKoala(nativeSecond[i])
	}

	err := test.IsSolved(&circuit, &assignment, ecc.BN254.ScalarField())
	require.NoError(t, err, "Second RandomField consistency check failed")
	t.Log("Second RandomField after safeguard matches on BN254")
}

// TestToBinaryFromBinary tests ToBinary + FromBinary path for emulated KoalaBear
type toBinaryCircuit struct {
	Input    koalagnark.Element
	Expected frontend.Variable // expected lowest 7 bits as integer
}

func (c *toBinaryCircuit) Define(api frontend.API) error {
	koalaAPI := koalagnark.NewAPI(api)
	bits := koalaAPI.ToBinary(c.Input)
	nbBits := 7
	reconstructed := api.FromBinary(bits[:nbBits]...)
	api.AssertIsEqual(reconstructed, c.Expected)
	return nil
}

func TestToBinaryFromBinaryBN254(t *testing.T) {
	testCases := []struct {
		value    uint64
		expected int // value & 127
	}{
		{49, 49},
		{128, 0},
		{255, 127},
		{1268550066, 1268550066 & 127},
		{0, 0},
		{1, 1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("value_%d", tc.value), func(t *testing.T) {
			var circuit toBinaryCircuit
			var assignment toBinaryCircuit

			var fe field.Element
			fe.SetUint64(tc.value)
			assignment.Input = koalagnark.NewElementFromKoala(fe)
			assignment.Expected = tc.expected

			err := test.IsSolved(&circuit, &assignment, ecc.BN254.ScalarField())
			require.NoError(t, err, "ToBinary/FromBinary failed for value %d", tc.value)
		})
	}
}

// TestFSConsistencyMultiRound tests multi-round FS operations
type fsMultiRoundCircuit struct {
	// Round 1 inputs
	Round1Inputs [8]koalagnark.Element
	// Round 2 inputs (extension elements, flattened)
	Round2Inputs [32]koalagnark.Element // 8 ext elements = 32 base elements
	// Expected FieldExt after round 1
	ExpectedCoin1 koalagnark.Ext
	// Expected FieldExt after round 2
	ExpectedCoin2 koalagnark.Ext
	// Expected integers after round 2
	ExpectedInts [4]frontend.Variable
}

func (c *fsMultiRoundCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	koalaAPI := koalagnark.NewAPI(api)

	// Round 1: update + sample FieldExt
	fs.Update(c.Round1Inputs[:]...)
	coin1 := fs.RandomFieldExt()
	koalaAPI.AssertIsEqualExt(coin1, c.ExpectedCoin1)

	// Round 2: update with extension data + sample FieldExt + sample IntegerVec
	fs.Update(c.Round2Inputs[:]...)
	coin2 := fs.RandomFieldExt()
	koalaAPI.AssertIsEqualExt(coin2, c.ExpectedCoin2)

	ints := fs.RandomManyIntegers(4, 128)
	for i := 0; i < 4; i++ {
		api.AssertIsEqual(ints[i], c.ExpectedInts[i])
	}

	return nil
}

func TestFSConsistencyMultiRoundBN254(t *testing.T) {
	if err := poseidon2_koalabear.RegisterGates(); err != nil {
		t.Fatal(err)
	}

	// Native computation
	nativeFS := NewFS()

	round1 := [8]field.Element{}
	for i := range round1 {
		round1[i].SetUint64(uint64(i*1111 + 2222))
	}
	nativeFS.Update(round1[:]...)
	nativeCoin1 := nativeFS.RandomFext()
	t.Logf("Native coin1: %v", nativeCoin1.String())

	// Round 2: extension elements flattened to base
	round2Ext := [8]fext.Element{}
	for i := range round2Ext {
		round2Ext[i].B0.A0.SetUint64(uint64(i*100 + 1))
		round2Ext[i].B0.A1.SetUint64(uint64(i*100 + 2))
		round2Ext[i].B1.A0.SetUint64(uint64(i*100 + 3))
		round2Ext[i].B1.A1.SetUint64(uint64(i*100 + 4))
	}
	// Feed as extension (how native verifier does it via UpdateSV → UpdateExt)
	nativeFS.UpdateExt(round2Ext[:]...)
	nativeCoin2 := nativeFS.RandomFext()
	t.Logf("Native coin2: %v", nativeCoin2.String())

	nativeInts := nativeFS.RandomManyIntegers(4, 128)
	t.Logf("Native ints: %v", nativeInts)

	// Build circuit witness
	var circuit fsMultiRoundCircuit
	var assignment fsMultiRoundCircuit

	for i := 0; i < 8; i++ {
		assignment.Round1Inputs[i] = koalagnark.NewElementFromKoala(round1[i])
	}
	// Flatten extension elements the same way the gnark verifier does
	for i := 0; i < 8; i++ {
		assignment.Round2Inputs[4*i] = koalagnark.NewElementFromKoala(round2Ext[i].B0.A0)
		assignment.Round2Inputs[4*i+1] = koalagnark.NewElementFromKoala(round2Ext[i].B0.A1)
		assignment.Round2Inputs[4*i+2] = koalagnark.NewElementFromKoala(round2Ext[i].B1.A0)
		assignment.Round2Inputs[4*i+3] = koalagnark.NewElementFromKoala(round2Ext[i].B1.A1)
	}
	assignment.ExpectedCoin1 = koalagnark.NewExt(nativeCoin1)
	assignment.ExpectedCoin2 = koalagnark.NewExt(nativeCoin2)
	for i := 0; i < 4; i++ {
		assignment.ExpectedInts[i] = nativeInts[i]
	}

	err := test.IsSolved(&circuit, &assignment, ecc.BN254.ScalarField())
	if err != nil {
		// Print detailed error
		fmt.Printf("coin1: %v\n", nativeCoin1.String())
		fmt.Printf("coin2: %v\n", nativeCoin2.String())
		fmt.Printf("ints: %v\n", nativeInts)
	}
	require.NoError(t, err, "Multi-round FS consistency check failed on BN254")
	t.Log("Multi-round GnarkFSKoalagnark on BN254 is consistent")
}
