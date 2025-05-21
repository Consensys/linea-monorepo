package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

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

// SerializeCompiledIOP marshals a wizard.CompiledIOP object into CBOR.
func SerializeCompiledIOP(comp *wizard.CompiledIOP) ([]byte, error) {
	numRounds := comp.NumRounds()
	if numRounds == 0 {
		return serializeAnyWithCborPkg(&rawCompiledIOP{})
	}

	// Initialize rawCompiledIOP with pre-allocated slices
	raw := &rawCompiledIOP{
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

	// Serialize non-round-specific data
	if err := serializeNonRoundData(comp, raw); err != nil {
		return nil, err
	}

	// Serialize round-specific data in parallel
	var mu sync.Mutex
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			cols, err := serializeColumns(comp, round)
			if err != nil {
				utils.Panic("round %d: serialize columns: %v", round, err)
			}

			qParams, err := serializeQueries(&comp.QueriesParams, round)
			if err != nil {
				utils.Panic("round %d: serialize queries params: %v", round, err)
			}

			qNoParams, err := serializeQueries(&comp.QueriesNoParams, round)
			if err != nil {
				utils.Panic("round %d: serialize queries no params: %v", round, err)
			}

			coins, err := serializeCoins(comp, round)
			if err != nil {
				utils.Panic("round %d: serialize coins: %v", round, err)
			}

			subProvers, err := serializeSubProverAction(comp, round)
			if err != nil {
				utils.Panic("round %d: serialize SubProvers: %v", round, err)
			}

			subVerifiers, err := serializeSubVerifierAction(comp, round)
			if err != nil {
				utils.Panic("round %d: serialize SubVerifiers: %v", round, err)
			}

			fiatShamirHooks, err := serializeFSHookPS(comp, round)
			if err != nil {
				utils.Panic("round %d: serialize FiatShamirHookPreSampling: %v", round, err)
			}

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
	}

	parallel.ExecuteChunky(numRounds, work)

	serIOP, err := serializeAnyWithCborPkg(raw)
	logrus.Info("Successfully serialized compiled IOP")
	return serIOP, err
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

	logrus.Infof("Number of rounds in SerDe CompiledIOP: %d", numRounds)

	// Set primitive fields
	comp.DummyCompiled = raw.DummyCompiled
	comp.SelfRecursionCount = raw.SelfRecursionCount

	var mu sync.Mutex

	// Register all column IDs first to ensure queries/subprovers can reference them
	if err := registerAllColumns(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("register columns: %w", err)
	}

	// Deserialize non-round-specific data (Precomputed, PcsCtxs, PublicInputs)
	if err := deserializeNonRoundData(raw, comp); err != nil {
		return nil, err
	}

	// Deserialize round-specific data
	if err := deserializeRoundData(raw, comp, numRounds, &mu); err != nil {
		return nil, err
	}

	// Deserialize FiatShamirSetup last, as it may depend on prior state
	if err := deserializeFiatShamirSetup(raw, comp); err != nil {
		return nil, fmt.Errorf("deserialize FiatShamirSetup: %w", err)
	}

	return comp, nil
}

// serializeNonRoundData serializes non-round-specific attributes.
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

// registerAllColumns registers all column IDs from Columns and Precomputed.
func registerAllColumns(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	// First pass: Register column IDs from Columns
	for round := 0; round < numRounds; round++ {
		if round >= len(raw.Columns) {
			return fmt.Errorf("round %d exceeds available columns data", round)
		}
		for _, rawCol := range raw.Columns[round] {
			mu.Lock()
			v, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp)
			if err != nil {
				mu.Unlock()
				return fmt.Errorf("round %d first-pass column: %w", round, err)
			}
			col := v.Interface().(column.Natural)
			colIDs := GetAllColIDs(col)
			for _, colID := range colIDs {
				if !comp.Columns.Exists(colID) {
					comp.InsertColumn(round, colID, col.Size(), col.Status())
					logrus.Infof("First-pass registered column %v in round %d", colID, round)
				}
			}
			mu.Unlock()
		}
	}

	// Register Precomputed columns
	if err := deserializePrecomputed(raw, comp); err != nil {
		return fmt.Errorf("deserialize Precomputed: %w", err)
	}
	for _, colID := range comp.Precomputed.ListAllKeys() {
		mu.Lock()
		if !comp.Columns.Exists(colID) {
			comp.InsertColumn(0, colID, 0, column.Committed)
			logrus.Infof("Registered precomputed column %v", colID)
		}
		mu.Unlock()
	}

	return nil
}

// deserializeNonRoundData deserializes non-round-specific attributes.
func deserializeNonRoundData(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	// Precomputed is already deserialized in registerAllColumns
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

// deserializeRoundData deserializes round-specific attributes in order.
func deserializeRoundData(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	for round := 0; round < numRounds; round++ {
		if round >= len(raw.Columns) || round >= len(raw.Coins) ||
			round >= len(raw.QueriesNoParams) || round >= len(raw.QueriesParams) ||
			round >= len(raw.Subprovers) || round >= len(raw.SubVerifiers) ||
			round >= len(raw.FiatShamirHooksPreSampling) {
			return fmt.Errorf("round %d exceeds available data in rawCompiledIOP", round)
		}

		// Columns
		if err := deserializeColumns(raw.Columns[round], round, comp, mu); err != nil {
			return fmt.Errorf("round %d columns: %w", round, err)
		}

		// Coins
		if err := deserializeCoins(raw.Coins[round], round, comp, mu); err != nil {
			return fmt.Errorf("round %d coins: %w", round, err)
		}

		// Queries
		if err := deserializeQuery(raw.QueriesParams[round], round, &comp.QueriesParams, comp); err != nil {
			return fmt.Errorf("round %d queriesParams: %w", round, err)
		}
		if err := deserializeQuery(raw.QueriesNoParams[round], round, &comp.QueriesNoParams, comp); err != nil {
			return fmt.Errorf("round %d queriesNoParams: %w", round, err)
		}

		// SubProvers
		if err := deserializeSubProvers(raw.Subprovers[round], round, comp, mu); err != nil {
			return fmt.Errorf("round %d subProvers: %w", round, err)
		}

		// SubVerifiers
		if err := deserializeSubVerifiers(raw.SubVerifiers[round], round, comp, mu); err != nil {
			return fmt.Errorf("round %d subVerifiers: %w", round, err)
		}

		// FiatShamirHooksPreSampling
		if err := deserializeFSHooks(raw.FiatShamirHooksPreSampling[round], round, comp, mu); err != nil {
			return fmt.Errorf("round %d fiatShamirHooksPreSampling: %w", round, err)
		}
	}
	return nil
}

// Round-Specific Serialization Functions

func serializeColumns(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	cols := comp.Columns.AllHandlesAtRound(round)
	rawCols := make([]json.RawMessage, 0, len(cols))
	seen := make(map[ifaces.ColID]bool)

	for _, col := range cols {
		colIDs := GetAllColIDs(col)
		for _, colID := range colIDs {
			if seen[colID] {
				continue
			}
			seen[colID] = true
			// logrus.Infof("Serializing column %v in round %d", colID, round)
			r, err := SerializeValue(reflect.ValueOf(&col), DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize column `%v`: %w", colID, err)
			}
			rawCols = append(rawCols, r)
		}
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

// Non-Round-Specific Serialization Functions

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

// serializeExtraData serializes the ExtraData map into raw.ExtraData as a JSON raw message.
func serializeExtraData(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if comp.ExtraData == nil {
		raw.ExtraData = json.RawMessage(NilString)
		logrus.Debug("Serialized ExtraData: nil or empty")
		return nil
	}

	data, err := json.Marshal(comp.ExtraData)
	if err != nil {
		return fmt.Errorf("serialize ExtraData: %w", err)
	}

	raw.ExtraData = data
	logrus.Debugf("Serialized ExtraData: %d bytes", len(data))
	return nil
}

// Non-Round-Specific Deserialization Functions

// deserializeExtraData deserializes raw.ExtraData into comp.ExtraData as a map[string]any.
func deserializeExtraData(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.ExtraData == nil || bytes.Equal(raw.ExtraData, []byte(NilString)) {
		comp.ExtraData = nil
		logrus.Debug("Deserialized ExtraData: nil")
		return nil
	}

	var extraData map[string]any
	if err := json.Unmarshal(raw.ExtraData, &extraData); err != nil {
		return fmt.Errorf("deserialize ExtraData: %w", err)
	}

	// Special case: key= VERIFYING_KEY
	// for key, val := range extraData {
	// 	if key == "VERIFYING_KEY" {
	// 		extraData[key] = val.(fr.Element)
	// 	}
	// }

	comp.ExtraData = extraData
	logrus.Debugf("Deserialized ExtraData: %d entries", len(extraData))
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
		// logrus.Infof("Deserialized precomputed column %v", entry.Key)
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

// Round-Specific Deserialization Functions

func deserializeColumns(rawCols []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for _, rawCol := range rawCols {
		mu.Lock()
		v, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp)
		if err != nil {
			mu.Unlock()
			return fmt.Errorf("failed to deserialize column: %w", err)
		}
		col := v.Interface().(column.Natural)
		colIDs := GetAllColIDs(col)
		for _, colID := range colIDs {
			if !comp.Columns.Exists(colID) {
				comp.InsertColumn(round, colID, col.Size(), col.Status())
				logrus.Infof("Registered column %v in round %d", colID, round)
			} else {
				logrus.Debugf("Column %v already exists in round %d", colID, round)
			}
		}
		mu.Unlock()
	}
	return nil
}

func deserializeCoins(rawCoins []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for _, rawCoin := range rawCoins {
		v, err := DeserializeValue(rawCoin, DeclarationMode, reflect.TypeOf(coin.Info{}), comp)
		if err != nil {
			return fmt.Errorf("failed to deserialize coin: %w", err)
		}
		coin := v.Interface().(coin.Info)
		mu.Lock()
		comp.Coins.AddToRound(round, coin.Name, coin)
		mu.Unlock()
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
		//logrus.Infof("Deserializing query %v in round %d", q.Name(), round)
		register.AddToRound(round, q.Name(), q)
	}
	return nil
}

func deserializeSubProvers(rawSubProvers []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for i, rawAction := range rawSubProvers {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.ProverAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize ProverAction %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.ProverAction)
		mu.Lock()
		comp.RegisterProverAction(round, action)
		mu.Unlock()
	}
	return nil
}

func deserializeSubVerifiers(rawSubVerifiers []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for i, rawAction := range rawSubVerifiers {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize VerifierAction %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.VerifierAction)
		mu.Lock()
		comp.RegisterVerifierAction(round, action)
		mu.Unlock()
	}
	return nil
}

func deserializeFSHooks(rawFSHooks []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for i, rawAction := range rawFSHooks {
		v, err := DeserializeValue(rawAction, DeclarationMode, reflect.TypeOf((*wizard.VerifierAction)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("could not deserialize FiatShamirHookPreSampling %d at round %d: %w", i, round, err)
		}
		action := v.Interface().(wizard.VerifierAction)
		mu.Lock()
		comp.FiatShamirHooksPreSampling.AppendToInner(round, action)
		mu.Unlock()
	}
	return nil
}
