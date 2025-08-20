package serialization_test

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// GetDW initializes a distributed wizard configuration using the
// ZkEVM's compiled IOP and a StandardModuleDiscoverer with preset parameters.
// The function compiles the necessary segments and produces a conglomerated
// distributed wizard, which is then returned.
func GetDW(cfg *config.Config) *distributed.DistributedWizard {

	var (
		traceLimits = cfg.TracesLimits
		zkEVM       = zkevm.FullZKEVMWithSuite(&traceLimits, zkevm.CompilationSuite{}, cfg)
		disc        = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 29,
			Predivision:  1,
			Advices:      zkevm.DiscoveryAdvices,
		}
		dw = distributed.DistributeWizard(zkEVM.WizardIOP, disc)
	)

	// These are the slow and expensive operations.
	dw.CompileSegments().Conglomerate(100)
	return dw
}

func TestSerdeDW(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
	cfg, err := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-limitless.toml")
	if err != nil {
		t.Fatalf("failed to read config file: %s", err)
	}
	dist := GetDW(cfg)

	t.Run("ModuleNames", func(t *testing.T) {
		runSerdeTest(t, dist.ModuleNames, "DW.ModuleNames", true, false)
	})

	for i := range dist.GLs {
		t.Run(fmt.Sprintf("DW.GLModule-%d", i), func(t *testing.T) {
			runSerdeTest(t, dist.GLs[i], fmt.Sprintf("DW.GLModule-%d", i), true, false)
		})
	}

	for i := range dist.LPPs {
		t.Run(fmt.Sprintf("DW.LPPModule-%d", i), func(t *testing.T) {
			runSerdeTest(t, dist.LPPs[i], fmt.Sprintf("DW.LPPModule-%d", i), true, false)
		})
	}

	t.Run("DefaultModule", func(t *testing.T) {
		runSerdeTest(t, dist.DefaultModule, "DW.DefaultModule", true, false)
	})

	t.Run("Bootstrapper", func(t *testing.T) {
		runSerdeTest(t, dist.Bootstrapper, "DW.Bootstrapper", true, false)
	})

	t.Run("Discoverer", func(t *testing.T) {
		runSerdeTest(t, dist.Disc, "DW.Discoverer", true, false)
	})

	t.Run("CompiledDefault", func(t *testing.T) {
		runSerdeTest(t, dist.CompiledDefault, "DW.CompiledDefault", true, false)
	})

	for i := range dist.CompiledGLs {
		t.Run(fmt.Sprintf("CompiledGL-%v", i), func(t *testing.T) {
			runSerdeTest(t, dist.CompiledGLs[i], fmt.Sprintf("DW.CompiledGL-%v", i), true, false)
		})
	}

	for i := range dist.CompiledLPPs {
		t.Run(fmt.Sprintf("CompiledLPP-%v", i), func(t *testing.T) {
			runSerdeTest(t, dist.CompiledLPPs[i], fmt.Sprintf("DW.CompiledLPP-%v", i), true, false)
		})
	}

	// To save memory
	cong := dist.CompiledConglomeration
	dist = nil
	runtime.GC()

	t.Run("CompiledConglomeration", func(t *testing.T) {
		runSerdeTest(t, cong, "DW.CompiledConglomeration", true, false)
	})
}

// TestSerdeDWPerf: This test assumes that serialization and deserialization (ser/de) have already been
// verified to be correct in the previous test (TestSerdeDW).
func TestSerdeDWPerf(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")
	cfg, err := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-limitless.toml")
	if err != nil {
		t.Fatalf("failed to read config file: %s", err)
	}

	var perfLogs profiling.PerfLogs
	dw := GetDW(cfg)

	t.Run("ModuleNames", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.ModuleNames, "DW.ModuleNames"))
	})

	for i := range dw.GLs {
		t.Run(fmt.Sprintf("GLModule-%d", i), func(t *testing.T) {
			perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.GLs[i], fmt.Sprintf("DW.GLModule-%v", i)))
		})
	}

	for i := range dw.LPPs {
		t.Run(fmt.Sprintf("LPPModule-%d", i), func(t *testing.T) {
			perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.LPPs[i], fmt.Sprintf("DW.LPPModule-%d", i)))
		})
	}

	t.Run("DefaultModule", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.DefaultModule, "DW.DefaultModule"))
	})

	t.Run("Bootstrapper", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.Bootstrapper, "DW.Bootstrapper"))
	})

	t.Run("Discoverer", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.Disc, "DW.Discoverer"))
	})

	t.Run("CompiledDefault", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.CompiledDefault, "DW.CompiledDefault"))
	})

	for i := range dw.CompiledGLs {
		t.Run(fmt.Sprintf("CompiledGL-%v", i), func(t *testing.T) {
			perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.CompiledGLs[i], fmt.Sprintf("DW.CompiledGL-%v", i)))
		})
	}

	for i := range dw.CompiledLPPs {
		t.Run(fmt.Sprintf("CompiledLPP-%v", i), func(t *testing.T) {
			perfLogs = append(perfLogs, runSerdeTestPerf(t, dw.CompiledLPPs[i], fmt.Sprintf("DW.CompiledLPP-%v", i)))
		})
	}

	// To save memory
	cong := dw.CompiledConglomeration
	dw = nil
	runtime.GC()

	t.Run("CompiledConglomeration", func(t *testing.T) {
		perfLogs = append(perfLogs, runSerdeTestPerf(t, cong, "DW.CompiledConglomeration"))
	})

	// Write performance logs to CSV
	if err := perfLogs.WritePerformanceLogsToCSV(path.Join("perf", "dw-perf-logs.csv")); err != nil {
		t.Fatalf("Error writing performance logs to csv: %v", err)
	}
}

func BenchmarkSerdeDW(b *testing.B) {

	b.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	cfg, err := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-sepolia-limitless.toml")
	if err != nil {
		b.Fatalf("failed to read config file: %s", err)
	}
	dist := GetDW(cfg)
	b.ResetTimer()

	b.Run("ModuleNames", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, dist.ModuleNames, "DW.ModuleNames")
		}
	})

	for i := range dist.GLs {
		b.Run(fmt.Sprintf("GLModule-%d", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				runSerdeBenchmark(b, dist.GLs[i], fmt.Sprintf("DW.GLModule-%d", i))
			}
		})
	}

	for i := range dist.LPPs {
		b.Run(fmt.Sprintf("LPPModule-%d", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				runSerdeBenchmark(b, dist.LPPs[i], fmt.Sprintf("DW.LPPModule-%d", i))
			}
		})
	}

	b.Run("DefaultModule", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, dist.DefaultModule, "DW.DefaultModule")
		}
	})

	b.Run("Bootstrapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, dist.Bootstrapper, "DW.Bootstrapper")
		}
	})

	b.Run("Discoverer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, dist.Disc, "DW.Discoverer")
		}
	})

	b.Run("CompiledDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, dist.CompiledDefault, "DW.CompiledDefault")
		}
	})

	for i := range dist.CompiledGLs {

		b.Run(fmt.Sprintf("CompiledGL-%v", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				runSerdeBenchmark(b, dist.CompiledGLs[i], fmt.Sprintf("DW.CompiledGL-%v", i))
			}
		})
	}

	for i := range dist.CompiledLPPs {

		b.Run(fmt.Sprintf("CompiledLPP-%v", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				runSerdeBenchmark(b, dist.CompiledLPPs[i], fmt.Sprintf("DW.CompiledLPP-%v", i))
			}
		})
	}

	// To save memory
	cong := dist.CompiledConglomeration
	dist = nil
	runtime.GC()

	b.Run("CompiledConglomeration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runSerdeBenchmark(b, cong, "DW.CompiledConglomeration")
		}
	})
}

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
func GetBasicDW() *distributed.DistributedWizard {

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
func TestSerdeDWCong(t *testing.T) {

	t.Skipf("the test is a development/debug/integration test. It is not needed for CI")

	// Setup
	distWizard := GetBasicDW()
	cong := distWizard.CompiledConglomeration
	distWizard = nil
	runtime.GC()

	// Subtests
	tests := []struct {
		name        string
		obj         any
		sanityCheck bool
		failfast    bool
	}{
		{name: "Wiop", obj: cong.Wiop, sanityCheck: true, failfast: true},
		{name: "Recursion", obj: cong.Recursion, sanityCheck: true, failfast: true},

		// All of these tests PASS
		{name: "MaxNbProofs", obj: cong.MaxNbProofs, sanityCheck: true, failfast: false},
		{name: "DefaultWitness", obj: cong.DefaultWitness, sanityCheck: true, failfast: true},
		{name: "DefaultIops", obj: cong.DefaultIops, sanityCheck: true, failfast: true},
		{name: "PrecomputedGLVks", obj: cong.PrecomputedGLVks, sanityCheck: true, failfast: false},
		{name: "PrecomputedLPPVks", obj: cong.PrecomputedLPPVks, sanityCheck: true, failfast: false},
		{name: "VerifyingKeyColumns", obj: cong.VerifyingKeyColumns, sanityCheck: true, failfast: false},
		{name: "HolisticLookupMappedLPPPostion", obj: cong.HolisticLookupMappedLPPPostion, sanityCheck: true, failfast: false},
		{name: "HolisticLookupMappedLPPVK", obj: cong.HolisticLookupMappedLPPVK, sanityCheck: true, failfast: false},
		{name: "IsGL", obj: cong.IsGL, sanityCheck: true, failfast: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSerdeTest(t, tt.obj, tt.name, tt.sanityCheck, tt.failfast)
		})
	}
}
