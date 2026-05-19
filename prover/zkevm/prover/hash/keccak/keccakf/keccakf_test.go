package keccakfkoalabear

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	kcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf/iokeccakf"
	"github.com/stretchr/testify/assert"
)

func TestKeccakf(t *testing.T) {

	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	numCases := 30
	maxNumKeccakf := 2
	// The -1 is here to prevent the generation of a padding block
	maxInputBytes := maxNumKeccakf*keccak.Rate - 1

	definer, prover := keccakfTestingModule(maxNumKeccakf)
	comp := wizard.Compile(definer, dummy.Compile)

	for i := 0; i < numCases; i++ {
		// Generate a random piece of data
		dataSize := rng.IntN(maxInputBytes + 1)
		data := make([]byte, dataSize)
		utils.ReadPseudoRand(rng, data)

		// Generate permutation traces for the data
		traces := keccak.PermTraces{}
		keccak.Hash(data, &traces)

		proof := wizard.Prove(comp, prover(t, traces))
		assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")
	}
}

func keccakfTestingModule(
	maxNumKeccakf int,
) (
	define wizard.DefineFunc,
	prover func(t *testing.T, traces keccak.PermTraces) wizard.MainProverStep,
) {

	var (
		mod    = &Module{}
		size   = NumRows(maxNumKeccakf)
		blocks = make([][kcommon.NumSlices]ifaces.Column, kcommon.NumLanesInBlock)
	)

	// The testing wizard uniquely calls the keccakf module
	define = func(b *wizard.Builder) {

		comp := b.CompiledIOP
		for m := 0; m < kcommon.NumLanesInBlock; m++ {
			for z := 0; z < kcommon.NumSlices; z++ {
				blocks[m][z] = comp.InsertCommit(0, ifaces.ColIDf("BLOCK_%v_%v", m, z), size, true)
			}
		}

		mod = NewModule(b.CompiledIOP, KeccakfInputs{
			Blocks:       blocks,
			IsBlock:      comp.InsertCommit(0, "IS_BLOCK", size, true),
			IsFirstBlock: comp.InsertCommit(0, "IS_FIRST_BLOCK", size, true),
			IsBlockBaseB: comp.InsertCommit(0, "IS_BLOCK_BASEB", size, true),
			IsActive:     comp.InsertCommit(0, "IS_ACTIVE", size, true),
			KeccakfSize:  size,
		})
	}

	// And the prover (instanciated for traces) is called
	prover = func(
		t *testing.T,
		traces keccak.PermTraces,
	) wizard.MainProverStep {
		return func(run *wizard.ProverRuntime) {
			// assign the input columns
			var (
				keccakfBlocks = iokeccakf.KeccakFBlocks{
					Blocks:        mod.Inputs.Blocks,
					IsBlockActive: mod.Inputs.IsActive,
					IsBlock:       mod.Inputs.IsBlock,
					IsFirstBlock:  mod.Inputs.IsFirstBlock,
					IsBlockBaseB:  mod.Inputs.IsBlockBaseB,
					KeccakfSize:   mod.Inputs.KeccakfSize,
				}
			)

			keccakfBlocks.AssignBlocks(run, traces)
			keccakfBlocks.AssignBlockFlags(run, traces)

			// Assigns the module
			mod.Assign(run, traces)

			// Asserts that the last value in aIota is the correct one. `pos` is
			// the last active row of the module (given the traces we got). We
			// use it to reconstruct what the module "believes" to be the final
			// keccak state. Then, we compare this value with one generated in
			// the traces.
			numPerm := len(traces.KeccakFInps)
			pos := numPerm*keccak.NumRound - 1
			expectedState := traces.KeccakFOuts[numPerm-1]
			extractedState := keccak.State{}
			for x := 0; x < 5; x++ {
				for y := 0; y < 5; y++ {
					var a [8]uint8
					for z := 0; z < kcommon.NumSlices; z++ {
						v := mod.BackToThetaOrOutput.StateNext[x][y][z].GetColAssignmentAt(run, pos)
						a[z] = uint8(v.Uint64())
					}
					extractedState[x][y] = binary.LittleEndian.Uint64(a[:])
				}
			}

			assert.Equal(t, expectedState, extractedState)
		}
	}

	return define, prover
}

// ConstraintCounts holds the counts of different constraint types in the wizard.
type ConstraintCounts struct {
	Global      int
	Local       int
	Inclusion   int
	Permutation int
	Projection  int
	Range       int
	Columns     int
	Cells       int
}

// countConstraints iterates over all queries in the CompiledIOP and counts them by type.
func countConstraints(comp *wizard.CompiledIOP) ConstraintCounts {
	counts := ConstraintCounts{}

	for _, qID := range comp.QueriesNoParams.AllKeys() {
		q := comp.QueriesNoParams.Data(qID)
		switch q.(type) {
		case query.GlobalConstraint:
			counts.Global++
		case query.LocalConstraint:
			counts.Local++
		case query.Inclusion:
			counts.Inclusion++
		case query.Permutation, query.FixedPermutation:
			counts.Permutation++
		case query.Projection:
			counts.Projection++
		case query.Range:
			counts.Range++
		}
	}

	for _, colID := range comp.Columns.AllKeys() {
		counts.Columns++
		counts.Cells += comp.Columns.GetSize(colID)
	}

	return counts
}

// keccakfBenchModule creates a keccakf module for benchmarking without *testing.T dependency.
// Returns the define function, a prover step generator, and the module pointer.
func keccakfBenchModule(maxNumKeccakf int) (
	define wizard.DefineFunc,
	proverStep func(traces keccak.PermTraces) wizard.MainProverStep,
	mod *Module,
) {
	var (
		size   = NumRows(maxNumKeccakf)
		blocks = make([][kcommon.NumSlices]ifaces.Column, kcommon.NumLanesInBlock)
	)
	mod = &Module{}

	define = func(b *wizard.Builder) {
		comp := b.CompiledIOP
		for m := 0; m < kcommon.NumLanesInBlock; m++ {
			for z := 0; z < kcommon.NumSlices; z++ {
				blocks[m][z] = comp.InsertCommit(0, ifaces.ColIDf("BLOCK_%v_%v", m, z), size, true)
			}
		}

		mod = NewModule(comp, KeccakfInputs{
			Blocks:       blocks,
			IsBlock:      comp.InsertCommit(0, "IS_BLOCK", size, true),
			IsFirstBlock: comp.InsertCommit(0, "IS_FIRST_BLOCK", size, true),
			IsBlockBaseB: comp.InsertCommit(0, "IS_BLOCK_BASEB", size, true),
			IsActive:     comp.InsertCommit(0, "IS_ACTIVE", size, true),
			KeccakfSize:  size,
		})
	}

	proverStep = func(traces keccak.PermTraces) wizard.MainProverStep {
		return func(run *wizard.ProverRuntime) {
			keccakfBlocks := iokeccakf.KeccakFBlocks{
				Blocks:        mod.Inputs.Blocks,
				IsBlockActive: mod.Inputs.IsActive,
				IsBlock:       mod.Inputs.IsBlock,
				IsFirstBlock:  mod.Inputs.IsFirstBlock,
				IsBlockBaseB:  mod.Inputs.IsBlockBaseB,
				KeccakfSize:   mod.Inputs.KeccakfSize,
			}

			keccakfBlocks.AssignBlocks(run, traces)
			keccakfBlocks.AssignBlockFlags(run, traces)
			mod.Assign(run, traces)
		}
	}

	return define, proverStep, mod
}

// generateKeccakTraces generates keccak permutation traces for a given number of permutations.
func generateKeccakTraces(numKeccakf int) keccak.PermTraces {
	// Each keccakf processes Rate bytes, so we need (numKeccakf * Rate - 1) bytes
	// to trigger exactly numKeccakf permutations (the -1 avoids an extra padding block)
	dataSize := numKeccakf*keccak.Rate - 1
	data := make([]byte, dataSize)

	// Fill with deterministic pseudo-random data
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	utils.ReadPseudoRand(rng, data)

	traces := keccak.PermTraces{}
	keccak.Hash(data, &traces)
	return traces
}

// BenchmarkKeccakfCompile benchmarks the time to build the keccakf module and register
// high-level constraints (global, local, lookup/inclusion, projection, range)
// for different numbers of permutations (1K, 10K, 20K, 30K).
// This measures only the constraint registration phase, not the prover time.
func BenchmarkKeccakfCompile(b *testing.B) {
	sizes := []int{20000, 30000, 40000, 50000}

	for _, numKeccakf := range sizes {
		name := fmt.Sprintf("%dK", numKeccakf/1000)
		b.Run(name, func(b *testing.B) {
			size := NumRows(numKeccakf)

			// Run once outside timing to get constraint counts
			var counts ConstraintCounts
			{
				blocks := make([][kcommon.NumSlices]ifaces.Column, kcommon.NumLanesInBlock)
				define := func(build *wizard.Builder) {
					comp := build.CompiledIOP
					for m := 0; m < kcommon.NumLanesInBlock; m++ {
						for z := 0; z < kcommon.NumSlices; z++ {
							blocks[m][z] = comp.InsertCommit(0, ifaces.ColIDf("BLOCK_%v_%v", m, z), size, true)
						}
					}
					_ = NewModule(comp, KeccakfInputs{
						Blocks:       blocks,
						IsBlock:      comp.InsertCommit(0, "IS_BLOCK", size, true),
						IsFirstBlock: comp.InsertCommit(0, "IS_FIRST_BLOCK", size, true),
						IsBlockBaseB: comp.InsertCommit(0, "IS_BLOCK_BASEB", size, true),
						IsActive:     comp.InsertCommit(0, "IS_ACTIVE", size, true),
						KeccakfSize:  size,
					})
				}
				comp := wizard.Compile(define)
				counts = countConstraints(comp)
			}

			b.Logf("numKeccakf=%d size=%d global=%d local=%d inclusion=%d permutation=%d projection=%d range=%d columns=%d cells=%d",
				numKeccakf, size, counts.Global, counts.Local, counts.Inclusion,
				counts.Permutation, counts.Projection, counts.Range, counts.Columns, counts.Cells)

			b.ReportMetric(float64(counts.Global), "global")
			b.ReportMetric(float64(counts.Local), "local")
			b.ReportMetric(float64(counts.Inclusion), "inclusion/lookup")
			b.ReportMetric(float64(counts.Permutation), "permutation")
			b.ReportMetric(float64(counts.Projection), "projection")
			b.ReportMetric(float64(counts.Range), "range")
			b.ReportMetric(float64(counts.Columns), "columns")
			b.ReportMetric(float64(counts.Cells), "cells")
			b.ReportMetric(float64(numKeccakf), "keccakf")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				blocks := make([][kcommon.NumSlices]ifaces.Column, kcommon.NumLanesInBlock)
				define := func(build *wizard.Builder) {
					comp := build.CompiledIOP
					for m := 0; m < kcommon.NumLanesInBlock; m++ {
						for z := 0; z < kcommon.NumSlices; z++ {
							blocks[m][z] = comp.InsertCommit(0, ifaces.ColIDf("BLOCK_%v_%v", m, z), size, true)
						}
					}
					_ = NewModule(comp, KeccakfInputs{
						Blocks:       blocks,
						IsBlock:      comp.InsertCommit(0, "IS_BLOCK", size, true),
						IsFirstBlock: comp.InsertCommit(0, "IS_FIRST_BLOCK", size, true),
						IsBlockBaseB: comp.InsertCommit(0, "IS_BLOCK_BASEB", size, true),
						IsActive:     comp.InsertCommit(0, "IS_ACTIVE", size, true),
						KeccakfSize:  size,
					})
				}
				// Only runs the define function to register constraints, no compilation
				_ = wizard.Compile(define)
			}
		})
	}
}

// BenchmarkKeccakfProver benchmarks the prover time (witness assignment) for the keccakf module
// for different numbers of permutations (1K, 10K, 20K, 30K).
// Each iteration includes generateKeccakTraces (keccak.Hash over the padded message)
// plus RunProver (column assignment). No cryptographic protocol compilation.
func BenchmarkKeccakfProver(b *testing.B) {
	sizes := []int{20000, 30000, 40000, 50000}

	for _, numKeccakf := range sizes {
		name := fmt.Sprintf("%dK", numKeccakf/1000)
		b.Run(name, func(b *testing.B) {
			// Setup: compile once (no compilers)
			define, proverStep, _ := keccakfBenchModule(numKeccakf)
			comp := wizard.Compile(define) // No compilers - just constraint registration

			b.Logf("numKeccakf=%d (each iteration: generateKeccakTraces + RunProver)", numKeccakf)
			b.ReportMetric(float64(numKeccakf), "keccakf")

			b.ResetTimer()
			for range b.N {
				traces := generateKeccakTraces(numKeccakf)
				_ = wizard.RunProver(comp, proverStep(traces), false)
			}
		})
	}
}
