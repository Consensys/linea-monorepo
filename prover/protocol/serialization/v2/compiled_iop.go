package v2

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
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

	columnObjectType // used to tag a backreference as referring to a column
	columnName
	coinObjectType
	queriesObjectType
)

// BackReference is a lightweight reference to an object that has been serialized elsewhere.
// For example, a column used in multiple rounds will be defined once, and reused via a BackReference.
type BackReference struct {
	What objectType `cbor:"w"` // kind of object being referenced
	ID   int        `cbor:"i"` // index into the corresponding object list (e.g., Columns)
}

// serializableColumnV2 is the minimal form of a column's metadata, used for unique definition.
// It can be used later to retrieve the full column data in wizard.CompiledIOP
type serializableColumnV2 struct {
	Name   ifaces.ColID  `cbor:"name"`     // column ID
	Round  int           `cbor:"round"`    // round in which it's defined
	Status column.Status `cbor:"status"`   // column status (e.g., committed, queried)
	Size   int           `cbor:"log2Size"` // size in log2 format
}

// Serializer holds state for encoding and decoding a CompiledIOP.
type Serializer struct {
	typeMap   map[string]int       // used for managing custom types (unused here, but prepared for extension)
	columnMap map[ifaces.ColID]int // maps column ID to its index in Columns

	Types []string `cbor:"types,omitempty"` // optional list of custom types (not used currently)

	// Columns stores the actual serialized form of each unique column.
	// Each column is only serialized once and referenced later.
	Columns [][]byte `cbor:"column_defs,omitempty"`

	// ColumnBackRefs contains per-round column backreferences. Each entry in the outer slice is a round;
	// Each inner slice contains CBOR-encoded BackReferences to Columns.
	ColumnBackRefs [][]cbor.RawMessage `cbor:"columns"`
}

// SerializeCompiledIOPV2 serializes a CompiledIOP into compact CBOR format using backreferences.
// It ensures that each column is only serialized once and referenced where needed.
func SerializeCompiledIOPV2(comp *wizard.CompiledIOP) ([]byte, error) {
	ser := &Serializer{
		typeMap:        make(map[string]int),
		columnMap:      make(map[ifaces.ColID]int),
		Columns:        make([][]byte, 0),
		ColumnBackRefs: make([][]cbor.RawMessage, comp.NumRounds()),
	}

	logrus.Infof("No. of rounds: %d", comp.NumRounds())

	for round := 0; round < comp.NumRounds(); round++ {
		cols := comp.Columns.AllHandlesAtRound(round)
		backRefs := make([]cbor.RawMessage, len(cols))

		for i, col := range cols {
			// Each column is either serialized (if first time seen) or replaced with a backreference.
			ref, err := ser.marshalColumn(col)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal column %q: %v", col.GetColID(), err)
			}
			backRefs[i] = ref
		}
		// Store all backreferences for this round.
		ser.ColumnBackRefs[round] = backRefs
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

	comp := serialization.NewEmptyCompiledIOP()

	for round, rawRefs := range ser.ColumnBackRefs {
		for _, ref := range rawRefs {
			// Decode each backreference and insert the referenced column into the CompiledIOP.
			if err := ser.unmarshalColumn(ref, comp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal column at round %d: %v", round, err)
			}
		}
	}
	return comp, nil
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
		return cbor.Marshal(BackReference{What: columnObjectType, ID: pos})
	}

	// logrus.Infof("Back ref does not exist for colID:%s", col.ID)

	// Construct a serializable form of the column.
	serCol := serializableColumnV2{
		Name:   col.GetColID(),
		Round:  col.Round(),
		Status: col.Status(),
		Size:   col.Size(),
	}

	// Serialize the full column definition.
	marshaled, err := cbor.Marshal(serCol)
	if err != nil {
		return nil, fmt.Errorf("could not marshal column %q: %v", col.ID, err)
	}

	// Register this new column in the Columns list and map.
	pos := len(ser.Columns)
	ser.Columns = append(ser.Columns, marshaled)
	ser.columnMap[col.ID] = pos

	// Return a backreference pointing to the stored column.
	return cbor.Marshal(BackReference{What: columnObjectType, ID: pos})
}

// unmarshalColumn decodes a backreference and reconstructs the actual column it points to.
func (ser *Serializer) unmarshalColumn(data []byte, comp *wizard.CompiledIOP) error {
	var ref BackReference
	if err := cbor.Unmarshal(data, &ref); err != nil {
		return fmt.Errorf("failed to decode BackReference: %v", err)
	}

	if ref.What != columnObjectType {
		return fmt.Errorf("invalid back-reference type: got %v, expected %v", ref.What, columnObjectType)
	}

	if ref.ID < 0 || ref.ID >= len(ser.Columns) {
		return fmt.Errorf("invalid back-reference ID: %d (Columns length: %d)", ref.ID, len(ser.Columns))
	}

	var serCol serializableColumnV2
	if err := cbor.Unmarshal(ser.Columns[ref.ID], &serCol); err != nil {
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
