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
	}
}

// SerializeCompiledIOP marshals a [wizard.CompiledIOP] object into JSON.
// This allows deserializing the IOP during prover time instead of recompiling everything.
// Example:
//
//	 	comp := wizard.Compile(myBuilder, myCompilerSuite...)
//		marshaled, err := SerializeCompiledIOP(comp)
//		if err != nil {
//			panic(err)
//		}
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

	// Serialize non-round specific params

	// Serialize Precomputed attribute
	if err := serializePrecomputed(comp, raw); err != nil {
		return nil, fmt.Errorf("serialize Precomputed: %w", err)
	}

	// Serialize PcsCtxs attribute
	if err := serializePcsCtxs(comp, raw); err != nil {
		return nil, fmt.Errorf("serialize PcsCtxs: %w", err)
	}

	// Serialize FiatShamirSetup attribute
	if err := serializeFiatShamirSetup(comp, raw); err != nil {
		return nil, err
	}

	// Serialize PublicInputs attribute
	if err := serializePublicInputs(comp, raw); err != nil {
		return nil, err
	}

	// Serialize round-specific attributes in parallel
	var mu sync.Mutex
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			var localErr error

			// Serialize columns
			cols, err := serializeColumns(comp, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize columns: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize queries with params
			qParams, err := serializeQueries(&comp.QueriesParams, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize queries params: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize queries without params
			qNoParams, err := serializeQueries(&comp.QueriesNoParams, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize queries no params: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize coins
			coins, err := serializeCoins(comp, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize coins: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize subProvers
			subProvers, err := serializeSubProverAction(comp, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize SubProvers: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize subVerifiers
			subVerifiers, err := serializeSubVerifierAction(comp, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize SubVerifiers: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Serialize FiatShamirHooksPresampling
			fiatShamirHooks, err := serializeFSHookPS(comp, round)
			if err != nil {
				localErr = fmt.Errorf("round %d: serialize FiatShamirHookPreSampling: %w", round, err)
				utils.Panic(localErr.Error())
			}

			// Safely assign results to raw
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

// DeserializeCompiledIOP unmarshals a [wizard.CompiledIOP] object or returns an error
// if the marshalled object does not have the right format.
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

	// Set primitive fields
	comp.DummyCompiled = raw.DummyCompiled
	comp.SelfRecursionCount = raw.SelfRecursionCount

	// Deserialize Precomputed attribute
	if err := deserializePrecomputed(raw, comp); err != nil {
		return nil, fmt.Errorf("deserialize Precomputed: %w", err)
	}

	// Deserialize PcsCtxs attribute
	if err := deserializePcsCtxs(raw, comp); err != nil {
		return nil, fmt.Errorf("deserialize PcsCtxs: %w", err)
	}

	// Deserialize Artefacts attribute
	// if err := deserializeArtefacts(raw, comp); err != nil {
	// 	return nil, fmt.Errorf("deserialize Artefacts: %w", err)
	// }

	// Deserialize FiatShamirSetup attribute
	if err := deserializeFiatShamirSetup(raw, comp); err != nil {
		return nil, err
	}

	// Deserialize round-specific attributes in parallel
	var mu sync.Mutex
	if err := deserializeColumnsAndCoins(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("columns and coins: %w", err)
	}
	if err := deserializeQueries(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("queries: %w", err)
	}
	if err := deserializeSubProverAction(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding subProverAction: %w", err)
	}
	if err := deserializeSubVerifierAction(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding subVerifierAction: %w", err)
	}
	if err := deserializeFSHookPS(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding FiatShamirHookPreSampling: %w", err)
	}

	// IMPORTANT: Deserialize PublicInputs attribute at the last since PublicInputs contains Acc
	// ifaces.Accessor (e.g., accessor.LocalOpening), which references column IDs.
	// If Columns are not deserialized first, calls to column.Store.GetHandle (via MustGet) fail,
	// causing the panic.
	if err := deserializePublicInputs(raw, comp); err != nil {
		return nil, err
	}
	return comp, nil
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
			return nil, fmt.Errorf("could not serialize query `%v` : %w", qName, err)
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
			return nil, fmt.Errorf("could not serialize coin `%v` : %w", coinName, err)
		}
		rawCoins[i] = r
	}

	return rawCoins, nil
}

func serializeSubProverAction(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	// Access ProverAction slice for the round
	actionsP := comp.SubProvers.GetOrEmpty(round)
	rawProverActions := make([]json.RawMessage, len(actionsP))
	for i, actionP := range actionsP {
		// MUST pass only pointer here: For example:
		// action kind:ptr type:*byte32cmp.Bytes32CmpProverAction
		// *action kind:ptr type:*wizard.ProverAction
		r, err := SerializeValue(reflect.ValueOf(&actionP), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize ProverAction %d at round %d: %w", i, round, err)
		}
		rawProverActions[i] = r
	}
	return rawProverActions, nil
}

func serializeSubVerifierAction(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	// Access VerifierAction slice for the round
	actionsV := comp.SubVerifiers.GetOrEmpty(round)
	rawVerifierActions := make([]json.RawMessage, len(actionsV))
	for i, actionV := range actionsV {
		r, err := SerializeValue(reflect.ValueOf(&actionV), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize VerifierAction %d at round %d: %w", i, round, err)
		}
		rawVerifierActions[i] = r
	}
	return rawVerifierActions, nil
}

func serializeFSHookPS(comp *wizard.CompiledIOP, round int) ([]json.RawMessage, error) {
	actionsFSHook := comp.FiatShamirHooksPreSampling.GetOrEmpty(round)
	rawFSHookPS := make([]json.RawMessage, len(actionsFSHook))
	for i, actionFS := range actionsFSHook {
		r, err := SerializeValue(reflect.ValueOf(&actionFS), DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize FSHook presampling %d at round %d: %w", i, round, err)
		}
		rawFSHookPS[i] = r
	}
	return rawFSHookPS, nil
}

// serializePrecomputed serializes the Precomputed attribute into the rawCompiledIOP.
func serializePrecomputed(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	precomputedEntries := make([]struct {
		Key   ifaces.ColID
		Value ifaces.ColAssignment
	}, 0, len(comp.Precomputed.ListAllKeys()))
	for _, key := range comp.Precomputed.ListAllKeys() {
		value := comp.Precomputed.MustGet(key)
		precomputedEntries = append(precomputedEntries, struct {
			Key   ifaces.ColID
			Value ifaces.ColAssignment
		}{Key: key, Value: value})
	}
	serPrecomputed, err := SerializeValue(reflect.ValueOf(precomputedEntries), DeclarationMode)
	if err != nil {
		return fmt.Errorf("serialize Precomputed: %w", err)
	}
	raw.Precomputed = serPrecomputed
	return nil
}

// serializePcsCtxs serializes the PcsCtxs attribute into the rawCompiledIOP.
func serializePcsCtxs(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if comp.PcsCtxs != nil {
		serPcsCtxs, err := SerializeValue(reflect.ValueOf(comp.PcsCtxs), DeclarationMode)
		if err != nil {
			return fmt.Errorf("serialize PcsCtxs: %w", err)
		}
		raw.PcsCtxs = serPcsCtxs
	} else {
		raw.PcsCtxs = json.RawMessage(NilString)
	}
	return nil
}

func serializeFiatShamirSetup(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if !comp.FiatShamirSetup.IsZero() {
		fiatShamirBytes := comp.FiatShamirSetup.Bytes()
		serFiatShamir, err := serializeAnyWithCborPkg(fiatShamirBytes[:])
		if err != nil {
			return fmt.Errorf("serialize FiatShamirSetup: %w", err)
		}
		raw.FiatShamirSetup = serFiatShamir
	} else {
		raw.FiatShamirSetup = json.RawMessage(NilString)
	}
	return nil
}

func serializePublicInputs(comp *wizard.CompiledIOP, raw *rawCompiledIOP) error {
	if len(comp.PublicInputs) > 0 {
		serPublicInputs, err := SerializeValue(reflect.ValueOf(comp.PublicInputs), DeclarationMode)
		if err != nil {
			return fmt.Errorf("serialize PublicInputs: %w", err)
		}
		raw.PublicInputs = serPublicInputs
	} else {
		raw.PublicInputs = json.RawMessage(NilString)
	}
	return nil
}

func deserializePublicInputs(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.PublicInputs != nil && !bytes.Equal(raw.PublicInputs, []byte(NilString)) {
		publicInputsVal, err := DeserializeValue(raw.PublicInputs, DeclarationMode, reflect.TypeOf([]wizard.PublicInput{}), comp)
		if err != nil {
			return fmt.Errorf("deserialize PublicInputs: %w", err)
		}
		comp.PublicInputs = publicInputsVal.Interface().([]wizard.PublicInput)
	}
	return nil
}

func deserializeFiatShamirSetup(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.FiatShamirSetup != nil && !bytes.Equal(raw.FiatShamirSetup, []byte(NilString)) {
		var fiatShamirBytes [field.Bytes]byte
		if err := deserializeAnyWithCborPkg(raw.FiatShamirSetup, &fiatShamirBytes); err != nil {
			return fmt.Errorf("deserialize FiatShamirSetup: %w", err)
		}
		var fiatShamir field.Element
		fiatShamir.SetBytes(fiatShamirBytes[:])
		comp.FiatShamirSetup = fiatShamir
	}
	return nil
}

// deserializePcsCtxs deserializes the PcsCtxs attribute from the rawCompiledIOP.
func deserializePcsCtxs(raw *rawCompiledIOP, comp *wizard.CompiledIOP) error {
	if raw.PcsCtxs != nil && !bytes.Equal(raw.PcsCtxs, []byte(NilString)) {
		pcsCtxsVal, err := DeserializeValue(raw.PcsCtxs, DeclarationMode, reflect.TypeOf((*any)(nil)).Elem(), comp)
		if err != nil {
			return fmt.Errorf("deserialize PcsCtxs: %w", err)
		}
		comp.PcsCtxs = pcsCtxsVal.Interface()
	}
	return nil
}

// deserializePrecomputed deserializes the Precomputed attribute from the rawCompiledIOP.
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
	precomputedEntries := preCompVal.Interface().([]struct {
		Key   ifaces.ColID
		Value ifaces.ColAssignment
	})
	for _, entry := range precomputedEntries {
		comp.Precomputed.InsertNew(entry.Key, entry.Value)
	}
	if len(comp.Precomputed.ListAllKeys()) == 0 && len(precomputedEntries) > 0 {
		logrus.Warn("Deserialized Precomputed mapping is empty despite having entries")
	}
	return nil
}

// deserializeSubProverAction deserializes SubProvers for all rounds in parallel
func deserializeSubProverAction(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			if round >= len(raw.Subprovers) {
				utils.Panic("invalid round number %d, exceeds available SubProvers rounds", round)
			}
			if err := deserializeSubProvers(raw.Subprovers[round], round, comp, mu); err != nil {
				utils.Panic("round %d SubProvers: %v", round, err)
			}
		}
	}
	parallel.ExecuteChunky(numRounds, work)
	return nil
}

// deserializeSubVerifierAction: deserializes Subverifiers action for all rounds in parallel
func deserializeSubVerifierAction(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			if round >= len(raw.SubVerifiers) {
				utils.Panic("invalid round number %d, exceeds available SubVerifiers rounds", round)
			}
			if err := deserializeSubVerifiers(raw.SubVerifiers[round], round, comp, mu); err != nil {
				utils.Panic("round %d SubProvers: %v", round, err)
			}
		}
	}
	parallel.ExecuteChunky(numRounds, work)
	return nil
}

// deserializeFSHookPS deserializes FiatShamirHooksPreSampling actions for all rounds in parallel
func deserializeFSHookPS(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			if round >= len(raw.FiatShamirHooksPreSampling) {
				utils.Panic("invalid round number %d, exceeds available FiatShamirHooksPreSampling rounds", round)
			}
			if err := deserializeFSHooks(raw.FiatShamirHooksPreSampling[round], round, comp, mu); err != nil {
				utils.Panic("round %d FiatShamirHooksPreSampling: %v", round, err)
			}
		}
	}
	parallel.ExecuteChunky(numRounds, work)
	return nil
}

// deserializeFSHooks deserializes FiatShamirHooksPreSampling actions for a single round
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

// deserializeSubProvers deserializes SubProvers for a single round
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

// deserializeSubVerifiers deserializes Subverifiers for a single round
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

func deserializeColumnsAndCoins(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	work := func(start, stop int) {
		var wg sync.WaitGroup
		for round := start; round < stop; round++ {
			if round >= len(raw.Columns) || round >= len(raw.Coins) {
				utils.Panic("invalid round number %d, exceeds available rounds", round)
			}

			wg.Add(2)

			// Run deserializeCoins in parallel
			go func(round int) {
				defer wg.Done()
				if err := deserializeCoins(raw.Coins[round], round, comp, mu); err != nil {
					utils.Panic("round %d coins: %v", round, err)
				}
			}(round)

			// Run deserializeColumns in parallel
			go func(round int) {
				defer wg.Done()
				if err := deserializeColumns(raw.Columns[round], comp, mu); err != nil {
					utils.Panic("round %d columns: %v", round, err)
				}
			}(round)
		}
		wg.Wait()
	}

	parallel.ExecuteChunky(numRounds, work)
	return nil
}

func deserializeQueries(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
	work := func(start, stop int) {
		for round := start; round < stop; round++ {
			if round >= len(raw.QueriesNoParams) || round >= len(raw.QueriesParams) {
				utils.Panic("invalid round number %d, exceeds available rounds", round)
			}
			if err := deserializeQuery(raw.QueriesNoParams[round], round, &comp.QueriesNoParams, comp, mu); err != nil {
				utils.Panic("round %d queriesNoParams: %v", round, err)
			}
			if err := deserializeQuery(raw.QueriesParams[round], round, &comp.QueriesParams, comp, mu); err != nil {
				utils.Panic("round %d queriesParams: %v", round, err)
			}
		}
	}

	parallel.ExecuteChunky(numRounds, work)
	return nil
}

func deserializeCoins(rawCoins []json.RawMessage, round int, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for _, rawCoin := range rawCoins {
		v, err := DeserializeValue(rawCoin, DeclarationMode, reflect.TypeOf(coin.Info{}), comp)
		if err != nil {
			return err
		}
		coin := v.Interface().(coin.Info)
		mu.Lock()
		comp.Coins.AddToRound(round, coin.Name, coin)
		mu.Unlock()
	}
	return nil
}

func deserializeColumns(rawCols []json.RawMessage, comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for _, rawCol := range rawCols {
		mu.Lock()
		if _, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp); err != nil {
			mu.Unlock()
			return err
		}
		mu.Unlock()
	}
	return nil
}

func deserializeQuery(rawQueries []json.RawMessage, round int, register *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query], comp *wizard.CompiledIOP, mu *sync.Mutex) error {
	for _, rawQ := range rawQueries {
		v, err := DeserializeValue(rawQ, DeclarationMode, queryType, comp)
		if err != nil {
			return err
		}
		q := v.Interface().(ifaces.Query)
		mu.Lock()
		register.AddToRound(round, q.Name(), q)
		mu.Unlock()
	}
	return nil
}
