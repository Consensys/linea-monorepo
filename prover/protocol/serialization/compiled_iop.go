package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/sirupsen/logrus"
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
	DummyCompiled              bool                `json:"dummyCompiled"`
}

func NewEmptyCompiledIOP() *wizard.CompiledIOP {
	return &wizard.CompiledIOP{
		Columns:                    column.NewStore(),
		QueriesParams:              wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		QueriesNoParams:            wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		Coins:                      wizard.NewRegister[coin.Name, coin.Info](),
		Precomputed:                collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		SubProvers:                 collection.VecVec[wizard.ProverAction]{},
		SubVerifiers:               collection.VecVec[wizard.VerifierAction]{},
		FiatShamirHooksPreSampling: collection.VecVec[wizard.VerifierAction]{},
	}
}

// SerializeCompiledIOP marshals a [wizard.CompiledIOP] object into JSON. This is
// meant to allow deserializing the IOP during the prover time instead of
// recompiling everything every time we want to run the prover.
//
// Example:
//
//	 	comp := wizard.Compile(myBuilder, myCompilerSuite...)
//		marshaled, err := SerializeCompiledIOP(comp)
//		if err != nil {
//			panic(err)
//		}
func SerializeCompiledIOP(comp *wizard.CompiledIOP) ([]byte, error) {
	numRounds := comp.NumRounds()

	logrus.Infof("Serializing IOP with numRounds:%d \n", numRounds)
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
	}

	// Mutex to protect slice assignments
	var mu sync.Mutex

	// Work function for ExecuteChunky
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

	// Run parallel execution
	parallel.ExecuteChunky(numRounds, work)

	serIOP, err := serializeAnyWithCborPkg(raw)
	logrus.Info("Successfully serializated compiled IOP")

	// Serialize the final rawCompiledIOP
	return serIOP, err
}

// DeserializeCompiledIOP unmarshals a [wizard.CompiledIOP] object or returns an error
// if the marshalled object does not have the right format. The
// deserialized object can then be used for proving or verification via the
// functions [wizard.Prove] or [wizard.Verify].
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

	// Mutex to protect updates to comp
	var mu sync.Mutex

	if err := deserializeColumnsAndCoins(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("columns and coins: %w", err)
	}
	if err := deserializeQueries(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("queries: %w", err)
	}

	if err := deserializeSubProverAction(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding subProverAction:%w", err)
	}

	if err := deserializeSubVerifierAction(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding subVerifierAction:%w", err)
	}

	if err := deserializeFSHookPS(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("error while decoding subVerifierAction:%w", err)
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
				if err := deserializeColumns(raw.Columns[round], comp); err != nil {
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

func deserializeColumns(rawCols []json.RawMessage, comp *wizard.CompiledIOP) error {
	for _, rawCol := range rawCols {
		if _, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp); err != nil {
			return err
		}
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
