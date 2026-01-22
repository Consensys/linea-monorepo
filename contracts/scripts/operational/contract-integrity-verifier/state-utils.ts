/**
 * Contract Integrity Verifier - State Verification Utilities
 *
 * Utilities for verifying contract state via view calls, storage slots,
 * and ERC-7201 namespaced storage.
 */

import { ethers } from "ethers";
import {
  StateVerificationConfig,
  StateVerificationResult,
  ViewCallConfig,
  ViewCallResult,
  SlotConfig,
  SlotResult,
  NamespaceConfig,
  NamespaceResult,
  AbiElement,
} from "./types";

// ============================================================================
// ERC-7201 Namespace Calculation
// ============================================================================

/**
 * Calculates the base storage slot for an ERC-7201 namespace.
 * Formula: keccak256(abi.encode(uint256(keccak256(id)) - 1)) & ~bytes32(uint256(0xff))
 */
export function calculateErc7201Slot(namespaceId: string): string {
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
// Storage Slot Reading and Decoding
// ============================================================================

/**
 * Reads and decodes a storage slot value.
 */
export async function readStorageSlot(
  provider: ethers.JsonRpcProvider,
  address: string,
  slot: string,
  type: string,
  offset: number = 0,
): Promise<unknown> {
  const rawValue = await provider.getStorage(address, slot);

  return decodeSlotValue(rawValue, type, offset);
}

/**
 * Decodes a raw storage slot value based on type and offset.
 */
export function decodeSlotValue(rawValue: string, type: string, offset: number = 0): unknown {
  // Remove 0x prefix and ensure 64 chars (32 bytes)
  const normalized = rawValue.slice(2).padStart(64, "0");

  // For packed storage, extract the relevant bytes
  // Storage is right-aligned (low bytes at the end)
  // Offset is from the right (low bytes)
  const typeBytes = getTypeBytes(type);
  const startByte = 32 - offset - typeBytes;
  const endByte = 32 - offset;
  const hexValue = normalized.slice(startByte * 2, endByte * 2);

  switch (type) {
    case "address":
      return ethers.getAddress("0x" + hexValue.slice(-40));
    case "bool":
      return hexValue !== "0".repeat(hexValue.length);
    case "uint8":
    case "uint32":
    case "uint64":
    case "uint128":
      return BigInt("0x" + hexValue).toString();
    case "uint256":
      return BigInt("0x" + hexValue).toString();
    case "bytes32":
      return "0x" + hexValue;
    default:
      return "0x" + hexValue;
  }
}

/**
 * Returns the byte size of a Solidity type.
 */
function getTypeBytes(type: string): number {
  switch (type) {
    case "address":
      return 20;
    case "bool":
    case "uint8":
      return 1;
    case "uint32":
      return 4;
    case "uint64":
      return 8;
    case "uint128":
      return 16;
    case "uint256":
    case "bytes32":
      return 32;
    default:
      return 32;
  }
}

// ============================================================================
// View Call Execution
// ============================================================================

/**
 * Executes a view function call and returns the result.
 */
export async function executeViewCall(
  provider: ethers.JsonRpcProvider,
  address: string,
  abi: AbiElement[],
  config: ViewCallConfig,
): Promise<ViewCallResult> {
  try {
    // Find function in ABI
    const funcAbi = abi.find((e) => e.type === "function" && e.name === config.function);
    if (!funcAbi) {
      return {
        function: config.function,
        params: config.params,
        expected: config.expected,
        actual: undefined,
        status: "fail",
        message: `Function '${config.function}' not found in ABI`,
      };
    }

    // Create interface and encode call
    const iface = new ethers.Interface([funcAbi]);
    const calldata = iface.encodeFunctionData(config.function, config.params ?? []);

    // Execute call
    const result = await provider.call({ to: address, data: calldata });

    // Decode result
    const decoded = iface.decodeFunctionResult(config.function, result);

    // Handle single vs multiple return values
    const actual = decoded.length === 1 ? formatValue(decoded[0]) : decoded.map(formatValue);

    // Compare
    const comparison = config.comparison ?? "eq";
    const pass = compareValues(actual, config.expected, comparison);

    return {
      function: config.function,
      params: config.params,
      expected: config.expected,
      actual,
      status: pass ? "pass" : "fail",
      message: pass
        ? `${config.function}() = ${formatForDisplay(actual)}`
        : `Expected ${formatForDisplay(config.expected)}, got ${formatForDisplay(actual)}`,
    };
  } catch (error) {
    return {
      function: config.function,
      params: config.params,
      expected: config.expected,
      actual: undefined,
      status: "fail",
      message: `Call failed: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}

/**
 * Formats a value for comparison (handles BigInt, etc.)
 */
function formatValue(value: unknown): unknown {
  if (typeof value === "bigint") {
    return value.toString();
  }
  if (Array.isArray(value)) {
    return value.map(formatValue);
  }
  return value;
}

/**
 * Formats a value for display in messages.
 */
function formatForDisplay(value: unknown): string {
  if (typeof value === "string" && value.length > 20) {
    return value.slice(0, 10) + "..." + value.slice(-8);
  }
  return String(value);
}

/**
 * Compares two values using the specified comparison mode.
 */
function compareValues(actual: unknown, expected: unknown, comparison: string): boolean {
  // Normalize for comparison
  const normalizedActual = normalizeForComparison(actual);
  const normalizedExpected = normalizeForComparison(expected);

  switch (comparison) {
    case "eq":
      return normalizedActual === normalizedExpected;
    case "gt":
      return BigInt(normalizedActual) > BigInt(normalizedExpected);
    case "gte":
      return BigInt(normalizedActual) >= BigInt(normalizedExpected);
    case "lt":
      return BigInt(normalizedActual) < BigInt(normalizedExpected);
    case "lte":
      return BigInt(normalizedActual) <= BigInt(normalizedExpected);
    case "contains":
      return String(normalizedActual).includes(String(normalizedExpected));
    default:
      return normalizedActual === normalizedExpected;
  }
}

/**
 * Normalizes a value for comparison (lowercase addresses, stringify numbers).
 */
function normalizeForComparison(value: unknown): string {
  if (typeof value === "string") {
    // Lowercase addresses
    if (value.startsWith("0x") && value.length === 42) {
      return value.toLowerCase();
    }
    return value;
  }
  if (typeof value === "bigint" || typeof value === "number") {
    return String(value);
  }
  if (typeof value === "boolean") {
    return String(value);
  }
  return String(value);
}

// ============================================================================
// Slot Verification
// ============================================================================

/**
 * Verifies a single storage slot.
 */
export async function verifySlot(
  provider: ethers.JsonRpcProvider,
  address: string,
  config: SlotConfig,
): Promise<SlotResult> {
  try {
    const actual = await readStorageSlot(provider, address, config.slot, config.type, config.offset ?? 0);

    const pass = compareValues(actual, config.expected, "eq");

    return {
      slot: config.slot,
      name: config.name,
      expected: config.expected,
      actual,
      status: pass ? "pass" : "fail",
      message: pass
        ? `${config.name} = ${formatForDisplay(actual)}`
        : `${config.name}: expected ${formatForDisplay(config.expected)}, got ${formatForDisplay(actual)}`,
    };
  } catch (error) {
    return {
      slot: config.slot,
      name: config.name,
      expected: config.expected,
      actual: undefined,
      status: "fail",
      message: `Failed to read slot: ${error instanceof Error ? error.message : String(error)}`,
    };
  }
}

// ============================================================================
// Namespace Verification
// ============================================================================

/**
 * Verifies variables within an ERC-7201 namespace.
 */
export async function verifyNamespace(
  provider: ethers.JsonRpcProvider,
  address: string,
  config: NamespaceConfig,
): Promise<NamespaceResult> {
  const baseSlot = calculateErc7201Slot(config.id);
  const baseSlotBigInt = BigInt(baseSlot);

  const variableResults: SlotResult[] = [];

  for (const variable of config.variables) {
    // Calculate actual slot: baseSlot + offset
    const actualSlotBigInt = baseSlotBigInt + BigInt(variable.offset);
    const actualSlot = "0x" + actualSlotBigInt.toString(16).padStart(64, "0");

    const result = await verifySlot(provider, address, {
      slot: actualSlot,
      type: variable.type,
      name: variable.name,
      expected: variable.expected,
    });

    variableResults.push(result);
  }

  const allPass = variableResults.every((r) => r.status === "pass");

  return {
    namespaceId: config.id,
    baseSlot,
    variables: variableResults,
    status: allPass ? "pass" : "fail",
  };
}

// ============================================================================
// Main State Verification
// ============================================================================

/**
 * Performs complete state verification for a contract.
 */
export async function verifyState(
  provider: ethers.JsonRpcProvider,
  address: string,
  abi: AbiElement[],
  config: StateVerificationConfig,
): Promise<StateVerificationResult> {
  const viewCallResults: ViewCallResult[] = [];
  const namespaceResults: NamespaceResult[] = [];
  const slotResults: SlotResult[] = [];

  // 1. Execute view calls
  if (config.viewCalls && config.viewCalls.length > 0) {
    for (const viewCall of config.viewCalls) {
      const result = await executeViewCall(provider, address, abi, viewCall);
      viewCallResults.push(result);
    }
  }

  // 2. Verify namespaces (ERC-7201)
  if (config.namespaces && config.namespaces.length > 0) {
    for (const namespace of config.namespaces) {
      const result = await verifyNamespace(provider, address, namespace);
      namespaceResults.push(result);
    }
  }

  // 3. Verify explicit slots
  if (config.slots && config.slots.length > 0) {
    for (const slot of config.slots) {
      const result = await verifySlot(provider, address, slot);
      slotResults.push(result);
    }
  }

  // Aggregate results
  const allViewCallsPass = viewCallResults.every((r) => r.status === "pass");
  const allNamespacesPass = namespaceResults.every((r) => r.status === "pass");
  const allSlotsPass = slotResults.every((r) => r.status === "pass");

  const totalChecks = viewCallResults.length + namespaceResults.length + slotResults.length;
  const passedChecks =
    viewCallResults.filter((r) => r.status === "pass").length +
    namespaceResults.filter((r) => r.status === "pass").length +
    slotResults.filter((r) => r.status === "pass").length;

  const allPass = allViewCallsPass && allNamespacesPass && allSlotsPass;

  return {
    status: allPass ? "pass" : "fail",
    message: allPass ? `All ${totalChecks} state checks passed` : `${passedChecks}/${totalChecks} state checks passed`,
    viewCallResults: viewCallResults.length > 0 ? viewCallResults : undefined,
    namespaceResults: namespaceResults.length > 0 ? namespaceResults : undefined,
    slotResults: slotResults.length > 0 ? slotResults : undefined,
  };
}
