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

// Note, this is a work in progress : not all fields are represented.
type rawCompiledIOP struct {
	Columns         [][]json.RawMessage `json:"columns"`
	QueriesParams   [][]json.RawMessage `json:"queriesParams"`
	QueriesNoParams [][]json.RawMessage `json:"queriesNoParams"`
	Coins           [][]json.RawMessage `json:"coins"`
	DummyCompiled   bool                `json:"dummyCompiled"`
}

// SerializeCompiledIOP marshals a [wizard.CompiledIOP] object into JSON. This is
// meant to allow deserializing the IOP during the prover runtime instead of
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

	raw := &rawCompiledIOP{}
	numRounds := comp.NumRounds()

	for round := 0; round < numRounds; round++ {

		var (
			cols         = comp.Columns.AllHandlesAtRound(round)
			qParams      = comp.QueriesParams.AllKeysAt(round)
			qNoParams    = comp.QueriesNoParams.AllKeysAt(round)
			coins        = comp.Coins.AllKeysAt(round)
			rawCols      = make([]json.RawMessage, len(cols))
			rawQParams   = make([]json.RawMessage, len(qParams))
			rawQNoParams = make([]json.RawMessage, len(qNoParams))
			rawCoins     = make([]json.RawMessage, len(coins))
		)

		for i := range cols {
			r, err := SerializeValue(reflect.ValueOf(&cols[i]), DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize column `%v` : %w", cols[i].GetColID(), err)
			}
			rawCols[i] = r
		}

		for i, qName := range qParams {
			q := comp.QueriesParams.Data(qName)
			// It's important that we provide a pointer to the query rather than
			// the query itself. Otherwise, the serializer will resolve the type
			// as the concrete type and we want it to resolve the Query interface
			// type.
			r, err := SerializeValue(reflect.ValueOf(&q), DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize query `%v` : %w", qName, err)
			}
			rawQParams[i] = r
		}

		for i, qName := range qNoParams {
			q := comp.QueriesNoParams.Data(qName)
			// It's important that we provide a pointer to the query rather than
			// the query itself. Otherwise, the serializer will resolve the type
			// as the concrete type and we want it to resolve the Query interface
			// type.
			r, err := SerializeValue(reflect.ValueOf(&q), DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize query `%v` : %w", qName, err)
			}
			rawQNoParams[i] = r
		}

		for i, cName := range coins {
			c := comp.Coins.Data(cName)
			r, err := SerializeValue(reflect.ValueOf(c), DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize coin `%v` : %w", cName, err)
			}
			rawCoins[i] = r
		}

		raw.Columns = append(raw.Columns, rawCols)
		raw.QueriesParams = append(raw.QueriesParams, rawQParams)
		raw.QueriesNoParams = append(raw.QueriesNoParams, rawQNoParams)
		raw.Coins = append(raw.Coins, rawCoins)
	}

	res, _ := serializeAnyWithCborPkg(raw)
	return res, nil
}

// DeserializeCompiledIOP unmarshals a [wizard.CompiledIOP] object or returns
// an error if the marshalled object does not have the right format. The
// deserialized object can then be used for proving or verification via the
// functions [wizard.Prove] or [wizard.Verify].
func DeserializeCompiledIOP(data []byte) (*wizard.CompiledIOP, error) {

	comp := newEmptyCompiledIOP()
	raw := &rawCompiledIOP{}
	if err := deserializeAnyWithCborPkg(data, raw); err != nil {
		return nil, err
	}

	numRounds := len(raw.Columns)

	// It is crucial that we first deserialize the columns and the coins before
	// we attempt
	// to deserialize the queries because the queries will make references to
	// columns.
	for round := 0; round < numRounds; round++ {
		for _, rawCoin := range raw.Coins[round] {
			v, err := DeserializeValue(rawCoin, DeclarationMode, reflect.TypeOf(coin.Info{}), comp)
			if err != nil {
				return nil, err
			}

			coin := v.Interface().(coin.Info)
			comp.Coins.AddToRound(round, coin.Name, coin)
		}

		for _, rawCol := range raw.Columns[round] {

			// Note: that this will register the column within comp already, so
			// we don't need to worry about doing it explicitly again in the
			// current function. That's why we ignore the returned reflect.Value
			if _, err := DeserializeValue(rawCol, DeclarationMode, columnType, comp); err != nil {
				return nil, err
			}
		}
	}

	for round := 0; round < numRounds; round++ {

		for _, rawQ := range raw.QueriesNoParams[round] {
			v, err := DeserializeValue(rawQ, DeclarationMode, queryType, comp)
			if err != nil {
				return nil, err
			}

			q := v.Interface().(ifaces.Query)
			comp.QueriesNoParams.AddToRound(round, q.Name(), q)
		}

		for _, rawQ := range raw.QueriesParams[round] {
			v, err := DeserializeValue(rawQ, DeclarationMode, queryType, comp)
			if err != nil {
				return nil, err
			}

			q := v.Interface().(ifaces.Query)
			comp.QueriesParams.AddToRound(round, q.Name(), q)
		}
	}

	return comp, nil
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
