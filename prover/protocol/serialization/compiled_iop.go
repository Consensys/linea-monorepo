package serialization

import (
	"reflect"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/fxamacker/cbor/v2"
)

// rawCompiledIOP represents the serialized form of CompiledIOP.
type serCompiledIOP struct {
	DummyCompiled          bool `cbor:"j"`
	WithStorePointerChecks bool `cbor:"o"`
	SelfRecursionCount     int  `cbor:"k"`

	FiatShamirSetup field.Element `cbor:"l"`

	Columns         BackReference     `cbor:"a"`
	QueriesParams   [][]BackReference `cbor:"b"`
	QueriesNoParams [][]BackReference `cbor:"c"`
	Coins           [][]BackReference `cbor:"d"`

	Subprovers                 [][]PackedStructObject `cbor:"e"`
	SubVerifiers               [][]PackedStructObject `cbor:"f"`
	FiatShamirHooksPreSampling [][]PackedStructObject `cbor:"g"`

	Precomputed []cbor.RawMessage `cbor:"h"`
	PcsCtxs     cbor.RawMessage   `cbor:"i"`

	PublicInputs cbor.RawMessage `cbor:"m"`
	ExtraData    cbor.RawMessage `cbor:"n"`
}

func (s *Serializer) PackCompiledIOPFast(comp *wizard.CompiledIOP) (BackReference, *serdeError) {

	if comp == nil {
		return 0, nil
	}
	// Fast cache hit
	if idx, ok := s.compiledIOPs[comp]; ok {
		return BackReference(idx), nil
	}
	// Reserve slot and cache BEFORE packing to break recursion
	idx := len(s.PackedObject.CompiledIOP)
	s.compiledIOPs[comp] = idx
	s.PackedObject.CompiledIOP = append(s.PackedObject.CompiledIOP, nil)

	// Now start ser. faster
	rawCompiledIOP := serCompiledIOP{}

	// Primitives
	rawCompiledIOP.SelfRecursionCount = comp.SelfRecursionCount
	rawCompiledIOP.DummyCompiled = comp.DummyCompiled
	rawCompiledIOP.WithStorePointerChecks = comp.WithStorePointerChecks

	// Field.Element
	rawCompiledIOP.FiatShamirSetup = comp.FiatShamirSetup

	for idx, colID := range comp.Precomputed.ListAllKeys() {

		backRefColID , err :=  s.PackColumnID(colID)
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-precomputed-colID)")
		}

		colAssignment := comp.Precomputed.MustGet(colID)
		


		rawCompiledIOP.Precomputed[idx] = 
	}

	// Round specific data
	numRounds := comp.NumRounds()
	backRefCol, err := s.PackStore(comp.Columns)
	if err != nil {
		return 0, err.wrapPath("(compiled-IOP-columns-store)")
	}
	rawCompiledIOP.Columns = backRefCol

	// Outer loop - round
	for round := 0; round < numRounds; round++ {

		// Pack coins faster
		coinNames := comp.Coins.AllKeysAt(round)
		for i, coinName := range coinNames {
			c := comp.Coins.Data(coinName)
			backRefCoin, err := s.PackCoin(c)
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-coins)")
			}
			rawCompiledIOP.Coins[round][i] = backRefCoin
		}

		// Pack query Params faster
		queries := comp.QueriesParams.AllKeysAt(round)
		for i, query := range queries {
			backRefQuery, err := s.PackQuery(comp.QueriesParams.Data(query))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-queries-params)")
			}
			rawCompiledIOP.QueriesParams[round][i] = backRefQuery
		}

		// Pack query NoParams faster
		queries = comp.QueriesNoParams.AllKeysAt(round)
		for i, query := range queries {
			backRefQuery, err := s.PackQuery(comp.QueriesNoParams.Data(query))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-queries-no-params)")
			}
			rawCompiledIOP.QueriesNoParams[round][i] = backRefQuery
		}

		// Pack subProverActions faster
		proverActions := comp.SubProvers.GetOrEmpty(round)
		for i, subProverAction := range proverActions {
			// ValueOf(subProverAction) => should pass the direct struct obj
			pActionObj, err := s.PackStructObject(reflect.ValueOf(subProverAction))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-subprovers)")
			}
			rawCompiledIOP.Subprovers[round][i] = pActionObj
		}

		// Pack verifierActions faster
		verifierActions := comp.SubVerifiers.GetOrEmpty(round)
		for i, verifierAction := range verifierActions {
			// ValueOf(verifierAction) => should pass the direct struct obj
			vActionObj, err := s.PackStructObject(reflect.ValueOf(verifierAction))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-subverifiers)")
			}
			rawCompiledIOP.SubVerifiers[round][i] = vActionObj
		}

		// Pack FSHookPreSampling faster
		hookActions := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
		for i, hookAction := range hookActions {
			// ValueOf(hookAction) => should pass the direct struct obj
			hookObj, err := s.PackStructObject(reflect.ValueOf(hookAction))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-fiatshamirhooks-pre-sampling)")
			}
			rawCompiledIOP.FiatShamirHooksPreSampling[round][i] = hookObj
		}
	}

}

func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (*wizard.CompiledIOP, *serdeError) {

}
