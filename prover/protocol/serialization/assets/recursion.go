package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
)

// rawRecursion represents the serialized form of Recursion
type rawRecursion struct {
	Name             string            `json:"name"`
	Subscript        string            `json:"subscript"`
	InputCompiledIOP json.RawMessage   `json:"inputCompiledIOP"`
	Round            int               `json:"round"`
	PlonkCtx         json.RawMessage   `json:"plonkCtx"`
	PcsCtx           []json.RawMessage `json:"pcsCtx"`
}

// SerializeRecursion serializes a Recursion struct
func SerializeRecursion(rec *recursion.Recursion) ([]byte, error) {
	if rec == nil {
		return []byte(serialization.NilString), nil
	}

	raw := rawRecursion{
		Name:      rec.Name,
		Subscript: rec.Subscript,
		Round:     rec.Round,
		PcsCtx:    make([]json.RawMessage, len(rec.PcsCtx)),
	}

	// Handle InputCompiledIOP
	if rec.InputCompiledIOP != nil {
		iopSer, err := serialization.SerializeCompiledIOP(rec.InputCompiledIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize InputCompiledIOP: %w", err)
		}
		raw.InputCompiledIOP = iopSer
	} else {
		raw.InputCompiledIOP = []byte(serialization.NilString)
	}

	// Handle PlonkCtx
	// if rec.plonkCtx != nil {
	// 	ctxSer, err := SerializePlonkCompilationCtx(rec.PlonkCtx)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to serialize PlonkCtx: %w", err)
	// 	}
	// 	raw.PlonkCtx = ctxSer
	// } else {
	// 	raw.PlonkCtx = []byte(serialization.NilString)
	// }

	// Handle PcsCtx
	if rec.PcsCtx != nil {
		raw.PcsCtx = make([]json.RawMessage, len(rec.PcsCtx))
		for i, ctx := range rec.PcsCtx {
			if ctx != nil {
				ctxSer, err := serialization.SerializeValue(reflect.ValueOf(ctx), serialization.DeclarationMode)
				if err != nil {
					return nil, fmt.Errorf("failed to serialize PcsCtx[%d]: %w", i, err)
				}
				raw.PcsCtx[i] = ctxSer
			} else {
				raw.PcsCtx[i] = []byte(serialization.NilString)
			}
		}
	}

	return serialization.SerializeAnyWithCborPkg(&raw)
}

// DeserializeRecursion deserializes into a Recursion struct
func DeserializeRecursion(data []byte) (*recursion.Recursion, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawRecursion
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize raw Recursion: %w", err)
	}

	rec := &recursion.Recursion{
		Name:      raw.Name,
		Subscript: raw.Subscript,
		Round:     raw.Round,
		PcsCtx:    make([]*vortex.Ctx, len(raw.PcsCtx)),
	}

	// Handle InputCompiledIOP
	if !bytes.Equal(raw.InputCompiledIOP, []byte(serialization.NilString)) {
		iop, err := serialization.DeserializeCompiledIOP(raw.InputCompiledIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize InputCompiledIOP: %w", err)
		}
		rec.InputCompiledIOP = iop
	}

	// Handle PlonkCtx
	// if !bytes.Equal(raw.PlonkCtx, []byte(serialization.NilString)) {
	// 	ctx, err := deserializeCompilationCtx(raw.PlonkCtx)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to deserialize PlonkCtx: %w", err)
	// 	}
	// 	rec.PlonkCtx = ctx
	// }

	// Handle PcsCtx
	if raw.PcsCtx != nil {
		rec.PcsCtx = make([]*vortex.Ctx, len(raw.PcsCtx))
		for i, ctxData := range raw.PcsCtx {
			if !bytes.Equal(ctxData, []byte(serialization.NilString)) {
				comp := serialization.NewEmptyCompiledIOP()
				ctxVal, err := serialization.DeserializeValue(ctxData, serialization.DeclarationMode, reflect.TypeOf(vortex.Ctx{}), comp)
				if err != nil {
					return nil, fmt.Errorf("failed to deserialize PcsCtx[%d]: %w", i, err)
				}
				rec.PcsCtx[i] = ctxVal.Interface().(*vortex.Ctx)
			}
		}
	}

	return rec, nil
}

// serializeCompilationCtx serializes CompilationCtx excluding Plonk
// func SerializePlonkCompilationCtx(ctx *plonkinternal.CompilationCtx) (json.RawMessage, error) {
// 	if ctx == nil {
// 		return []byte(serialization.NilString), nil
// 	}

// 	raw := rawCompilationCtx{
// 		Name:           ctx.name,
// 		Subscript:      ctx.subscript,
// 		Round:          ctx.round,
// 		MaxNbInstances: ctx.maxNbInstances,
// 	}

// 	// Serialize comp
// 	if ctx.comp != nil {
// 		compSer, err := serialization.SerializeCompiledIOP(ctx.comp)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to serialize Comp: %w", err)
// 		}
// 		raw.Comp = compSer
// 	} else {
// 		raw.Comp = []byte(serialization.NilString)
// 	}

// 	// Serialize Columns
// 	columnsSer, err := serialization.SerializeValue(reflect.ValueOf(ctx.Columns), serialization.DeclarationMode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to serialize Columns: %w", err)
// 	}
// 	raw.Columns = columnsSer

// 	// Serialize RangeCheckOption
// 	rangeCheckSer, err := serialization.SerializeValue(reflect.ValueOf(ctx.RangeCheckOption), serialization.DeclarationMode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to serialize RangeCheckOption: %w", err)
// 	}
// 	raw.RangeCheckOption = rangeCheckSer

// 	// Serialize FixedNbRowsOption
// 	fixedRowsSer, err := serialization.SerializeValue(reflect.ValueOf(ctx.FixedNbRowsOption), serialization.DeclarationMode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to serialize FixedNbRowsOption: %w", err)
// 	}
// 	raw.FixedNbRowsOption = fixedRowsSer

// 	// Serialize ExternalHasherOption
// 	hasherSer, err := serialization.SerializeValue(reflect.ValueOf(ctx.ExternalHasherOption), serialization.DeclarationMode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to serialize ExternalHasherOption: %w", err)
// 	}
// 	raw.ExternalHasherOption = hasherSer

// 	return serialization.SerializeAnyWithCborPkg(&raw)
// }
