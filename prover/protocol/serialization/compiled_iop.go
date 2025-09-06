package serialization

import (
	"bytes"
	"math/big"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// RawCompiledIOP represents the serialized form of CompiledIOP.
type RawCompiledIOP struct {
	DummyCompiled          bool `cbor:"j"`
	WithStorePointerChecks bool `cbor:"o"`
	SelfRecursionCount     int  `cbor:"k"`

	FiatShamirSetup *big.Int `cbor:"l"`

	Columns         BackReference     `cbor:"a"`
	QueriesParams   [][]BackReference `cbor:"b"`
	QueriesNoParams [][]BackReference `cbor:"c"`
	Coins           [][]BackReference `cbor:"d"`

	Subprovers                 [][]PackedStructObject `cbor:"e"`
	SubVerifiers               [][]PackedStructObject `cbor:"f"`
	FiatShamirHooksPreSampling [][]PackedStructObject `cbor:"g"`
	Precomputed                []PackedStructObject   `cbor:"h"`
	PublicInputs               []PackedStructObject   `cbor:"m"`
	ExtraData                  []PackedStructObject   `cbor:"n"`

	PcsCtxs any `cbor:"i"`
}

func initRawCompiledIOP(comp *wizard.CompiledIOP) *RawCompiledIOP {

	numRounds := comp.NumRounds()
	return &RawCompiledIOP{
		SelfRecursionCount:     comp.SelfRecursionCount,
		DummyCompiled:          comp.DummyCompiled,
		WithStorePointerChecks: comp.WithStorePointerChecks,

		FiatShamirSetup: comp.FiatShamirSetup.BigInt(fieldToSmallBigInt(comp.FiatShamirSetup)),

		// Preallocate arrays to reduce allocations
		QueriesParams:              make([][]BackReference, numRounds),
		QueriesNoParams:            make([][]BackReference, numRounds),
		Coins:                      make([][]BackReference, numRounds),
		Subprovers:                 make([][]PackedStructObject, numRounds),
		SubVerifiers:               make([][]PackedStructObject, numRounds),
		FiatShamirHooksPreSampling: make([][]PackedStructObject, numRounds),

		Precomputed:  make([]PackedStructObject, len(comp.Precomputed.ListAllKeys())),
		PublicInputs: make([]PackedStructObject, len(comp.PublicInputs)),
		ExtraData:    make([]PackedStructObject, len(comp.ExtraData)),
	}
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
	refIdx := len(s.PackedObject.CompiledIOPFast)
	s.compiledIOPsFast[comp] = refIdx
	s.PackedObject.CompiledIOPFast = append(s.PackedObject.CompiledIOPFast, nil)

	rawCompiledIOP := initRawCompiledIOP(comp)

	// Marshal precomputed data
	type rawPrecomputed struct {
		_         struct{}             `cbor:"toarray"`
		ColID     ifaces.ColID         `cbor:"k"`
		ColAssign ifaces.ColAssignment `cbor:"v"`
	}

	for idx, colID := range comp.Precomputed.ListAllKeys() {

		// This should not work becoz of interfaces
		// if err := encodeWithCBORToBuffer(&buf, preComputed); err != nil {
		// 	return 0, newSerdeErrorf(err.Error())
		// }
		// rawCompiledIOP.Precomputed[idx] = buf.Bytes()
		// buf.Reset()
		preComputed := rawPrecomputed{ColID: colID, ColAssign: comp.Precomputed.MustGet(colID)}
		packedPrecomputedObj, err := s.PackStructObject(reflect.ValueOf(preComputed))
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-precomputed)")
		}
		rawCompiledIOP.Precomputed[idx] = packedPrecomputedObj
	}

	// marshall pcsctx
	pcsAny, err := s.PackInterface(reflect.ValueOf(comp.PcsCtxs))
	if err != nil {
		return 0, err.wrapPath("(compiled-IOP-pcs-ctx)")
	}
	rawCompiledIOP.PcsCtxs = pcsAny

	// marshall public inputs
	type rawPublicInput struct {
		_    struct{}        `cbor:"toarray"`
		Name string          `cbor:"n"`
		Acc  ifaces.Accessor `cbor:"a"`
	}
	for idx, pubInp := range comp.PublicInputs {
		rawPubInp := rawPublicInput{Name: pubInp.Name, Acc: pubInp.Acc}
		packedPubInp, err := s.PackStructObject(reflect.ValueOf(rawPubInp))
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-public-inputs)")
		}
		rawCompiledIOP.PublicInputs[idx] = packedPubInp
	}

	// marshall extra data
	type rawExtraData struct {
		_     struct{} `cbor:"toarray"`
		Key   string   `cbor:"k"`
		Value any      `cbor:"v"`
	}
	mapIdx := 0
	for key, extraDataAny := range comp.ExtraData {
		rawExtraData := rawExtraData{Key: key, Value: extraDataAny}
		packedExtraData, err := s.PackStructObject(reflect.ValueOf(rawExtraData))
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-extra-data)")
		}
		rawCompiledIOP.ExtraData[mapIdx] = packedExtraData
		mapIdx++
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

		// Reflection is unavoidable here atm
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
			// ValueOf(subProverAction) => should pass the direct struct obj i.e. concrete type and not interface type
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

	var buf bytes.Buffer
	if err := encodeWithCBORToBuffer(&buf, rawCompiledIOP); err != nil {
		return 0, newSerdeErrorf(err.Error())
	}

	s.PackedObject.CompiledIOPFast[refIdx] = buf.Bytes()
	return BackReference(refIdx), nil
}

// func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (*wizard.CompiledIOP, *serdeError) {

// }
