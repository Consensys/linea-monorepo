package invalidity

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	invalidityPI "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
)

// CreateTxHashLimbs splits a 32-byte hash into 16 big-endian 2-byte limbs.
func CreateTxHashLimbs(b [32]byte) [16]field.Element {
	var limbs [16]field.Element
	for i := 0; i < 16; i++ {
		limbs[i].SetBytes([]byte{b[i*2], b[i*2+1]})
	}
	return limbs
}

// CreateFromLimbs splits a 20-byte address into 10 big-endian 2-byte limbs.
func CreateFromLimbs(b [20]byte) [10]field.Element {
	var limbs [10]field.Element
	for i := 0; i < 10; i++ {
		limbs[i].SetBytes([]byte{b[i*2], b[i*2+1]})
	}
	return limbs
}

// MockZkevmPI creates a minimal ZkEvm with limb-based public inputs for BadPrecompileCircuit.
// It registers columns and public inputs matching the InvalidityPIExtractor layout,
// then proves them with the provided input values.
func MockZkevmPI(rng *rand.Rand, in invalidityPI.Inputs) (*wizard.CompiledIOP, wizard.Proof) {
	define := func(b *wizard.Builder) {
		var (
			stateRootHashCols [8]ifaces.Column
			txHashCols        [16]ifaces.Column
			fromAddressCols   [10]ifaces.Column
		)

		for i := 0; i < 8; i++ {
			stateRootHashCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("STATE_ROOT_HASH_%d", i), 1, true)
		}

		for i := 0; i < 16; i++ {
			txHashCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("TX_HASH_%d", i), 1, true)
		}

		for i := 0; i < 10; i++ {
			fromAddressCols[i] = b.CompiledIOP.InsertProof(0, ifaces.ColIDf("FROM_ADDRESS_%d", i), 1, true)
		}

		hasBadPrecompileCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), 1, true)
		nbL2LogsCol := b.CompiledIOP.InsertProof(0, ifaces.ColID("NB_L2_LOGS_COL"), 1, true)

		extractor := &invalidityPI.InvalidityPIExtractor{}

		for i := 0; i < 8; i++ {
			extractor.StateRootHash[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("StateRootHash_BE_%d", i),
				accessors.NewFromPublicColumn(stateRootHashCols[i], 0),
			)
		}

		for i := 0; i < 16; i++ {
			extractor.TxHash[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("TxHash_BE_%d", i),
				accessors.NewFromPublicColumn(txHashCols[i], 0),
			)
		}

		for i := 0; i < 10; i++ {
			extractor.FromAddress[i] = b.CompiledIOP.InsertPublicInput(
				fmt.Sprintf("From_BE_%d", i),
				accessors.NewFromPublicColumn(fromAddressCols[i], 0),
			)
		}

		extractor.HasBadPrecompile = b.CompiledIOP.InsertPublicInput("HasBadPrecompile", accessors.NewFromPublicColumn(hasBadPrecompileCol, 0))
		extractor.NbL2Logs = b.CompiledIOP.InsertPublicInput("NbL2Logs", accessors.NewFromPublicColumn(nbL2LogsCol, 0))

		b.CompiledIOP.ExtraData[invalidityPI.InvalidityPIExtractorMetadata] = extractor
	}

	comp := wizard.Compile(define, dummy.Compile)

	prove := func(run *wizard.ProverRuntime) {
		for i := 0; i < 8; i++ {
			run.AssignColumn(ifaces.ColIDf("STATE_ROOT_HASH_%d", i), smartvectors.NewConstant(in.StateRootLimbs[i], 1))
		}

		for i := 0; i < 16; i++ {
			run.AssignColumn(ifaces.ColIDf("TX_HASH_%d", i), smartvectors.NewConstant(in.TxHashLimbs[i], 1))
		}

		for i := 0; i < 10; i++ {
			run.AssignColumn(ifaces.ColIDf("FROM_ADDRESS_%d", i), smartvectors.NewConstant(in.FromLimbs[i], 1))
		}

		var hasBadPrecompileVal field.Element
		if in.CaseInputs.HasBadPrecompile {
			one := field.One()
			hasBadPrecompileVal = field.PseudoRand(rng)
			hasBadPrecompileVal.Add(&hasBadPrecompileVal, &one)
		} else {
			hasBadPrecompileVal = field.Zero()
		}

		run.AssignColumn(ifaces.ColID("HASH_BAD_PRECOMPILE_COL"), smartvectors.NewConstant(hasBadPrecompileVal, 1))
		run.AssignColumn(ifaces.ColID("NB_L2_LOGS_COL"), smartvectors.NewConstant(field.NewElement(uint64(in.CaseInputs.NumL2Logs)), 1))
	}

	proof := wizard.Prove(comp, prove)

	return comp, proof
}
