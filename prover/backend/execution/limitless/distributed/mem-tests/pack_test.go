package limitless

import (
	"fmt"
	"sort"
	"testing"
)

// worker represents a single GL or LPP worker.
type worker struct {
	modules []string
	memUsed float64 // GiB
}

// greedyPack packs modules into workers based on a memory limit (GiB).
func greedyPack(modules map[string]float64, memLimit float64) []worker {
	// Sort modules by decreasing memory usage
	type kv struct {
		name string
		mem  float64
	}
	var list []kv
	for k, v := range modules {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].mem > list[j].mem
	})

	var workers []worker
	for _, mod := range list {
		placed := false
		// Try to fit into existing workers
		for i := range workers {
			if workers[i].memUsed+mod.mem <= memLimit {
				workers[i].modules = append(workers[i].modules, mod.name)
				workers[i].memUsed += mod.mem
				placed = true
				break
			}
		}
		// Otherwise, create a new worker
		if !placed {
			workers = append(workers, worker{
				modules: []string{mod.name},
				memUsed: mod.mem,
			})
		}
	}
	return workers
}

// TestOptimalWorkerPacking prints an optimal GL/LPP worker assignment
func TestOptimalWorkerPacking(t *testing.T) {
	// GL module ΔRSS in GiB (from your measurements)
	glModules := map[string]float64{
		"ARITH-OPS":       31.660,
		"ECDSA":           12.534,
		"ELLIPTIC_CURVES": 18.235,
		"G2_CHECK":        6.132,
		"HUB-KECCAK":      17.549,
		"MODEXP_256":      10.682,
		"MODEXP_4096":     128.443,
		"SHA2":            17.179,
		"STATIC":          0.958,
		"TINY-STUFFS":     0.012,
	}
	// LPP module ΔRSS in GiB
	lppModules := map[string]float64{
		"ARITH-OPS":       28.905,
		"ECDSA":           7.683,
		"ELLIPTIC_CURVES": 11.068,
		"G2_CHECK":        10.103,
		"HUB-KECCAK":      12.525,
		"MODEXP_256":      1.714,
		"MODEXP_4096":     11.999,
		"SHA2":            17.098,
		"STATIC":          17.679,
		"TINY-STUFFS":     0.000,
	}

	// Effective memory limits
	glLimit := 128.0 // GiB for subproof production
	lppLimit := 64.0 // GiB for LPP workers

	fmt.Println("\n=== GL Workers ===")
	glWorkers := greedyPack(glModules, glLimit)
	for i, w := range glWorkers {
		fmt.Printf("Worker %d: memUsed=%.3f GiB, modules=%v\n", i+1, w.memUsed, w.modules)
	}

	fmt.Println("\n=== LPP Workers ===")
	lppWorkers := greedyPack(lppModules, lppLimit)
	for i, w := range lppWorkers {
		fmt.Printf("Worker %d: memUsed=%.3f GiB, modules=%v\n", i+1, w.memUsed, w.modules)
	}
}
