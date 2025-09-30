package gnarkutil

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark/frontend"
)

func TestCountVariables(t *testing.T) {

	testcases := []struct {
		Circ         any
		NbPub, NbSec int
	}{
		{
			Circ: struct {
				A T `gnark:",public"`
				B T `gnark:",secret"`
			}{},
			NbPub: 1,
			NbSec: 1,
		},
		{
			Circ: struct {
				A T `gnark:",secret"`
				B T `gnark:",public"`
			}{},
			NbPub: 1,
			NbSec: 1,
		},
		{
			Circ: struct {
				A [3]T `gnark:",secret"`
				B T    `gnark:",public"`
			}{},
			NbPub: 1,
			NbSec: 3,
		},
		{
			Circ: struct {
				A T    `gnark:",secret"`
				B [3]T `gnark:",public"`
			}{},
			NbPub: 3,
			NbSec: 1,
		},
	}

	for i := range testcases {

		testcaseName := fmt.Sprintf("case-%v", i)

		t.Run(testcaseName, func(t *testing.T) {

			nbPub, nbSec := CountVariables(testcases[i].Circ)
			if nbPub != testcases[i].NbPub {
				t.Errorf("nbPub: %d != %d", nbPub, testcases[i].NbPub)
			}
			if nbSec != testcases[i].NbSec {
				t.Errorf("nbSec: %d != %d", nbSec, testcases[i].NbSec)
			}

		})

	}

}
