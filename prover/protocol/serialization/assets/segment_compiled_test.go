package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

// rawRecursedSegmentCompilation represents the serialized form of RecursedSegmentCompilation.
type rawRecursedSegmentCompilation struct {
	ModuleGL      json.RawMessage `json:"moduleGL"`
	ModuleLPP     json.RawMessage `json:"moduleLPP"`
	DefaultModule json.RawMessage `json:"defaultModule"`
	Recursion     json.RawMessage `json:"recursion"`
	RecursionComp json.RawMessage `json:"recursionComp"`
}

var recurSegComp = dw.CompiledGLs[0]

// SerializeRecursedSegmentCompilation serializes a RecursedSegmentCompilation instance field-by-field.
func SerializeRecursedSegmentCompilation(segComp *distributed.RecursedSegmentCompilation) ([]byte, error) {
	if segComp == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawRecursedSegmentCompilation{}

	// Serialize ModuleGL
	if segComp.ModuleGL != nil {
		glSer, err := SerializeModuleGL(segComp.ModuleGL)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleGL: %w", err)
		}
		raw.ModuleGL = glSer
	} else {
		raw.ModuleGL = []byte(serialization.NilString)
	}

	// Serialize ModuleLPP
	if segComp.ModuleLPP != nil {
		lppSer, err := SerializeModuleLPP(segComp.ModuleLPP)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleLPP: %w", err)
		}
		raw.ModuleLPP = lppSer
	} else {
		raw.ModuleLPP = []byte(serialization.NilString)
	}

	// Serialize DefaultModule
	if segComp.DefaultModule != nil {
		defModSer, err := SerializeDWDefMods(segComp.DefaultModule)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize DefaultModule: %w", err)
		}
		raw.DefaultModule = defModSer
	} else {
		raw.DefaultModule = []byte(serialization.NilString)
	}

	// Serialize Recursion
	if segComp.Recursion != nil {
		fmt.Printf("Need for Serde *Recursion")
		// recSer, err := serialization.SerializeRecursion(segComp.Recursion)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to serialize Recursion: %w", err)
		// }
		// raw.Recursion = recSer
	} else {
		fmt.Printf("NOOO Need for Serde *Recursion")
		raw.Recursion = []byte(serialization.NilString)
	}

	// Serialize RecursionComp
	if segComp.RecursionComp != nil {
		compSer, err := serialization.SerializeCompiledIOP(segComp.RecursionComp)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize RecursionComp: %w", err)
		}
		raw.RecursionComp = compSer
	} else {
		raw.RecursionComp = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeRecursedSegmentCompilation deserializes a RecursedSegmentCompilation instance field-by-field.
func DeserializeRecursedSegmentCompilation(data []byte) (*distributed.RecursedSegmentCompilation, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawRecursedSegmentCompilation
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize raw RecursedSegmentCompilation: %w", err)
	}

	segComp := &distributed.RecursedSegmentCompilation{}

	// Deserialize ModuleGL
	if !bytes.Equal(raw.ModuleGL, []byte(serialization.NilString)) {
		gl, err := DeserializeModuleGL(raw.ModuleGL)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleGL: %w", err)
		}
		segComp.ModuleGL = gl
	}

	// Deserialize ModuleLPP
	if !bytes.Equal(raw.ModuleLPP, []byte(serialization.NilString)) {
		lpp, err := DeserializeModuleLPP(raw.ModuleLPP)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleLPP: %w", err)
		}
		segComp.ModuleLPP = lpp
	}

	// Deserialize DefaultModule
	if !bytes.Equal(raw.DefaultModule, []byte(serialization.NilString)) {
		defMod, err := DeserializeDWDefMods(raw.DefaultModule)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize DefaultModule: %w", err)
		}
		segComp.DefaultModule = defMod
	}

	// Deserialize Recursion
	// if !bytes.Equal(raw.Recursion, []byte(serialization.NilString)) {
	// 	rec, err := serialization.DeserializeRecursion(raw.Recursion)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to deserialize Recursion: %w", err)
	// 	}
	// 	segComp.Recursion = rec
	// }

	// Deserialize RecursionComp
	if !bytes.Equal(raw.RecursionComp, []byte(serialization.NilString)) {
		comp, err := serialization.DeserializeCompiledIOP(raw.RecursionComp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize RecursionComp: %w", err)
		}
		segComp.RecursionComp = comp
	}

	return segComp, nil
}

// TestSerdeRecursedSegmentCompilation tests serialization and deserialization of a RecursedSegmentCompilation.
func TestSerdeRecursedSegmentCompilation(t *testing.T) {
	if dw == nil {
		t.Fatal("DistributedWizard is nil")
	}

	// Serialize the RecursedSegmentCompilation
	serializedData, err := SerializeRecursedSegmentCompilation(recurSegComp)
	if err != nil {
		t.Fatalf("Failed to serialize RecursedSegmentCompilation: %v", err)
	}

	// Deserialize the RecursedSegmentCompilation
	deserializedSegComp, err := DeserializeRecursedSegmentCompilation(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize RecursedSegmentCompilation: %v", err)
	}

	// Compare exported fields
	if !test_utils.CompareExportedFields(recurSegComp, deserializedSegComp) {
		t.Errorf("Mismatch in exported fields after RecursedSegmentCompilation serde")
	}
}
