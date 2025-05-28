package v2

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization/cmn"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// serializableColumnV2 is the minimal form of a column's metadata, used for unique definition.
// It can be used later to retrieve the full column data in wizard.CompiledIOP
type serializableColumnV2 struct {
	Name   ifaces.ColID  `cbor:"name"`     // column ID
	Round  int           `cbor:"round"`    // round in which it's defined
	Status column.Status `cbor:"status"`   // column status (e.g., committed, queried)
	Size   int           `cbor:"log2Size"` // size in log2 format
}

// marshalColumn serializes a column if it's not already serialized.
// Returns a CBOR-encoded BackReference to the column (existing or newly created).
func (ser *Serializer) marshalColumn(iCol ifaces.Column) ([]byte, error) {
	col, ok := iCol.(column.Natural)
	if !ok {
		return nil, fmt.Errorf("cannot convert column of type: %T into column.Natural", iCol)
	}

	// Check if the column has already been serialized.
	if pos, ok := ser.columnMap[col.ID]; ok {
		logrus.Infof("Back ref already exists for colID:%s", col.ID)
		return cmn.SerializeAnyWithCborPkg(BackReference{What: columnObjectType, ID: pos})
	}

	// Construct a serializable form of the column.
	serCol := serializableColumnV2{
		Name:   col.GetColID(),
		Round:  col.Round(),
		Status: col.Status(),
		Size:   col.Size(),
	}

	// Serialize the full column definition.
	marshaled, err := cmn.SerializeAnyWithCborPkg(serCol)
	if err != nil {
		return nil, fmt.Errorf("could not marshal column %q: %v", col.ID, err)
	}

	// Register this new column in the Columns list and map.
	pos := len(ser.Columns)
	ser.Columns = append(ser.Columns, marshaled)
	ser.columnMap[col.ID] = pos

	// Return a backreference pointing to the stored column.
	return cmn.SerializeAnyWithCborPkg(BackReference{What: columnObjectType, ID: pos})
}

// unmarshalColumn decodes a backreference and reconstructs the actual column it points to.
func (ser *Serializer) unmarshalColumn(data []byte, comp *wizard.CompiledIOP) error {
	var ref BackReference
	if err := cmn.DeserializeAnyWithCborPkg(data, &ref); err != nil {
		return fmt.Errorf("failed to decode BackReference: %v", err)
	}

	if ref.What != columnObjectType {
		return fmt.Errorf("invalid back-reference type: got %v, expected %v", ref.What, columnObjectType)
	}

	if ref.ID < 0 || ref.ID >= len(ser.Columns) {
		return fmt.Errorf("invalid back-reference ID: %d (Columns length: %d)", ref.ID, len(ser.Columns))
	}

	var serCol serializableColumnV2
	if err := cmn.DeserializeAnyWithCborPkg(ser.Columns[ref.ID], &serCol); err != nil {
		return fmt.Errorf("failed to decode serializableColumn at ID %d: %v", ref.ID, err)
	}

	// Reconstruct and insert the column into the CompiledIOP.
	comp.InsertColumn(
		int(serCol.Round),
		ifaces.ColID(serCol.Name),
		int(serCol.Size),
		column.Status(serCol.Status),
	)
	return nil
}

// marshalCoin serializes a coin if it's not already serialized.
// Returns a CBOR-encoded BackReference to the coin (existing or newly created).
func (ser *Serializer) marshalCoin(coin coin.Info) ([]byte, error) {
	// Check if the coin has already been serialized.
	if pos, ok := ser.coinMap[coin.Name]; ok {
		logrus.Infof("Back ref already exists for coin:%s", coin.Name)
		return cmn.SerializeAnyWithCborPkg(BackReference{What: coinObjectType, ID: pos})
	}

	marshaled, err := cmn.SerializeAnyWithCborPkg(coin)
	if err != nil {
		return nil, fmt.Errorf("could not marshal coin %q: %v", coin.Name, err)
	}

	pos := len(ser.Coins)
	ser.Coins = append(ser.Coins, marshaled)
	ser.coinMap[coin.Name] = pos

	// Return a backreference pointing to the stored coin.
	return cmn.SerializeAnyWithCborPkg(BackReference{What: coinObjectType, ID: pos})
}

// unmarshalCoin decodes a backreference and reconstructs the actual coin it points to.
func (ser *Serializer) unmarshalCoin(data []byte, comp *wizard.CompiledIOP) error {
	var ref BackReference
	if err := cmn.DeserializeAnyWithCborPkg(data, &ref); err != nil {
		return fmt.Errorf("failed to decode BackReference: %v", err)
	}

	if ref.What != coinObjectType {
		return fmt.Errorf("invalid back-reference type: got %v, expected %v", ref.What, coinObjectType)
	}

	if ref.ID < 0 || ref.ID >= len(ser.Coins) {
		return fmt.Errorf("invalid back-reference ID: %d (Coins length: %d)", ref.ID, len(ser.Coins))
	}

	var serCoin coin.Info
	if err := cmn.DeserializeAnyWithCborPkg(ser.Coins[ref.ID], &serCoin); err != nil {
		return fmt.Errorf("failed to decode serializableCoin at ID %d: %v", ref.ID, err)
	}

	// Reconstruct and insert the coin into the CompiledIOP.
	comp.Coins.AddToRound(
		serCoin.Round,
		serCoin.Name,
		serCoin,
	)

	return nil
}
