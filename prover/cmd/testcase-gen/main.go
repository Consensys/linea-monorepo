package main

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/coordinator"
	"github.com/consensys/accelerated-crypto-monorepo/backend/coordinator/testcase_gen"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/dummycircuit"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func main() {

	testcase_gen.Initialize()

	// Run the setup
	pp := dummycircuit.GenPublicParamsUnsafe()
	generator := testcase_gen.MakeGeneratorFromCLI()

	output := &coordinator.ProverOutput{}
	generator.PopulateCoordOutput(output)
	output.ComputeProofInput()

	var x fr.Element
	x.SetString(output.DebugData.FinalHash)

	output.Proof = dummycircuit.MakeProof(pp, x)
	output.WriteInFile(testcase_gen.Ofile())
}
