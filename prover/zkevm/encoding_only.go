package zkevm

import (
	"fmt"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

var (
	encodeOnlyZkevm            *ZkEvm
	onceEncodeOnlyZkevm        = sync.Once{}
	encodeOnlyCompilationSuite = compilationSuite{
		mimc.CompileMiMC,
		specialqueries.RangeProof,
		lookup.CompileLogDerivative,
	}
)

func EncodeOnlyZkEvm(tl *config.TracesLimits) *ZkEvm {
	onceEncodeOnlyZkevm.Do(func() {
		encodeOnlyZkevm = fullZKEVMWithSuite(tl, encodeOnlyCompilationSuite)
	})

	return encodeOnlyZkevm
}

func (z *ZkEvm) AssignAndEncodeInFile(filepath string, input *Witness) {
	run := wizard.ProverOnlyFirstRound(z.WizardIOP, z.prove(input))
	t := time.Now()
	f := files.MustOverwrite(filepath)
	fmt.Printf("[%v] encoding the assignment\n", time.Now())
	b := serialization.SerializeAssignment(run.Columns)
	fmt.Printf("[%v] writing the encoded assignment\n", time.Now())
	f.Write(b)
	f.Close()
	fmt.Printf("[%v] blob total size %v bytes, took %v sec to encode and write\n", time.Now(), len(b), time.Since(t).Seconds())
}
