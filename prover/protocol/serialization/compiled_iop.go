package serialization

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

type RawPrecomputed struct {
	_         struct{}             `cbor:"toarray"`
	ColID     ifaces.ColID         `cbor:"k"`
	ColAssign ifaces.ColAssignment `cbor:"v"`
}

type RawPublicInput struct {
	_    struct{}        `cbor:"toarray"`
	Name string          `cbor:"n"`
	Acc  ifaces.Accessor `cbor:"a"`
}
type RawExtraData struct {
	_     struct{} `cbor:"toarray"`
	Key   string   `cbor:"k"`
	Value any      `cbor:"v"`
}

var (
	typeofRawPrecomputedData = reflect.TypeOf(RawPrecomputed{})
	typeofRawPublicInput     = reflect.TypeOf(RawPublicInput{})
	typeofRawExtraData       = reflect.TypeOf(RawExtraData{})
	typeofPcsCtxs            = reflect.TypeOf((*vortex.Ctx)(nil))
)

// RawCompiledIOP represents the serialized form of CompiledIOP.
type RawCompiledIOP struct {
	DummyCompiled          bool `cbor:"j"`
	NumRounds              int  `cbor:"r"`
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
		NumRounds:              numRounds,
		SelfRecursionCount:     comp.SelfRecursionCount,
		DummyCompiled:          comp.DummyCompiled,
		WithStorePointerChecks: comp.WithStorePointerChecks,
		FiatShamirSetup:        comp.FiatShamirSetup.BigInt(fieldToSmallBigInt(comp.FiatShamirSetup)),

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

// packActionValue packs a single action value which can be either a struct or a slice.
// If the original value was a nil pointer, this returns (zero, nil, true) indicating the caller should skip
// (same behaviour as original continue).
// pathFmt should be a format string that accepts one %d for the index when producing error paths.
func (s *Serializer) packActionValue(val any, pathFmt string, idx int) (PackedStructObject, *serdeError, bool) {
	v := reflect.ValueOf(val)
	// Handle nil pointer early: preserve behavior of original `continue`
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return PackedStructObject{}, nil, true
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		obj, err := s.PackStructObject(v)
		if err != nil {
			return PackedStructObject{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}
		return obj, nil, false
	case reflect.Slice:
		obj, err := s.PackArrayOrSlice(v)
		if err != nil {
			return PackedStructObject{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}
		return obj, nil, false
	default:
		return PackedStructObject{}, newSerdeErrorf("invalid action type: %v, expected struct or slice", v.Type()).wrapPath(fmt.Sprintf(pathFmt, idx)), false
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
	s.PackedObject.CompiledIOPFast = append(s.PackedObject.CompiledIOPFast, RawCompiledIOP{})

	rawCompiledIOP := initRawCompiledIOP(comp)
	// ------- Non-round specific data

	// Marshal precomputed data
	{
		for idx, colID := range comp.Precomputed.ListAllKeys() {
			preComputed := RawPrecomputed{ColID: colID, ColAssign: comp.Precomputed.MustGet(colID)}
			packedPrecomputedObj, err := s.PackStructObject(reflect.ValueOf(preComputed))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-precomputed)")
			}
			rawCompiledIOP.Precomputed[idx] = packedPrecomputedObj
		}
	}

	// Marshall pcsctx
	{
		if comp.PcsCtxs != nil {
			pcsAny, err := s.PackInterface(reflect.ValueOf(comp.PcsCtxs))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-pcs-ctx)")
			}
			rawCompiledIOP.PcsCtxs = pcsAny
		}
	}

	// Marshall public inputs
	{
		for idx, pubInp := range comp.PublicInputs {
			rawPubInp := RawPublicInput{Name: pubInp.Name, Acc: pubInp.Acc}
			packedPubInp, err := s.PackStructObject(reflect.ValueOf(rawPubInp))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-public-inputs)")
			}
			rawCompiledIOP.PublicInputs[idx] = packedPubInp
		}
	}

	// Marshall extra data
	{

		mapIdx := 0
		for key, extraDataAny := range comp.ExtraData {
			rawExtraData := RawExtraData{Key: key, Value: extraDataAny}
			packedExtraData, err := s.PackStructObject(reflect.ValueOf(rawExtraData))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-extra-data)")
			}
			rawCompiledIOP.ExtraData[mapIdx] = packedExtraData
			mapIdx++
		}
	}

	// ------- Round specific data

	{
		backRefCol, err := s.PackStore(comp.Columns)
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-columns-store)")
		}
		rawCompiledIOP.Columns = backRefCol

		// Outer loop - round
		for round := 0; round < rawCompiledIOP.NumRounds; round++ {
			// Pack coins faster
			coinNames := comp.Coins.AllKeysAt(round)
			rawCompiledIOP.Coins[round] = make([]BackReference, len(coinNames))
			for i, coinName := range coinNames {
				c := comp.Coins.Data(coinName)
				backRefCoin, err := s.PackCoin(c)
				if err != nil {
					return 0, err.wrapPath("(compiled-IOP-coins)")
				}
				rawCompiledIOP.Coins[round][i] = backRefCoin
			}

			// Pack query Params faster
			queriesParams := comp.QueriesParams.AllKeysAt(round)
			rawCompiledIOP.QueriesParams[round] = make([]BackReference, len(queriesParams))
			for i, query := range queriesParams {
				backRefQuery, err := s.PackQuery(comp.QueriesParams.Data(query))
				if err != nil {
					return 0, err.wrapPath("(compiled-IOP-queries-params)")
				}
				rawCompiledIOP.QueriesParams[round][i] = backRefQuery
			}

			// Pack query NoParams faster
			queriesNoParams := comp.QueriesNoParams.AllKeysAt(round)
			rawCompiledIOP.QueriesNoParams[round] = make([]BackReference, len(queriesNoParams))
			for i, query := range queriesNoParams {
				backRefQuery, err := s.PackQuery(comp.QueriesNoParams.Data(query))
				if err != nil {
					return 0, err.wrapPath("(compiled-IOP-queries-no-params)")
				}
				rawCompiledIOP.QueriesNoParams[round][i] = backRefQuery
			}

			// Pack subProverActions faster
			proverActions := comp.SubProvers.GetOrEmpty(round)
			rawCompiledIOP.Subprovers[round] = make([]PackedStructObject, len(proverActions))
			for i, subProverAction := range proverActions {
				obj, serr, skipped := s.packActionValue(subProverAction, "(compiled-IOP-subprovers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				rawCompiledIOP.Subprovers[round][i] = obj
			}

			// Pack verifierActions faster
			verifierActions := comp.SubVerifiers.GetOrEmpty(round)
			rawCompiledIOP.SubVerifiers[round] = make([]PackedStructObject, len(verifierActions))
			for i, verifierAction := range verifierActions {
				obj, serr, skipped := s.packActionValue(verifierAction, "(compiled-IOP-subverifiers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				rawCompiledIOP.SubVerifiers[round][i] = obj
			}

			// Pack FSHookPreSampling faster
			hookActions := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
			rawCompiledIOP.FiatShamirHooksPreSampling[round] = make([]PackedStructObject, len(hookActions))
			for i, hookAction := range hookActions {
				obj, serr, skipped := s.packActionValue(hookAction, "(compiled-IOP-fiatshamirhooks-pre-sampling[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				rawCompiledIOP.FiatShamirHooksPreSampling[round][i] = obj
			}
		}
	}

	s.PackedObject.CompiledIOPFast[refIdx] = *rawCompiledIOP
	return BackReference(refIdx), nil
}

func newEmptyCompiledIOP(rawCompIOP RawCompiledIOP) *wizard.CompiledIOP {
	return &wizard.CompiledIOP{
		DummyCompiled:              rawCompIOP.DummyCompiled,
		SelfRecursionCount:         rawCompIOP.SelfRecursionCount,
		WithStorePointerChecks:     rawCompIOP.WithStorePointerChecks,
		Columns:                    column.NewStore(),
		QueriesParams:              wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		QueriesNoParams:            wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		Coins:                      wizard.NewRegister[coin.Name, coin.Info](),
		SubProvers:                 collection.VecVec[wizard.ProverAction]{},
		SubVerifiers:               collection.VecVec[wizard.VerifierAction]{},
		FiatShamirHooksPreSampling: collection.VecVec[wizard.VerifierAction]{},
		Precomputed:                collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		PublicInputs:               make([]wizard.PublicInput, len(rawCompIOP.PublicInputs), cap(rawCompIOP.PublicInputs)),
		ExtraData:                  make(map[string]any, len(rawCompIOP.ExtraData)),
		PcsCtxs:                    nil,
	}
}

func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (reflect.Value, *serdeError) {

	if v < 0 || int(v) >= len(d.PackedObject.CompiledIOP) {
		return reflect.Value{}, newSerdeErrorf("invalid compiled-IOP backreference: %v", v)
	}

	// Cache Hit first => Already deserialized
	if d.CompiledIOPsFast[v] != nil {
		return reflect.ValueOf(d.CompiledIOPsFast[v]), nil
	}

	// Unpack the raw compiled iop
	packedRawCompiledIOP := d.PackedObject.CompiledIOPFast[v]

	// Reserve the cache
	deCompIOP := newEmptyCompiledIOP(packedRawCompiledIOP)
	d.CompiledIOPsFast[v] = deCompIOP

	// Start unpacking the raw compiled iop

	// Unmarshall Field.Element
	{
		f, err := unmarshalBigInt(nil, packedRawCompiledIOP.FiatShamirSetup, TypeOfBigInt)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser compiled-IOP-fiatshamirsetup)")
		}
		deCompIOP.FiatShamirSetup.SetBigInt(f.Interface().(*big.Int))
	}

	// Unpack column store
	{
		storeVal, err := d.UnpackStore(packedRawCompiledIOP.Columns)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser. compiled-IOP-columns)")
		}
		deCompIOP.Columns = storeVal.Interface().(*column.Store)
	}

	// Non-round specific data

	// Unmarshall precomputed data
	{
		for _, rawPrecomp := range packedRawCompiledIOP.Precomputed {
			res, err := d.UnpackStructObject(rawPrecomp, typeofRawPrecomputedData)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw pre-computed data: %w", err)
			}

			precomp, ok := res.Interface().(RawPrecomputed)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to RawPrecomputed: %w", err)
			}

			deCompIOP.Precomputed.InsertNew(precomp.ColID, precomp.ColAssign)
		}
	}

	// Unmarshall pcsctx
	{
		val, err := d.UnpackInterface(packedRawCompiledIOP.PcsCtxs, typeofPcsCtxs)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser compiled-IOP-pcsctxs)")
		}

		var ok bool
		deCompIOP.PcsCtxs, ok = val.Interface().(*vortex.Ctx)
		if !ok {
			return reflect.Value{}, newSerdeErrorf("could not cast to *vortex.Ctx: %w", err)
		}
	}

	// Unmarshall public input
	{
		for idx, rawPubInp := range packedRawCompiledIOP.PublicInputs {
			res, err := d.UnpackStructObject(rawPubInp, typeofRawPublicInput)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw public input: %w", err)
			}

			pubInp, ok := res.Interface().(RawPublicInput)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw public input: %w", err)
			}
			deCompIOP.PublicInputs[idx] = wizard.PublicInput{
				Name: pubInp.Name,
				Acc:  pubInp.Acc,
			}
		}
	}

	// Unmarshall extradata
	{
		for _, rawExtraData := range packedRawCompiledIOP.ExtraData {
			res, err := d.UnpackStructObject(rawExtraData, typeofRawExtraData)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw extra data: %w", err)
			}
			extraData, ok := res.Interface().(RawExtraData)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw extra data: %w", err)
			}
			deCompIOP.ExtraData[extraData.Key] = extraData.Value
		}
	}

	// ------- Round specific data
	{
		for round := 0; round < packedRawCompiledIOP.NumRounds; round++ {

			// Unpack coins
			{
				for _, backRef := range packedRawCompiledIOP.Coins[round] {
					val, err := d.UnpackCoin(backRef)
					if err != nil {
						return reflect.Value{}, err.wrapPath("(deser compiled-IOP-coins)")
					}

					coinInfo, ok := val.Interface().(*coin.Info)
					if !ok {
						return reflect.Value{}, newSerdeErrorf("could not cast to *coin.Info: %w", err)
					}
					deCompIOP.Coins.AddToRound(round, coinInfo.Name, *coinInfo)
				}
			}

		}
	}

}
