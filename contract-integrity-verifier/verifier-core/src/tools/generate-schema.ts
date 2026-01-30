/**
 * Contract Integrity Verifier - Schema Generator
 *
 * Generates storage schema JSON from Solidity storage layout files.
 * Uses the CryptoAdapter pattern for framework-agnostic crypto operations.
 *
 * @packageDocumentation
 */

import type { CryptoAdapter } from "../adapter";

// ============================================================================
// Types
// ============================================================================

export interface FieldDef {
  slot: number;
  type: string;
  byteOffset?: number;
}

export interface StructDef {
  namespace?: string;
  baseSlot?: string;
  fields: Record<string, FieldDef>;
}

export interface Schema {
  $comment?: string;
  structs: Record<string, StructDef>;
}

export interface SchemaGeneratorOptions {
  /** Optional comment to add to the schema */
  comment?: string;
  /** Whether to validate calculated baseSlots against explicit constants */
  validateConstants?: boolean;
  /** Whether to include explicit byteOffset: 0 for first field in packed slots (default: true) */
  includeExplicitZeroOffset?: boolean;
}

export interface ParseResult {
  schema: Schema;
  warnings: string[];
}

// ============================================================================
// Type Sizes
// ============================================================================

const TYPE_SIZES: Record<string, number> = {
  // Boolean
  bool: 1,
  // Unsigned integers
  uint8: 1,
  uint16: 2,
  uint24: 3,
  uint32: 4,
  uint40: 5,
  uint48: 6,
  uint56: 7,
  uint64: 8,
  uint72: 9,
  uint80: 10,
  uint88: 11,
  uint96: 12,
  uint104: 13,
  uint112: 14,
  uint120: 15,
  uint128: 16,
  uint136: 17,
  uint144: 18,
  uint152: 19,
  uint160: 20,
  uint168: 21,
  uint176: 22,
  uint184: 23,
  uint192: 24,
  uint200: 25,
  uint208: 26,
  uint216: 27,
  uint224: 28,
  uint232: 29,
  uint240: 30,
  uint248: 31,
  uint256: 32,
  // Signed integers
  int8: 1,
  int16: 2,
  int24: 3,
  int32: 4,
  int40: 5,
  int48: 6,
  int56: 7,
  int64: 8,
  int72: 9,
  int80: 10,
  int88: 11,
  int96: 12,
  int104: 13,
  int112: 14,
  int120: 15,
  int128: 16,
  int136: 17,
  int144: 18,
  int152: 19,
  int160: 20,
  int168: 21,
  int176: 22,
  int184: 23,
  int192: 24,
  int200: 25,
  int208: 26,
  int216: 27,
  int224: 28,
  int232: 29,
  int240: 30,
  int248: 31,
  int256: 32,
  // Fixed-size byte arrays
  bytes1: 1,
  bytes2: 2,
  bytes3: 3,
  bytes4: 4,
  bytes5: 5,
  bytes6: 6,
  bytes7: 7,
  bytes8: 8,
  bytes9: 9,
  bytes10: 10,
  bytes11: 11,
  bytes12: 12,
  bytes13: 13,
  bytes14: 14,
  bytes15: 15,
  bytes16: 16,
  bytes17: 17,
  bytes18: 18,
  bytes19: 19,
  bytes20: 20,
  bytes21: 21,
  bytes22: 22,
  bytes23: 23,
  bytes24: 24,
  bytes25: 25,
  bytes26: 26,
  bytes27: 27,
  bytes28: 28,
  bytes29: 29,
  bytes30: 30,
  bytes31: 31,
  bytes32: 32,
  // Address
  address: 20,
};

// ============================================================================
// Helper Functions
// ============================================================================

function hexToBytes(hex: string): Uint8Array {
  const normalized = hex.startsWith("0x") ? hex.slice(2) : hex;
  const bytes = new Uint8Array(normalized.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(normalized.slice(i * 2, i * 2 + 2), 16);
  }
  return bytes;
}

/**
 * Calculate ERC-7201 base slot from namespace ID.
 * Formula: keccak256(abi.encode(uint256(keccak256(id)) - 1)) & ~bytes32(uint256(0xff))
 */
export function calculateErc7201BaseSlot(adapter: CryptoAdapter, namespaceId: string): string {
  // Step 1: keccak256(id)
  const idHash = adapter.keccak256(namespaceId);
  const hashBigInt = BigInt(idHash);

  // Step 2: uint256(hash) - 1
  const decremented = hashBigInt - 1n;

  // Step 3: keccak256(abi.encode(decremented))
  const encoded = adapter.encodeAbiParameters(["uint256"], [decremented]);
  const finalHash = adapter.keccak256(hexToBytes(encoded));

  // Step 4: Mask off the last byte (& ~0xff)
  const finalBigInt = BigInt(finalHash);
  const masked = finalBigInt & ~0xffn;

  return "0x" + masked.toString(16).padStart(64, "0");
}

/**
 * Extract namespace ID from NatSpec comment.
 * Looks for: @custom:storage-location erc7201:linea.storage.SomeName
 */
function extractNamespace(comments: string): string | undefined {
  const match = comments.match(/@custom:storage-location\s+erc7201:([^\s*]+)/);
  return match ? match[1] : undefined;
}

/**
 * Known struct names discovered during parsing.
 * Used to distinguish struct types from enums.
 */
let knownStructNames: Set<string> = new Set();

/**
 * Known enum names and their byte sizes.
 * Solidity enums use the smallest uint type that can hold all values.
 */
let knownEnums: Map<string, number> = new Map();

/**
 * Reset known struct names (call before parsing new files).
 */
function resetKnownStructs(): void {
  knownStructNames = new Set();
}

/**
 * Reset known enums (call before parsing new files).
 */
function resetKnownEnums(): void {
  knownEnums = new Map();
}

/**
 * Register a struct name as known.
 */
function registerStructName(name: string): void {
  knownStructNames.add(name);
}

/**
 * Calculate the byte size needed for an enum based on value count.
 * Solidity uses the smallest uint type that can hold all values.
 */
function calculateEnumSize(valueCount: number): number {
  if (valueCount <= 256) return 1; // uint8
  if (valueCount <= 65536) return 2; // uint16
  if (valueCount <= 16777216) return 3; // uint24
  if (valueCount <= 4294967296) return 4; // uint32
  return 32; // fallback to uint256
}

/**
 * Parse enum definitions from Solidity source and register them.
 */
function parseEnums(source: string): void {
  // Match enum definitions: enum Name { VALUE1, VALUE2, ... }
  // Use limited whitespace quantifier to prevent ReDoS
  const enumRegex = /enum[ \t]+(\w+)[ \t]*\{([^}]+)\}/g;
  let match;
  while ((match = enumRegex.exec(source)) !== null) {
    const enumName = match[1];
    const valuesStr = match[2];
    // Count values (split by comma, filter empty)
    const values = valuesStr.split(",").filter((v) => v.trim().length > 0);
    const valueCount = values.length;
    const byteSize = calculateEnumSize(valueCount);
    knownEnums.set(enumName, byteSize);
  }
}

/**
 * Parse a Solidity type and return normalized type string.
 * Converts enums to their equivalent uint type, preserves struct names.
 */
function normalizeType(solidityType: string): string {
  const trimmed = solidityType.trim();

  // Handle arrays first
  if (trimmed.endsWith("[]")) {
    const baseType = trimmed.slice(0, -2);
    return `${normalizeType(baseType)}[]`;
  }

  // Handle nested mappings: mapping(KeyType => mapping(...))
  // Need to handle balanced parentheses for nested mappings
  // Use limited whitespace quantifier to prevent ReDoS
  if (trimmed.startsWith("mapping")) {
    const innerMatch = trimmed.match(/^mapping[ \t]*\([ \t]*([^=>\s]+)[ \t]*=>[ \t]*(.+)\)$/);
    if (innerMatch) {
      const keyType = normalizeType(innerMatch[1]);
      const valueType = normalizeType(innerMatch[2].trim());
      return `mapping(${keyType} => ${valueType})`;
    }
  }

  // Check if it's a known struct type (preserve it)
  if (knownStructNames.has(trimmed)) {
    return trimmed;
  }

  // Check if it's a known enum type (convert to uintN)
  if (knownEnums.has(trimmed)) {
    const byteSize = knownEnums.get(trimmed)!;
    return `uint${byteSize * 8}`;
  }

  // Handle primitive types
  if (
    trimmed.startsWith("uint") ||
    trimmed.startsWith("int") ||
    trimmed.startsWith("bytes") ||
    trimmed === "address" ||
    trimmed === "bool" ||
    trimmed === "string"
  ) {
    return trimmed;
  }

  // PascalCase that's not a known struct or enum - could be interface or unknown type
  if (/^[A-Z][a-zA-Z0-9_]*$/.test(trimmed)) {
    // Check if it looks like an interface (starts with I followed by uppercase)
    if (/^I[A-Z]/.test(trimmed)) {
      return "address"; // Interfaces are addresses
    }
    // Unknown type - preserve the name but it will be treated as full slot
    return trimmed;
  }

  return trimmed;
}

/**
 * Get the byte size of a type (for packing calculation).
 */
function getTypeSize(type: string): number {
  return TYPE_SIZES[type] ?? 32;
}

/**
 * Extract a mapping type from a line, handling nested mappings with balanced parentheses.
 */
function extractMappingType(line: string): { typeStr: string; name: string } | null {
  // Find "mapping(" at the start (after whitespace)
  const mappingStart = line.search(/\bmapping\s*\(/);
  if (mappingStart === -1) return null;

  // Count parentheses to find the end of the mapping type
  let depth = 0;
  let mappingEnd = -1;

  for (let i = mappingStart; i < line.length; i++) {
    if (line[i] === "(") depth++;
    if (line[i] === ")") {
      depth--;
      if (depth === 0) {
        mappingEnd = i + 1;
        break;
      }
    }
  }

  if (mappingEnd === -1) return null;

  const typeStr = line.slice(mappingStart, mappingEnd);
  const rest = line.slice(mappingEnd).trim();

  // Extract the variable name
  const nameMatch = rest.match(/^(\w+)\s*;/);
  if (!nameMatch) return null;

  return { typeStr, name: nameMatch[1] };
}

/**
 * Parse struct fields from Solidity source.
 * @param structBody - The content inside the struct braces
 * @param includeExplicitZeroOffset - Whether to include byteOffset: 0 for first packed field
 */
function parseStructFields(structBody: string, includeExplicitZeroOffset: boolean = true): Record<string, FieldDef> {
  const fields: Record<string, FieldDef> = {};
  let currentSlot = 0;
  let currentByteOffset = 0;
  let isSlotPacked = false; // Track if current slot has multiple fields

  const lines = structBody.split("\n");

  // First pass: determine which slots will have packed fields
  const slotFieldCounts: Map<number, number> = new Map();
  let tempSlot = 0;
  let tempOffset = 0;

  for (const line of lines) {
    if (line.trim().startsWith("//") || line.trim().startsWith("*")) continue;

    const mappingResult = extractMappingType(line);
    if (mappingResult) {
      if (tempOffset > 0) {
        tempSlot++;
        tempOffset = 0;
      }
      tempSlot++;
      continue;
    }

    const fieldMatch = line.match(/^\s*([A-Z]?[a-zA-Z0-9_]+(?:\[\])?)\s+(\w+)\s*;/);
    if (fieldMatch) {
      const [, typeStr] = fieldMatch;
      const normalizedType = normalizeType(typeStr);
      const typeSize = getTypeSize(normalizedType);

      const isDynamic1 =
        normalizedType.endsWith("[]") || normalizedType.startsWith("mapping") || knownStructNames.has(normalizedType);

      if (isDynamic1) {
        if (tempOffset > 0) {
          tempSlot++;
          tempOffset = 0;
        }
        tempSlot++;
        continue;
      }

      if (tempOffset + typeSize > 32) {
        tempSlot++;
        tempOffset = 0;
      }

      slotFieldCounts.set(tempSlot, (slotFieldCounts.get(tempSlot) || 0) + 1);
      tempOffset += typeSize;

      if (tempOffset >= 32) {
        tempSlot++;
        tempOffset = 0;
      }
    }
  }

  // Second pass: actually build the fields
  for (const line of lines) {
    // Skip comments
    if (line.trim().startsWith("//") || line.trim().startsWith("*")) {
      continue;
    }

    // Try mapping first (with proper nested mapping support)
    const mappingResult = extractMappingType(line);
    if (mappingResult) {
      const { typeStr, name } = mappingResult;
      // Mappings always start a new slot
      if (currentByteOffset > 0) {
        currentSlot++;
        currentByteOffset = 0;
      }
      fields[name] = {
        slot: currentSlot,
        type: normalizeType(typeStr),
      };
      currentSlot++;
      isSlotPacked = false;
      continue;
    }

    // Try regular field (handles types like IMessageService, uint256, address[], etc.)
    const fieldMatch = line.match(/^\s*([A-Z]?[a-zA-Z0-9_]+(?:\[\])?)\s+(\w+)\s*;/);
    if (fieldMatch) {
      const [, typeStr, name] = fieldMatch;
      const normalizedType = normalizeType(typeStr);
      const typeSize = getTypeSize(normalizedType);

      // Dynamic types (arrays, mappings, structs) always start a new slot
      const isDynamic =
        normalizedType.endsWith("[]") || normalizedType.startsWith("mapping") || knownStructNames.has(normalizedType);

      if (isDynamic) {
        if (currentByteOffset > 0) {
          currentSlot++;
          currentByteOffset = 0;
        }
        fields[name] = {
          slot: currentSlot,
          type: normalizedType,
        };
        currentSlot++;
        isSlotPacked = false;
        continue;
      }

      // Check if this field fits in the current slot
      if (currentByteOffset + typeSize > 32) {
        // Start new slot
        currentSlot++;
        currentByteOffset = 0;
      }

      // Determine if this slot is packed (has multiple fields)
      const slotFieldCount = slotFieldCounts.get(currentSlot) || 1;
      isSlotPacked = slotFieldCount > 1;

      // Include byteOffset if not at offset 0, or if slot is packed and includeExplicitZeroOffset is true
      const shouldIncludeOffset = currentByteOffset > 0 || (isSlotPacked && includeExplicitZeroOffset && typeSize < 32);

      fields[name] = {
        slot: currentSlot,
        type: normalizedType,
        ...(shouldIncludeOffset ? { byteOffset: currentByteOffset } : {}),
      };

      currentByteOffset += typeSize;

      // If we filled the slot exactly, move to next
      if (currentByteOffset >= 32) {
        currentSlot++;
        currentByteOffset = 0;
        isSlotPacked = false;
      }
    }
  }

  return fields;
}

/**
 * Extract preceding comments (both multi-line and single-line styles) for a struct.
 */
function extractPrecedingComments(source: string, structIndex: number): string {
  const beforeStruct = source.slice(0, structIndex);
  const lines = beforeStruct.split("\n");
  const commentLines: string[] = [];

  for (let i = lines.length - 1; i >= 0; i--) {
    const line = lines[i].trim();

    if (line === "") continue;

    if (line.startsWith("///")) {
      commentLines.unshift(line);
      continue;
    }

    if (line.endsWith("*/")) {
      let blockStart = i;
      while (blockStart >= 0 && !lines[blockStart].includes("/**")) {
        commentLines.unshift(lines[blockStart]);
        blockStart--;
      }
      if (blockStart >= 0) {
        commentLines.unshift(lines[blockStart]);
      }
      break;
    }

    break;
  }

  return commentLines.join("\n");
}

/**
 * Extract explicit storage location constants from source.
 * Returns a map of struct name -> explicit slot value.
 */
function extractExplicitConstants(source: string): Map<string, string> {
  const constants = new Map<string, string>();
  const constantPattern = /bytes32\s+(?:private\s+)?constant\s+(\w+StorageLocation)\s*=\s*(0x[a-fA-F0-9]+)/g;
  let match;
  while ((match = constantPattern.exec(source)) !== null) {
    const [, constantName, slotValue] = match;
    const structName = constantName.replace("StorageLocation", "");
    constants.set(structName, slotValue.toLowerCase());
  }
  return constants;
}

// ============================================================================
// Main Functions
// ============================================================================

/**
 * Parse a single Solidity file and extract storage schema.
 *
 * @param adapter - CryptoAdapter for hashing operations
 * @param source - Solidity source code
 * @param fileName - Optional filename for error messages
 * @param options - Parser options
 */
export function parseSoliditySource(
  adapter: CryptoAdapter,
  source: string,
  fileName?: string,
  options: SchemaGeneratorOptions = {},
): ParseResult {
  // Reset known types for this single-file parsing session
  resetKnownStructs();
  resetKnownEnums();

  // First pass: discover all struct names in this file
  // Use limited whitespace quantifier to prevent ReDoS
  const structNamePattern = /struct[ \t]+(\w+)[ \t]*\{/g;
  let nameMatch;
  while ((nameMatch = structNamePattern.exec(source)) !== null) {
    registerStructName(nameMatch[1]);
  }

  // First pass: discover all enum definitions in this file
  parseEnums(source);

  // Parse the source
  return parseSoliditySourceInternal(adapter, source, fileName, options);
}

/**
 * Merge multiple schemas into one.
 * Later schemas override earlier ones for conflicting struct names.
 */
export function mergeSchemas(schemas: Schema[]): Schema {
  const merged: Schema = {
    $comment: "Auto-generated storage schema",
    structs: {},
  };

  for (const schema of schemas) {
    if (schema.$comment && schema.$comment !== "Auto-generated storage schema") {
      merged.$comment = schema.$comment;
    }
    for (const [name, def] of Object.entries(schema.structs)) {
      if (merged.structs[name]) {
        // Merge fields, prefer newer baseSlot/namespace
        merged.structs[name] = {
          ...merged.structs[name],
          ...def,
          fields: { ...merged.structs[name].fields, ...def.fields },
        };
      } else {
        merged.structs[name] = def;
      }
    }
  }

  return merged;
}

/**
 * Generate a storage schema from multiple Solidity sources.
 *
 * @param adapter - CryptoAdapter for hashing operations
 * @param sources - Array of { source, fileName } objects
 * @param options - Generator options
 */
export function generateSchema(
  adapter: CryptoAdapter,
  sources: Array<{ source: string; fileName?: string }>,
  options: SchemaGeneratorOptions = {},
): ParseResult {
  const allWarnings: string[] = [];
  const schemas: Schema[] = [];

  // Reset known types for this multi-file parsing session
  resetKnownStructs();
  resetKnownEnums();

  // First pass: discover all struct names across ALL sources
  // Use limited whitespace quantifier to prevent ReDoS
  const structNamePattern = /struct[ \t]+(\w+)[ \t]*\{/g;
  for (const { source } of sources) {
    let nameMatch;
    while ((nameMatch = structNamePattern.exec(source)) !== null) {
      registerStructName(nameMatch[1]);
    }
  }

  // First pass: discover all enum definitions across ALL sources
  for (const { source } of sources) {
    parseEnums(source);
  }

  // Second pass: parse each source (without resetting type names)
  for (const { source, fileName } of sources) {
    const { schema, warnings } = parseSoliditySourceInternal(adapter, source, fileName, options);
    schemas.push(schema);
    allWarnings.push(...warnings);
  }

  const finalSchema = mergeSchemas(schemas);
  if (options.comment) {
    finalSchema.$comment = options.comment;
  }

  return { schema: finalSchema, warnings: allWarnings };
}

/**
 * Internal parsing function that doesn't reset known structs.
 * Used by generateSchema for multi-file parsing.
 */
function parseSoliditySourceInternal(
  adapter: CryptoAdapter,
  source: string,
  fileName?: string,
  options: SchemaGeneratorOptions = {},
): ParseResult {
  const schema: Schema = {
    $comment: options.comment ?? "Auto-generated storage schema",
    structs: {},
  };
  const warnings: string[] = [];

  // Default to including explicit zero offset for consistency
  const includeExplicitZeroOffset = options.includeExplicitZeroOffset !== false;

  // Extract all explicit storage location constants
  const explicitConstants = extractExplicitConstants(source);

  // Parse struct definitions with field details
  // Use limited whitespace quantifier to prevent ReDoS
  const structPattern = /struct[ \t]+(\w+)[ \t]*\{([^}]+)\}/g;

  let match;
  while ((match = structPattern.exec(source)) !== null) {
    const [, structName, structBody] = match;
    const structIndex = match.index;

    // Extract comments that precede this struct
    const comments = extractPrecedingComments(source, structIndex);

    const namespace = extractNamespace(comments);
    const fields = parseStructFields(structBody, includeExplicitZeroOffset);

    const structDef: StructDef = { fields };

    if (namespace) {
      structDef.namespace = namespace;
      const calculatedSlot = calculateErc7201BaseSlot(adapter, namespace);
      structDef.baseSlot = calculatedSlot;

      // Validate against explicit constant if present
      const explicitSlot = explicitConstants.get(structName);
      if (explicitSlot && options.validateConstants !== false) {
        if (explicitSlot !== calculatedSlot) {
          const fileInfo = fileName ? ` in ${fileName}` : "";
          warnings.push(
            `Calculated baseSlot for ${structName} (${calculatedSlot}) does not match ` +
              `explicit constant${fileInfo} (${explicitSlot}). Using explicit value.`,
          );
          structDef.baseSlot = explicitSlot;
        }
      }
    } else {
      // No namespace annotation, but check for explicit constant
      const explicitSlot = explicitConstants.get(structName);
      if (explicitSlot) {
        structDef.baseSlot = explicitSlot;
      }
    }

    schema.structs[structName] = structDef;
  }

  return { schema, warnings };
}
