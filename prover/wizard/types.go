package wizard

// id is a general identifier type for every Wizard object (column, query, etc..).
// The id can be parsed as: <Obj type (1 byte) || Obj indentifier (48 bytes)>.
//
// The id is purely internal
type id uint64
