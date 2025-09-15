package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/sirupsen/logrus"
)

func main() {

	// allow override via environment variable
	maxRlpByteSize := 1 << 10
	if v := os.Getenv("MAX_RLP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxRlpByteSize = n
		}
	}
	fmt.Println("Compiling with maxRlpByteSize =", maxRlpByteSize)

	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}
	)

	// generate keccak proof for the circuit
	maxNumKeccakF := maxRlpByteSize/136 + 1 // each keccakF can hash 136 bytes.
	colSize := maxRlpByteSize/16 + 1        // each limb is 16 bytes.
	// @azam do we still need this ?
	size := utils.NextPowerOfTwo(colSize)

	gdm := generic.GenDataModule{}
	gim := generic.GenInfoModule{}

	definer := func(builder *wizard.Builder) {
		comp := builder.CompiledIOP
		gdm = invalidity.CreateGenDataModule(comp, size)
		gim = invalidity.CreateGenInfoModule(comp, size)

		inp := keccak.KeccakSingleProviderInput{
			MaxNumKeccakF: maxNumKeccakF,
			Provider: generic.GenericByteModule{
				Data: gdm,
				Info: gim},
		}
		keccak.NewKeccakSingleProvider(comp, inp)
	}
	comp := wizard.Compile(definer, dummy.Compile)
	logrus.Info("keccak circuit compiled")

	// define the circuit
	circuit := invalidity.CircuitInvalidity{
		SubCircuit: &invalidity.BadNonceBalanceCircuit{},
	}

	// allocate the circuit
	circuit.Allocate(invalidity.Config{
		Depth:             config.Depth,
		KeccakCompiledIOP: comp,
		MaxRlpByteSize:    maxRlpByteSize,
	})

	// compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&circuit,
	)

	if err != nil {
		utils.Panic("circuit compilation failed: %v", err)
	}

	logrus.WithField("constraints", scs.GetNbConstraints()).Info("circuit compiled")

}
