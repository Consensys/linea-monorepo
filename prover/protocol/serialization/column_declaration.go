package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// serializableColumnDecl is used to represent a "natural" column, meaning a
// column that is explicitly registered as part of the scheme. This is in
// opposition with [serializableColumnRef] where such columns are encoded
// by just citing their names.
//
// Concretely, we need this because [column.Natural] has a complex structure
// that is deeply nested within [column.Store]. And this prevents directly
// applying the generic reflection-based serialization logic to it.
type serializableColumnDecl struct {
	Name   ifaces.ColID
	Round  int
	Status column.Status
	Size   int
}

// The function takes a Natural column as parameter rather than an
// [ifaces.Column]
func intoSerializableColDecl(c *column.Natural) *serializableColumnDecl {
	return &serializableColumnDecl{
		Name:   c.ID,
		Round:  c.Round(),
		Status: c.Status(),
		Size:   c.Size(),
	}
}

type serializableManuallyShifted struct {
	Natural *serializableColumnDecl
	Root    *serializableColumnDecl
	Offset  int
}

func intoSerializableManuallyShifted(d *dedicated.ManuallyShifted) *serializableManuallyShifted {
	rootNatural, ok := d.Root.(column.Natural)
	if !ok {
		panic(fmt.Errorf("root is not a column.Natural, got %T", d.Root))
	}
	return &serializableManuallyShifted{
		Natural: intoSerializableColDecl(&d.Natural),
		Root:    intoSerializableColDecl(&rootNatural),
		Offset:  d.Offset,
	}
}

func (s *serializableManuallyShifted) intoManuallyShifted(comp *wizard.CompiledIOP) *dedicated.ManuallyShifted {
	natural := s.Natural.intoNaturalAndRegister(comp)
	root := s.Root.intoNaturalAndRegister(comp)
	return &dedicated.ManuallyShifted{
		Natural: natural.(column.Natural),
		Root:    root,
		Offset:  s.Offset,
	}
}

// Converts a serializableColumnDecl back into a column.Natural and registers it in a
// wizard.CompiledIOP context, returning an ifaces.Column interface. Used during deserialization
// to reconstruct the column structure after loading CBOR-encoded metadata.
func (c *serializableColumnDecl) intoNaturalAndRegister(comp *wizard.CompiledIOP) ifaces.Column {
	if comp.Columns.Exists(c.Name) {
		return comp.Columns.GetHandle(c.Name)
	}
	return comp.InsertColumn(c.Round, c.Name, c.Size, c.Status)
}

// serializeColumnInterface handles serialization of column interfaces in DeclarationMode.
// Column types can be either Natural or ManuallyShifted
func serializeColumnInterface(v reflect.Value, mode Mode) (json.RawMessage, error) {
	concrete := v.Elem()

	var data json.RawMessage
	var err error
	switch concrete.Type() {
	case naturalType:
		col := v.Interface().(column.Natural)
		decl := intoSerializableColDecl(&col)
		data, err = SerializeValue(reflect.ValueOf(decl), mode)
		if err != nil {
			return nil, err
		}

	case manuallyShiftedType:
		shifted := v.Interface().(*dedicated.ManuallyShifted)
		decl := intoSerializableManuallyShifted(shifted)
		data, err = SerializeValue(reflect.ValueOf(decl), mode)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported column type in DeclarationMode: %s", concrete.Type().String())
	}

	//	fmt.Printf("SER Column Interface type:%s \n", concrete.Type().String())
	raw := map[string]interface{}{
		"type":  concrete.Type().String(),
		"value": data,
	}

	return serializeAnyWithCborPkg(raw)
}

// Helper function to deserialize column.Natural
func deserializeColumnNatural(value json.RawMessage, mode Mode, comp *wizard.CompiledIOP, ifaceValue reflect.Value) (reflect.Value, error) {
	rawType := reflect.TypeOf(&serializableColumnDecl{})
	v, err := DeserializeValue(value, mode, rawType, comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
	}
	nat := v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)
	ifaceValue.Set(reflect.ValueOf(nat))
	return ifaceValue, nil
}

// Helper function to deserialize dedicated.ManuallyShifted
func deserializeManuallyShifted(value json.RawMessage, mode Mode, comp *wizard.CompiledIOP, ifaceValue reflect.Value) (reflect.Value, error) {
	rawType := reflect.TypeOf(&serializableManuallyShifted{})
	v, err := DeserializeValue(value, mode, rawType, comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize ManuallyShifted: %w", err)
	}
	shifted := v.Interface().(*serializableManuallyShifted).intoManuallyShifted(comp)
	ifaceValue.Set(reflect.ValueOf(shifted))
	return ifaceValue, nil
}
