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
	"github.com/sirupsen/logrus"
)

// expandUser expands ~ to the current user's home dir

func TestMemUsageCompiledMods(t *testing.T) {
	// Path to the prover assets directory (adjust if needed)
	assetsDir := "/home/ubuntu/linea-monorepo/prover/prover-assets/6.0.3/mainnet/execution-limitless"

	files, err := os.ReadDir(assetsDir)
	if err != nil {
		t.Fatalf("failed to read assets dir: %v", err)
	}

	var glModules []string
	var lppModules [][]string

	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, "dw-compiled-gl-") && strings.HasSuffix(name, ".bin") {
			base := strings.TrimSuffix(strings.TrimPrefix(name, "dw-compiled-gl-"), ".bin")
			glModules = append(glModules, base)
		}
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

	var beforeGL, afterGL, afterLPP runtime.MemStats

	// Measure memory before GL
	runtime.GC()
	runtime.ReadMemStats(&beforeGL)

	// Load all GL modules
	for _, mod := range glModules {
		logrus.Infof("Loading compiled GL module %s", mod)
		_, err := zkevm.LoadCompiledGL(cfg, distributed.ModuleName(mod)) // pass cfg if needed
		if err != nil {
			t.Fatalf("failed to load compiled GL %s: %v", mod, err)
		}
	}

	// Measure memory after GL
	runtime.GC()
	runtime.ReadMemStats(&afterGL)

	// Load all LPP modules
	for _, mods := range lppModules {
		logrus.Infof("Loading compiled LPP module %s", mods)
		_, err := zkevm.LoadCompiledLPP(cfg, []distributed.ModuleName{distributed.ModuleName(mods[0])}) // pass cfg if needed
		if err != nil {
			t.Fatalf("failed to load compiled LPP %v: %v", mods, err)
		}
	}

	// Measure memory after LPP
	runtime.GC()
	runtime.ReadMemStats(&afterLPP)

	// Convert to GiB
	toGiB := func(b uint64) float64 {
		return float64(b) / (1024 * 1024 * 1024)
	}

	glDelta := toGiB(afterGL.Alloc - beforeGL.Alloc)
	lppDelta := toGiB(afterLPP.Alloc - afterGL.Alloc)
	totalDelta := toGiB(afterLPP.Alloc - beforeGL.Alloc)

	fmt.Printf("Memory usage delta (GiB):\n")
	fmt.Printf("  Compiled GL:   %.3f GiB\n", glDelta)
	fmt.Printf("  Compiled LPP:  %.3f GiB\n", lppDelta)
	fmt.Printf("  Total:         %.3f GiB\n", totalDelta)
}
