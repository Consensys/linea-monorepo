// Package common provides a simple code generator for the gpu/plonk2 packages.
//
// It reads Go text/template files, executes them with curve-specific data,
// formats the output with go/format, and writes the result to the target directory.
package common

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator executes templates and writes formatted Go source files.
type Generator struct {
	generatedBy string
}

// New creates a Generator that stamps each output file with the given generatedBy label.
func New(generatedBy string) *Generator {
	return &Generator{generatedBy: generatedBy}
}

// Execute renders templateSrc with data, formats the result as Go source, and
// writes it to outputPath. The directory is created if it does not exist.
func (g *Generator) Execute(outputPath string, templateSrc string, data any) error {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"mul":     func(a, b int) int { return a * b },
		"add":     func(a, b int) int { return a + b },
	}).Parse(templateSrc)
	if err != nil {
		return fmt.Errorf("parse template for %s: %w", outputPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template for %s: %w", outputPath, err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		// Emit unformatted source for debugging.
		_ = os.WriteFile(outputPath+".broken", buf.Bytes(), 0o644)
		return fmt.Errorf("format source for %s: %w\n\nunformatted written to %s.broken", outputPath, err, outputPath)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(outputPath), err)
	}

	return os.WriteFile(outputPath, src, 0o644)
}
