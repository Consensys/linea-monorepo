#!/usr/bin/env ts-node
/**
 * Contract Integrity Verifier - Test Runner
 * Run with: npx ts-node scripts/operational/contract-integrity-verifier/__tests__/run-tests.ts
 */

import { ethers } from "ethers";
import {
  calculateErc7201Slot,
  decodeSlotValue,
  executeViewCall,
  verifySlot,
  verifyNamespace,
  verifyState,
} from "../state-utils";
import { detectArtifactFormat, extractSelectorsFromArtifact } from "../abi-utils";
import { compareBytecode } from "../bytecode-utils";
import { calculateErc7201BaseSlot, parsePath, computeSlot } from "../storage-path-utils";
import {
  AbiElement,
  ViewCallConfig,
  SlotConfig,
  NamespaceConfig,
  StateVerificationConfig,
  NormalizedArtifact,
  ImmutableReference,
  StorageSchema,
} from "../types";

// ============================================================================
// Test Utilities
// ============================================================================

let testsPassed = 0;
let testsFailed = 0;

function assert(condition: boolean, message: string): void {
  if (condition) {
    console.log(`  ✓ ${message}`);
    testsPassed++;
  } else {
    console.log(`  ✗ ${message}`);
    testsFailed++;
  }
}

function assertEqual<T>(actual: T, expected: T, message: string): void {
  const normalizedActual = typeof actual === "string" ? actual.toLowerCase() : actual;
  const normalizedExpected = typeof expected === "string" ? expected.toLowerCase() : expected;
  assert(normalizedActual === normalizedExpected, `${message}: expected ${expected}, got ${actual}`);
}

// ============================================================================
// Mock Provider
// ============================================================================

class MockProvider {
  private storage: Map<string, string> = new Map();
  private callResults: Map<string, string> = new Map();

  setStorage(slot: string, value: string): void {
    this.storage.set(slot.toLowerCase(), value);
  }

  setCallResult(calldata: string, result: string): void {
    this.callResults.set(calldata.toLowerCase(), result);
  }

  async getStorage(_address: string, slot: string): Promise<string> {
    return this.storage.get(slot.toLowerCase()) ?? "0x" + "0".repeat(64);
  }

  async call(tx: { to: string; data: string }): Promise<string> {
    const result = this.callResults.get(tx.data.toLowerCase());
    if (!result) {
      throw new Error(`No mock result for calldata: ${tx.data}`);
    }
    return result;
  }
}

// ============================================================================
// Test Data
// ============================================================================

const TEST_ADDRESS = "0x1234567890123456789012345678901234567890";
const TEST_OWNER = "0xabcdef1234567890abcdef1234567890abcdef12";

const MOCK_ABI: AbiElement[] = [
  {
    type: "function",
    name: "owner",
    inputs: [],
    outputs: [{ internalType: "address", name: "", type: "address" }],
    stateMutability: "view",
  },
  {
    type: "function",
    name: "hasRole",
    inputs: [
      { internalType: "bytes32", name: "role", type: "bytes32" },
      { internalType: "address", name: "account", type: "address" },
    ],
    outputs: [{ internalType: "bool", name: "", type: "bool" }],
    stateMutability: "view",
  },
];

// ============================================================================
// Tests
// ============================================================================

async function testErc7201SlotCalculation(): Promise<void> {
  console.log("\n📦 Testing ERC-7201 Slot Calculation...");

  const slot = calculateErc7201Slot("example.main");
  assert(slot.match(/^0x[0-9a-f]{64}$/) !== null, "Slot is valid 32-byte hex");
  assert(slot.slice(-2) === "00", "Last byte is masked to 0x00");

  const slot1 = calculateErc7201Slot("linea.storage.YieldManager");
  const slot2 = calculateErc7201Slot("linea.storage.LineaRollup");
  assert(slot1 !== slot2, "Different namespaces produce different slots");
}

async function testDecodeSlotValue(): Promise<void> {
  console.log("\n📦 Testing Slot Value Decoding...");

  // Address
  const addressValue = "0x000000000000000000000000" + TEST_OWNER.slice(2);
  const decodedAddress = decodeSlotValue(addressValue, "address");
  assertEqual(decodedAddress, ethers.getAddress(TEST_OWNER), "Decode address");

  // uint256
  const uint256Value = "0x" + "0".repeat(62) + "64"; // 100 in hex
  const decodedUint256 = decodeSlotValue(uint256Value, "uint256");
  assertEqual(decodedUint256, "100", "Decode uint256");

  // uint8
  const uint8Value = "0x" + "0".repeat(62) + "06";
  const decodedUint8 = decodeSlotValue(uint8Value, "uint8");
  assertEqual(decodedUint8, "6", "Decode uint8");

  // bool true
  const boolTrueValue = "0x" + "0".repeat(62) + "01";
  const decodedBoolTrue = decodeSlotValue(boolTrueValue, "bool");
  assertEqual(decodedBoolTrue, true, "Decode bool (true)");

  // bool false
  const boolFalseValue = "0x" + "0".repeat(64);
  const decodedBoolFalse = decodeSlotValue(boolFalseValue, "bool");
  assertEqual(decodedBoolFalse, false, "Decode bool (false)");
}

async function testViewCalls(): Promise<void> {
  console.log("\n📦 Testing View Calls...");

  const mockProvider = new MockProvider();
  const iface = new ethers.Interface(MOCK_ABI);

  // Setup mock for owner()
  const ownerCalldata = iface.encodeFunctionData("owner", []);
  const ownerReturnData = iface.encodeFunctionResult("owner", [TEST_OWNER]);
  mockProvider.setCallResult(ownerCalldata, ownerReturnData);

  const config: ViewCallConfig = {
    function: "owner",
    expected: TEST_OWNER,
  };

  const result = await executeViewCall(
    mockProvider as unknown as ethers.JsonRpcProvider,
    TEST_ADDRESS,
    MOCK_ABI,
    config,
  );

  assertEqual(result.status, "pass", "View call owner() passes");

  // Test with parameters (hasRole)
  const role = "0x" + "0".repeat(64);
  const hasRoleCalldata = iface.encodeFunctionData("hasRole", [role, TEST_OWNER]);
  const hasRoleReturnData = iface.encodeFunctionResult("hasRole", [true]);
  mockProvider.setCallResult(hasRoleCalldata, hasRoleReturnData);

  const hasRoleConfig: ViewCallConfig = {
    function: "hasRole",
    params: [role, TEST_OWNER],
    expected: true,
  };

  const hasRoleResult = await executeViewCall(
    mockProvider as unknown as ethers.JsonRpcProvider,
    TEST_ADDRESS,
    MOCK_ABI,
    hasRoleConfig,
  );

  assertEqual(hasRoleResult.status, "pass", "View call hasRole() with params passes");
}

async function testSlotVerification(): Promise<void> {
  console.log("\n📦 Testing Slot Verification...");

  const mockProvider = new MockProvider();
  mockProvider.setStorage("0x0", "0x" + "0".repeat(62) + "06");

  const config: SlotConfig = {
    slot: "0x0",
    type: "uint8",
    name: "_initialized",
    expected: "6",
  };

  const result = await verifySlot(mockProvider as unknown as ethers.JsonRpcProvider, TEST_ADDRESS, config);

  assertEqual(result.status, "pass", "Slot verification passes with correct value");
  assertEqual(result.actual, "6", "Actual value is correct");
}

async function testNamespaceVerification(): Promise<void> {
  console.log("\n📦 Testing Namespace Verification...");

  const mockProvider = new MockProvider();
  const baseSlot = calculateErc7201Slot("test.storage.MyContract");
  const baseSlotBigInt = BigInt(baseSlot);

  // Set storage for offset 0 (address)
  const slot0 = "0x" + baseSlotBigInt.toString(16).padStart(64, "0");
  mockProvider.setStorage(slot0, "0x000000000000000000000000" + TEST_OWNER.slice(2));

  const config: NamespaceConfig = {
    id: "test.storage.MyContract",
    variables: [{ offset: 0, type: "address", name: "admin", expected: TEST_OWNER }],
  };

  const result = await verifyNamespace(mockProvider as unknown as ethers.JsonRpcProvider, TEST_ADDRESS, config);

  assertEqual(result.status, "pass", "Namespace verification passes");
  assertEqual(result.variables[0].status, "pass", "Variable verification passes");
}

async function testFullStateVerification(): Promise<void> {
  console.log("\n📦 Testing Full State Verification...");

  const mockProvider = new MockProvider();
  const iface = new ethers.Interface(MOCK_ABI);

  // Setup view call mock
  const ownerCalldata = iface.encodeFunctionData("owner", []);
  const ownerReturnData = iface.encodeFunctionResult("owner", [TEST_OWNER]);
  mockProvider.setCallResult(ownerCalldata, ownerReturnData);

  // Setup slot mock
  mockProvider.setStorage("0x0", "0x" + "0".repeat(62) + "06");

  // Setup namespace mock
  const baseSlot = calculateErc7201Slot("linea.storage.YieldManager");
  const baseSlotBigInt = BigInt(baseSlot);
  const slot0 = "0x" + baseSlotBigInt.toString(16).padStart(64, "0");
  mockProvider.setStorage(slot0, "0x000000000000000000000000" + TEST_OWNER.slice(2));

  const config: StateVerificationConfig = {
    viewCalls: [{ function: "owner", expected: TEST_OWNER }],
    slots: [{ slot: "0x0", type: "uint8", name: "_initialized", expected: "6" }],
    namespaces: [
      {
        id: "linea.storage.YieldManager",
        variables: [{ offset: 0, type: "address", name: "messageService", expected: TEST_OWNER }],
      },
    ],
  };

  const result = await verifyState(mockProvider as unknown as ethers.JsonRpcProvider, TEST_ADDRESS, MOCK_ABI, config);

  assertEqual(result.status, "pass", "Full state verification passes");
  assert(result.message.includes("3 state checks passed"), "Message indicates 3 checks passed");
}

// ============================================================================
// Foundry Artifact Tests
// ============================================================================

async function testArtifactFormatDetection(): Promise<void> {
  console.log("\n📦 Testing Artifact Format Detection...");

  // Hardhat format
  const hardhatArtifact = {
    contractName: "TestContract",
    abi: [],
    bytecode: "0x6080604052",
    deployedBytecode: "0x6080604052",
  };
  assertEqual(detectArtifactFormat(hardhatArtifact), "hardhat", "Detects Hardhat format");

  // Foundry format
  const foundryArtifact = {
    abi: [],
    bytecode: { object: "0x6080604052" },
    deployedBytecode: { object: "0x6080604052" },
  };
  assertEqual(detectArtifactFormat(foundryArtifact), "foundry", "Detects Foundry format");
}

async function testFoundryMethodIdentifiers(): Promise<void> {
  console.log("\n📦 Testing Foundry Method Identifiers...");

  const artifact: NormalizedArtifact = {
    format: "foundry",
    contractName: "TestContract",
    abi: [],
    bytecode: "0x",
    deployedBytecode: "0x",
    immutableReferences: undefined,
    methodIdentifiers: new Map([
      ["8da5cb5b", "owner()"],
      ["91d14854", "hasRole(bytes32,address)"],
    ]),
  };

  const selectors = extractSelectorsFromArtifact(artifact);
  assertEqual(selectors.size, 2, "Extracts 2 selectors from Foundry artifact");
  assertEqual(selectors.get("8da5cb5b"), "owner()", "Correct owner selector mapping");
  assertEqual(selectors.get("91d14854"), "hasRole(bytes32,address)", "Correct hasRole selector mapping");
}

async function testBytecodeComparisonWithKnownImmutables(): Promise<void> {
  console.log("\n📦 Testing Bytecode Comparison with Known Immutables...");

  // Create bytecode with a known difference at position 100 (32 bytes)
  const baseBytecode = "60806040" + "00".repeat(96); // 100 bytes of padding
  const localImmutable = "0".repeat(64); // 32 bytes of zeros (placeholder)
  const remoteImmutable = "abcdef1234567890".repeat(4); // 32 bytes of actual value
  const suffix = "00".repeat(50); // More padding

  const localBytecode = baseBytecode + localImmutable + suffix;
  const remoteBytecode = baseBytecode + remoteImmutable + suffix;

  // Known immutable at position 100, length 32
  const knownImmutables: ImmutableReference[] = [{ start: 100, length: 32 }];

  const result = compareBytecode(localBytecode, remoteBytecode, knownImmutables);

  assertEqual(result.status, "pass", "Passes when differences are at known immutable positions");
  assert(result.onlyImmutablesDiffer === true, "Correctly identifies only immutables differ");
  assert(result.message.includes("known positions"), "Message indicates known positions: " + result.message);
}

async function testBytecodeComparisonWithUnknownDifferences(): Promise<void> {
  console.log("\n📦 Testing Bytecode Comparison with Unknown Differences...");

  // Create bytecode with difference NOT at immutable position
  const baseBytecode = "60806040" + "00".repeat(50);
  const localPart = "aa".repeat(10);
  const remotePart = "bb".repeat(10);
  const suffix = "00".repeat(100);

  const localBytecode = baseBytecode + localPart + suffix;
  const remoteBytecode = baseBytecode + remotePart + suffix;

  // Known immutable at position 200 (not where the difference is)
  const knownImmutables: ImmutableReference[] = [{ start: 200, length: 32 }];

  const result = compareBytecode(localBytecode, remoteBytecode, knownImmutables);

  assertEqual(result.status, "fail", "Fails when differences are not at known immutable positions");
  assert(result.onlyImmutablesDiffer === false, "Correctly identifies unexpected differences");
}

// ============================================================================
// Main
// ============================================================================

async function main(): Promise<void> {
  console.log("🧪 Running Contract Integrity Verifier Tests\n");
  console.log("=".repeat(50));

  // State verification tests
  await testErc7201SlotCalculation();
  await testDecodeSlotValue();
  await testViewCalls();
  await testSlotVerification();
  await testNamespaceVerification();
  await testFullStateVerification();

  // Foundry artifact tests
  await testArtifactFormatDetection();
  await testFoundryMethodIdentifiers();
  await testBytecodeComparisonWithKnownImmutables();
  await testBytecodeComparisonWithUnknownDifferences();

  // Storage path tests
  await testStoragePathParsing();
  await testStoragePathSlotComputation();
  await testErc7201BaseSlotCalculation();

  console.log("\n" + "=".repeat(50));
  console.log(`\n📊 Results: ${testsPassed} passed, ${testsFailed} failed`);

  if (testsFailed > 0) {
    process.exit(1);
  }
}

// ============================================================================
// Storage Path Tests
// ============================================================================

async function testStoragePathParsing(): Promise<void> {
  console.log("\n🧪 Testing storage path parsing...");

  // Simple field
  const simple = parsePath("MyStruct:myField");
  assertEqual(simple.structName, "MyStruct", "Simple path struct name");
  assertEqual(simple.segments.length, 1, "Simple path segment count");
  assertEqual(simple.segments[0].type, "field", "Simple path segment type");
  if (simple.segments[0].type === "field") {
    assertEqual(simple.segments[0].name, "myField", "Simple path field name");
  }

  // Nested field
  const nested = parsePath("Storage:config.value");
  assertEqual(nested.structName, "Storage", "Nested path struct name");
  assertEqual(nested.segments.length, 2, "Nested path segment count");

  // Array index
  const array = parsePath("Storage:items[0]");
  assertEqual(array.segments.length, 2, "Array path segment count");
  if (array.segments[1].type === "arrayIndex") {
    assertEqual(array.segments[1].index, 0, "Array index value");
  }

  // Array length
  const length = parsePath("Storage:items.length");
  assertEqual(length.segments.length, 2, "Array length segment count");
  assertEqual(length.segments[1].type, "arrayLength", "Array length segment type");

  // Mapping key
  const mapping = parsePath("Storage:balances[0x1234567890123456789012345678901234567890]");
  assertEqual(mapping.segments.length, 2, "Mapping path segment count");
  if (mapping.segments[1].type === "mappingKey") {
    assert(mapping.segments[1].key.startsWith("0x"), "Mapping key is address");
  }
}

async function testStoragePathSlotComputation(): Promise<void> {
  console.log("\n🧪 Testing storage path slot computation...");

  const testSchema: StorageSchema = {
    structs: {
      TestStorage: {
        namespace: "test.storage.TestStorage",
        fields: {
          firstField: { slot: 0, type: "address" },
          secondField: { slot: 1, type: "uint256" },
          packedField: { slot: 2, type: "bool", byteOffset: 0 },
        },
      },
    },
  };

  // First field at slot 0
  const path1 = parsePath("TestStorage:firstField");
  const slot1 = computeSlot(path1, testSchema);
  assertEqual(slot1.type, "address", "First field type");
  assertEqual(slot1.byteOffset, 0, "First field byte offset");

  // Second field at slot 1
  const path2 = parsePath("TestStorage:secondField");
  const slot2 = computeSlot(path2, testSchema);
  assertEqual(slot2.type, "uint256", "Second field type");

  // Verify slot offset from base
  const baseSlot = BigInt(slot1.slot);
  const secondSlot = BigInt(slot2.slot);
  assertEqual(secondSlot - baseSlot, 1n, "Second field is 1 slot after first");

  // Packed field byte offset
  const path3 = parsePath("TestStorage:packedField");
  const slot3 = computeSlot(path3, testSchema);
  assertEqual(slot3.type, "bool", "Packed field type");
  assertEqual(slot3.byteOffset, 0, "Packed field byte offset");
}

async function testErc7201BaseSlotCalculation(): Promise<void> {
  console.log("\n🧪 Testing ERC-7201 base slot calculation...");

  // Test with known namespace - LineaRollupYieldExtension
  // The expected slot can be verified using Solidity:
  // bytes32 slot = keccak256(abi.encode(uint256(keccak256("linea.storage.LineaRollupYieldExtension")) - 1)) & ~bytes32(uint256(0xff));
  const namespace = "linea.storage.LineaRollupYieldExtension";
  const slot = calculateErc7201BaseSlot(namespace);

  // Should be a valid 32-byte hex string
  assert(slot.startsWith("0x"), "Slot starts with 0x");
  assertEqual(slot.length, 66, "Slot is 66 chars (0x + 64 hex)");

  // Last byte should be 00 (masked)
  assert(slot.endsWith("00"), "Last byte is 00 (masked)");

  // Different namespaces should produce different slots
  const slot2 = calculateErc7201BaseSlot("linea.storage.Different");
  assert(slot !== slot2, "Different namespaces produce different slots");
}

main().catch((error) => {
  console.error("Test error:", error);
  process.exit(1);
});
