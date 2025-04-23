// Package serialization provides utilities for serializing and deserializing
// the Wizard protocol's state, including CompiledIOP, column assignments, and
// symbolic expressions, using CBOR encoding for compactness and efficiency.
// It supports the Linea ZK rollup's prover by enabling state persistence and
// test file generation.
//
// Files:
// - cbor.go: Low-level CBOR encoding/decoding.
// - column_assignment.go: Serializes column assignments with compression.
// - column_declaration.go: Serializes column metadata.
// - compiled_iop.go: Serializes CompiledIOP structure.
// - serialization.go: Core recursive serialization logic with modes.
// - implementation_registry.go: Type registry for interface deserialization.
// - pure_expression.go: Serializes symbolic expressions for testing.
package serialization
