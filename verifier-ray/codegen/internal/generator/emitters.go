package generator

import (
	"fmt"
	"strings"

	codegen "github.com/consensys/linea-monorepo/verifier-ray/codegen"
)

// emitVanishingSystem writes the vanishing.System constant and its supporting
// local declarations (expression arrays, bucket arrays, module array) into w.
// The file-level imports for `field` and `vanishing` are emitted by the
// caller (Generate); this function does not re-emit them.
// vanishingSystemIndex is the template index used when emitting the single
// vanishing system. It determines the generated Zig name (system_N).
const vanishingSystemIndex = 0

// vanishingSystemName is the Zig identifier that emitVanishingSystem emits for
// the vanishing system constant. Generate uses this name in the systems literal.
const vanishingSystemName = "system_0"

func emitVanishingSystem(w *Writer, sys codegen.VanishingSystem) error {
	var buf strings.Builder
	opts := codegen.VanishingZigOptions{
		// Field/vanishing imports are already in the file header.
		FieldImport:     `@import("../field/koalabear.zig")`,
		VanishingImport: `@import("../query/vanishing.zig")`,
		EmitHeader:      false,
		EmitSystemsList: false,
	}
	named := codegen.NamedVanishingSystem{
		Name:   sys.SourceName,
		System: sys,
	}
	if err := codegen.WriteVanishingSystemZigWithOptions(&buf, vanishingSystemIndex, named, opts); err != nil {
		return fmt.Errorf("emit vanishing system: %w", err)
	}
	w.Raw(buf.String())
	return nil
}
