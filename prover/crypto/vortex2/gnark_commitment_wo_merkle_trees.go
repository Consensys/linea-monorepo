package vortex2

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
)

// Final ciruit - commitment without merkle trees

type VerifyOpeningCircuit struct {
	Proof       GProofWoMerkle        `gnark:"public"`
	X           frontend.Variable     `gnark:"public"`
	RandomCoin  frontend.Variable     `gnark:"public"`
	Ys          [][]frontend.Variable `gnark:"public"`
	EntryList   []frontend.Variable   `gnark:"public"`
	Commitments [][]frontend.Variable `gnark:"public"`
	Params      GParams
}

func (circuit *VerifyOpeningCircuit) Define(api frontend.API) error {

	err := GnarkVerify(
		api,
		circuit.Params,
		circuit.Commitments,
		circuit.Proof,
		circuit.X,
		circuit.Ys,
		circuit.RandomCoin,
		circuit.EntryList,
	)
	return err

}

// AllocateCircuitVariables allocate the slices of the verifier circuit
func AllocateCircuitVariables(
	verifyCircuit *VerifyOpeningCircuit,
	proof Proof,
	ys [][]field.Element,
	entryList []int,
	commitments []Commitment) {

	verifyCircuit.Proof.LinearCombination = make([]frontend.Variable, proof.LinearCombination.Len())

	verifyCircuit.Proof.Columns = make([][][]frontend.Variable, len(proof.Columns))
	for i := 0; i < len(proof.Columns); i++ {
		verifyCircuit.Proof.Columns[i] = make([][]frontend.Variable, len(proof.Columns[i]))
		for j := 0; j < len(proof.Columns[i]); j++ {
			verifyCircuit.Proof.Columns[i][j] = make([]frontend.Variable, len(proof.Columns[i][j]))
		}
	}

	verifyCircuit.EntryList = make([]frontend.Variable, len(entryList))

	verifyCircuit.Ys = make([][]frontend.Variable, len(ys))
	for i := 0; i < len(ys); i++ {
		verifyCircuit.Ys[i] = make([]frontend.Variable, len(ys[i]))
	}

	verifyCircuit.Commitments = make([][]frontend.Variable, len(commitments))
	for i := 0; i < len(commitments); i++ {
		verifyCircuit.Commitments[i] = make([]frontend.Variable, len(commitments[i]))
	}
}

// AssignCicuitVariables assign the witnesses of the slices of the verifier circuit
func AssignCicuitVariables(
	verifyCircuit *VerifyOpeningCircuit,
	proof Proof,
	ys [][]field.Element,
	entryList []int,
	commitments []Commitment) {

	frLinComb := make([]fr.Element, proof.LinearCombination.Len())
	proof.LinearCombination.WriteInSlice(frLinComb)
	for i := 0; i < proof.LinearCombination.Len(); i++ {
		verifyCircuit.Proof.LinearCombination[i] = frLinComb[i].String()
	}

	for i := 0; i < len(proof.Columns); i++ {
		for j := 0; j < len(proof.Columns[i]); j++ {
			for k := 0; k < len(proof.Columns[i][j]); k++ {
				verifyCircuit.Proof.Columns[i][j][k] = proof.Columns[i][j][k].String()
			}
		}
	}

	for i := 0; i < len(entryList); i++ {
		verifyCircuit.EntryList[i] = entryList[i]
	}

	for i := 0; i < len(ys); i++ {
		for j := 0; j < len(ys[i]); j++ {
			verifyCircuit.Ys[i][j] = ys[i][j].String()
		}
	}

	for i := 0; i < len(commitments); i++ {
		for j := 0; j < len(commitments[i]); j++ {
			verifyCircuit.Commitments[i][j] = commitments[i][j].String()
		}
	}

}
