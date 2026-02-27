package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"

	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/sirupsen/logrus"
)

func main() {

	// allow override via environment variable
	maxRlpByteSize := 1 << 17 // 128 KB
	if v := os.Getenv("MAX_RLP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxRlpByteSize = n
		}
	}

	depth := 40
	if v := os.Getenv("Depth"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			depth = n
		}
	}

	fmt.Printf("Compiling with maxRlpByteSize = %v and depth = %v\n", maxRlpByteSize, depth)

	var (
		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    depth,
		}
	)

	// generate keccak proof for the circuit
	maxNumKeccakF := maxRlpByteSize/136 + 1 // 136 bytes is the number of bytes absorbed per permutation keccakF.
	colSize := maxRlpByteSize/16 + 1        // each limb is 16 bytes.
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
	comp := wizard.Compile(definer, zkevm.FullCompilationSuite...)
	logrus.Info("keccak circuit compiled")

	// define the circuit
	circuit := invalidity.CircuitInvalidity{
		SubCircuit: &invalidity.BadNonceBalanceCircuit{},
	}
	p := profile.Start()
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
	p.Stop()
	fmt.Println(p.Top())

}
