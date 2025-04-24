package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// rawCompiledIOP represents the serialized form of CompiledIOP.
type rawCompiledIOP struct {
	Columns         [][]json.RawMessage `json:"columns"`
	QueriesParams   [][]json.RawMessage `json:"queriesParams"`
	QueriesNoParams [][]json.RawMessage `json:"queriesNoParams"`
	Coins           [][]json.RawMessage `json:"coins"`
	DummyCompiled   bool                `json:"dummyCompiled"`
}

// SerializeCompiledIOP marshals a [wizard.CompiledIOP] object into JSON.
func SerializeCompiledIOP(comp *wizard.CompiledIOP) ([]byte, error) {
	raw := &rawCompiledIOP{}
	numRounds := comp.NumRounds()

	for round := 0; round < numRounds; round++ {
		rawCols, err := serializeColumns(comp, round)
		if err != nil {
			return nil, err
		}
		rawQParams, err := serializeQueries(&comp.QueriesParams, round)
		if err != nil {
			return nil, err
		}
		rawQNoParams, err := serializeQueries(&comp.QueriesNoParams, round)
		if err != nil {
			return nil, err
		}
		rawCoins, err := serializeCoins(comp, round)
		if err != nil {
			return nil, err
		}

		raw.Columns = append(raw.Columns, rawCols)
		raw.QueriesParams = append(raw.QueriesParams, rawQParams)
		raw.QueriesNoParams = append(raw.QueriesNoParams, rawQNoParams)
		raw.Coins = append(raw.Coins, rawCoins)
	}

	res, _ := serializeAnyWithCborPkg(raw)
	return res, nil
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

// DeserializeCompiledIOP unmarshals a [wizard.CompiledIOP] object or returns an error if the marshalled object does not have the right format.
func DeserializeCompiledIOP(data []byte) (*wizard.CompiledIOP, error) {
	comp := newEmptyCompiledIOP()
	raw := &rawCompiledIOP{}
	if err := deserializeAnyWithCborPkg(data, raw); err != nil {
		return nil, err
	}

	numRounds := len(raw.Columns)

	if err := deserializeColumnsAndCoins(raw, comp, numRounds); err != nil {
		return nil, err
	}

	if err := deserializeQueries(raw, comp, numRounds); err != nil {
		return nil, err
	}

	return comp, nil
}

func deserializeColumnsAndCoins(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int) error {
	for round := 0; round < numRounds; round++ {
		if err := deserializeCoins(raw.Coins[round], round, comp); err != nil {
			return err
		}
		if err := deserializeColumns(raw.Columns[round], comp); err != nil {
			return err
		}
	}
	return nil
}

func deserializeCoins(rawCoins []json.RawMessage, round int, comp *wizard.CompiledIOP) error {
	for _, rawCoin := range rawCoins {
		v, err := DeserializeValue(rawCoin, DeclarationMode, reflect.TypeOf(coin.Info{}), comp)
		if err != nil {
			return err
		}
		coin := v.Interface().(coin.Info)
		comp.Coins.AddToRound(round, coin.Name, coin)
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

func deserializeQueries(raw *rawCompiledIOP, comp *wizard.CompiledIOP, numRounds int) error {
	for round := 0; round < numRounds; round++ {
		if err := deserializeQuery(raw.QueriesNoParams[round], round, &comp.QueriesNoParams, comp); err != nil {
			return err
		}
		if err := deserializeQuery(raw.QueriesParams[round], round, &comp.QueriesParams, comp); err != nil {
			return err
		}
	}
	return nil
}

func deserializeQuery(rawQueries []json.RawMessage, round int, register *wizard.ByRoundRegister[ifaces.QueryID, ifaces.Query], comp *wizard.CompiledIOP) error {
	for _, rawQ := range rawQueries {
		v, err := DeserializeValue(rawQ, DeclarationMode, queryType, comp)
		if err != nil {
			return err
		}
		q := v.Interface().(ifaces.Query)
		register.AddToRound(round, q.Name(), q)
	}
	return nil
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
