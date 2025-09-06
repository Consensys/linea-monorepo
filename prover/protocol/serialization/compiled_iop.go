package serialization

import (
	"fmt"
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

	// Marshal precomputed data
	type rawPrecomputed struct {
		_         struct{}             `cbor:"toarray"`
		ColID     ifaces.ColID         `cbor:"k"`
		ColAssign ifaces.ColAssignment `cbor:"v"`
	}
	for idx, colID := range comp.Precomputed.ListAllKeys() {
		preComputed := rawPrecomputed{ColID: colID, ColAssign: comp.Precomputed.MustGet(colID)}
		packedPrecomputedObj, err := s.PackStructObject(reflect.ValueOf(preComputed))
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-precomputed)")
		}
		rawCompiledIOP.Precomputed[idx] = packedPrecomputedObj
	}

	// Marshall pcsctx
	if comp.PcsCtxs != nil {
		pcsAny, err := s.PackInterface(reflect.ValueOf(comp.PcsCtxs))
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-pcs-ctx)")
		}
		rawCompiledIOP.PcsCtxs = pcsAny
	}

	// Marshall public inputs
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

	// Marshall extra data
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
		rawCompiledIOP.Coins[round] = make([]BackReference, len(coinNames))
		for i, coinName := range coinNames {
			c := comp.Coins.Data(coinName)
			backRefCoin, err := s.PackCoin(c)
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-coins)")
			}
			rawCompiledIOP.Coins[round][i] = backRefCoin
		}

		// Reflection is unavoidable here atm. To avoid reflection here, we might have to
		// write a custom interface dispatcher. In essence this would entail writing a serdec
		// compiler that auto generates Pack/Unpack methods for all struct objects within the
		// implementation regsitry.

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
			v := reflect.ValueOf(subProverAction)
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					continue // Skip nil pointers
				}
				v = v.Elem() // Dereference to get the underlying value
			}
			switch v.Kind() {
			case reflect.Struct:
				pActionObj, err := s.PackStructObject(v)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-subprovers[%d])", i))
				}
				rawCompiledIOP.Subprovers[round][i] = pActionObj
			case reflect.Slice:
				pActionObj, err := s.PackArrayOrSlice(v)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-subprovers[%d])", i))
				}
				rawCompiledIOP.Subprovers[round][i] = pActionObj
			default:
				return 0, newSerdeErrorf("invalid subprover type: %v, expected struct or slice", v.Type()).wrapPath(fmt.Sprintf("(compiled-IOP-subprovers[%d])", i))
			}
		}

		// Pack verifierActions faster
		verifierActions := comp.SubVerifiers.GetOrEmpty(round)
		rawCompiledIOP.SubVerifiers[round] = make([]PackedStructObject, len(verifierActions))
		for i, verifierAction := range verifierActions {
			concreteVerifierActVal := reflect.ValueOf(verifierAction)
			if concreteVerifierActVal.Kind() == reflect.Ptr {
				if concreteVerifierActVal.IsNil() {
					continue // Skip nil pointers
				}
				concreteVerifierActVal = concreteVerifierActVal.Elem()
			}
			switch concreteVerifierActVal.Kind() {
			case reflect.Struct:
				vActionObj, err := s.PackStructObject(concreteVerifierActVal)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-subverifiers[%d])", i))
				}
				rawCompiledIOP.SubVerifiers[round][i] = vActionObj
			case reflect.Slice:
				vActionObj, err := s.PackArrayOrSlice(concreteVerifierActVal)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-subverifiers[%d])", i))
				}
				rawCompiledIOP.SubVerifiers[round][i] = vActionObj
			default:
				return 0, newSerdeErrorf("invalid verifier type: %v, expected struct or slice", concreteVerifierActVal.Type()).wrapPath(fmt.Sprintf("(compiled-IOP-subverifiers[%d])", i))
			}
		}

		// Pack FSHookPreSampling faster
		hookActions := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
		rawCompiledIOP.FiatShamirHooksPreSampling[round] = make([]PackedStructObject, len(hookActions))
		for i, hookAction := range hookActions {
			concreteHookActVal := reflect.ValueOf(hookAction)
			if concreteHookActVal.Kind() == reflect.Ptr {
				if concreteHookActVal.IsNil() {
					continue // Skip nil pointers
				}
				concreteHookActVal = concreteHookActVal.Elem()
			}
			switch concreteHookActVal.Kind() {
			case reflect.Struct:
				hookObj, err := s.PackStructObject(concreteHookActVal)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-fiatshamirhooks-pre-sampling[%d])", i))
				}
				rawCompiledIOP.FiatShamirHooksPreSampling[round][i] = hookObj
			case reflect.Slice:
				hookObj, err := s.PackArrayOrSlice(concreteHookActVal)
				if err != nil {
					return 0, err.wrapPath(fmt.Sprintf("(compiled-IOP-fiatshamirhooks-pre-sampling[%d])", i))
				}
				rawCompiledIOP.FiatShamirHooksPreSampling[round][i] = hookObj
			default:
				return 0, newSerdeErrorf("invalid hook type: %v, expected struct or slice", concreteHookActVal.Type()).wrapPath(fmt.Sprintf("(compiled-IOP-fiatshamirhooks-pre-sampling[%d])", i))
			}
		}
	}

	s.PackedObject.CompiledIOPFast[refIdx] = *rawCompiledIOP
	return BackReference(refIdx), nil
}

// func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (*wizard.CompiledIOP, *serdeError) {

// }
