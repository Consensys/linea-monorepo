package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// rawDistWizard represents the serialized form of DistributedWizard.
type rawDistWizard struct {
	ModuleNames            json.RawMessage `json:"moduleNames"`
	LPPs                   json.RawMessage `json:"lpps"`
	GLs                    json.RawMessage `json:"gls"`
	DefaultModule          json.RawMessage `json:"defaultModule"`
	Bootstrapper           json.RawMessage `json:"bootstrapper"`
	Disc                   json.RawMessage `json:"disc"`
	CompiledGLs            json.RawMessage `json:"compiledGLs"`
	CompiledLPPs           json.RawMessage `json:"compiledLPPs"`
	CompiledDefault        json.RawMessage `json:"compiledDefault"`
	CompiledConglomeration json.RawMessage `json:"compiledConglomeration"`
}

// SerializeDistWizard serializes a DistributedWizard instance field-by-field.
func SerializeDistWizard(dw *distributed.DistributedWizard) ([]byte, error) {
	if dw == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawDistWizard{}

	// Serialize ModuleNames
	if dw.ModuleNames != nil {
		mnSer, err := serialization.SerializeValue(reflect.ValueOf(dw.ModuleNames), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleNames: %w", err)
		}
		raw.ModuleNames = mnSer
	} else {
		raw.ModuleNames = []byte(serialization.NilString)
	}

	// Serialize LPPs
	if dw.LPPs != nil {
		lppsSer, err := SerializeModuleLPPs(dw.LPPs)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize LPPs: %w", err)
		}
		raw.LPPs = lppsSer
	} else {
		raw.LPPs = []byte(serialization.NilString)
	}

	// Serialize GLs
	if dw.GLs != nil {
		glsSer, err := SerializeModuleGLs(dw.GLs)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize GLs: %w", err)
		}
		raw.GLs = glsSer
	} else {
		raw.GLs = []byte(serialization.NilString)
	}

	// Serialize DefaultModule
	if dw.DefaultModule != nil {
		defSer, err := SerializeDWDefMods(dw.DefaultModule)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize DefaultModule: %w", err)
		}
		raw.DefaultModule = defSer
	} else {
		raw.DefaultModule = []byte(serialization.NilString)
	}

	// Serialize Bootstrapper
	if dw.Bootstrapper != nil {
		bootSer, err := serialization.SerializeCompiledIOP(dw.Bootstrapper)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Bootstrapper: %w", err)
		}
		raw.Bootstrapper = bootSer
	} else {
		raw.Bootstrapper = []byte(serialization.NilString)
	}

	// Serialize Disc
	if dw.Disc != nil {
		discSer, err := SerializeDisc(dw.Disc)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Disc: %w", err)
		}
		raw.Disc = discSer
	} else {
		raw.Disc = []byte(serialization.NilString)
	}

	/*
		// Serialize CompiledGLs
		if dw.CompiledGLs != nil {
			cglsSer, err := SerializeRecursedSegmentCompilations(dw.CompiledGLs)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize CompiledGLs: %w", err)
			}
			raw.CompiledGLs = cglsSer
		} else {
			raw.CompiledGLs = []byte(serialization.NilString)
		}

		// Serialize CompiledLPPs
		if dw.CompiledLPPs != nil {
			clppsSer, err := SerializeRecursedSegmentCompilations(dw.CompiledLPPs)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize CompiledLPPs: %w", err)
			}
			raw.CompiledLPPs = clppsSer
		} else {
			raw.CompiledLPPs = []byte(serialization.NilString)
		}

		// Serialize CompiledDefault
		if dw.CompiledDefault != nil {
			cdefSer, err := SerializeRecursedSegmentCompilation(dw.CompiledDefault)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize CompiledDefault: %w", err)
			}
			raw.CompiledDefault = cdefSer
		} else {
			raw.CompiledDefault = []byte(serialization.NilString)
		}

		// Serialize CompiledConglomeration
		if dw.CompiledConglomeration != nil {
			ccongSer, err := SerializeConglomeratorCompilation(dw.CompiledConglomeration)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize CompiledConglomeration: %w", err)
			}
			raw.CompiledConglomeration = ccongSer
		} else {
			raw.CompiledConglomeration = []byte(serialization.NilString)
		} */

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeDistWizard deserializes a DistributedWizard instance from CBOR-encoded data.
func DeserializeDistWizard(data []byte) (*distributed.DistributedWizard, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawDistWizard
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize raw DistributedWizard: %w", err)
	}

	dw := &distributed.DistributedWizard{}

	// Deserialize Bootstrapper first to use as context
	var comp *wizard.CompiledIOP
	if !bytes.Equal(raw.Bootstrapper, []byte(serialization.NilString)) {
		boot, err := serialization.DeserializeCompiledIOP(raw.Bootstrapper)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Bootstrapper: %w", err)
		}
		dw.Bootstrapper = boot
		comp = boot
	} else {
		comp = serialization.NewEmptyCompiledIOP()
	}

	// Deserialize ModuleNames
	if !bytes.Equal(raw.ModuleNames, []byte(serialization.NilString)) {
		mnVal, err := serialization.DeserializeValue(raw.ModuleNames, serialization.DeclarationMode, reflect.TypeOf([]distributed.ModuleName{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleNames: %w", err)
		}
		dw.ModuleNames = mnVal.Interface().([]distributed.ModuleName)
	}

	// Deserialize LPPs
	if !bytes.Equal(raw.LPPs, []byte(serialization.NilString)) {
		lpps, err := DeserializeModuleLPPs(raw.LPPs)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize LPPs: %w", err)
		}
		dw.LPPs = lpps
	}

	// Deserialize GLs
	if !bytes.Equal(raw.GLs, []byte(serialization.NilString)) {
		gls, err := DeserializeModuleGLs(raw.GLs)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize GLs: %w", err)
		}
		dw.GLs = gls
	}

	// Deserialize DefaultModule
	if !bytes.Equal(raw.DefaultModule, []byte(serialization.NilString)) {
		def, err := DeserializeDWDefMods(raw.DefaultModule)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize DefaultModule: %w", err)
		}
		dw.DefaultModule = def
	}

	// Deserialize Disc
	if !bytes.Equal(raw.Disc, []byte(serialization.NilString)) {
		disc, err := DeserializeDisc(raw.Disc)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Disc: %w", err)
		}
		dw.Disc = disc
	}

	/*
		// Deserialize CompiledGLs
		if !bytes.Equal(raw.CompiledGLs, []byte(serialization.NilString)) {
			cgls, err := DeserializeRecursedSegmentCompilations(raw.CompiledGLs)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize CompiledGLs: %w", err)
			}
			dw.CompiledGLs = cgls
		}

		// Deserialize CompiledLPPs
		if !bytes.Equal(raw.CompiledLPPs, []byte(serialization.NilString)) {
			clpps, err := DeserializeRecursedSegmentCompilations(raw.CompiledLPPs)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize CompiledLPPs: %w", err)
			}
			dw.CompiledLPPs = clpps
		}

		// Deserialize CompiledDefault
		if !bytes.Equal(raw.CompiledDefault, []byte(serialization.NilString)) {
			cdef, err := DeserializeRecursedSegmentCompilation(raw.CompiledDefault)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize CompiledDefault: %w", err)
			}
			dw.CompiledDefault = cdef
		}

		// Deserialize CompiledConglomeration
		if !bytes.Equal(raw.CompiledConglomeration, []byte(serialization.NilString)) {
			ccong, err := DeserializeConglomeratorCompilation(raw.CompiledConglomeration)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize CompiledConglomeration: %w", err)
			}
			dw.CompiledConglomeration = ccong
		}
	*/

	return dw, nil
}

// rawModuleLPP represents the serialized form of ModuleLPP.
type rawModuleLPP struct {
	CompiledIOP            json.RawMessage `json:"compiledIOP"`
	Disc                   json.RawMessage `json:"disc"`
	DefinitionInputs       json.RawMessage `json:"definitionInputs"`
	InitialFiatShamirState json.RawMessage `json:"initialFiatShamirState"`
	N0Hash                 json.RawMessage `json:"n0Hash"`
	N1Hash                 json.RawMessage `json:"n1Hash"`
	LogDerivativeSum       json.RawMessage `json:"logDerivativeSum"`
	GrandProduct           json.RawMessage `json:"grandProduct"`
	Horner                 json.RawMessage `json:"horner"`
}

// serializeModuleLPP serializes a single ModuleLPP instance field-by-field.
func SerializeModuleLPP(lpp *distributed.ModuleLPP) ([]byte, error) {
	if lpp == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawModuleLPP{}

	// Serialize CompiledIOP first (includes Columns store)
	comp := lpp.GetModuleTranslator().Wiop
	if comp == nil {
		return nil, fmt.Errorf("ModuleLPP has nil CompiledIOP")
	}

	compSer, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP: %w", err)
	}
	raw.CompiledIOP = compSer

	// Serialize disc
	disc := lpp.GetModuleTranslator().Disc
	serComp, err := SerializeDisc(disc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize LPP module discoverer:%w", err)
	}

	raw.Disc = serComp

	// Serialize InitialFiatShamirState
	if lpp.InitialFiatShamirState != nil {
		ifsSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.InitialFiatShamirState), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize InitialFiatShamirState: %w", err)
		}
		raw.InitialFiatShamirState = ifsSer
	} else {
		raw.InitialFiatShamirState = []byte(serialization.NilString)
	}

	// Serialize N0Hash
	if lpp.N0Hash != nil {
		n0Ser, err := serialization.SerializeValue(reflect.ValueOf(lpp.N0Hash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize N0Hash: %w", err)
		}
		raw.N0Hash = n0Ser
	} else {
		raw.N0Hash = []byte(serialization.NilString)
	}

	// Serialize N1Hash
	if lpp.N1Hash != nil {
		n1Ser, err := serialization.SerializeValue(reflect.ValueOf(lpp.N1Hash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize N1Hash: %w", err)
		}
		raw.N1Hash = n1Ser
	} else {
		raw.N1Hash = []byte(serialization.NilString)
	}

	// Serialize LogDerivativeSum
	if !reflect.ValueOf(lpp.LogDerivativeSum).IsZero() {
		ldsSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.LogDerivativeSum), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize LogDerivativeSum: %w", err)
		}
		raw.LogDerivativeSum = ldsSer
	} else {
		raw.LogDerivativeSum = []byte(serialization.NilString)
	}

	// Serialize GrandProduct
	if !reflect.ValueOf(lpp.GrandProduct).IsZero() {
		gpSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.GrandProduct), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize GrandProduct: %w", err)
		}
		raw.GrandProduct = gpSer
	} else {
		raw.GrandProduct = []byte(serialization.NilString)
	}

	// Serialize Horner
	if !reflect.ValueOf(lpp.Horner).IsZero() {
		hSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.Horner), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Horner: %w", err)
		}
		raw.Horner = hSer
	} else {
		raw.Horner = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeModuleLPP deserializes a single ModuleLPP instance field-by-field.
func DeserializeModuleLPP(data []byte) (*distributed.ModuleLPP, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawModuleLPP
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleLPP raw data: %w", err)
	}

	// Initialize ModuleLPP
	lpp := &distributed.ModuleLPP{}

	// Deserialize CompiledIOP first (includes Columns store)
	comp, err := serialization.DeserializeCompiledIOP(raw.CompiledIOP)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP: %w", err)
	}

	disc, err := DeserializeDisc(raw.Disc)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize LPP module discoverer: %w", err)
	}
	lpp.SetModuleTranslator(comp, disc)

	// Deserialize InitialFiatShamirState (depends on Columns)
	if !bytes.Equal(raw.InitialFiatShamirState, []byte(serialization.NilString)) {
		ifsVal, err := serialization.DeserializeValue(raw.InitialFiatShamirState, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize InitialFiatShamirState: %w", err)
		}
		lpp.InitialFiatShamirState = ifsVal.Interface().(ifaces.Column)
	}

	// Deserialize N0Hash (depends on Columns)
	if !bytes.Equal(raw.N0Hash, []byte(serialization.NilString)) {
		n0Val, err := serialization.DeserializeValue(raw.N0Hash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N0Hash: %w", err)
		}
		lpp.N0Hash = n0Val.Interface().(ifaces.Column)
	}

	// Deserialize N1Hash (depends on Columns)
	if !bytes.Equal(raw.N1Hash, []byte(serialization.NilString)) {
		n1Val, err := serialization.DeserializeValue(raw.N1Hash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N1Hash: %w", err)
		}
		lpp.N1Hash = n1Val.Interface().(ifaces.Column)
	}

	// Deserialize LogDerivativeSum
	if !bytes.Equal(raw.LogDerivativeSum, []byte(serialization.NilString)) {
		// comp = lpp.GetModuleTranslator().Wiop
		ldsVal, err := serialization.DeserializeValue(raw.LogDerivativeSum, serialization.DeclarationMode, reflect.TypeOf(query.LogDerivativeSum{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize LogDerivativeSum: %w", err)
		}
		lpp.LogDerivativeSum = ldsVal.Interface().(query.LogDerivativeSum)
	}

	// Deserialize GrandProduct
	if !bytes.Equal(raw.GrandProduct, []byte(serialization.NilString)) {
		gpVal, err := serialization.DeserializeValue(raw.GrandProduct, serialization.DeclarationMode, reflect.TypeOf(query.GrandProduct{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize GrandProduct: %w", err)
		}
		lpp.GrandProduct = gpVal.Interface().(query.GrandProduct)
	}

	// Deserialize Horner
	if !bytes.Equal(raw.Horner, []byte(serialization.NilString)) {
		hVal, err := serialization.DeserializeValue(raw.Horner, serialization.DeclarationMode, reflect.TypeOf(query.Horner{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Horner: %w", err)
		}
		lpp.Horner = hVal.Interface().(query.Horner)
	}

	return lpp, nil
}

// SerializeModuleLPPs serializes a slice of ModuleLPP instances.
func SerializeModuleLPPs(lpps []*distributed.ModuleLPP) ([]byte, error) {
	rawLPPs := make([]json.RawMessage, len(lpps))
	for i, lpp := range lpps {
		lppSer, err := SerializeModuleLPP(lpp)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleLPP at index %d: %w", i, err)
		}
		rawLPPs[i] = lppSer
	}
	return serialization.SerializeAnyWithCborPkg(rawLPPs)
}

// DeserializeModuleLPPs deserializes a slice of ModuleLPP instances.
func DeserializeModuleLPPs(data []byte) ([]*distributed.ModuleLPP, error) {
	var rawLPPs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawLPPs); err != nil {
		return nil, fmt.Errorf("failed to deserialize LPPs raw slice: %w", err)
	}

	lpps := make([]*distributed.ModuleLPP, len(rawLPPs))
	for i, raw := range rawLPPs {
		lpp, err := DeserializeModuleLPP(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleLPP at index %d: %w", i, err)
		}
		lpps[i] = lpp
	}
	return lpps, nil
}

// rawModuleGL represents the serialized form of ModuleGL.
type rawModuleGL struct {
	CompiledIOP              json.RawMessage `json:"compiledIOP"`
	Disc                     json.RawMessage `json:"disc"`
	DefinitionInput          json.RawMessage `json:"definitionInput"`
	IsFirst                  json.RawMessage `json:"isFirst"`
	IsLast                   json.RawMessage `json:"isLast"`
	SentValuesGlobal         json.RawMessage `json:"sentValuesGlobal"`
	SentValuesGlobalHash     json.RawMessage `json:"sentValuesGlobalHash"`
	SentValuesGlobalMap      json.RawMessage `json:"sentValuesGlobalMap"`
	ReceivedValuesGlobal     json.RawMessage `json:"receivedValuesGlobal"`
	ReceivedValuesGlobalAccs json.RawMessage `json:"receivedValuesGlobalAccs"`
	ReceivedValuesGlobalHash json.RawMessage `json:"receivedValuesGlobalHash"`
	ReceivedValuesGlobalMap  json.RawMessage `json:"receivedValuesGlobalMap"`
}

// SerializeModuleGL serializes a single ModuleGL instance field-by-field.
func SerializeModuleGL(gl *distributed.ModuleGL) ([]byte, error) {
	if gl == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawModuleGL{}

	// Serialize CompiledIOP
	comp := gl.GetModuleTranslator().Wiop
	if comp == nil {
		return nil, fmt.Errorf("ModuleGL has nil CompiledIOP")
	}

	compSer, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP: %w", err)
	}
	raw.CompiledIOP = compSer

	// Serialize Disc
	disc := gl.GetModuleTranslator().Disc
	serComp, err := SerializeDisc(disc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize GL module discoverer: %w", err)
	}
	raw.Disc = serComp

	// Serialize IsFirst
	if gl.IsFirst != nil {
		isFirstSer, err := serialization.SerializeValue(reflect.ValueOf(gl.IsFirst), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize IsFirst: %w", err)
		}
		raw.IsFirst = isFirstSer
	} else {
		raw.IsFirst = []byte(serialization.NilString)
	}

	// Serialize IsLast
	if gl.IsLast != nil {
		isLastSer, err := serialization.SerializeValue(reflect.ValueOf(gl.IsLast), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize IsLast: %w", err)
		}
		raw.IsLast = isLastSer
	} else {
		raw.IsLast = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobal
	if len(gl.SentValuesGlobal) > 0 {
		svgSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobal), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobal: %w", err)
		}
		raw.SentValuesGlobal = svgSer
	} else {
		raw.SentValuesGlobal = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobalHash
	if gl.SentValuesGlobalHash != nil {
		svgHashSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobalHash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobalHash: %w", err)
		}
		raw.SentValuesGlobalHash = svgHashSer
	} else {
		raw.SentValuesGlobalHash = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobalMap
	if len(gl.SentValuesGlobalMap) > 0 {
		svgMapSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobalMap), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobalMap: %w", err)
		}
		raw.SentValuesGlobalMap = svgMapSer
	} else {
		raw.SentValuesGlobalMap = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobal
	if gl.ReceivedValuesGlobal != nil {
		rvgSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobal), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobal: %w", err)
		}
		raw.ReceivedValuesGlobal = rvgSer
	} else {
		raw.ReceivedValuesGlobal = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalAccs
	if len(gl.ReceivedValuesGlobalAccs) > 0 {
		rvgAccsSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalAccs), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalAccs: %w", err)
		}
		raw.ReceivedValuesGlobalAccs = rvgAccsSer
	} else {
		raw.ReceivedValuesGlobalAccs = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalHash
	if gl.ReceivedValuesGlobalHash != nil {
		rvgHashSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalHash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalHash: %w", err)
		}
		raw.ReceivedValuesGlobalHash = rvgHashSer
	} else {
		raw.ReceivedValuesGlobalHash = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalMap
	if len(gl.ReceivedValuesGlobalMap) > 0 {
		rvgMapSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalMap), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalMap: %w", err)
		}
		raw.ReceivedValuesGlobalMap = rvgMapSer
	} else {
		raw.ReceivedValuesGlobalMap = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeModuleGL deserializes a single ModuleGL instance field-by-field.
func DeserializeModuleGL(data []byte) (*distributed.ModuleGL, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawModuleGL
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleGL raw data: %w", err)
	}

	// Initialize ModuleGL
	gl := &distributed.ModuleGL{}

	// Deserialize CompiledIOP
	comp, err := serialization.DeserializeCompiledIOP(raw.CompiledIOP)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP: %w", err)
	}

	disc, err := DeserializeDisc(raw.Disc)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize GL module discoverer: %w", err)
	}
	gl.SetModuleTranslator(comp, disc)

	// Deserialize IsFirst
	if !bytes.Equal(raw.IsFirst, []byte(serialization.NilString)) {
		isFirstVal, err := serialization.DeserializeValue(raw.IsFirst, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize IsFirst: %w", err)
		}
		gl.IsFirst = isFirstVal.Interface().(ifaces.Column)
	}

	// Deserialize IsLast
	if !bytes.Equal(raw.IsLast, []byte(serialization.NilString)) {
		isLastVal, err := serialization.DeserializeValue(raw.IsLast, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize IsLast: %w", err)
		}
		gl.IsLast = isLastVal.Interface().(ifaces.Column)
	}

	// Deserialize SentValuesGlobal
	if !bytes.Equal(raw.SentValuesGlobal, []byte(serialization.NilString)) {
		svgVal, err := serialization.DeserializeValue(raw.SentValuesGlobal, serialization.DeclarationMode, reflect.TypeOf([]query.LocalOpening{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobal: %w", err)
		}
		gl.SentValuesGlobal = svgVal.Interface().([]query.LocalOpening)
	}

	// Deserialize SentValuesGlobalHash
	if !bytes.Equal(raw.SentValuesGlobalHash, []byte(serialization.NilString)) {
		svgHashVal, err := serialization.DeserializeValue(raw.SentValuesGlobalHash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobalHash: %w", err)
		}
		gl.SentValuesGlobalHash = svgHashVal.Interface().(ifaces.Column)
	}

	// Deserialize SentValuesGlobalMap
	if !bytes.Equal(raw.SentValuesGlobalMap, []byte(serialization.NilString)) {
		svgMapVal, err := serialization.DeserializeValue(raw.SentValuesGlobalMap, serialization.DeclarationMode, reflect.TypeOf(map[string]int{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobalMap: %w", err)
		}
		gl.SentValuesGlobalMap = svgMapVal.Interface().(map[string]int)
	}

	// Deserialize ReceivedValuesGlobal
	if !bytes.Equal(raw.ReceivedValuesGlobal, []byte(serialization.NilString)) {
		rvgVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobal, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobal: %w", err)
		}
		gl.ReceivedValuesGlobal = rvgVal.Interface().(ifaces.Column)
	}

	// Deserialize ReceivedValuesGlobalAccs
	if !bytes.Equal(raw.ReceivedValuesGlobalAccs, []byte(serialization.NilString)) {
		rvgAccsVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalAccs, serialization.DeclarationMode, reflect.TypeOf([]ifaces.Accessor{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalAccs: %w", err)
		}
		gl.ReceivedValuesGlobalAccs = rvgAccsVal.Interface().([]ifaces.Accessor)
	}

	// Deserialize ReceivedValuesGlobalHash
	if !bytes.Equal(raw.ReceivedValuesGlobalHash, []byte(serialization.NilString)) {
		rvgHashVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalHash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalHash: %w", err)
		}
		gl.ReceivedValuesGlobalHash = rvgHashVal.Interface().(ifaces.Column)
	}

	// Deserialize ReceivedValuesGlobalMap
	if !bytes.Equal(raw.ReceivedValuesGlobalMap, []byte(serialization.NilString)) {
		rvgMapVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalMap, serialization.DeclarationMode, reflect.TypeOf(map[string]int{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalMap: %w", err)
		}
		gl.ReceivedValuesGlobalMap = rvgMapVal.Interface().(map[string]int)
	}

	return gl, nil
}

// SerializeModuleGLs serializes a slice of ModuleGL instances.
func SerializeModuleGLs(gls []*distributed.ModuleGL) ([]byte, error) {
	rawGLs := make([]json.RawMessage, len(gls))
	for i, gl := range gls {
		glSer, err := SerializeModuleGL(gl)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleGL at index %d: %w", i, err)
		}
		rawGLs[i] = glSer
	}
	return serialization.SerializeAnyWithCborPkg(rawGLs)
}

// DeserializeModuleGLs deserializes a slice of ModuleGL instances.
func DeserializeModuleGLs(data []byte) ([]*distributed.ModuleGL, error) {
	var rawGLs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawGLs); err != nil {
		return nil, fmt.Errorf("failed to deserialize GLs raw slice: %w", err)
	}

	gls := make([]*distributed.ModuleGL, len(rawGLs))
	for i, raw := range rawGLs {
		gl, err := DeserializeModuleGL(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleGL at index %d: %w", i, err)
		}
		gls[i] = gl
	}
	return gls, nil
}

type rawDWDefMod struct {
	Cols        json.RawMessage `json:"cols"`
	CompiledIOP json.RawMessage `json:"compiledIOP"`
}

func SerializeDWDefMods(dm *distributed.DefaultModule) ([]byte, error) {
	// Serialize attributes individually
	cols, compIOP := dm.Column, dm.Wiop

	serCol, err := serialization.SerializeValue(reflect.ValueOf(cols), serialization.DeclarationMode)
	if err != nil {
		return nil, err
	}

	serCompIOP, err := serialization.SerializeCompiledIOP(compIOP)
	if err != nil {
		return nil, err
	}

	serDefmod := rawDWDefMod{
		Cols:        serCol,
		CompiledIOP: serCompIOP,
	}

	return serialization.SerializeAnyWithCborPkg(serDefmod)
}

func DeserializeDWDefMods(data []byte) (*distributed.DefaultModule, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var rawDWDM rawDWDefMod
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawDWDM); err != nil {
		return nil, err
	}

	dm := &distributed.DefaultModule{}

	// Deserialize columns first
	comp := serialization.NewEmptyCompiledIOP()
	cols, err := serialization.DeserializeValue(rawDWDM.Cols, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
	if err != nil {
		return nil, err
	}

	dm.Column = cols.Interface().(ifaces.Column)

	iop, err := serialization.DeserializeCompiledIOP(rawDWDM.CompiledIOP)
	if err != nil {
		return nil, err
	}

	dm.Wiop = iop

	return dm, nil
}
