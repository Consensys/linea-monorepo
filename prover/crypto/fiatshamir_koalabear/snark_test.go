package fiatshamir

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
)

type SafeGuardUpdateCircuit struct {
	A, B   frontend.Variable
	R1, R2 poseidon2_koalabear.Octuplet
}

func (c *SafeGuardUpdateCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)
	fs.Update(c.A)
	a := fs.RandomField()
	fs.Update(c.B)
	b := fs.RandomField()
	api.AssertIsDifferent(a, b)
	return nil
}

func GetCircuitWitnessSafeGuardUpdateCircuit() (*SafeGuardUpdateCircuit, *SafeGuardUpdateCircuit) {

	h := poseidon2_koalabear.NewMDHasher()
	var a, b, c koalabear.Element
	a.SetRandom()
	b.SetRandom()
	Update(h, a)
	c = Random

	var circuit, witness SafeGuardUpdateCircuit
	circuit.A = a.String()
	circuit.B = b.String()
	return &circuit, &witness
}

// func TestGnarkRandomVec(t *testing.T) {

// 	for _, testCase := range randomIntVecTestCases {
// 		testName := fmt.Sprintf("%v-integers-of-%v-bits", testCase.NumIntegers, testCase.IntegerBitSize)
// 		t.Run(testName, func(t *testing.T) {

// 			f := func(api frontend.API) error {

// 				gnarkFs := NewGnarkFS(api, nil)
// 				fs := NewMiMCFiatShamir()

// 				fs.Update(field.NewElement(2))
// 				gnarkFs.Update(field.NewElement(2))

// 				a := fs.RandomManyIntegers(testCase.NumIntegers, 1<<testCase.IntegerBitSize)
// 				aGnark := gnarkFs.RandomManyIntegers(testCase.NumIntegers, 1<<testCase.IntegerBitSize)

// 				for i := range a {
// 					api.AssertIsEqual(a[i], aGnark[i])
// 				}

// 				return nil
// 			}

// 			gnarkutil.AssertCircuitSolved(t, f)
// 		})
// 	}
// }

// func TestGnarkFiatShamirEmpty(t *testing.T) {

// 	f := func(api frontend.API) error {
// 		Y := NewMiMCFiatShamir().RandomFext()
// 		fs := NewGnarkFS(api, nil)
// 		y := fs.RandomField()
// 		api.AssertIsEqual(Y, y)
// 		return nil
// 	}

// 	gnarkutil.AssertCircuitSolved(t, f)
// }

// func TestGnarkUpdateVec(t *testing.T) {

// 	f := func(api frontend.API) error {
// 		fs := NewMiMCFiatShamir()
// 		fs.UpdateVec(vector.ForTest(2, 2, 1, 2))
// 		y1 := fs.RandomFext()

// 		fs2 := NewGnarkFS(api, nil)
// 		fs2.UpdateVec([]zk.WrappedVariable{2, 2, 1, 2})
// 		y2 := fs2.RandomField()

// 		api.AssertIsEqual(y1, y2)
// 		return nil
// 	}

// 	gnarkutil.AssertCircuitSolved(t, f)
// }
