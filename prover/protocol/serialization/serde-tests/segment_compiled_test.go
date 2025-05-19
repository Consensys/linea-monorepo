package serdetests

import (
	"encoding/json"
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
)

// rawRecursedSegmentCompilation represents the serialized form of RecursedSegmentCompilation.
type rawRecursedSegmentCompilation struct {
	ModuleGL      json.RawMessage `json:"moduleGL"`
	ModuleLPP     json.RawMessage `json:"moduleLPP"`
	DefaultModule json.RawMessage `json:"defaultModule"`
	Recursion     json.RawMessage `json:"recursion"`
	RecursionComp json.RawMessage `json:"recursionComp"`
}

var recursionSegComp = dw.CompiledGLs[0]

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
	// if segComp.Recursion != nil {
	// 	recSer, err := serialization.SerializeRecursion(segcomp.Recursion)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to serialize Recursion: %w", err)
	// 	}
	// 	raw.Recursion = recSer
	// } else {
	// 	raw.Recursion = []byte(serialization.NilString)
	// }

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
