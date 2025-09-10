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

type PackedPublicInput struct {
	_    struct{}        `cbor:",toarray" serde:"omit"`
	Name string          `cbor:"n"`
	Acc  ifaces.Accessor `cbor:"a"`
}

type PackedExtradata struct {
	_     struct{} `cbor:",toarray" serde:"omit"`
	Key   string   `cbor:"k"`
	Value any      `cbor:"v"`
}

type PackedQuery struct {
	_             struct{}      `cbor:",toarray" serde:"omit"`
	BackReference BackReference `cbor:"b"`
	ConcreteType  int           `cbor:"t"`
	Round         int           `cbor:"r"`

	IsIgnored                       bool `cbor:"i"`
	IsSkippedFromProverTranscript   bool `cbor:"s"`
	IsSkippedFromVerifierTranscript bool `cbor:"v"`
}

type PackedRawData struct {
	_             struct{}           `cbor:",toarray" serde:"omit"`
	ConcreteType  int                `cbor:"t"`
	ConcreteValue PackedStructObject `cbor:"v"`
}

type packedIOPMetadata struct {
	dummyCompiled          bool
	withStorePointerChecks bool

	numRounds          int
	coinsLen           int
	queryParamsLen     int
	queryNoParamsLen   int
	selfRecursionCount int
	precomputedLen     int
	publicinputLen     int
	extradataLen       int

	packedFSSetup *big.Int
}

var (
	typeofRawPrecomputedData = reflect.TypeOf(RawPrecomputed{})
	typeofRawPublicInput     = reflect.TypeOf(PackedPublicInput{})
	typeofRawExtraData       = reflect.TypeOf(PackedExtradata{})
	typeofPcsCtxs            = reflect.TypeOf((*vortex.Ctx)(nil))
	typeOfProverAction       = reflect.TypeOf((*wizard.ProverAction)(nil)).Elem()
	typeOfVerifierAction     = reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem()
)

// PackedCompiledIOP represents the serialized form of CompiledIOP.
type PackedCompiledIOP struct {
	DummyCompiled          bool     `cbor:"j"`
	NumRounds              int      `cbor:"r"`
	WithStorePointerChecks bool     `cbor:"o"`
	SelfRecursionCount     int      `cbor:"k"`
	FiatShamirSetup        *big.Int `cbor:"l"`

	Columns            BackReference     `cbor:"a"`
	Coins              []BackReference   `cbor:"d"`
	QueriesParams      []PackedQuery     `cbor:"b"`
	QueriesNoParams    []PackedQuery     `cbor:"c"`
	SubProvers         [][]PackedRawData `cbor:"e"`
	SubVerifiers       [][]PackedRawData `cbor:"f"`
	FSHooksPreSampling [][]PackedRawData `cbor:"g"`

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
	s.PackedObject.CompiledIOPFast = append(s.PackedObject.CompiledIOPFast, PackedCompiledIOP{})

	var (
		coinNames    = comp.Coins.AllKeys()
		qIDsParams   = comp.QueriesParams.AllKeys()
		qIDsNoParams = comp.QueriesNoParams.AllKeys()
		precomputed  = comp.Precomputed.ListAllKeys()
	)

	metadata := packedIOPMetadata{
		numRounds:              comp.NumRounds(),
		coinsLen:               len(coinNames),
		queryParamsLen:         len(qIDsParams),
		queryNoParamsLen:       len(qIDsNoParams),
		selfRecursionCount:     comp.SelfRecursionCount,
		precomputedLen:         len(precomputed),
		publicinputLen:         len(comp.PublicInputs),
		extradataLen:           len(comp.ExtraData),
		dummyCompiled:          comp.DummyCompiled,
		withStorePointerChecks: comp.WithStorePointerChecks,
		packedFSSetup:          comp.FiatShamirSetup.BigInt(fieldToSmallBigInt(comp.FiatShamirSetup)),
	}
	packedComp := newPackedCompiledIOP(metadata)

	// Precomputed
	{
		for idx, colID := range precomputed {
			pre := RawPrecomputed{ColID: colID, ColAssign: comp.Precomputed.MustGet(colID)}
			obj, err := s.PackStructObject(reflect.ValueOf(pre))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-precomputed)")
			}
			packedComp.Precomputed[idx] = obj
		}
	}

	// PcsCtxs
	{
		if comp.PcsCtxs != nil {
			pcsAny, err := s.PackValue(reflect.ValueOf(comp.PcsCtxs))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-pcs-ctx)")
			}
			packedComp.PcsCtxs = pcsAny
		}
	}

	// Public inputs
	{
		for i, pi := range comp.PublicInputs {
			rawPI := PackedPublicInput{Name: pi.Name, Acc: pi.Acc}
			obj, err := s.PackStructObject(reflect.ValueOf(rawPI))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-public-inputs)")
			}
			packedComp.PublicInputs[i] = obj
		}
	}

	// Extra data (map order not fixed; preserved as-is)
	{
		i := 0
		for k, v := range comp.ExtraData {
			rawED := PackedExtradata{Key: k, Value: v}
			obj, err := s.PackStructObject(reflect.ValueOf(rawED))
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-extra-data)")
			}
			packedComp.ExtraData[i] = obj
			i++
		}
	}

	// Columns
	{
		backRefCol, err := s.PackStore(comp.Columns)
		if err != nil {
			return 0, err.wrapPath("(compiled-IOP-columns-store)")
		}
		packedComp.Columns = backRefCol
	}

	// Coins
	{

		for i, name := range coinNames {
			c := comp.Coins.Data(name)
			br, err := s.PackCoin(c)
			if err != nil {
				return 0, err.wrapPath("(compiled-IOP-coins)")
			}
			packedComp.Coins[i] = br
		}
	}

	// Queries
	{
		if list, e := s.packAllQueries(&comp.QueriesParams, qIDsParams, "query-params"); e != nil {
			return 0, e
		} else {
			packedComp.QueriesParams = list
		}
		if list, e := s.packAllQueries(&comp.QueriesNoParams, qIDsNoParams, "query-no-params"); e != nil {
			return 0, e
		} else {
			packedComp.QueriesNoParams = list
		}
	}

	// Prover actions
	{
		proverActions := comp.SubProvers.GetInner()
		for round, actPerRound := range proverActions {
			packedComp.SubProvers[round] = make([]PackedRawData, len(actPerRound))
			for i, act := range actPerRound {
				obj, serr, skipped := s.packProverAction(act, "(compiled-IOP-subprovers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				packedComp.SubProvers[round][i] = obj
			}
		}
	}

	// FS hooks pre-sampling
	{
		fsHookActions := comp.FiatShamirHooksPreSampling.GetInner()
		for round, actPerRound := range fsHookActions {
			packedComp.FSHooksPreSampling[round] = make([]PackedRawData, len(actPerRound))
			for i, act := range actPerRound {
				obj, serr, skipped := s.packVerifierAction(act, "(compiled-IOP-fs-hooks-pre-sampling[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				packedComp.FSHooksPreSampling[round][i] = obj
			}
		}
	}

	// Verifier actions
	{
		verifierActions := comp.SubVerifiers.GetInner()
		for round, actPerRound := range verifierActions {
			packedComp.SubVerifiers[round] = make([]PackedRawData, len(actPerRound))
			for i, act := range actPerRound {
				obj, serr, skipped := s.packVerifierAction(act, "(compiled-IOP-subverifiers[%d])", i)
				if serr != nil {
					return 0, serr
				}
				if skipped {
					continue
				}
				packedComp.SubVerifiers[round][i] = obj
			}
		}
	}

	s.PackedObject.CompiledIOPFast[refIdx] = *packedComp
	return BackReference(refIdx), nil
}

// -------------------- unpacking --------------------

func (d *Deserializer) UnpackCompiledIOPFast(v BackReference) (reflect.Value, *serdeError) {
	if v < 0 || int(v) >= len(d.PackedObject.CompiledIOPFast) {
		return reflect.Value{}, newSerdeErrorf("invalid compiled-IOP backreference: %v", v)
	}
	if d.compiledIOPsFast[v] != nil {
		return reflect.ValueOf(d.compiledIOPsFast[v]), nil
	}
	packedCompIOP := d.PackedObject.CompiledIOPFast[v]

	// Reserve the cache and outer shapes up-front
	deComp := newEmptyCompiledIOP(packedCompIOP)
	d.compiledIOPsFast[v] = deComp

	// defer func() {
	// 	deComp.Columns.ReserveFor(packedCompIOP.NumRounds)
	// 	deComp.Coins.ReserveFor(packedCompIOP.NumRounds)
	// 	deComp.QueriesParams.ReserveFor(packedCompIOP.NumRounds)
	// 	deComp.QueriesNoParams.ReserveFor(packedCompIOP.NumRounds)
	// 	deComp.SubProvers.Reserve(packedCompIOP.NumRounds)
	// 	deComp.SubVerifiers.Reserve(packedCompIOP.NumRounds)
	// 	deComp.FiatShamirHooksPreSampling.Reserve(packedCompIOP.NumRounds)
	// }()

	// FiatShamirSetup
	{
		f, err := unmarshalBigInt(d, *packedCompIOP.FiatShamirSetup, TypeOfBigInt)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser compiled-IOP-fiatshamirsetup)")
		}
		deComp.FiatShamirSetup.SetBigInt(f.Interface().(*big.Int))
	}

	// Columns
	{
		storeVal, err := d.UnpackStore(packedCompIOP.Columns)
		if err != nil {
			return reflect.Value{}, err.wrapPath("(deser. compiled-IOP-columns)")
		}
		deComp.Columns = storeVal.Interface().(*column.Store)
	}

	// Precomputed
	{
		for _, rp := range packedCompIOP.Precomputed {
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
		if packedCompIOP.PcsCtxs != nil {
			pcsVal, err := d.UnpackValue(packedCompIOP.PcsCtxs, typeofPcsCtxs)
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
		for i, rpi := range packedCompIOP.PublicInputs {
			res, err := d.UnpackStructObject(rpi, typeofRawPublicInput)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw public input: %w", err)
			}
			pi, ok := res.Interface().(PackedPublicInput)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw public input: %w", err)
			}
			deComp.PublicInputs[i] = wizard.PublicInput{Name: pi.Name, Acc: pi.Acc}
		}
	}

	// Extra data
	{
		for _, red := range packedCompIOP.ExtraData {
			res, err := d.UnpackStructObject(red, typeofRawExtraData)
			if err != nil {
				return reflect.Value{}, newSerdeErrorf("could not unpack struct object for raw extra data: %w", err)
			}
			ed, ok := res.Interface().(PackedExtradata)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("could not cast to raw extra data: %w", err)
			}
			deComp.ExtraData[ed.Key] = ed.Value
		}
	}

	// Coins
	{
		for _, br := range packedCompIOP.Coins {
			val, err := d.UnpackCoin(br)
			if err != nil {
				return reflect.Value{}, err.wrapPath("(deser compiled-IOP-coins)")
			}
			info, ok := val.Interface().(coin.Info)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("illegal cast to *coin.Info: %w", err)
			}
			deComp.Coins.AddToRound(info.Round, info.Name, info)
		}
	}

	// Queries
	{
		if se := d.unpackAllQueries(&deComp.QueriesParams, packedCompIOP.QueriesParams, "params"); se != nil {
			return reflect.Value{}, se
		}
		if se := d.unpackAllQueries(&deComp.QueriesNoParams, packedCompIOP.QueriesNoParams, "no-params"); se != nil {
			return reflect.Value{}, se
		}
	}

	// Actions and hooks
	{
		deComp.SubProvers.Inner = make([][]wizard.ProverAction, len(packedCompIOP.SubProvers))
		if se := d.unpackAllProverActions(deComp.SubProvers.Inner, packedCompIOP.SubProvers); se != nil {
			return reflect.Value{}, se
		}

		// logrus.Printf("Before unpacking FSHOOK  len of packedCompIOP.FSHooksPreSampling: %v", len(packedCompIOP.FSHooksPreSampling))

		deComp.FiatShamirHooksPreSampling.Inner = make([][]wizard.VerifierAction, len(packedCompIOP.FSHooksPreSampling))
		if se := d.unpackAllFSHooksPreSampling(deComp.FiatShamirHooksPreSampling.Inner, packedCompIOP.FSHooksPreSampling); se != nil {
			return reflect.Value{}, se
		}
		// logrus.Printf("AFTER unpacking FSHOOK  len of packedCompIOP.FSHooksPreSampling: %v", len(packedCompIOP.FSHooksPreSampling))

		deComp.SubVerifiers.Inner = make([][]wizard.VerifierAction, len(packedCompIOP.SubVerifiers))
		if se := d.unpackAllVerifierActions(deComp.SubVerifiers.Inner, packedCompIOP.SubVerifiers); se != nil {
			return reflect.Value{}, se
		}
	}

	return reflect.ValueOf(d.compiledIOPsFast[v]), nil
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
func (s *Serializer) packActionCommon(val any, pathFmt string, idx int) (PackedRawData, *serdeError, bool) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return PackedRawData{}, nil, true
		}
		v = v.Elem()
	}

	concreteTypeStr := getPkgPathAndTypeNameIndirect(v.Interface())
	if serr, skipped := checkRegisteredOrWarn(concreteTypeStr, v.Type()); serr != nil || skipped {
		return PackedRawData{}, serr, true
	}
	typeIdx := s.packTypeIndex(concreteTypeStr)

	switch v.Kind() {
	case reflect.Struct:
		obj, err := s.PackStructObject(v)
		if err != nil {
			return PackedRawData{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}
		return PackedRawData{ConcreteType: typeIdx, ConcreteValue: obj}, nil, false
	case reflect.Slice, reflect.Array:
		logrus.Printf("Concrete type while packing action: %v", concreteTypeStr)
		obj, err := s.PackArrayOrSlice(v)
		if err != nil {
			return PackedRawData{}, err.wrapPath(fmt.Sprintf(pathFmt, idx)), false
		}

		return PackedRawData{ConcreteType: typeIdx, ConcreteValue: obj}, nil, false
	default:
		return PackedRawData{}, newSerdeErrorf("invalid action type: %v, expected struct or slice", v.Type()).wrapPath(fmt.Sprintf(pathFmt, idx)), false
	}
}

func (s *Serializer) packProverAction(val wizard.ProverAction, pathFmt string, idx int) (PackedRawData, *serdeError, bool) {
	return s.packActionCommon(val, pathFmt, idx)
}
func (s *Serializer) packVerifierAction(val wizard.VerifierAction, pathFmt string, idx int) (PackedRawData, *serdeError, bool) {
	return s.packActionCommon(val, pathFmt, idx)
}

func newPackedCompiledIOP(packedMetadata packedIOPMetadata) *PackedCompiledIOP {
	return &PackedCompiledIOP{
		NumRounds:              packedMetadata.numRounds,
		SelfRecursionCount:     packedMetadata.selfRecursionCount,
		DummyCompiled:          packedMetadata.dummyCompiled,
		WithStorePointerChecks: packedMetadata.withStorePointerChecks,
		FiatShamirSetup:        packedMetadata.packedFSSetup,
		Coins:                  make([]BackReference, packedMetadata.coinsLen),
		QueriesParams:          make([]PackedQuery, packedMetadata.queryParamsLen),
		QueriesNoParams:        make([]PackedQuery, packedMetadata.queryNoParamsLen),
		SubProvers:             make([][]PackedRawData, packedMetadata.numRounds),
		SubVerifiers:           make([][]PackedRawData, packedMetadata.numRounds),
		FSHooksPreSampling:     make([][]PackedRawData, packedMetadata.numRounds),
		Precomputed:            make([]PackedStructObject, packedMetadata.precomputedLen),
		PublicInputs:           make([]PackedStructObject, packedMetadata.publicinputLen),
		ExtraData:              make([]PackedStructObject, packedMetadata.extradataLen),
	}
}

// Query packers per register (no closures inside functions)
func (s *Serializer) packAllQueries(reg *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query], qIDs []ifaces.QueryID,
	contextLabel string) ([]PackedQuery, *serdeError) {

	if len(qIDs) == 0 {
		return nil, nil
	}

	out := make([]PackedQuery, len(qIDs))
	for i, id := range qIDs {
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
		out[i] = PackedQuery{
			BackReference:                   backRef,
			ConcreteType:                    typeIdx,
			Round:                           reg.Round(id),
			IsIgnored:                       reg.IsIgnored(id),
			IsSkippedFromProverTranscript:   reg.IsSkippedFromProverTranscript(id),
			IsSkippedFromVerifierTranscript: reg.IsSkippedFromVerifierTranscript(id),
		}
	}
	return out, nil
}

// -------------------- constructor with Reserve --------------------

func newEmptyCompiledIOP(packedCompIOP PackedCompiledIOP) *wizard.CompiledIOP {
	deComp := &wizard.CompiledIOP{
		DummyCompiled:              packedCompIOP.DummyCompiled,
		SelfRecursionCount:         packedCompIOP.SelfRecursionCount,
		WithStorePointerChecks:     packedCompIOP.WithStorePointerChecks,
		Columns:                    column.NewStore(),
		Coins:                      wizard.NewRegister[coin.Name, coin.Info](),
		QueriesParams:              wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		QueriesNoParams:            wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		SubProvers:                 collection.VecVec[wizard.ProverAction]{},
		SubVerifiers:               collection.VecVec[wizard.VerifierAction]{},
		FiatShamirHooksPreSampling: collection.VecVec[wizard.VerifierAction]{},
		Precomputed:                collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		PublicInputs:               make([]wizard.PublicInput, len(packedCompIOP.PublicInputs), cap(packedCompIOP.PublicInputs)),
		ExtraData:                  make(map[string]any, len(packedCompIOP.ExtraData)),
		PcsCtxs:                    nil,
	}

	// Preallocate spaces
	deComp.Coins.ByRounds.Inner = make([][]coin.Name, packedCompIOP.NumRounds)
	deComp.QueriesParams.ByRounds.Inner = make([][]ifaces.QueryID, packedCompIOP.NumRounds)
	deComp.QueriesNoParams.ByRounds.Inner = make([][]ifaces.QueryID, packedCompIOP.NumRounds)

	// deComp.SubProvers.Inner = make([][]wizard.ProverAction, len(packedCompIOP.SubProvers))
	// deComp.SubVerifiers.Inner = make([][]wizard.VerifierAction, len(packedCompIOP.SubVerifiers))
	// deComp.FiatShamirHooksPreSampling.Inner = make([][]wizard.VerifierAction, len(packedCompIOP.FSHooksPreSampling))

	return deComp
}

// -------------------- query/action unpack helpers --------------------

// Query unpackers per register (no closures inside functions)
func (d *Deserializer) unpackAllQueries(reg *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query],
	raws []PackedQuery, contextLabel string) *serdeError {

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
		reg.AddToRound(rq.Round, qID, q)
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
func (d *Deserializer) unpackAllProverActions(deCompProverActions [][]wizard.ProverAction, actions2D [][]PackedRawData) *serdeError {
	var (
		res   reflect.Value
		se    *serdeError
		ct    reflect.Type
		ctStr string
		err   error
	)

	for round, actions := range actions2D {
		if len(actions) == 0 {
			deCompProverActions[round] = nil
		}
		deCompProverActions[round] = make([]wizard.ProverAction, len(actions))
		for idx, action := range actions {
			ctStr = d.PackedObject.Types[action.ConcreteType]
			ct, err = findRegisteredImplementation(ctStr)
			if err != nil {
				return newSerdeErrorf("could not find registered implementation for prover action: %w", err)
			}

			// logrus.Printf("Interface type passed for %s-action is %v", contextLabel, ifaceType)
			// logrus.Printf("**** Before Checking for concrete type: %v kind:%v", ct, ct.Kind())

			switch ct.Kind() {
			case reflect.Struct:
				res, se = d.UnpackStructObject(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser struct compiled-IOP-prover-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers
				val := res.Addr().Interface()

				// DEBUG Purposes only
				// if !TypeTracker[ctStr] {
				// 	logrus.Printf("Concrete type:%v implements iface type", ct)
				// 	TypeTracker[ctStr] = true
				// }

				act, ok := val.(wizard.ProverAction)
				if !ok {
					return newSerdeErrorf("illegal cast of type %v with string rep %s to prover action", ct, ctStr)
				}

				deCompProverActions[round][idx] = act

			case reflect.Slice, reflect.Array:

				logrus.Printf("** Before registering prover action of kind slice/array ct:%v ct.Elem():%v", ct, ct.Elem())

				res, se = d.UnpackArrayOrSlice(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser slice/array compiled-IOP-prover-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers

				val := res.Addr().Interface()

				// DEBUG Purposes only
				if !TypeTracker[ctStr] {
					logrus.Printf("Concrete type:%v implements iface type", ct)
					TypeTracker[ctStr] = true
				}
				act, ok := val.(wizard.ProverAction)
				if !ok {
					return newSerdeErrorf("illegal cast of %v with string rep %s to prover action", ct, ctStr)
				}
				deCompProverActions[round][idx] = act

				//logrus.Printf("** After registering prover action of kind slice/array")

			default:
				return newSerdeErrorf("unsupported kind:%v for prover action", ct.Kind())
			}
		}
	}
	return nil
}

func (d *Deserializer) unpackAllVerifierActions(deCompVerifierActions [][]wizard.VerifierAction, actions2D [][]PackedRawData) *serdeError {

	var (
		res   reflect.Value
		se    *serdeError
		ct    reflect.Type
		ctStr string
		err   error
	)

	for round, actions := range actions2D {
		if len(actions) == 0 {
			deCompVerifierActions[round] = nil
		}
		deCompVerifierActions[round] = make([]wizard.VerifierAction, len(actions))
		for idx, action := range actions {
			ctStr = d.PackedObject.Types[action.ConcreteType]
			ct, err = findRegisteredImplementation(ctStr)
			if err != nil {
				return newSerdeErrorf("could not find registered implementation for verifier action: %w", err)
			}

			// logrus.Printf("Interface type passed for %s-action is %v", contextLabel, ifaceType)
			// logrus.Printf("**** Before Checking for concrete type: %v kind:%v", ct, ct.Kind())

			switch ct.Kind() {
			case reflect.Struct:
				res, se = d.UnpackStructObject(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser struct compiled-IOP-verifier-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers
				val := res.Addr().Interface()
				// DEBUG Purposes only
				// if !TypeTracker[ctStr] {
				// 	logrus.Printf("Concrete type:%v implements iface type", ct)
				// 	TypeTracker[ctStr] = true
				// }

				act, ok := val.(wizard.VerifierAction)
				if !ok {
					return newSerdeErrorf("illegal cast of type %v with string rep %s to verifier action", ct, ctStr)
				}

				deCompVerifierActions[round][idx] = act

			case reflect.Slice, reflect.Array:
				res, se = d.UnpackArrayOrSlice(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser slice/array compiled-IOP-verifier-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers
				val := res.Addr().Interface()
				// DEBUG Purposes only
				if !TypeTracker[ctStr] {
					logrus.Printf("Concrete type:%v implements iface type", ct)
					TypeTracker[ctStr] = true
				}

				act, ok := val.(wizard.VerifierAction)
				if !ok {
					return newSerdeErrorf("illegal cast of %v with string rep %s to verifier action", ct, ctStr)
				}
				deCompVerifierActions[round][idx] = act

			default:
				return newSerdeErrorf("unsupported kind:%v for verifier action", ct.Kind())
			}
		}
	}
	return nil
}

func (d *Deserializer) unpackAllFSHooksPreSampling(deCompFSHooksPreSampling [][]wizard.VerifierAction, actions2D [][]PackedRawData) *serdeError {

	var (
		res   reflect.Value
		se    *serdeError
		ct    reflect.Type
		ctStr string
		err   error
	)

	// logrus.Printf("Outer Deser FSHOOK Length of actions2D: %v", len(actions2D))
	// logrus.Printf("Actions 2d: %v", actions2D)

	for round, actions := range actions2D {
		if len(actions) == 0 {
			deCompFSHooksPreSampling[round] = nil
		}
		deCompFSHooksPreSampling[round] = make([]wizard.VerifierAction, len(actions))
		for idx, action := range actions {
			ctStr = d.PackedObject.Types[action.ConcreteType]
			ct, err = findRegisteredImplementation(ctStr)
			if err != nil {
				return newSerdeErrorf("could not find registered implementation for FsHook verifier action: %w", err)
			}

			// logrus.Printf("Interface type passed for %s-action is %v", contextLabel, ifaceType)
			// logrus.Printf("**** Before Checking for concrete type: %v kind:%v", ct, ct.Kind())

			switch ct.Kind() {
			case reflect.Struct:
				res, se = d.UnpackStructObject(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser struct compiled-IOP-fshook-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers
				val := res.Addr().Interface()
				// DEBUG Purposes only
				// if !TypeTracker[ctStr] {
				// 	logrus.Printf("Concrete type:%v implements iface type", ct)
				// 	TypeTracker[ctStr] = true
				// }

				act, ok := val.(wizard.VerifierAction)
				if !ok {
					return newSerdeErrorf("illegal cast of type %v with string rep %s to fs hook verifier action", ct, ctStr)
				}

				deCompFSHooksPreSampling[round][idx] = act

			case reflect.Slice, reflect.Array:
				res, se = d.UnpackArrayOrSlice(action.ConcreteValue, ct)
				if se != nil {
					return se.wrapPath("(deser slice/array compiled-IOP-fshook-actions)")
				}

				// IMPORTANT: The concrete type must be a pointer to a struct because implicity
				// go always returns a pointer type when reflect.TypeOf(...) for either pointer or value receivers
				val := res.Addr().Interface()
				// DEBUG Purposes only
				if !TypeTracker[ctStr] {
					logrus.Printf("Concrete type:%v implements iface type", ct)
					TypeTracker[ctStr] = true
				}

				act, ok := val.(wizard.VerifierAction)
				if !ok {
					return newSerdeErrorf("illegal cast of %v with string rep %s to fshook action", ct, ctStr)
				}
				deCompFSHooksPreSampling[round][idx] = act

			default:
				return newSerdeErrorf("unsupported kind:%v for fshook action", ct.Kind())
			}
		}
	}
	return nil
}
