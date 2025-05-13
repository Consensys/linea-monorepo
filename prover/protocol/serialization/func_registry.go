package serialization

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"

// 	"github.com/consensys/linea-monorepo/prover/maths/field"
// )

// // FuncRegistryEntry holds an identifier and a function
// type FuncRegistryEntry struct {
// 	ID   string
// 	Func interface{}
// 	Type reflect.Type
// }

// // funcRegistry maps identifiers to function implementations
// var funcRegistry = make(map[string]FuncRegistryEntry)

// // RegisterFunc registers a function with a unique identifier
// func RegisterFunc(id string, fn interface{}) {
// 	if fn == nil {
// 		panic("RegisterFunc: cannot register nil function")
// 	}
// 	fnType := reflect.TypeOf(fn)
// 	if fnType.Kind() != reflect.Func {
// 		panic(fmt.Sprintf("RegisterFunc: expected a function, got %v", fnType))
// 	}
// 	if _, exists := funcRegistry[id]; exists {
// 		panic(fmt.Sprintf("Function with id %q already registered", id))
// 	}
// 	funcRegistry[id] = FuncRegistryEntry{
// 		ID:   id,
// 		Func: fn,
// 		Type: fnType,
// 	}
// }

// // ExistsInFuncRegistry checks if a function identifier exists
// func ExistsInFuncRegistry(id string) bool {
// 	_, exists := funcRegistry[id]
// 	return exists
// }

// // SerializeInputFiller serializes an InputFiller function
// func SerializeInputFiller(v reflect.Value) (json.RawMessage, error) {
// 	if !v.IsValid() || v.IsNil() {
// 		return json.RawMessage(`"nil"`), nil
// 	}
// 	if v.Type() != reflect.TypeOf((func(int, int) field.Element)(nil)) {
// 		return nil, fmt.Errorf("value is not an InputFiller function: %v", v.Type())
// 	}
// 	for _, entry := range funcRegistry {
// 		if entry.Type == v.Type() && reflect.ValueOf(entry.Func).Pointer() == v.Pointer() {
// 			return json.RawMessage(fmt.Sprintf(`"%s"`, entry.ID)), nil
// 		}
// 	}
// 	return nil, fmt.Errorf("unknown InputFiller function")
// }

// // DeserializeInputFiller deserializes an InputFiller function
// func DeserializeInputFiller(data json.RawMessage, t reflect.Type) (reflect.Value, error) {
// 	if string(data) == `"nil"` {
// 		return reflect.Zero(t), nil
// 	}
// 	if t != reflect.TypeOf((func(int, int) field.Element)(nil)) {
// 		return reflect.Value{}, fmt.Errorf("type is not an InputFiller function: %v", t)
// 	}
// 	var id string
// 	if err := json.Unmarshal(data, &id); err != nil {
// 		return reflect.Value{}, fmt.Errorf("failed to unmarshal InputFiller identifier: %w", err)
// 	}
// 	if !ExistsInFuncRegistry(id) {
// 		return reflect.Value{}, fmt.Errorf("unknown InputFiller identifier: %s", id)
// 	}
// 	entry := funcRegistry[id]
// 	if entry.Type != t {
// 		return reflect.Value{}, fmt.Errorf("type mismatch for InputFiller %s: expected %v, got %v", id, t, entry.Type)
// 	}
// 	return reflect.ValueOf(entry.Func), nil
// }
