package plonkutil_test

import (
	"fmt"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover/plonkutil"
	plonkBn254 "github.com/consensys/gnark/backend/plonk/bn254"
)

func TestReadPP(t *testing.T) {

	t.Setenv("PROVER_PKEY_FILE", "../../../docker/setup/light/proving_key.bin")
	t.Setenv("PROVER_VKEY_FILE", "../../../docker/setup/light/verifying_key.bin")
	t.Setenv("PROVER_R1CS_FILE", "../../../docker/setup/light/circuit.bin")
	t.Setenv("PROVER_CONFLATED_TRACES_DIR", "not-needed")
	t.Setenv("PROVER_VERSION", "none")

	setup := make(chan plonkutil.Setup, 1)
	go plonkutil.ReadPPFromConfig(setup)
	// wait for the setup to finish before returning to catch any errors.
	<-setup
}

func TestRecreateVerifierSol(t *testing.T) {

	t.SkipNow()

	types := []string{"full", "full-large"}

	for _, t := range types {
		vkeyFile := fmt.Sprintf("../../../docker/setup/%v/verifying_key.bin", t)
		exportTo := fmt.Sprintf("../../../docker/setup/%v/Verifier.sol", t)
		vk := &plonkBn254.VerifyingKey{}

		// Read the proving key
		f := files.MustRead(vkeyFile)
		vk.ReadFrom(f)
		f.Close()

		// Write the verification key
		f = files.MustOverwrite(exportTo)
		vk.ExportSolidity(f)
		f.Close()
	}

}
