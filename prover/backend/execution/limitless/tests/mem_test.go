package limitless

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/zkevm"

	"github.com/shirou/gopsutil/v3/process"
)

// getRSSBytes returns current process RSS in bytes.
func getRSSBytes() uint64 {
	p, _ := process.NewProcess(int32(os.Getpid()))
	mem, _ := p.MemoryInfo()
	return mem.RSS
}

func toGiB(b uint64) float64 {
	return float64(b) / (1 << 30)
}

// ---- Test for GL modules ----
func TestMemUseAllCompGL(t *testing.T) {
	assetsDir := "/home/ubuntu/linea-monorepo/prover/prover-assets/6.0.3/mainnet/execution-limitless"
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("failed to read assets dir: %v", err)
	}

	var glModules []string
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "dw-compiled-gl-") && strings.HasSuffix(name, ".bin") {
			base := strings.TrimSuffix(strings.TrimPrefix(name, "dw-compiled-gl-"), ".bin")
			glModules = append(glModules, base)
		}
	}

	cfg, err := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml")
	if err != nil {
		t.Fatalf("failed to read config file: %s", err)
	}

	dw := &distributed.DistributedWizard{
		CompiledGLs: make([]*distributed.RecursedSegmentCompilation, len(glModules)),
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	baseTotalAlloc := ms.TotalAlloc
	baseAlloc := ms.Alloc
	baseRSS := getRSSBytes()
	maxRSS := baseRSS

	for i, mod := range glModules {
		// snapshot before
		runtime.ReadMemStats(&ms)
		beforeTotal := ms.TotalAlloc
		beforeAlloc := ms.Alloc
		beforeRSS := getRSSBytes()

		// load
		compGL, err := zkevm.LoadCompiledGL(cfg, distributed.ModuleName(mod))
		if err != nil {
			t.Fatalf("failed to load compiled GL %s: %v", mod, err)
		}
		dw.CompiledGLs[i] = compGL

		// snapshot after
		runtime.ReadMemStats(&ms)
		afterTotal := ms.TotalAlloc
		afterAlloc := ms.Alloc
		afterRSS := getRSSBytes()
		if afterRSS > maxRSS {
			maxRSS = afterRSS
		}

		fmt.Printf("[GL %s] TotalAlloc Δ: %.3f GiB | LiveHeap Δ: %.3f GiB | RSS Δ: %.3f GiB\n",
			mod,
			toGiB(afterTotal-beforeTotal),
			toGiB(afterAlloc-beforeAlloc),
			toGiB(afterRSS-beforeRSS),
		)
	}

	// final stats
	runtime.ReadMemStats(&ms)
	finalTotal := ms.TotalAlloc
	finalAlloc := ms.Alloc
	finalRSS := getRSSBytes()

	fmt.Printf("\n[GL Totals]\n")
	fmt.Printf("Peak RSS observed: %.3f GiB\n", toGiB(maxRSS))
	fmt.Printf("Total Go allocs (TotalAlloc Δ): %.3f GiB\n", toGiB(finalTotal-baseTotalAlloc))
	fmt.Printf("Final live Go heap (Alloc Δ): %.3f GiB\n", toGiB(finalAlloc-baseAlloc))
	fmt.Printf("Final RSS Δ: %.3f GiB\n\n", toGiB(finalRSS-baseRSS))
}

// ---- Test for LPP modules ----
func TestMemUseAllCompLPP(t *testing.T) {
	assetsDir := "/home/ubuntu/linea-monorepo/prover/prover-assets/6.0.3/mainnet/execution-limitless"
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("failed to read assets dir: %v", err)
	}

	var lppModules [][]string
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "dw-compiled-lpp-") && strings.HasSuffix(name, ".bin") {
			base := strings.TrimSuffix(strings.TrimPrefix(name, "dw-compiled-lpp-"), ".bin")
			base = strings.Trim(base, "[]")
			lppModules = append(lppModules, []string{base})
		}
	}

	cfg, err := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml")
	if err != nil {
		t.Fatalf("failed to read config file: %s", err)
	}

	dw := &distributed.DistributedWizard{
		CompiledLPPs: make([]*distributed.RecursedSegmentCompilation, len(lppModules)),
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	baseTotalAlloc := ms.TotalAlloc
	baseAlloc := ms.Alloc
	baseRSS := getRSSBytes()
	maxRSS := baseRSS

	for i, mods := range lppModules {
		// snapshot before
		runtime.ReadMemStats(&ms)
		beforeTotal := ms.TotalAlloc
		beforeAlloc := ms.Alloc
		beforeRSS := getRSSBytes()

		// load
		compLPP, err := zkevm.LoadCompiledLPP(cfg, []distributed.ModuleName{distributed.ModuleName(mods[0])})
		if err != nil {
			t.Fatalf("failed to load compiled LPP %v: %v", mods, err)
		}
		dw.CompiledLPPs[i] = compLPP

		// snapshot after
		runtime.ReadMemStats(&ms)
		afterTotal := ms.TotalAlloc
		afterAlloc := ms.Alloc
		afterRSS := getRSSBytes()
		if afterRSS > maxRSS {
			maxRSS = afterRSS
		}

		fmt.Printf("[LPP %s] TotalAlloc Δ: %.3f GiB | LiveHeap Δ: %.3f GiB | RSS Δ: %.3f GiB\n",
			mods[0],
			toGiB(afterTotal-beforeTotal),
			toGiB(afterAlloc-beforeAlloc),
			toGiB(afterRSS-beforeRSS),
		)
	}

	// final stats
	runtime.ReadMemStats(&ms)
	finalTotal := ms.TotalAlloc
	finalAlloc := ms.Alloc
	finalRSS := getRSSBytes()

	fmt.Printf("\n[LPP Totals]\n")
	fmt.Printf("Peak RSS observed: %.3f GiB\n", toGiB(maxRSS))
	fmt.Printf("Total Go allocs (TotalAlloc Δ): %.3f GiB\n", toGiB(finalTotal-baseTotalAlloc))
	fmt.Printf("Final live Go heap (Alloc Δ): %.3f GiB\n", toGiB(finalAlloc-baseAlloc))
	fmt.Printf("Final RSS Δ: %.3f GiB\n\n", toGiB(finalRSS-baseRSS))
}

// --- Measure per GL module ---
func TestMemUsePerCompGL(t *testing.T) {
	assetsDir := "/home/ubuntu/linea-monorepo/prover/prover-assets/6.0.3/mainnet/execution-limitless"
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("failed to read assets dir: %v", err)
	}

	var glModules []string
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "dw-compiled-gl-") && strings.HasSuffix(name, ".bin") {
			base := strings.TrimSuffix(strings.TrimPrefix(name, "dw-compiled-gl-"), ".bin")
			glModules = append(glModules, base)
		}
	}

	cfg, _ := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml")
	dw := &distributed.DistributedWizard{
		CompiledGLs: make([]*distributed.RecursedSegmentCompilation, len(glModules)),
	}

	var prevRSS = getRSSBytes()
	maxRSS := prevRSS
	fmt.Printf("\n[GL Modules RSS Delta]\n")
	for i, mod := range glModules {
		compGL, err := zkevm.LoadCompiledGL(cfg, distributed.ModuleName(mod))
		if err != nil {
			t.Fatalf("failed to load compiled GL %s: %v", mod, err)
		}
		dw.CompiledGLs[i] = compGL

		currRSS := getRSSBytes()
		delta := currRSS - prevRSS

		prevRSS = currRSS
		if currRSS > maxRSS {
			maxRSS = currRSS
		}
		fmt.Printf("GL[%s]: ΔRSS = %.3f  Peak RSS = %.3f GiB\n", mod, toGiB(delta), toGiB(maxRSS))
	}
}

// --- Measure per LPP module ---
func TestMemUsePerCompLPP(t *testing.T) {
	assetsDir := "/home/ubuntu/linea-monorepo/prover/prover-assets/6.0.3/mainnet/execution-limitless"
	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("failed to read assets dir: %v", err)
	}

	var lppModules [][]string
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "dw-compiled-lpp-") && strings.HasSuffix(name, ".bin") {
			base := strings.TrimSuffix(strings.TrimPrefix(name, "dw-compiled-lpp-"), ".bin")
			base = strings.Trim(base, "[]")
			lppModules = append(lppModules, []string{base})
		}
	}

	cfg, _ := config.NewConfigFromFileUnchecked("/home/ubuntu/linea-monorepo/prover/config/config-mainnet-limitless.toml")
	dw := &distributed.DistributedWizard{
		CompiledLPPs: make([]*distributed.RecursedSegmentCompilation, len(lppModules)),
	}

	var prevRSS = getRSSBytes()
	maxRSS := prevRSS
	fmt.Printf("\n[LPP Modules RSS Delta]\n")
	for i, mods := range lppModules {
		compLPP, err := zkevm.LoadCompiledLPP(cfg, []distributed.ModuleName{distributed.ModuleName(mods[0])})
		if err != nil {
			t.Fatalf("failed to load compiled LPP %v: %v", mods, err)
		}
		dw.CompiledLPPs[i] = compLPP

		currRSS := getRSSBytes()
		delta := currRSS - prevRSS

		prevRSS = currRSS
		if currRSS > maxRSS {
			maxRSS = currRSS
		}
		fmt.Printf("LPP[%s]: ΔRSS = %.3f GiB, Peak RSS = %.3f GiB\n", mods[0], toGiB(delta), toGiB(maxRSS))
	}
}
