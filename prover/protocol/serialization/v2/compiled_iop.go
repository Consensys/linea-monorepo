package v2

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/fxamacker/cbor/v2"
	"github.com/sirupsen/logrus"
)

// objectType is an enum to identify types of serialized objects in backreferences.
type objectType uint8

const (
	typeObjectType objectType = iota

	// Used to tag a backreference types
	columnObjectType
	coinObjectType
	queriesObjectType
)

// BackReference is a lightweight reference to an object that has been serialized elsewhere.
// For example, a column used in multiple rounds will be defined once, and reused via a BackReference.
type BackReference struct {
	What objectType `cbor:"w"` // kind of object being referenced
	ID   int        `cbor:"i"` // index into the corresponding object list (e.g., Columns)
}

// Serializer holds state for encoding and decoding a CompiledIOP.
type Serializer struct {
	typeMap   map[string]int       // Used for managing custom types (unused here, but prepared for extension)
	columnMap map[ifaces.ColID]int // Maps column ID to its index in Columns
	coinMap   map[coin.Name]int    // Maps coin name to its index in coins

	Types []string `cbor:"types,omitempty"` // optional list of custom types (not used currently)

	// Columns stores the actual serialized form of each unique column.
	// Each column is only serialized once and referenced later.
	Columns [][]byte `cbor:"columns,omitempty"`
	Coins   [][]byte `cbor:"coins,omitempty"`

	// BackRefs contains per-round column backreferences. Each entry in the outer slice is a round;
	// Each inner slice contains CBOR-encoded BackReferences to its object such as Columns, Coins, Queries etc.
	ColumnBackRefs [][]cbor.RawMessage `cbor:"column_ref,omitempty"`
	CoinBackRefs   [][]cbor.RawMessage `cbor:"coin_ref,omitempty"`
}

func initSerializer(numRounds int) *Serializer {
	return &Serializer{
		typeMap: make(map[string]int),

		columnMap:      make(map[ifaces.ColID]int),
		Columns:        make([][]byte, 0),
		ColumnBackRefs: make([][]cbor.RawMessage, numRounds),

		coinMap:      make(map[coin.Name]int),
		Coins:        make([][]byte, 0),
		CoinBackRefs: make([][]cbor.RawMessage, numRounds),
	}
}

// SerializeCompiledIOPV2 serializes a CompiledIOP into compact CBOR format using backreferences.
// It ensures that each column is only serialized once and referenced where needed.
func SerializeCompiledIOPV2(comp *wizard.CompiledIOP) ([]byte, error) {
	ser := initSerializer(comp.NumRounds())

	logrus.Infof("No. of rounds: %d", comp.NumRounds())

	for round := 0; round < comp.NumRounds(); round++ {

		var (
			cols        = comp.Columns.AllHandlesAtRound(round)
			backColRefs = make([]cbor.RawMessage, len(cols))

			coins        = comp.Coins.AllKeysAt(round)
			backCoinRefs = make([]cbor.RawMessage, len(coins))
		)

		for i, col := range cols {
			// Each column is either serialized (if first time seen) or replaced with a backreference.
			colRef, err := ser.marshalColumn(col)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal column %q: %v", col.GetColID(), err)
			}
			backColRefs[i] = colRef
		}

		for i, coinName := range coins {
			coinData := comp.Coins.Data(coinName)
			coinRef, err := ser.marshalCoin(coinData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal coin %q: %v", coinName, err)
			}
			backCoinRefs[i] = coinRef
		}

		// Store all backreferences for this round.
		ser.ColumnBackRefs[round] = backColRefs
		ser.CoinBackRefs[round] = backCoinRefs
	}

	// Final CBOR encoding of the full Serializer structure.
	return serialization.SerializeAnyWithCborPkg(ser)
}

// DeserializeCompiledIOPV2 reconstructs a CompiledIOP from the compact CBOR-encoded form.
// It uses the column definitions and the per-round backreferences to rebuild the full structure.
func DeserializeCompiledIOPV2(data []byte) (*wizard.CompiledIOP, error) {
	var ser Serializer
	if err := serialization.DeserializeAnyWithCborPkg(data, &ser); err != nil {
		return nil, fmt.Errorf("failed to decode Serializer: %v", err)
	}

	var (
		comp      = serialization.NewEmptyCompiledIOP()
		numRounds = len(ser.ColumnBackRefs)
	)

	for round := 0; round < numRounds; round++ {

		for _, ref := range ser.ColumnBackRefs[round] {
			// Decode each backreference and insert the referenced column into the CompiledIOP.
			if err := ser.unmarshalColumn(ref, comp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal column at round %d: %v", round, err)
			}
		}

		for _, ref := range ser.CoinBackRefs[round] {
			// Decode each backreference and insert the referenced coin into the CompiledIOP.
			if err := ser.unmarshalCoin(ref, comp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal coin at round %d: %v", round, err)
			}
		}
	}
	return comp, nil
}
