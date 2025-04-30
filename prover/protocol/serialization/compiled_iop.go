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
)

// rawCompiledIOP represents the serialized form of CompiledIOP.
type rawCompiledIOP struct {
	Columns         [][]json.RawMessage `json:"columns"`
	QueriesParams   [][]json.RawMessage `json:"queriesParams"`
	QueriesNoParams [][]json.RawMessage `json:"queriesNoParams"`
	Coins           [][]json.RawMessage `json:"coins"`
	DummyCompiled   bool                `json:"dummyCompiled"`
}

func newEmptyCompiledIOP() *wizard.CompiledIOP {
	return &wizard.CompiledIOP{
		Columns:         column.NewStore(),
		QueriesParams:   wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		QueriesNoParams: wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
		Coins:           wizard.NewRegister[coin.Name, coin.Info](),
		Precomputed:     collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
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
	if numRounds == 0 {
		return serializeAnyWithCborPkg(&rawCompiledIOP{})
	}

	// Initialize rawCompiledIOP with pre-allocated slices
	raw := &rawCompiledIOP{
		Columns:         make([][]json.RawMessage, numRounds),
		QueriesParams:   make([][]json.RawMessage, numRounds),
		QueriesNoParams: make([][]json.RawMessage, numRounds),
		Coins:           make([][]json.RawMessage, numRounds),
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

			// Safely assign results to raw
			mu.Lock()
			raw.Columns[round] = cols
			raw.QueriesParams[round] = qParams
			raw.QueriesNoParams[round] = qNoParams
			raw.Coins[round] = coins
			mu.Unlock()
		}
	}

	// Run parallel execution
	parallel.ExecuteChunky(numRounds, work)

	// Serialize the final rawCompiledIOP
	return serializeAnyWithCborPkg(raw)
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

// DeserializeCompiledIOP unmarshals a [wizard.CompiledIOP] object or returns an error
// if the marshalled object does not have the right format. The
// deserialized object can then be used for proving or verification via the
// functions [wizard.Prove] or [wizard.Verify].
func DeserializeCompiledIOP(data []byte) (*wizard.CompiledIOP, error) {
	comp := newEmptyCompiledIOP()
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

	if err := deserializeColumnsAndCoinsParallel(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("columns and coins: %w", err)
	}
	if err := deserializeQueriesParallel(raw, comp, numRounds, &mu); err != nil {
		return nil, fmt.Errorf("queries: %w", err)
	}

	return comp, nil
}

func deserializeColumnsAndCoinsParallel(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
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

func deserializeQueriesParallel(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int, mu *sync.Mutex) error {
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
