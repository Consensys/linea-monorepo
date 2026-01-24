/**
 * Contract Integrity Verifier - Storage Path Utilities
 *
 * Utilities for parsing storage paths and computing ERC-7201 storage slots.
 * Supports schema-based verification with human-readable paths.
 */

import { readFileSync } from "fs";
import { resolve } from "path";
import { ethers } from "ethers";
import {
  StorageSchema,
  StorageStructDef,
  StorageFieldDef,
  StoragePathConfig,
  StoragePathResult,
  SolidityType,
} from "../types";

// ============================================================================
// Schema Loading
// ============================================================================

/**
 * Validates a storage schema structure.
 * @throws Error if schema is malformed
 */
function validateStorageSchema(schema: unknown, path: string): asserts schema is StorageSchema {
  if (!schema || typeof schema !== "object") {
    throw new Error(`Invalid schema at ${path}: expected object`);
  }

  const obj = schema as Record<string, unknown>;
  if (!obj.structs || typeof obj.structs !== "object") {
    throw new Error(`Invalid schema at ${path}: missing 'structs' object`);
  }

  for (const [structName, structDef] of Object.entries(obj.structs)) {
    if (!structDef || typeof structDef !== "object") {
      throw new Error(`Invalid schema at ${path}: struct '${structName}' must be an object`);
    }

    const struct = structDef as Record<string, unknown>;
    if (!struct.fields || typeof struct.fields !== "object") {
      throw new Error(`Invalid schema at ${path}: struct '${structName}' missing 'fields' object`);
    }

    // Validate that struct has either namespace or baseSlot (for root structs)
    // Non-root structs (accessed via mapping) may have neither
    for (const [fieldName, fieldDef] of Object.entries(struct.fields as Record<string, unknown>)) {
      if (!fieldDef || typeof fieldDef !== "object") {
        throw new Error(`Invalid schema at ${path}: field '${structName}.${fieldName}' must be an object`);
      }

      const field = fieldDef as Record<string, unknown>;
      if (typeof field.slot !== "number") {
        throw new Error(`Invalid schema at ${path}: field '${structName}.${fieldName}' missing numeric 'slot'`);
      }
      if (typeof field.type !== "string") {
        throw new Error(`Invalid schema at ${path}: field '${structName}.${fieldName}' missing string 'type'`);
      }
    }
  }
}

/**
 * Loads a storage schema from a JSON file.
 * @throws Error if file cannot be read or schema is malformed
 */
export function loadStorageSchema(schemaPath: string, configDir: string): StorageSchema {
  const resolvedPath = resolve(configDir, schemaPath);
  let content: string;
  try {
    content = readFileSync(resolvedPath, "utf-8");
  } catch (err) {
    throw new Error(
      `Failed to read schema file at ${resolvedPath}: ${err instanceof Error ? err.message : String(err)}`,
    );
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(content);
  } catch (err) {
    throw new Error(
      `Failed to parse schema JSON at ${resolvedPath}: ${err instanceof Error ? err.message : String(err)}`,
    );
  }

  validateStorageSchema(parsed, resolvedPath);
  return parsed;
}

// ============================================================================
// ERC-7201 Slot Calculation
// ============================================================================

/**
 * Calculates the base storage slot for an ERC-7201 namespace.
 * Formula: keccak256(abi.encode(uint256(keccak256(id)) - 1)) & ~bytes32(uint256(0xff))
 */
export function calculateErc7201BaseSlot(namespaceId: string): string {
  // Step 1: keccak256(id)
  const idHash = ethers.keccak256(ethers.toUtf8Bytes(namespaceId));

  // Step 2: uint256(hash) - 1
  const hashBigInt = BigInt(idHash);
  const decremented = hashBigInt - 1n;

  // Step 3: keccak256(abi.encode(decremented))
  const encoded = ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [decremented]);
  const finalHash = ethers.keccak256(encoded);

  // Step 4: Mask off the last byte (& ~0xff)
  const finalBigInt = BigInt(finalHash);
  const masked = finalBigInt & ~0xffn;

  return "0x" + masked.toString(16).padStart(64, "0");
}

// ============================================================================
// Path Parsing
// ============================================================================

interface ParsedPath {
  structName: string;
  segments: PathSegment[];
}

type PathSegment =
  | { type: "field"; name: string }
  | { type: "arrayIndex"; index: number }
  | { type: "arrayLength" }
  | { type: "mappingKey"; key: string };

/**
 * Parses a storage path into components.
 * Format: "StructName:field.subfield[key].nested"
 * @throws Error if path is empty, malformed, or contains invalid characters
 */
export function parsePath(path: string): ParsedPath {
  if (!path || typeof path !== "string") {
    throw new Error("Invalid path: path must be a non-empty string");
  }

  const trimmedPath = path.trim();
  if (trimmedPath.length === 0) {
    throw new Error("Invalid path: path cannot be empty or whitespace");
  }

  const colonIndex = trimmedPath.indexOf(":");
  if (colonIndex === -1) {
    throw new Error(`Invalid path format: ${path}. Expected "StructName:path.to.field"`);
  }

  if (colonIndex === 0) {
    throw new Error(`Invalid path format: ${path}. Struct name cannot be empty`);
  }

  const structName = trimmedPath.slice(0, colonIndex);
  const pathPart = trimmedPath.slice(colonIndex + 1);

  // Validate struct name (must be a valid identifier)
  if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(structName)) {
    throw new Error(`Invalid struct name: ${structName}. Must be a valid identifier`);
  }

  if (pathPart.length === 0) {
    throw new Error(`Invalid path format: ${path}. Field path cannot be empty`);
  }

  const segments: PathSegment[] = [];

  // Tokenize the path
  let current = "";
  let i = 0;

  while (i < pathPart.length) {
    const char = pathPart[i];

    if (char === ".") {
      if (current) {
        segments.push({ type: "field", name: current });
        current = "";
      }
      i++;
    } else if (char === "[") {
      // Handle current field before bracket
      if (current) {
        segments.push({ type: "field", name: current });
        current = "";
      }

      // Find closing bracket
      const closeIndex = pathPart.indexOf("]", i);
      if (closeIndex === -1) {
        throw new Error(`Unclosed bracket in path: ${path}`);
      }

      const bracketContent = pathPart.slice(i + 1, closeIndex);

      // Check if it's "length"
      if (bracketContent === "length") {
        segments.push({ type: "arrayLength" });
      }
      // Check if it's a number (array index)
      else if (/^\d+$/.test(bracketContent)) {
        segments.push({ type: "arrayIndex", index: parseInt(bracketContent, 10) });
      }
      // Otherwise it's a mapping key
      else {
        segments.push({ type: "mappingKey", key: bracketContent });
      }

      i = closeIndex + 1;
    } else {
      current += char;
      i++;
    }
  }

  // Handle remaining field
  if (current) {
    if (current === "length") {
      segments.push({ type: "arrayLength" });
    } else {
      segments.push({ type: "field", name: current });
    }
  }

  return { structName, segments };
}

// ============================================================================
// Slot Computation
// ============================================================================

interface ComputedSlot {
  slot: string;
  type: SolidityType;
  byteOffset: number;
}

/**
 * Computes the storage slot for a parsed path.
 */
export function computeSlot(parsedPath: ParsedPath, schema: StorageSchema): ComputedSlot {
  const structDef = schema.structs[parsedPath.structName];
  if (!structDef) {
    throw new Error(`Unknown struct: ${parsedPath.structName}`);
  }

  // Get base slot
  let baseSlot: bigint;
  if (structDef.baseSlot) {
    baseSlot = BigInt(structDef.baseSlot);
  } else if (structDef.namespace) {
    baseSlot = BigInt(calculateErc7201BaseSlot(structDef.namespace));
  } else {
    throw new Error(`Struct ${parsedPath.structName} has no baseSlot or namespace`);
  }

  let currentSlot = baseSlot;
  let currentType: SolidityType = "uint256";
  let currentByteOffset = 0;
  let currentStructDef: StorageStructDef | undefined = structDef;

  for (let i = 0; i < parsedPath.segments.length; i++) {
    const segment = parsedPath.segments[i];

    if (segment.type === "field") {
      if (!currentStructDef) {
        throw new Error(`Cannot access field ${segment.name} on non-struct type`);
      }

      const field: StorageFieldDef | undefined = currentStructDef.fields[segment.name];
      if (!field) {
        throw new Error(`Unknown field: ${segment.name} in struct`);
      }

      currentSlot = currentSlot + BigInt(field.slot);
      currentType = field.type;
      currentByteOffset = field.byteOffset ?? 0;

      // Check if this field references another struct
      const fieldType: SolidityType = field.type;
      if (fieldType.startsWith("mapping(")) {
        // Keep currentStructDef for mapping value lookup
      } else if (!isPrimitiveType(fieldType)) {
        // Look up nested struct
        currentStructDef = schema.structs[fieldType];
      } else {
        currentStructDef = undefined;
      }
    } else if (segment.type === "arrayLength") {
      // Array length is stored at the slot itself
      currentType = "uint256";
      currentByteOffset = 0;
    } else if (segment.type === "arrayIndex") {
      // Array element: keccak256(slot) + index
      const dataSlot = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [currentSlot]));
      currentSlot = BigInt(dataSlot) + BigInt(segment.index);
      // Extract element type from array type
      currentType = extractArrayElementType(currentType);
      currentByteOffset = 0;
    } else if (segment.type === "mappingKey") {
      // Mapping: keccak256(abi.encode(key, slot))
      const keyType = extractMappingKeyType(currentType);
      const valueType = extractMappingValueType(currentType);

      const encodedKey = encodeKey(segment.key, keyType);

      // keccak256(abi.encode(key, slot))
      const combined = ethers.AbiCoder.defaultAbiCoder().encode([keyType, "uint256"], [encodedKey, currentSlot]);
      const mappingSlot = ethers.keccak256(combined);

      currentSlot = BigInt(mappingSlot);
      currentType = valueType as SolidityType;
      currentByteOffset = 0;

      // Check if value type is a struct
      if (!isPrimitiveType(valueType)) {
        currentStructDef = schema.structs[valueType];
      }
    }
  }

  return {
    slot: "0x" + currentSlot.toString(16).padStart(64, "0"),
    type: currentType,
    byteOffset: currentByteOffset,
  };
}

function isPrimitiveType(type: string): boolean {
  return (
    type.startsWith("uint") ||
    type.startsWith("int") ||
    type.startsWith("bytes") ||
    type === "address" ||
    type === "bool"
  );
}

function extractArrayElementType(arrayType: SolidityType): SolidityType {
  if (arrayType.endsWith("[]")) {
    return arrayType.slice(0, -2) as SolidityType;
  }
  return "uint256";
}

function extractMappingKeyType(mappingType: SolidityType): string {
  // mapping(address => bool) -> address
  const match = mappingType.match(/^mapping\((\w+)\s*=>/);
  return match ? match[1] : "address";
}

function extractMappingValueType(mappingType: SolidityType): string {
  // mapping(address => bool) -> bool
  // mapping(address => SomeStruct) -> SomeStruct
  const match = mappingType.match(/=>\s*(.+)\)$/);
  return match ? match[1].trim() : "uint256";
}

function encodeKey(key: string, keyType: string): unknown {
  if (keyType === "address") {
    // Validate address format
    if (!/^0x[a-fA-F0-9]{40}$/.test(key)) {
      throw new Error(`Invalid address key: ${key}. Expected 0x followed by 40 hex characters.`);
    }
    return key;
  }
  if (keyType.startsWith("uint") || keyType.startsWith("int")) {
    // Validate numeric format
    if (!/^-?\d+$/.test(key) && !/^0x[a-fA-F0-9]+$/.test(key)) {
      throw new Error(`Invalid numeric key: ${key}. Expected decimal or hex number.`);
    }
    try {
      return BigInt(key);
    } catch {
      throw new Error(`Failed to convert key to BigInt: ${key}`);
    }
  }
  if (keyType === "bytes32") {
    // Validate bytes32 format
    if (!/^0x[a-fA-F0-9]{64}$/.test(key)) {
      throw new Error(`Invalid bytes32 key: ${key}. Expected 0x followed by 64 hex characters.`);
    }
    return key;
  }
  // For unknown types, return as-is but warn
  return key;
}

// ============================================================================
// Storage Reading
// ============================================================================

/**
 * Reads and decodes a storage value at a computed slot.
 */
export async function readStorageAtPath(
  provider: ethers.JsonRpcProvider,
  address: string,
  computedSlot: ComputedSlot,
): Promise<unknown> {
  const rawValue = await provider.getStorage(address, computedSlot.slot);
  return decodeStorageValue(rawValue, computedSlot.type, computedSlot.byteOffset);
}

/**
 * Decodes a raw storage value based on type and byte offset.
 */
function decodeStorageValue(rawValue: string, type: SolidityType, byteOffset: number): unknown {
  const normalized = rawValue.slice(2).padStart(64, "0");
  const typeBytes = getTypeBytes(type);

  // Storage is right-aligned (low bytes at the end)
  const startByte = 32 - byteOffset - typeBytes;
  const endByte = 32 - byteOffset;
  const hexValue = normalized.slice(startByte * 2, endByte * 2);

  if (type === "address") {
    return ethers.getAddress("0x" + hexValue.slice(-40));
  }
  if (type === "bool") {
    return hexValue !== "0".repeat(hexValue.length);
  }
  if (type.startsWith("uint") || type.startsWith("int")) {
    return BigInt("0x" + hexValue).toString();
  }
  if (type.startsWith("bytes")) {
    return "0x" + hexValue;
  }
  return "0x" + hexValue;
}

function getTypeBytes(type: SolidityType): number {
  if (type === "address") return 20;
  if (type === "bool") return 1;
  if (type === "uint8" || type === "int8") return 1;
  if (type === "uint16" || type === "int16") return 2;
  if (type === "uint32" || type === "int32") return 4;
  if (type === "uint64" || type === "int64") return 8;
  if (type === "uint96" || type === "int96") return 12;
  if (type === "uint128" || type === "int128") return 16;
  if (type === "uint256" || type === "int256") return 32;
  if (type === "bytes32") return 32;
  if (type === "bytes4") return 4;
  return 32;
}

// ============================================================================
// Path Verification
// ============================================================================

/**
 * Verifies a storage path against expected value.
 */
export async function verifyStoragePath(
  provider: ethers.JsonRpcProvider,
  address: string,
  config: StoragePathConfig,
  schema: StorageSchema,
): Promise<StoragePathResult> {
  try {
    const parsed = parsePath(config.path);
    const computed = computeSlot(parsed, schema);
    const actual = await readStorageAtPath(provider, address, computed);

    const pass = compareValues(actual, config.expected, config.comparison ?? "eq");

    return {
      path: config.path,
      computedSlot: computed.slot,
      type: computed.type,
      expected: config.expected,
      actual,
      status: pass ? "pass" : "fail",
      message: pass
        ? `${config.path} = ${formatValue(actual)}`
        : `${config.path}: expected ${formatValue(config.expected)}, got ${formatValue(actual)}`,
    };
  } catch (error) {
    return {
      path: config.path,
      computedSlot: "error",
      type: "unknown",
      expected: config.expected,
      actual: undefined,
      status: "fail",
      message: `Error: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}

/**
 * Checks if a string can be safely converted to BigInt.
 */
function isNumericString(value: string): boolean {
  // Match decimal integers (with optional negative sign) or hex strings
  return /^-?\d+$/.test(value) || /^0x[0-9a-fA-F]+$/.test(value);
}

function compareValues(actual: unknown, expected: unknown, comparison: string): boolean {
  const normalizedActual = normalizeValue(actual);
  const normalizedExpected = normalizeValue(expected);

  switch (comparison) {
    case "eq":
      return normalizedActual === normalizedExpected;
    case "gt":
    case "gte":
    case "lt":
    case "lte": {
      // Validate both values are numeric before BigInt conversion
      if (!isNumericString(normalizedActual) || !isNumericString(normalizedExpected)) {
        // Fall back to string comparison for non-numeric values
        return normalizedActual === normalizedExpected;
      }
      const actualBigInt = BigInt(normalizedActual);
      const expectedBigInt = BigInt(normalizedExpected);
      if (comparison === "gt") return actualBigInt > expectedBigInt;
      if (comparison === "gte") return actualBigInt >= expectedBigInt;
      if (comparison === "lt") return actualBigInt < expectedBigInt;
      return actualBigInt <= expectedBigInt;
    }
    default:
      return normalizedActual === normalizedExpected;
  }
}

function normalizeValue(value: unknown): string {
  if (typeof value === "string") {
    if (value.startsWith("0x") && value.length === 42) {
      return value.toLowerCase();
    }
    return value;
  }
  return String(value);
}

function formatValue(value: unknown): string {
  if (typeof value === "string" && value.length > 20) {
    return value.slice(0, 10) + "..." + value.slice(-6);
  }
  return String(value);
}
