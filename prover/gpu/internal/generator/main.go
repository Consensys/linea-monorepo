// Command generator generates the typed per-curve GPU packages in gpu/plonk2.
//
// Run from this directory:
//
//	go run .
package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/consensys/linea-monorepo/prover/gpu/internal/generator/common"
	"github.com/consensys/linea-monorepo/prover/gpu/internal/generator/config"
	"github.com/consensys/linea-monorepo/prover/gpu/internal/generator/plonk"
)

func main() {
	// Resolve output base relative to this file's directory so the generator
	// works correctly regardless of the working directory it is invoked from.
	_, thisFile, _, _ := runtime.Caller(0)
	thisDir := filepath.Dir(thisFile)
	plonk2Dir := filepath.Join(thisDir, "..", "..", "plonk2")

	gen := common.New("gpu/internal/generator")

	curves := []config.Curve{
		config.BN254,
		config.BLS12377,
		config.BW6761,
	}

	for _, curve := range curves {
		outputDir := filepath.Join(plonk2Dir, curve.Package)
		log.Printf("generating %s → %s", curve.Name, outputDir)
		if err := plonk.Generate(curve, outputDir, gen); err != nil {
			log.Fatalf("generate %s: %v", curve.Name, err)
		}
	}

	log.Println("done")
}
