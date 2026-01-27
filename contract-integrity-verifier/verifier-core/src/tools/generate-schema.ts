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
}

export interface ParseResult {
  schema: Schema;
  warnings: string[];
}

// ============================================================================
// Type Sizes
// ============================================================================

const TYPE_SIZES: Record<string, number> = {
  bool: 1,
  uint8: 1,
  int8: 1,
  uint16: 2,
  int16: 2,
  uint32: 4,
  int32: 4,
  uint64: 8,
  int64: 8,
  uint96: 12,
  int96: 12,
  uint128: 16,
  int128: 16,
  uint256: 32,
  int256: 32,
  address: 20,
  bytes32: 32,
  bytes4: 4,
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
 * Parse a Solidity type and return normalized type string.
 */
function normalizeType(solidityType: string): string {
  // Handle enums as uint8
  if (
    !solidityType.startsWith("uint") &&
    !solidityType.startsWith("int") &&
    !solidityType.startsWith("bytes") &&
    solidityType !== "address" &&
    solidityType !== "bool"
  ) {
    // Check if it looks like an enum (single word, PascalCase)
    if (/^[A-Z][a-zA-Z0-9]*$/.test(solidityType)) {
      return "uint8"; // Enums are uint8 by default
    }
  }

  // Handle arrays
  if (solidityType.endsWith("[]")) {
    const baseType = solidityType.slice(0, -2);
    return `${normalizeType(baseType)}[]`;
  }

  // Handle mappings: mapping(KeyType => ValueType)
  const mappingMatch = solidityType.match(/^mapping\s*\(\s*(\w+).*=>\s*(.+)\)$/);
  if (mappingMatch) {
    const keyType = normalizeType(mappingMatch[1]);
    const valueType = normalizeType(mappingMatch[2].trim());
    return `mapping(${keyType} => ${valueType})`;
  }

  return solidityType;
}

/**
 * Get the byte size of a type (for packing calculation).
 */
function getTypeSize(type: string): number {
  return TYPE_SIZES[type] ?? 32;
}

/**
 * Parse struct fields from Solidity source.
 */
function parseStructFields(structBody: string): Record<string, FieldDef> {
  const fields: Record<string, FieldDef> = {};
  let currentSlot = 0;
  let currentByteOffset = 0;

  const lines = structBody.split("\n");

  for (const line of lines) {
    // Skip comments
    if (line.trim().startsWith("//") || line.trim().startsWith("*")) {
      continue;
    }

    // Try mapping first
    const mappingMatch = line.match(/^\s*(mapping\s*\([^)]+\))\s+(\w+)\s*;/);
    if (mappingMatch) {
      const [, typeStr, name] = mappingMatch;
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
      continue;
    }

    // Try regular field (handles types like IMessageService, uint256, address[], etc.)
    const fieldMatch = line.match(/^\s*([A-Z]?[a-zA-Z0-9_]+(?:\[\])?)\s+(\w+)\s*;/);
    if (fieldMatch) {
      const [, typeStr, name] = fieldMatch;
      const normalizedType = normalizeType(typeStr);
      const typeSize = getTypeSize(normalizedType);

      // Dynamic types (arrays, mappings) always start a new slot
      if (normalizedType.endsWith("[]") || normalizedType.startsWith("mapping")) {
        if (currentByteOffset > 0) {
          currentSlot++;
          currentByteOffset = 0;
        }
        fields[name] = {
          slot: currentSlot,
          type: normalizedType,
        };
        currentSlot++;
        continue;
      }

      // Check if this field fits in the current slot
      if (currentByteOffset + typeSize > 32) {
        // Start new slot
        currentSlot++;
        currentByteOffset = 0;
      }

      fields[name] = {
        slot: currentSlot,
        type: normalizedType,
        ...(currentByteOffset > 0 ? { byteOffset: currentByteOffset } : {}),
      };

      currentByteOffset += typeSize;

      // If we filled the slot exactly, move to next
      if (currentByteOffset >= 32) {
        currentSlot++;
        currentByteOffset = 0;
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
  const schema: Schema = {
    $comment: options.comment ?? "Auto-generated storage schema",
    structs: {},
  };
  const warnings: string[] = [];

  // First, extract all explicit storage location constants
  const explicitConstants = extractExplicitConstants(source);

  // Find struct definitions
  const structPattern = /struct\s+(\w+)\s*\{([^}]+)\}/g;

  let match;
  while ((match = structPattern.exec(source)) !== null) {
    const [, structName, structBody] = match;
    const structIndex = match.index;

    // Extract comments that precede this struct
    const comments = extractPrecedingComments(source, structIndex);

    const namespace = extractNamespace(comments);
    const fields = parseStructFields(structBody);

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

  for (const { source, fileName } of sources) {
    const { schema, warnings } = parseSoliditySource(adapter, source, fileName, options);
    schemas.push(schema);
    allWarnings.push(...warnings);
  }

  const finalSchema = mergeSchemas(schemas);
  if (options.comment) {
    finalSchema.$comment = options.comment;
  }

  return { schema: finalSchema, warnings: allWarnings };
}
