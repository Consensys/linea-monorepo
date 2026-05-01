// Package plonk drives code generation for the gpu/plonk2 per-curve packages.
package plonk

import (
	"path/filepath"

	"github.com/consensys/linea-monorepo/prover/gpu/internal/generator/common"
	"github.com/consensys/linea-monorepo/prover/gpu/internal/generator/config"
	tmpl "github.com/consensys/linea-monorepo/prover/gpu/internal/generator/plonk/template"
)

// Generate renders all templates for curve c into outputDir.
func Generate(c config.Curve, outputDir string, gen *common.Generator) error {
	entries := []struct {
		file string
		src  string
	}{
		{filepath.Join(outputDir, "doc.go"), tmpl.DocTemplate},
		{filepath.Join(outputDir, "cgo.go"), tmpl.CgoTemplate},
		// Phase 2: FrVector
		{filepath.Join(outputDir, "fr.go"), tmpl.FrTemplate},
		{filepath.Join(outputDir, "fr_stub.go"), tmpl.FrStubTemplate},
		{filepath.Join(outputDir, "fr_test.go"), tmpl.FrTestTemplate},
		// Phase 3: FFTDomain
		{filepath.Join(outputDir, "fft.go"), tmpl.FFTTemplate},
		{filepath.Join(outputDir, "fft_stub.go"), tmpl.FFTStubTemplate},
		{filepath.Join(outputDir, "fft_test.go"), tmpl.FFTTestTemplate},
		// Phase 4: MSM
		{filepath.Join(outputDir, "msm.go"), tmpl.MSMTemplate},
		{filepath.Join(outputDir, "msm_stub.go"), tmpl.MSMStubTemplate},
		{filepath.Join(outputDir, "msm_test.go"), tmpl.MSMTestTemplate},
	}

	for _, e := range entries {
		if err := gen.Execute(e.file, e.src, c); err != nil {
			return err
		}
	}
	return nil
}
