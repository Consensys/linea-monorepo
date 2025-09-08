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
	"github.com/sirupsen/logrus"
)

var (
	TypeTracker = make(map[string]bool)
)

type RawPrecomputed struct {
	_         struct{}             `cbor:",toarray" serde:"omit"`
	ColID     ifaces.ColID         `cbor:"k"`
	ColAssign ifaces.ColAssignment `cbor:"v"`
}

type RawPublicInput struct {
	_    struct{}        `cbor:",toarray" serde:"omit"`
	Name string          `cbor:"n"`
	Acc  ifaces.Accessor `cbor:"a"`
}

type RawExtraData struct {
	_     struct{} `cbor:",toarray" serde:"omit"`
	Key   string   `cbor:"k"`
	Value any      `cbor:"v"`
}

type RawQuery struct {
	_                               struct{}      `cbor:",toarray" serde:"omit"`
	BackReference                   BackReference `cbor:"b"`
	ConcreteType                    int           `cbor:"t"`
	IsIgnored                       bool          `cbor:"i"`
	IsSkippedFromProverTranscript   bool          `cbor:"s"`
	IsSkippedFromVerifierTranscript bool          `cbor:"v"`
}

type RawData struct {
	_             struct{}           `cbor:",toarray" serde:"omit"`
	ConcreteType  int                `cbor:"t"`
	ConcreteValue PackedStructObject `cbor:"v"`
}

var (
	typeofRawPrecomputedData = reflect.TypeOf(RawPrecomputed{})
	typeofRawPublicInput     = reflect.TypeOf(RawPublicInput{})
	typeofRawExtraData       = reflect.TypeOf(RawExtraData{})
	typeofPcsCtxs            = reflect.TypeOf((*vortex.Ctx)(nil))
	typeOfProverAction       = reflect.TypeOf((*wizard.ProverAction)(nil)).Elem()
	typeOfVerifierAction     = reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem()
)

// RawCompiledIOP represents the serialized form of CompiledIOP.
type RawCompiledIOP struct {
	DummyCompiled          bool     `cbor:"j"`
	NumRounds              int      `cbor:"r"`
	WithStorePointerChecks bool     `cbor:"o"`
	SelfRecursionCount     int      `cbor:"k"`
	FiatShamirSetup        *big.Int `cbor:"l"`

	Columns BackReference     `cbor:"a"`
	Coins   [][]BackReference `cbor:"d"`

	QueriesParams              [][]RawQuery `cbor:"b"`
	QueriesNoParams            [][]RawQuery `cbor:"c"`
	Subprovers                 [][]RawData  `cbor:"e"`
	SubVerifiers               [][]RawData  `cbor:"f"`
	FiatShamirHooksPreSampling [][]RawData  `cbor:"g"`

	Precomputed  []PackedStructObject `cbor:"h"`
	PublicInputs []PackedStructObject `cbor:"m"`
	ExtraData    []PackedStructObject `cbor:"n"`

	PcsCtxs any `cbor:"i"`
}

// -------------------- Packing --------------------

func (s *Serializer) PackCompiledIOPFast(comp *wizard.CompiledIOP) (BackReference, *serdeError) {
	if comp == nil {
		return 0, nil
	}
	if idx, ok := s.compiledIOPsFast[comp]; ok {
		return BackReference(idx), nil
	}
	refIdx := len(s.PackedObject.CompiledIOPFast)
	s.compiledIOPsFast[comp] = refIdx
	s.PackedObject.CompiledIOPFast = append(s.PackedObject.CompiledIOPFast, RawCompiledIOP{})

	raw := initRawCompiledIOP(comp)

	// Precomputed

	{
		for idx, colID := range comp.Precomputed.ListAllKeys() {
			pre := RawPrecomputed{ColID: colID, ColAssign: comp.Precomputed.MustGet(colID)}
			obj, err := s.PackStructObject(reflect.ValueOf(pre))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-precomputed)")
			}
			raw.Precomputed[idx] = obj
		}
	}

	// PcsCtxs
	{
		logrus.Println("Packing PcsCtxs")
		if comp.PcsCtxs != nil {
			pcsAny, err := s.PackValue(reflect.ValueOf(comp.PcsCtxs))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-pcs-ctx)")
			}
			raw.PcsCtxs = pcsAny
		}
	}

	// Public inputs
	{
		for i, pi := range comp.PublicInputs {
			rawPI := RawPublicInput{Name: pi.Name, Acc: pi.Acc}
			obj, err := s.PackStructObject(reflect.ValueOf(rawPI))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-public-inputs)")
			}
			raw.PublicInputs[i] = obj
		}
	}

	// Extra data (map order not fixed; preserved as-is)
	{
		i := 0
		for k, v := range comp.ExtraData {
			rawED := RawExtraData{Key: k, Value: v}
			obj, err := s.PackStructObject(reflect.ValueOf(rawED))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-extra-data)")
			}
			raw.ExtraData[i] = obj
			i++
		}
	}

	// Columns
	{
		backRefCol, err := s.PackStore(comp.Columns)
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-columns-store)")
		}
		raw.Columns = backRefCol
	}

	for round := 0; round < raw.NumRounds; round++ {

		// Coins
		{
			coinNames := comp.Coins.AllKeysAt(round)
			raw.Coins[round] = make([]BackReference, len(coinNames))
			for i, name := range coinNames {
				c := comp.Coins.Data(name)
				br, err := s.PackCoin(c)
				if err != nil {
					return 0, err.wrapPath("(compiled-IOP-coins)")
				}
				raw.Coins[round][i] = br
			}
		}

		// Queries
		{
			if list, e := s.packQueriesParamsRound(comp, round); e != nil {
				return 0, e
			} else {
				raw.QueriesParams[round] = list
			}
			if list, e := s.packQueriesNoParamsRound(comp, round); e != nil {
				return 0, e
			} else {
				raw.QueriesNoParams[round] = list
			}
		}

		// Prover actions
		{
			provers := comp.SubProvers.GetOrEmpty(round)
			raw.Subprovers[round] = make([]RawData, len(provers))
			for i, act := range provers {
				obj, serr, skipped := s.packProverAction(act, "(compiled-IOP-subprovers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				raw.Subprovers[round][i] = obj
			}
		}

		// FS hooks pre-sampling
		{
			hooks := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
			raw.FiatShamirHooksPreSampling[round] = make([]RawData, len(hooks))
			for i, hook := range hooks {
				obj, serr, skipped := s.packVerifierAction(hook, "(compiled-IOP-fiatshamirhooks-pre-sampling[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				raw.FiatShamirHooksPreSampling[round][i] = obj
			}
		}

		// Verifier actions
		{
			verifiers := comp.SubVerifiers.GetOrEmpty(round)
			raw.SubVerifiers[round] = make([]RawData, len(verifiers))
			for i, act := range verifiers {
				obj, serr, skipped := s.packVerifierAction(act, "(compiled-IOP-subverifiers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				raw.SubVerifiers[round][i] = obj
			}
		}
	}

	s.PackedObject.CompiledIOPFast[refIdx] = *raw
	return BackReference(refIdx), nil
}

// -------------------- unpacking --------------------

func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (reflect.Value, *serdeError) {
	if v < 0 || int(v) >= len(d.PackedObject.CompiledIOPFast) {
		return reflect.Value{}, newSerdeErrorf("invalid compiled-IOP backreference: %v", v)
	}
	if d.CompiledIOPsFast[v] != nil {
		return reflect.ValueOf(d.CompiledIOPsFast[v]), nil
	}
	raw := d.PackedObject.CompiledIOPFast[v]

	// Reserve the cache and outer shapes up-front
	deComp := newEmptyCompiledIOP(raw)
	d.CompiledIOPsFast[v] = deComp

	// FiatShamirSetup
	{
		f, err := unmarshalBigInt(d, *raw.FiatShamirSetup, TypeOfBigInt)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser compiled-IOP-fiatshamirsetup)")
		}
		deComp.FiatShamirSetup.SetBigInt(f.Interface().(*big.Int))
	}

	// Columns
	{
		storeVal, err := d.UnpackStore(raw.Columns)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser. compiled-IOP-columns)")
		}
		deComp.Columns = storeVal.Interface().(*column.Store)
	}

	// Precomputed
	{
		for _, rp := range raw.Precomputed {
			res, err := d.UnpackStructObject(rp, typeofRawPrecomputedData)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw pre-computed data: %w", err)
			}
			pre, ok := res.Interface().(RawPrecomputed)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to RawPrecomputed: %w", err)
			}
			deComp.Precomputed.InsertNew(pre.ColID, pre.ColAssign)
		}
	}

	// PcsCtxs
	{
		logrus.Println("Unpacking Pcsctxs")
		if raw.PcsCtxs != nil {
			pcsVal, err := d.UnpackValue(raw.PcsCtxs, typeofPcsCtxs)
			if err != nil {
				return reflect.Value{}, err.wrapPath("(deser compiled-IOP-pcsctxs)")
			}
			deComp.PcsCtxs = pcsVal.Interface()
		} else {
			deComp.PcsCtxs = nil
		}
	}

	// Public inputs
	{
		logrus.Println("Unpacking Public inputs")
		for i, rpi := range raw.PublicInputs {
			res, err := d.UnpackStructObject(rpi, typeofRawPublicInput)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw public input: %w", err)
			}
			pi, ok := res.Interface().(RawPublicInput)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw public input: %w", err)
			}
			deComp.PublicInputs[i] = wizard.PublicInput{Name: pi.Name, Acc: pi.Acc}
		}
	}

	// Extra data
	{
		for _, red := range raw.ExtraData {
			res, err := d.UnpackStructObject(red, typeofRawExtraData)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw extra data: %w", err)
			}
			ed, ok := res.Interface().(RawExtraData)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw extra data: %w", err)
			}
			deComp.ExtraData[ed.Key] = ed.Value
		}
	}

	// Rounds
	for round := 0; round < raw.NumRounds; round++ {

		// Coins
		{
			for _, br := range raw.Coins[round] {
				val, err := d.UnpackCoin(br)
				if err != nil {
					return reflect.Value{}, err.wrapPath("(deser compiled-IOP-coins)")
				}
				info, ok := val.Interface().(coin.Info)
				if !ok {
					return reflect.Value{}, newSerdeErrorf("illegal cast to *coin.Info: %w", err)
				}
				deComp.Coins.AddToRound(round, info.Name, info)
			}
		}

		// Queries
		{
			if se := d.unpackQueriesParamsRound(deComp, raw.QueriesParams[round], round); se != nil {
				return reflect.Value{}, se
			}
			if se := d.unpackQueriesNoParamsRound(deComp, raw.QueriesNoParams[round], round); se != nil {
				return reflect.Value{}, se
			}
		}

		// Actions and hooks
		{
			if se := d.unpackProverActionsRound(deComp, raw.Subprovers[round], round); se != nil {
				return reflect.Value{}, se
			}
			if se := d.unpackFSHooksRound(deComp, raw.FiatShamirHooksPreSampling[round], round); se != nil {
				return reflect.Value{}, se
			}
			if se := d.unpackVerifierActionsRound(deComp, raw.SubVerifiers[round], round); se != nil {
				return reflect.Value{}, se
			}
		}
	}

	return reflect.ValueOf(d.CompiledIOPsFast[v]), nil
}

// -------------------- Helper functions --------------------

func (s *Serializer) packTypeIndex(concreteTypeStr string) int {
	if idx, ok := s.typeMap[concreteTypeStr]; ok {
		return idx
	}
	s.PackedObject.Types = append(s.PackedObject.Types, concreteTypeStr)
	idx := len(s.PackedObject.Types) - 1
	s.typeMap[concreteTypeStr] = idx
	return idx
}

func checkRegisteredOrWarn(concreteTypeStr string, t reflect.Type) (*serdeError, bool) {
	if _, err := findRegisteredImplementation(concreteTypeStr); err != nil {
		if !TypeTracker[concreteTypeStr] {
			logrus.Warnf("attempted to serialize unregistered type repr=%q type=%v: %v", concreteTypeStr, t.String(), err)
			TypeTracker[concreteTypeStr] = true
		}
		// return newSerdeErrorf("attempted to serialize unregistered type repr=%q type=%v: %v", concreteTypeStr, t.String(), err), true
	}
	return nil, false
}

// Shared packing for prover/verifier actions (struct or slice), preserving messages and path formats.
func (s *Serializer) packActionCommon(val any, pathFmt string, idx int) (RawData, *serdeError, bool) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return RawData{}, nil, true
		}
		v = v.Elem()
	}

	concreteTypeStr := getPkgPathAndTypeNameIndirect(v.Interface())
	if serr, skipped := checkRegisteredOrWarn(concreteTypeStr, v.Type()); serr != nil || skipped {
		return RawData{}, serr, true
	}
	typeIdx := s.packTypeIndex(concreteTypeStr)

	switch v.Kind() {
	case reflect.Struct:
		obj, err := s.PackStructObject(v)
		if err != nil {
			return RawData{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}
		return RawData{ConcreteType: typeIdx, ConcreteValue: obj}, nil, false
	case reflect.Slice, reflect.Array:
		obj, err := s.PackArrayOrSlice(v)
		if err != nil {
			return RawData{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}
		return RawData{ConcreteType: typeIdx, ConcreteValue: obj}, nil, false
	default:
		return RawData{}, newSerdeErrorf("invalid action type: %v, expected struct or slice", v.Type()).wrapPath(fmt.Sprintf(pathFmt, idx)), false
	}
}

func (s *Serializer) packProverAction(val wizard.ProverAction, pathFmt string, idx int) (RawData, *serdeError, bool) {
	return s.packActionCommon(val, pathFmt, idx)
}
func (s *Serializer) packVerifierAction(val wizard.VerifierAction, pathFmt string, idx int) (RawData, *serdeError, bool) {
	return s.packActionCommon(val, pathFmt, idx)
}

func initRawCompiledIOP(comp *wizard.CompiledIOP) *RawCompiledIOP {
	numRounds := comp.NumRounds()
	return &RawCompiledIOP{
		NumRounds:                  numRounds,
		SelfRecursionCount:         comp.SelfRecursionCount,
		DummyCompiled:              comp.DummyCompiled,
		WithStorePointerChecks:     comp.WithStorePointerChecks,
		FiatShamirSetup:            comp.FiatShamirSetup.BigInt(fieldToSmallBigInt(comp.FiatShamirSetup)),
		Coins:                      make([][]BackReference, numRounds),
		QueriesParams:              make([][]RawQuery, numRounds),
		QueriesNoParams:            make([][]RawQuery, numRounds),
		Subprovers:                 make([][]RawData, numRounds),
		SubVerifiers:               make([][]RawData, numRounds),
		FiatShamirHooksPreSampling: make([][]RawData, numRounds),
		Precomputed:                make([]PackedStructObject, len(comp.Precomputed.ListAllKeys())),
		PublicInputs:               make([]PackedStructObject, len(comp.PublicInputs)),
		ExtraData:                  make([]PackedStructObject, len(comp.ExtraData)),
	}
}

// Query packers per register (no closures inside functions)
func (s *Serializer) packQueriesRound(comp *wizard.CompiledIOP,
	round int, contextLabel string) ([]RawQuery, *serdeError) {

	var reg wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query]
	switch contextLabel {
	case "params":
		reg = comp.QueriesParams
	case "no-params":
		reg = comp.QueriesNoParams
	default:
		return nil, newSerdeErrorf("invalid contextLabel during packing query: %v", contextLabel)
	}

	// IMPORTANT: This `AllKeysAt` internally calls `Reserve(round+1)` which is why in the corresponding
	// deserializer we need to do the same.
	ids := reg.AllKeysAt(round)
	out := make([]RawQuery, len(ids))
	for i, id := range ids {
		q := reg.Data(id)
		backRef, err := s.PackQuery(q)
		if err != nil {
			return nil, err.wrapPath("(ser-compiled-IOP-queries-" + contextLabel + ")")
		}
		v := reflect.ValueOf(q)
		ct := getPkgPathAndTypeNameIndirect(v.Interface())
		if serr, _ := checkRegisteredOrWarn(ct, v.Type()); serr != nil {
			return nil, serr
		}
		typeIdx := s.packTypeIndex(ct)
		out[i] = RawQuery{
			BackReference:                   backRef,
			ConcreteType:                    typeIdx,
			IsIgnored:                       reg.IsIgnored(id),
			IsSkippedFromProverTranscript:   reg.IsSkippedFromProverTranscript(id),
			IsSkippedFromVerifierTranscript: reg.IsSkippedFromVerifierTranscript(id),
		}
	}
	return out, nil
}

// Wrappers for clarity and call-site compatibility:
func (s *Serializer) packQueriesParamsRound(comp *wizard.CompiledIOP, round int) ([]RawQuery, *serdeError) {
	return s.packQueriesRound(comp, round, "params")
}

func (s *Serializer) packQueriesNoParamsRound(comp *wizard.CompiledIOP, round int) ([]RawQuery, *serdeError) {
	return s.packQueriesRound(comp, round, "no-params")
}

// -------------------- constructor with Reserve --------------------

func newEmptyCompiledIOP(rawCompIOP RawCompiledIOP) *wizard.CompiledIOP {
	comp := &wizard.CompiledIOP{
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
	// Ensure outer length equals NumRounds even if some rounds have zero entries
	comp.QueriesParams.ReserveFor(rawCompIOP.NumRounds)
	comp.QueriesNoParams.ReserveFor(rawCompIOP.NumRounds)
	comp.SubProvers.Reserve(rawCompIOP.NumRounds)
	comp.SubVerifiers.Reserve(rawCompIOP.NumRounds)
	comp.FiatShamirHooksPreSampling.Reserve(rawCompIOP.NumRounds)
	return comp
}

// -------------------- query/action unpack helpers --------------------

// Query unpackers per register (no closures inside functions)
func (d *Deserializer) unpackQueriesRound(deComp *wizard.CompiledIOP,
	raws []RawQuery, round int, contextLabel string) *serdeError {

	var reg wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query]
	switch contextLabel {
	case "params":
		reg = deComp.QueriesParams
	case "no-params":
		reg = deComp.QueriesNoParams
	default:
		return newSerdeErrorf("invalid contextLabel during unpacking query: %v", contextLabel)
	}

	for _, rq := range raws {
		ctStr := d.PackedObject.Types[rq.ConcreteType]
		ct, err := findRegisteredImplementation(ctStr)
		if err != nil {
			return newSerdeErrorf("could not find registered implementation for concrete type: %w", err)
		}

		val, se := d.UnpackQuery(rq.BackReference, ct)
		if se != nil {
			return se.wrapPath("(deser compiled-IOP-queries-" + contextLabel + ")")
		}

		q, ok := val.Interface().(ifaces.Query)
		if !ok {
			return newSerdeErrorf("illegal cast to ifaces.Query")
		}

		qID := q.Name()
		reg.AddToRound(round, qID, q)
		if rq.IsIgnored {
			reg.MarkAsIgnored(qID)
		}
		if rq.IsSkippedFromProverTranscript {
			reg.MarkAsSkippedFromProverTranscript(qID)
		}
		if rq.IsSkippedFromVerifierTranscript {
			reg.MarkAsSkippedFromVerifierTranscript(qID)
		}
	}
	return nil
}

// Wrappers for clarity and call-site compatibility:
func (d *Deserializer) unpackQueriesParamsRound(deComp *wizard.CompiledIOP, raws []RawQuery, round int) *serdeError {
	return d.unpackQueriesRound(deComp, raws, round, "params")
}

func (d *Deserializer) unpackQueriesNoParamsRound(deComp *wizard.CompiledIOP, raws []RawQuery, round int) *serdeError {
	return d.unpackQueriesRound(deComp, raws, round, "no-params")
}

// Common helper for prover/verifier actions per round
func (d *Deserializer) unpackActionsRound(deComp *wizard.CompiledIOP, raws []RawData,
	round int, contextLabel string, ifaceType reflect.Type,
	castFn func(any) (any, bool), registerFn func(*wizard.CompiledIOP, int, any)) *serdeError {
	for _, r := range raws {
		ctStr := d.PackedObject.Types[r.ConcreteType]
		ct, err := findRegisteredImplementation(ctStr)
		if err != nil {
			return newSerdeErrorf("could not find registered implementation for %s action: %w", contextLabel, err)
		}

		var (
			res reflect.Value
			se  *serdeError
		)

		// logrus.Printf("Interface type passed for %s-action is %v", contextLabel, ifaceType)
		// logrus.Printf("**** Before Checking for concrete type: %v kind:%v", ct, ct.Kind())

		switch ct.Kind() {
		case reflect.Struct:
			res, se = d.UnpackStructObject(r.ConcreteValue, ct)
			if se != nil {
				return se.wrapPath("(deser struct compiled-IOP-" + contextLabel + "-actions)")
			}

			val := res.Interface()
			if !ct.Implements(ifaceType) && ct.Kind() == reflect.Struct {
				// return newSerdeErrorf("concrete type:%v must implement iface type %v", ct, ifaceType)
				//logrus.Printf("Concrete type does not implement iface type. Trying pointer form")
				val = res.Addr().Interface()
			} else {
				if !TypeTracker[ctStr] {
					logrus.Printf("Concrete type:%v implements iface type", ct)
					TypeTracker[ctStr] = true
				}
			}

			act, ok := castFn(val)
			if !ok {
				return newSerdeErrorf("illegal cast to %s action", contextLabel)
			}
			registerFn(deComp, round, act)

		case reflect.Slice, reflect.Array:
			res, se = d.UnpackArrayOrSlice(r.ConcreteValue, ct)
			if se != nil {
				return se.wrapPath("(deser slice/array compiled-IOP-" + contextLabel + "-actions)")
			}

			// logrus.Printf("** Before registering prover action of kind slice/array")
			val := res.Interface()
			act, ok := castFn(val)
			if !ok {
				return newSerdeErrorf("illegal cast to %s action", contextLabel)
			}
			registerFn(deComp, round, act)

			//logrus.Printf("** After registering prover action of kind slice/array")

		default:
			return newSerdeErrorf("unsupported kind for %s action: %v", contextLabel, ct.Kind())
		}

		/*
			logrus.Printf("**** After Checking for concrete type: %v kind:%v", ct, ct.Kind())
			val := res.Interface()
			logrus.Printf("BEFORE deser prover val type: %v kind:%v addr:%v  type:%v kind:%v", reflect.TypeOf(val), reflect.TypeOf(val).Kind(), res.Addr(), reflect.TypeOf(res.Addr().Interface()), res.Addr().Kind())
			if res.Kind() != reflect.Ptr && res.CanAddr() {
				val = res.Addr().Interface()
			}

			logrus.Printf("AFTER deser prover val type: %v kind:%v addr:%v  type:%v kind:%v", reflect.TypeOf(val), reflect.TypeOf(val).Kind(), res.Addr(), reflect.TypeOf(res.Addr().Interface()), res.Addr().Kind())
			act, ok := castFn(val)
			if !ok {
				return newSerdeErrorf("illegal cast to %s action", contextLabel)
			}
			registerFn(deComp, round, act)

		*/

		/*
			// Try cast directly
			if act, ok := castFn(val); ok {
				registerFn(deComp, round, act)
				continue
			}

			// If it didn’t work, but it’s addressable, try pointer form
			if res.CanAddr() {
				valPtr := res.Addr().Interface()
				if act, ok := castFn(valPtr); ok {
					registerFn(deComp, round, act)
					continue
				}
			}
			return newSerdeErrorf("illegal cast to %s action (type %T)", contextLabel, val)
		*/

	}
	return nil
}

// Wrappers for clarity and call-site compatibility:
func (d *Deserializer) unpackProverActionsRound(de *wizard.CompiledIOP, raws []RawData, round int) *serdeError {
	return d.unpackActionsRound(
		de, raws, round, "prover", typeOfProverAction,
		func(v any) (any, bool) { act, ok := v.(wizard.ProverAction); return act, ok },
		func(c *wizard.CompiledIOP, r int, act any) { c.RegisterProverAction(r, act.(wizard.ProverAction)) },
	)
}

func (d *Deserializer) unpackVerifierActionsRound(de *wizard.CompiledIOP, raws []RawData, round int) *serdeError {
	return d.unpackActionsRound(
		de, raws, round, "verifier", typeOfVerifierAction,
		func(v any) (any, bool) { act, ok := v.(wizard.VerifierAction); return act, ok },
		func(c *wizard.CompiledIOP, r int, act any) { c.RegisterVerifierAction(r, act.(wizard.VerifierAction)) },
	)
}

func (d *Deserializer) unpackFSHooksRound(de *wizard.CompiledIOP, raws []RawData, round int) *serdeError {
	for _, r := range raws {
		ctStr := d.PackedObject.Types[r.ConcreteType]
		ct, err := findRegisteredImplementation(ctStr)
		if err != nil {
			return newSerdeErrorf("could not find registered implementation for prover action: %w", err)
		}
		res, se := d.UnpackStructObject(r.ConcreteValue, ct)
		if se != nil {
			return newSerdeErrorf("could not unpack struct object for verifier action: %w", se)
		}
		h, ok := res.Addr().Interface().(wizard.VerifierAction)
		if !ok {
			return newSerdeErrorf("could not deser because illegal cast to verifier action")
		}
		de.FiatShamirHooksPreSampling.AppendToInner(round, h)
	}
	return nil
}
