package assets

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	utils_limitless "github.com/consensys/linea-monorepo/prover/utils/limitless"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

type distributeTestCase struct {
	numRow int
}

func (d distributeTestCase) define(comp *wizard.CompiledIOP) {

	// Define the first module
	a0 := comp.InsertCommit(0, "a0", d.numRow)
	b0 := comp.InsertCommit(0, "b0", d.numRow)
	c0 := comp.InsertCommit(0, "c0", d.numRow)

	// Importantly, the second module must be slightly different than the first
	// one because else it will create a wierd edge case in the conglomeration:
	// as we would have two GL modules with the same verifying key and we would
	// not be able to infer a module from a VK.
	//
	// We differentiate the modules by adding a duplicate constraints for GL0
	a1 := comp.InsertCommit(0, "a1", d.numRow)
	b1 := comp.InsertCommit(0, "b1", d.numRow)
	c1 := comp.InsertCommit(0, "c1", d.numRow)

	comp.InsertGlobal(0, "global-0", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-duplicate", symbolic.Sub(c0, b0, a0))
	comp.InsertGlobal(0, "global-1", symbolic.Sub(c1, b1, a1))

	comp.InsertInclusion(0, "inclusion-0", []ifaces.Column{c0, b0, a0}, []ifaces.Column{c1, b1, a1})
}

// GetDistWizard initializes a distributed wizard configuration using the
// ZkEVM's compiled IOP and a StandardModuleDiscoverer with preset parameters.
// The function compiles the necessary segments and produces a conglomerated
// distributed wizard, which is then returned.
func GetDistWizard() *distributed.DistributedWizard {
	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   utils_limitless.GetAffinities(zkevm),
			Predivision:  1,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(zkevm.WizardIOP, disc).
				CompileSegments().
				Conglomerate(20)
	)

	return distWizard
}

func GetBasicDistWizard() *distributed.DistributedWizard {

	var (
		numRow = 1 << 10
		tc     = distributeTestCase{numRow: numRow}
		disc   = &distributed.StandardModuleDiscoverer{
			TargetWeight: 3 * numRow,
			Predivision:  1,
		}
		comp = wizard.Compile(func(build *wizard.Builder) {
			tc.define(build.CompiledIOP)
		})

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(comp, disc).
				CompileSegments().
				Conglomerate(20)
	)

	return distWizard
}

func TestSerdeDistWizard(t *testing.T) {
	dist := GetDistWizard()

	t.Run("Bootstrapper", func(t *testing.T) {
		runSerdeTest(t, dist.Bootstrapper, "DistributedWizard.Bootstrapper", true)
	})

	t.Run("Discoverer", func(t *testing.T) {
		runSerdeTest(t, dist.Disc, "DistributedWizard.Discoverer", true)
	})

	t.Run("CompiledDefault", func(t *testing.T) {
		runSerdeTest(t, dist.CompiledDefault, "DistributedWizard.CompiledDefault", true)
	})

	for i := range dist.CompiledGLs {
		t.Run(fmt.Sprintf("CompiledGL-%v", i), func(t *testing.T) {
			runSerdeTest(t, dist.CompiledGLs[i], fmt.Sprintf("DistributedWizard.CompiledGL-%v", i), true)
		})
	}

	for i := range dist.CompiledLPPs {
		t.Run(fmt.Sprintf("CompiledLPP-%v", i), func(t *testing.T) {
			runSerdeTest(t, dist.CompiledLPPs[i], fmt.Sprintf("DistributedWizard.CompiledLPP-%v", i), true)
		})
	}

	// To save memory
	cong := dist.CompiledConglomeration
	dist = nil
	runtime.GC()

	runSerdeTest(t, cong, "DistributedWizard.CompiledConglomeration", true)
}

func TestSerdeDWCong(t *testing.T) {
	distWizard := GetDistWizard()
	cong := distWizard.CompiledConglomeration
	distWizard = nil
	runtime.GC()
	unsafeBinDump(t, cong, true)
}

func unsafeBinDump(t *testing.T, cong *distributed.ConglomeratorCompilation, sanityCheck bool) {
	var buffer bytes.Buffer

	// Serialize
	serTime := profiling.TimeIt(func() {
		// Write architecture marker
		err := unsafe.WriteMarker(&buffer)
		if err != nil {
			t.Fatalf("could not write marker: %v", err)
		}
		// Serialize the slice
		err = unsafe.WriteSlice(&buffer, []*distributed.ConglomeratorCompilation{cong})
		if err != nil {
			t.Fatalf("could not marshal array of compiled IOP: %v", err)
		}
	})

	// Write to file using buffer's data
	err := utils.WriteToFile("dw-cong-dump.bin", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("could not write to file: %v", err)
	}

	// Create a new buffer for deserialization
	var readBuffer bytes.Buffer
	var deCong []*distributed.ConglomeratorCompilation

	// Deserialize
	deserTime := profiling.TimeIt(func() {
		// Read from file into readBuffer
		err = utils.ReadFromFile("cong-dump.bin", &readBuffer)
		if err != nil {
			t.Fatalf("could not read from file: %v", err)
		}
		// Verify architecture marker
		err = unsafe.ReadMarker(&readBuffer)
		if err != nil {
			t.Fatalf("could not read marker: %v", err)
		}
		// Deserialize the slice
		deCong, _, err = unsafe.ReadSlice[[]*distributed.ConglomeratorCompilation](&readBuffer)
		if err != nil {
			t.Fatalf("could not unmarshal array of compiled IOP: %v", err)
		}
	})

	t.Logf("%s serialization=%v deserialization=%v buffer-size=%v \n", "conglomeration", serTime, deserTime, readBuffer.Len())
	t.Logf("(ser)   No. of rounds in cong.WIOP:%d \n", cong.Wiop.NumRounds())
	t.Logf("(deser) No. of rounds in cong.WIOP:%d \n", deCong[0].Wiop.NumRounds())
	t.Logf("(ser)   QueriesParams mapping inner length in cong.WIOP:%d \n", len(cong.Wiop.QueriesParams.Mapping.InnerMap))
	t.Logf("(deser) QueriesParams mapping inner length in cong.WIOP:%d \n", len(deCong[0].Wiop.QueriesParams.Mapping.InnerMap))

	if sanityCheck {
		t.Logf("Running sanity check on ser/de cong. object")
		if !test_utils.CompareExportedFields(cong, deCong[0]) {
			t.Fatalf("Ser/de conglomerator compilation are not equal")
		}
	}
}
