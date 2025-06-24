package assets

import (
	"bytes"
	"fmt"
	"runtime"
	"slices"
	"testing"

	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	utils_limitless "github.com/consensys/linea-monorepo/prover/utils/limitless"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

// GetDistributed initializes a distributed wizard configuration using the
// ZkEVM's compiled IOP and a StandardModuleDiscoverer with preset parameters.
// The function compiles the necessary segments and produces a conglomerated
// distributed wizard, which is then returned.
func GetDistributed() *distributed.DistributedWizard {

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
				Conglomerate(5)
	)

	return distWizard
}

func TestSerdeDistWizard(t *testing.T) {
	dist := GetDistributed()

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

func TestSerdeCongWIOP(t *testing.T) {

	var (
		numRow = 1 << 10
		tc     = DistributeTestCase{numRow: numRow}
		disc   = &distributed.StandardModuleDiscoverer{
			TargetWeight: 3 * numRow,
			Predivision:  1,
		}
		comp = wizard.Compile(func(build *wizard.Builder) {
			tc.Define(build.CompiledIOP)
		})

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(comp, disc).
				CompileSegments().
				Conglomerate(20)
	)

	// runSerdeTest(t, comp, "initialWizard.CompiledIOP", true) => PASS
	// runSerdeTest(t, distWizard.CompiledConglomeration.Wiop, "DistributedWizard.CompiledConglomeration.WIOP", true)

	buffer := &bytes.Buffer{}
	iop := distWizard.CompiledConglomeration.Wiop

	serTime := profiling.TimeIt(func() {
		err := unsafe.WriteSlice(buffer, []*wizard.CompiledIOP{iop})
		if err != nil {
			t.Fatalf("could not marshal array of compiled IOP: %v", err)
		}
	})

	var iop2 []*wizard.CompiledIOP
	var err error
	deserTime := profiling.TimeIt(func() {
		readBuf := bytes.NewReader(buffer.Bytes())
		iop2, _, err = unsafe.ReadSlice[[]*wizard.CompiledIOP](readBuf)
		if err != nil {
			t.Fatalf("could not unmarshal array of compiled IOP: %v", err)
		}
	})

	t.Logf("%s serialization=%v deserialization=%v buffer-size=%v \n", "conglomeration", serTime, deserTime, len(buffer.Bytes()))
	t.Logf("Running sanity check on ser/de cong. object")
	if !test_utils.CompareExportedFields(iop, iop2[0]) {
		t.Fatalf("IOPs are not equal")
	}
}

type DistributeTestCase struct {
	numRow int
}

func (d DistributeTestCase) Define(comp *wizard.CompiledIOP) {

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

func TestUnsafeSerde(t *testing.T) {
	var buf bytes.Buffer

	// Slice with pointers (strings)
	data := []string{"hello", "unsafe", "world!"}

	err := unsafe.WriteSlice(&buf, data)
	if err != nil {
		t.Fatalf("could not marshal array of strings: %v", err)
	}

	readBuf := bytes.NewReader(buf.Bytes())
	data2, _, err := unsafe.ReadSlice[[]string](readBuf)
	if err != nil {
		t.Fatalf("could not unmarshal array of strings: %v", err)
	}

	if !slices.Equal(data, data2) {
		t.Fatalf("strings are not equal")
	}
}
