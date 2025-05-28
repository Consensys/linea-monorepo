package v2

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/fxamacker/cbor/v2"
	"github.com/sirupsen/logrus"
)

type objectType uint8

const (
	typeObjectType objectType = iota
	columnObjectType
	columnName
	coinObjectType
	queriesObjectType
)

type Serializer struct {
	typeMap map[string]int
	Types   []string `cbor:"types,omitempty"`

	columnMap map[ifaces.ColID]int
	Columns   [][]byte `cbor:"columns,omitempty"`

	MainObject *rawCompiledIOPV2 `cbor:"main_object,omitempty"`
}

type BackReference struct {
	What objectType `cbor:"w"`
	ID   int        `cbor:"i"`
}

type rawCompiledIOPV2 struct {
	Columns [][]cbor.RawMessage `cbor:"columns"`
}

type SerializableInterface struct {
	T string `cbor:"t"`
	V []byte `cbor:"v"`
}

type serializableColumnV2 struct {
	Name     string `cbor:"name"`
	Round    int8   `cbor:"round"`
	Status   int8   `cbor:"status"`
	Log2Size int8   `cbor:"log2Size"`
}

func SerializeCompiledIOPV2(comp *wizard.CompiledIOP) ([]byte, error) {
	ser := &Serializer{
		typeMap:    make(map[string]int),
		columnMap:  make(map[ifaces.ColID]int),
		MainObject: &rawCompiledIOPV2{},
	}
	numRound := comp.NumRounds()
	logrus.Printf("No. of rounds:%d \n", numRound)

	for round := 0; round < numRound; round++ {
		cols := comp.Columns.AllHandlesAtRound(round)
		rawCols := make([]cbor.RawMessage, len(cols))
		for i, col := range cols {
			serCol, err := ser.MarshalColumn(col)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal column %q: %v", col.GetColID(), err)
			}
			rawCols[i] = serCol
		}
		ser.MainObject.Columns = append(ser.MainObject.Columns, rawCols)
	}
	return serialization.SerializeAnyWithCborPkg(ser)
}

func DeserializeCompiledIOPV2(data []byte) (*wizard.CompiledIOP, error) {
	var ser Serializer
	if err := serialization.DeserializeAnyWithCborPkg(data, &ser); err != nil {
		return nil, fmt.Errorf("failed to decode Serializer: %v", err)
	}

	comp := serialization.NewEmptyCompiledIOP()
	for round, rawCols := range ser.MainObject.Columns {
		for _, colData := range rawCols {
			if err := ser.UnmarshalColumn(colData, comp); err != nil {
				return nil, fmt.Errorf("failed to unmarshal column at round %d: %v", round, err)
			}
		}
	}

	return comp, nil
}

func (ser *Serializer) MarshalColumn(iCol ifaces.Column) ([]byte, error) {
	col, ok := iCol.(column.Natural)
	if !ok {
		return nil, fmt.Errorf("cannot convert column of type:%T into column.Natural", col)
	}

	if _, ok := ser.columnMap[col.ID]; !ok {
		logrus.Printf("Back ref does not exist for colID:%s \n", col.ID)
		newPos := len(ser.Columns)
		ser.columnMap[col.ID] = newPos
		serColumn := serializableColumnV2{
			Name:     string(col.GetColID()),
			Round:    int8(col.Round()),
			Status:   int8(col.Status()),
			Log2Size: int8(utils.Log2Floor(col.Size())),
		}

		marshaled, err := cbor.Marshal(serColumn)
		if err != nil {
			return nil, fmt.Errorf("could not marshal column, n=%q err=%v", col.ID, err)
		}
		ser.Columns = append(ser.Columns, marshaled)
	} else {
		logrus.Printf("Back ref already exists for colID:%s \n", col.ID)
	}

	pos := ser.columnMap[col.ID]
	backRef := &BackReference{
		What: columnObjectType,
		ID:   pos,
	}
	return cbor.Marshal(backRef)
}

func (ser *Serializer) UnmarshalColumn(data []byte, comp *wizard.CompiledIOP) error {
	var backRef BackReference
	if err := serialization.DeserializeAnyWithCborPkg(data, &backRef); err != nil {
		return fmt.Errorf("failed to decode BackReference: %v", err)
	}

	if backRef.What != columnObjectType {
		return fmt.Errorf("invalid back-reference type: got %v, expected %v", backRef.What, columnObjectType)
	}

	if backRef.ID < 0 || backRef.ID >= len(ser.Columns) {
		return fmt.Errorf("invalid back-reference ID: %d (Columns length: %d)", backRef.ID, len(ser.Columns))
	}

	colData := ser.Columns[backRef.ID]
	var serCol serializableColumnV2
	if err := serialization.DeserializeAnyWithCborPkg(colData, &serCol); err != nil {
		return fmt.Errorf("failed to decode serializableColumn at ID %d: %v", backRef.ID, err)
	}

	comp.InsertColumn(
		int(serCol.Round),
		ifaces.ColID(serCol.Name),
		int(serCol.Log2Size),
		column.Status(serCol.Status),
	)
	return nil
}
