package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
)

var (
	compiledIOPType = reflect.TypeOf((*wizard.CompiledIOP)(nil))
)

// rawCompiledIOP represents the serialized form of CompiledIOP.
type rawCompiledIOP struct {
	Columns                    [][]json.RawMessage `json:"columns"`
	QueriesParams              [][]json.RawMessage `json:"queriesParams"`
	QueriesNoParams            [][]json.RawMessage `json:"queriesNoParams"`
	Coins                      [][]json.RawMessage `json:"coins"`
	Subprovers                 [][]json.RawMessage `json:"subProvers"`
	SubVerifiers               [][]json.RawMessage `json:"subVerifiers"`
	FiatShamirHooksPreSampling [][]json.RawMessage `json:"fiatShamirHooksPreSampling"`
	Precomputed                json.RawMessage     `json:"precomputed"`
	PcsCtxs                    json.RawMessage     `json:"pcsCtxs"`
	DummyCompiled              bool                `json:"dummyCompiled"`
	SelfRecursionCount         int                 `json:"selfRecursionCount"`
	Artefacts                  json.RawMessage     `json:"artefacts"`
	FiatShamirSetup            json.RawMessage     `json:"fiatShamirSetup"`
	PublicInputs               json.RawMessage     `json:"publicInputs"`
	ExtraData                  json.RawMessage     `json:"extraData"`

	mu sync.Mutex
}

func initializeRawCompiledIOP(comp *wizard.CompiledIOP, numRounds int) *rawCompiledIOP {
	return &rawCompiledIOP{
		Columns:                    make([][]json.RawMessage, numRounds),
		QueriesParams:              make([][]json.RawMessage, numRounds),
		QueriesNoParams:            make([][]json.RawMessage, numRounds),
		Coins:                      make([][]json.RawMessage, numRounds),
		Subprovers:                 make([][]json.RawMessage, numRounds),
		SubVerifiers:               make([][]json.RawMessage, numRounds),
		FiatShamirHooksPreSampling: make([][]json.RawMessage, numRounds),
		DummyCompiled:              comp.DummyCompiled,
		SelfRecursionCount:         comp.SelfRecursionCount,
	}
}

// NewEmptyCompiledIOP initializes an empty CompiledIOP object.
func NewEmptyCompiledIOP() *wizard.CompiledIOP {
	return &wizard.CompiledIOP{
		Columns:                    column.NewStore(),
		QueriesParams:              wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		QueriesNoParams:            wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		Coins:                      wizard.NewRegister[coin.Name, coin.Info](),
		SubProvers:                 collection.VecVec[wizard.ProverAction]{},
		SubVerifiers:               collection.VecVec[wizard.VerifierAction]{},
		FiatShamirHooksPreSampling: collection.VecVec[wizard.VerifierAction]{},
		Precomputed:                collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		PcsCtxs:                    nil,
		ExtraData:                  make(map[string]any),
	}
}

// SerializeCompiledIOP marshals a wizard.CompiledIOP object into CBOR. This is
// meant to allow deserializing the IOP during the prover time instead of
// recompiling everything every time we want to run the prover.
//
// Example:
//
//	 	comp := wizard.Compile(myBuilder, myCompilerSuite...)
//		marshaled, err := SerializeCompiledIOP(comp)
//		if err != nil {
//			panic(err)
func SerializeCompiledIOP(comp *wizard.CompiledIOP) ([]byte, error) {
	serTime := time.Now()
	numRounds := comp.NumRounds()
	if numRounds == 0 {
		return serializeAnyWithCborPkg(&rawCompiledIOP{})
	}

	raw := initializeRawCompiledIOP(comp, numRounds)

	if err := serializeNonRoundData(comp, raw); err != nil {
		return nil, err
	}

	serializeRoundSpecificData(comp, raw, numRounds)
	serIOP, err := serializeAnyWithCborPkg(raw)
	logrus.Infof("Successfully serialized compiled IOP and took %vs", time.Since(serTime).Seconds())
	return serIOP, err
}

func serializeRoundSpecificData(comp *wizard.CompiledIOP, raw *rawCompiledIOP, numRounds int) {
	var mu sync.Mutex
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			serializeRoundData(comp, raw, round, &mu)
		}
	}
	parallel.ExecuteChunky(numRounds, work)
}

func serializeRoundData(comp *wizard.CompiledIOP, raw *rawCompiledIOP, round int, mu *sync.Mutex) {
	cols, err := serializeColumns(comp, round)
	handleSerializationError("columns", round, err)

	qParams, err := serializeQueries(&comp.QueriesParams, round)
	handleSerializationError("queries params", round, err)

	qNoParams, err := serializeQueries(&comp.QueriesNoParams, round)
	handleSerializationError("queries no params", round, err)

	coins, err := serializeCoins(comp, round)
	handleSerializationError("coins", round, err)

	subProvers, err := serializeSubProverAction(comp, round)
	handleSerializationError("SubProvers", round, err)

	subVerifiers, err := serializeSubVerifierAction(comp, round)
	handleSerializationError("SubVerifiers", round, err)

	fiatShamirHooks, err := serializeFSHookPS(comp, round)
	handleSerializationError("FiatShamirHookPreSampling", round, err)

	mu.Lock()
	raw.Columns[round] = cols
	raw.QueriesParams[round] = qParams
	raw.QueriesNoParams[round] = qNoParams
	raw.Coins[round] = coins
	raw.Subprovers[round] = subProvers
	raw.SubVerifiers[round] = subVerifiers
	raw.FiatShamirHooksPreSampling[round] = fiatShamirHooks
	mu.Unlock()
}

// DeserializeCompiledIOP unmarshals a wizard.CompiledIOP object from CBOR.
func DeserializeCompiledIOP(data []byte) (*wizard.CompiledIOP, error) {
	comp := NewEmptyCompiledIOP()
	raw := &rawCompiledIOP{}
	if err := deserializeAnyWithCborPkg(data, raw); err != nil {
		return nil, fmt.Errorf("CBOR unmarshal failed: %w", err)
	}

	numRounds := len(raw.Columns)
	if numRounds == 0 {
		return comp, nil
	}

	logrus.Infof("Number of rounds to serde in CompiledIOP: %d", numRounds)

	comp.DummyCompiled = raw.DummyCompiled
	comp.SelfRecursionCount = raw.SelfRecursionCount

	// Register all column IDs first to ensure queries/subprovers can reference them. Structure deserialization
	// to pre-register all referenced data (e.g., columns) before processing dependent fields.
	if err := registerAllColumns(raw, comp, numRounds); err != nil {
		return nil, fmt.Errorf("register columns: %w", err)
	}

	if err := deserializeNonRoundData(raw, comp); err != nil {
		return nil, err
	}

	if err := deserializeRoundData(raw, comp, numRounds); err != nil {
		return nil, err
	}

	if err := deserializeFiatShamirSetup(raw, comp); err != nil {
		return nil, fmt.Errorf("deserialize FiatShamirSetup: %w", err)
	}

	logrus.Println("Successfully deserialized CompiledIOP")
	return comp, nil
}

func serializeNonRoundData(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if err := serializePrecomputed(comp, raw); err != nil {
		return fmt.Errorf("serialize Precomputed: %w", err)
	}
	if err := serializePcsCtxs(comp, raw); err != nil {
		return fmt.Errorf("serialize PcsCtxs: %w", err)
	}
	if err := serializePublicInputs(comp, raw); err != nil {
		return fmt.Errorf("serialize PublicInputs: %w", err)
	}

	if err := serializeFiatShamirSetup(comp, raw); err != nil {
		return fmt.Errorf("serialize FiatShamirSetup: %w", err)
	}

	if err := serializeExtraData(comp, raw); err != nil {
		return fmt.Errorf("serialize ExtraData: %w", err)
	}

	return nil
}

func registerAllColumns(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int) error {
	startTime := time.Now()
	for round := 0; round < numRounds; round++ {
		if round >= len(raw.Columns) {
			return fmt.Errorf("round %d exceeds available columns data", round)
		}
		for _, rawCol := range raw.Columns[round] {
			_, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp)
			if err != nil {
				return fmt.Errorf("round %d first-pass column: %w", round, err)
			}
		}
	}
	logrus.Printf("RegisterAllColumns took %vs \n", time.Since(startTime).Seconds())
	return nil
}

func deserializeNonRoundData(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if err := deserializePrecomputed(raw, comp); err != nil {
		return fmt.Errorf("deserialize Precomputed: %w", err)
	}

	if err := deserializePcsCtxs(raw, comp); err != nil {
		return fmt.Errorf("deserialize PcsCtxs: %w", err)
	}
	if err := deserializePublicInputs(raw, comp); err != nil {
		return fmt.Errorf("deserialize PublicInputs: %w", err)
	}

	if err := deserializeExtraData(raw, comp); err != nil {
		return fmt.Errorf("deserialize ExtraData: %w", err)
	}
	return nil
}

func deserializeRoundData(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int) error {
	for round := 0; round < numRounds; round++ {
		if round >= len(raw.Columns) || round >= len(raw.Coins) ||
			round >= len(raw.QueriesNoParams) || round >= len(raw.QueriesParams) ||
			round >= len(raw.Subprovers) || round >= len(raw.SubVerifiers) ||
			round >= len(raw.FiatShamirHooksPreSampling) {
			return fmt.Errorf("round %d exceeds available data in rawCompiledIOP", round)
		}

		if err := deserializeColumns(raw.Columns[round], comp); err != nil {
			return fmt.Errorf("round %d columns: %w", round, err)
		}

		if err := deserializeCoins(raw.Coins[round], round, comp); err != nil {
			return fmt.Errorf("round %d coins: %w", round, err)
		}

		if err := deserializeQuery(raw.QueriesParams[round], round, &comp.QueriesParams, comp); err != nil {
			return fmt.Errorf("round %d queriesParams: %w", round, err)
		}
		if err := deserializeQuery(raw.QueriesNoParams[round], round, &comp.QueriesNoParams, comp); err != nil {
			return fmt.Errorf("round %d queriesNoParams: %w", round, err)
		}

		if err := deserializeSubProvers(raw.Subprovers[round], round, comp); err != nil {
			return fmt.Errorf("round %d subProvers: %w", round, err)
		}

		if err := deserializeSubVerifiers(raw.SubVerifiers[round], round, comp); err != nil {
			return fmt.Errorf("round %d subVerifiers: %w", round, err)
		}

		if err := deserializeFSHooks(raw.FiatShamirHooksPreSampling[round], round, comp); err != nil {
			return fmt.Errorf("round %d fiatShamirHooksPreSampling: %w", round, err)
		}
	}
	return nil
}

func serializeColumns(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	cols := comp.Columns.AllHandlesAtRound(round)
	rawCols := make([]json.RawMessage, len(cols))

	for i, col := range cols {
		r, err := SerializeValue(reflect.ValueOf(&col), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize column `%v` : %w", col.GetColID(), err)
		}
		rawCols[i] = r
	}
	return rawCols, nil
}

func serializeQueries(register *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query], round int) ([]json.RawMessage, error) {
	qNames := register.AllKeysAt(round)
	rawQueries := make([]json.RawMessage, len(qNames))

	for i, qName := range qNames {
		q := register.Data(qName)
		r, err := SerializeValue(reflect.ValueOf(&q), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize query `%v`: %w", qName, err)
		}
		rawQueries[i] = r
	}
	return rawQueries, nil
}

func serializeCoins(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	coinNames := comp.Coins.AllKeysAt(round)
	rawCoins := make([]json.RawMessage, len(coinNames))

	for i, coinName := range coinNames {
		c := comp.Coins.Data(coinName)
		r, err := SerializeValue(reflect.ValueOf(c), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize coin `%v`: %w", coinName, err)
		}
		rawCoins[i] = r
	}
	return rawCoins, nil
}

func serializeSubProverAction(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	actions := comp.SubProvers.GetOrEmpty(round)
	rawActions := make([]json.RawMessage, len(actions))
	for i, action := range actions {
		// logrus.Printf("Type of step:%d is :%v \n", i, reflect.TypeOf(action))
		// logrus.Printf("Type of *step:%d is: %v \n", i, reflect.TypeOf(&action))

		r, err := SerializeValue(reflect.ValueOf(&action), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize ProverAction %d at round %d: %w", i, round, err)
		}
		rawActions[i] = r
	}
	return rawActions, nil
}

func serializeSubVerifierAction(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	actions := comp.SubVerifiers.GetOrEmpty(round)
	rawActions := make([]json.RawMessage, len(actions))
	for i, action := range actions {
		r, err := SerializeValue(reflect.ValueOf(&action), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize VerifierAction %d at round %d: %w", i, round, err)
		}
		rawActions[i] = r
	}
	return rawActions, nil
}

func serializeFSHookPS(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	actions := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
	rawActions := make([]json.RawMessage, len(actions))
	for i, action := range actions {
		r, err := SerializeValue(reflect.ValueOf(&action), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize FSHook presampling %d at round %d: %w", i, round, err)
		}
		rawActions[i] = r
	}
	return rawActions, nil
}

func serializePrecomputed(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	entries := make([]struct {
		Key   ifaces.ColID
		Value ifaces.ColAssignment
	}, 0, len(comp.Precomputed.ListAllKeys()))
	for _, key := range comp.Precomputed.ListAllKeys() {
		value := comp.Precomputed.MustGet(key)
		entries = append(entries, struct {
			Key   ifaces.ColID
			Value ifaces.ColAssignment
		}{Key: key, Value: value})
	}
	serPrecomputed, err := SerializeValue(reflect.ValueOf(entries), DeclarationMode)
	if err != nil {
		return fmt.Errorf("serialize Precomputed: %w", err)
	}
	raw.Precomputed = serPrecomputed
	return nil
}

func serializePcsCtxs(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if comp.PcsCtxs == nil {
		raw.PcsCtxs = json.RawMessage(NilString)
		logrus.Debug("Serialized PcsCtxs: nil")
		return nil
	}
	serPcsCtxs, err := SerializeValue(reflect.ValueOf(comp.PcsCtxs), DeclarationMode)
	if err != nil {
		return fmt.Errorf("serialize PcsCtxs: %w", err)
	}
	if len(serPcsCtxs) == 0 {
		logrus.Warn("Serialized PcsCtxs resulted in empty data")
		raw.PcsCtxs = json.RawMessage(NilString)
		return nil
	}
	logrus.Debugf("Serialized PcsCtxs: %v bytes", len(serPcsCtxs))
	raw.PcsCtxs = serPcsCtxs
	return nil
}

func serializePublicInputs(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if len(comp.PublicInputs) == 0 {
		raw.PublicInputs = json.RawMessage(NilString)
		return nil
	}
	serPublicInputs, err := SerializeValue(reflect.ValueOf(comp.PublicInputs), DeclarationMode)
	if err != nil {
		return fmt.Errorf("serialize PublicInputs: %w", err)
	}
	raw.PublicInputs = serPublicInputs
	return nil
}

func serializeFiatShamirSetup(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if comp.FiatShamirSetup.IsZero() {
		raw.FiatShamirSetup = json.RawMessage(NilString)
		return nil
	}
	fiatShamirBytes := comp.FiatShamirSetup.Bytes()
	serFiatShamir, err := serializeAnyWithCborPkg(fiatShamirBytes[:])
	if err != nil {
		return fmt.Errorf("serialize FiatShamirSetup: %w", err)
	}
	raw.FiatShamirSetup = serFiatShamir
	return nil
}

func serializeExtraData(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if comp.ExtraData == nil {
		raw.ExtraData = json.RawMessage(NilString)
		return nil
	}
	data, err := serializeMap(reflect.ValueOf(comp.ExtraData), DeclarationMode)
	if err != nil {
		return fmt.Errorf("serialize ExtraData: %w", err)
	}
	raw.ExtraData = data
	return nil
}

func deserializeExtraData(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.ExtraData == nil || bytes.Equal(raw.ExtraData, []byte(NilString)) {
		comp.ExtraData = nil
		return nil
	}
	deSerExtradata, err := deserializeMap(raw.ExtraData, DeclarationMode, reflect.TypeOf(comp.ExtraData), comp)
	if err != nil {
		return fmt.Errorf("deserialize Extradata: %w", err)
	}
	comp.ExtraData = deSerExtradata.Interface().(map[string]any)
	return nil
}

func deserializePrecomputed(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.Precomputed == nil {
		logrus.Warn("Precomputed attribute is nil")
		return nil
	}
	preCompType := reflect.SliceOf(reflect.TypeOf(struct {
		Key   ifaces.ColID
		Value ifaces.ColAssignment
	}{}))
	preCompVal, err := DeserializeValue(raw.Precomputed, DeclarationMode, preCompType, comp)
	if err != nil {
		return fmt.Errorf("deserialize Precomputed: %w", err)
	}
	entries := preCompVal.Interface().([]struct {
		Key   ifaces.ColID
		Value ifaces.ColAssignment
	})
	for _, entry := range entries {
		comp.Precomputed.InsertNew(entry.Key, entry.Value)
	}
	if len(comp.Precomputed.ListAllKeys()) == 0 && len(entries) > 0 {
		logrus.Warn("Deserialized Precomputed mapping is empty despite having entries")
	}
	return nil
}

func deserializePcsCtxs(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.PcsCtxs == nil || bytes.Equal(raw.PcsCtxs, []byte(NilString)) {
		comp.PcsCtxs = nil
		logrus.Debug("Deserialized PcsCtxs: nil")
		return nil
	}
	pcsCtxsType := reflect.TypeOf((*vortex.Ctx)(nil))
	pcsCtxsVal, err := DeserializeValue(raw.PcsCtxs, DeclarationMode, pcsCtxsType, comp)
	if err != nil {
		return fmt.Errorf("deserialize PcsCtxs: %w", err)
	}
	comp.PcsCtxs = pcsCtxsVal.Interface().(*vortex.Ctx)
	logrus.Debug("Deserialized PcsCtxs: non-nil *vortex.Ctx")
	return nil
}

func deserializePublicInputs(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.PublicInputs == nil || bytes.Equal(raw.PublicInputs, []byte(NilString)) {
		return nil
	}
	publicInputsVal, err := DeserializeValue(raw.PublicInputs, DeclarationMode, reflect.TypeOf([]wizard.PublicInput{}), comp)
	if err != nil {
		return fmt.Errorf("deserialize PublicInputs: %w", err)
	}
	comp.PublicInputs = publicInputsVal.Interface().([]wizard.PublicInput)
	return nil
}

func deserializeFiatShamirSetup(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.FiatShamirSetup == nil || bytes.Equal(raw.FiatShamirSetup, []byte(NilString)) {
		return nil
	}
	var fiatShamirBytes [field.Bytes]byte
	if err := deserializeAnyWithCborPkg(raw.FiatShamirSetup, &fiatShamirBytes); err != nil {
		return fmt.Errorf("deserialize FiatShamirSetup: %w", err)
	}
	var fiatShamir field.Element
	fiatShamir.SetBytes(fiatShamirBytes[:])
	comp.FiatShamirSetup = fiatShamir
	return nil
}

func deserializeColumns(rawCols []json.RawMessage, comp *wizard.CompiledIOP) error {
	for _, rawCol := range rawCols {
		if _, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp); err != nil {
			return err
		}
	}
	return nil
}

func deserializeCoins(rawCoins []json.RawMessage, round int, comp *wizard.CompiledIOP) error {
	for _, rawCoin := range rawCoins {
		v, err := DeserializeValue(rawCoin, DeclarationMode, reflect.TypeOf(coin.Info{}), comp)
		if err != nil {
			return fmt.Errorf("failed to deserialize coin: %w", err)
		}
		coin := v.Interface().(coin.Info)
		comp.Coins.AddToRound(round, coin.Name, coin)
	}
	return nil
}

func deserializeQuery(rawQueries []json.RawMessage, round int, register *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query], comp *wizard.CompiledIOP) error {
	for _, rawQ := range rawQueries {
		v, err := DeserializeValue(rawQ, DeclarationMode, queryType, comp)
		if err != nil {
			return fmt.Errorf("failed to deserialize query: %w", err)
		}
		q := v.Interface().(ifaces.Query)
		register.AddToRound(round, q.Name(), q)
	}
	return nil
}

func deserializeSubProvers(rawSubProvers []json.RawMessage, round int, comp *wizard.CompiledIOP) error {
	for i, rawAction := range rawSubProvers {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.ProverAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize ProverAction %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.ProverAction)
		comp.RegisterProverAction(round, action)
	}
	return nil
}

func deserializeSubVerifiers(rawSubVerifiers []json.RawMessage, round int, comp *wizard.CompiledIOP) error {
	for i, rawAction := range rawSubVerifiers {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize VerifierAction %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.VerifierAction)
		comp.RegisterVerifierAction(round, action)
	}
	return nil
}

func deserializeFSHooks(rawFSHooks []json.RawMessage, round int, comp *wizard.CompiledIOP) error {
	for i, rawAction := range rawFSHooks {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize FiatShamirHookPreSampling %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.VerifierAction)
		comp.FiatShamirHooksPreSampling.AppendToInner(round, action)
	}
	return nil
}

func RegisterColumns(from, to *wizard.CompiledIOP) error {
	startTime := time.Now()
	numRounds := from.Columns.NumRounds()
	for round := 0; round < numRounds; round++ {
		// Retrieve all colIDs
		colIDs := from.Columns.AllKeysAt(round)
		for _, colID := range colIDs {
			if !to.Columns.Exists(colID) {
				col := from.Columns.GetHandle(colID).(column.Natural)
				to.InsertColumn(round, colID, col.Size(), col.Status())
			}
		}
	}
	logrus.Printf("Registering from -> to compiled IOP took %vs", time.Since(startTime).Seconds())
	return nil
}

func handleSerializationError(context string, round int, err error) {
	if err != nil {
		utils.Panic("round %d: serialize %s: %v", round, context, err)
	}
}
