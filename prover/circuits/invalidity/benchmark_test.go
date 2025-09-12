package invalidity_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
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

func BenchmarkInvalidity(b *testing.B) {
	const maxRlpByteSize = 1024
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
	comp := wizard.Compile(definer, zkevm.FullCompilationSuite...)

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
		b.Fatal(err)
	}
	logrus.WithField("constraints", scs.GetNbConstraints()).Info("circuit compiled")

}
