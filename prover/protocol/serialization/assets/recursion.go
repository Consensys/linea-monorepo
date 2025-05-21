package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/internal/plonkinternal"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
)

// rawRecursion represents the serialized form of recursion.Recursion.
type rawRecursion struct {
	Name             string
	Subscript        string
	InputCompiledIOP json.RawMessage
	Round            int
	PlonkCtx         json.RawMessage
	PcsCtx           []json.RawMessage
}

// SerializeRecursion serializes a recursion.Recursion instance field-by-field.
func SerializeRecursion(r *recursion.Recursion) ([]byte, error) {
	if r == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawRecursion{
		Name:      r.Name,
		Subscript: r.Subscript,
		Round:     r.Round,
	}

	// Serialize InputCompiledIOP
	if r.InputCompiledIOP != nil {
		iopSer, err := serialization.SerializeCompiledIOP(r.InputCompiledIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize InputCompiledIOP: %w", err)
		}
		raw.InputCompiledIOP = iopSer
	} else {
		raw.InputCompiledIOP = []byte(serialization.NilString)
	}

	// Serialize PlonkCtx
	if r.PlonkCtx != nil {
		//plonkSer, err := SerializeCompilationCtx(r.PlonkCtx)
		plonkSer, err := serialization.SerializeValue(reflect.ValueOf(r.PlonkCtx), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize PlonkCtx: %w", err)
		}
		raw.PlonkCtx = plonkSer
	} else {
		raw.PlonkCtx = []byte(serialization.NilString)
	}

	// Serialize PcsCtx
	if r.PcsCtx != nil {
		raw.PcsCtx = make([]json.RawMessage, len(r.PcsCtx))
		for i, ctx := range r.PcsCtx {
			if ctx != nil {
				// ctxSer, err := SerializeVortexCtx(ctx)
				ctxSer, err := serialization.SerializeValue(reflect.ValueOf(ctx), serialization.DeclarationMode)
				if err != nil {
					return nil, fmt.Errorf("failed to serialize PcsCtx[%d]: %w", i, err)
				}
				raw.PcsCtx[i] = ctxSer
			} else {
				raw.PcsCtx[i] = []byte(serialization.NilString)
			}
		}
	} else {
		raw.PcsCtx = nil
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeRecursion deserializes a recursion.Recursion instance from CBOR-encoded data.
func DeserializeRecursion(data []byte) (*recursion.Recursion, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawRecursion
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize raw Recursion: %w", err)
	}

	r := &recursion.Recursion{
		Name:      raw.Name,
		Subscript: raw.Subscript,
		Round:     raw.Round,
	}

	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize InputCompiledIOP
	if !bytes.Equal(raw.InputCompiledIOP, []byte(serialization.NilString)) {
		iop, err := serialization.DeserializeCompiledIOP(raw.InputCompiledIOP)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize InputCompiledIOP: %w", err)
		}
		r.InputCompiledIOP = iop
		comp = iop
	}

	// Deserialize PlonkCtx
	if !bytes.Equal(raw.PlonkCtx, []byte(serialization.NilString)) {
		// plonk, err := DeserializeCompilationCtx(raw.PlonkCtx)
		plonk, err := serialization.DeserializeValue(raw.PlonkCtx, serialization.DeclarationMode, reflect.TypeOf(&plonkinternal.CompilationCtx{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize PlonkCtx: %w", err)
		}
		r.PlonkCtx = plonk.Interface().(*plonkinternal.CompilationCtx)
	}

	// Deserialize PcsCtx
	if raw.PcsCtx != nil {
		r.PcsCtx = make([]*vortex.Ctx, len(raw.PcsCtx))
		for i, ctxSer := range raw.PcsCtx {
			if !bytes.Equal(ctxSer, []byte(serialization.NilString)) {
				//ctx, err := DeserializeVortexCtx(ctxSer)
				ctx, err := serialization.DeserializeValue(ctxSer, serialization.DeclarationMode, reflect.TypeOf(&vortex.Ctx{}), comp)
				if err != nil {
					return nil, fmt.Errorf("failed to deserialize PcsCtx[%d]: %w", i, err)
				}
				r.PcsCtx[i] = ctx.Interface().(*vortex.Ctx)
			}
		}
	}

	return r, nil
}
