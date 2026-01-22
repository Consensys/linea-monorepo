/**
 * Contract Integrity Verifier
 *
 * A tool to verify deployed smart contract integrity (bytecode, ABI, and state)
 * against local artifact files.
 *
 * @packageDocumentation
 */

// Core exports
export { loadConfig, checkArtifactExists } from "./config";
export { runVerification, printSummary } from "./verifier";

// Utility exports
export {
  compareBytecode,
  stripCborMetadata,
  extractSelectorsFromBytecode,
  validateImmutablesAgainstArgs,
} from "./utils/bytecode";

export {
  loadArtifact,
  detectArtifactFormat,
  extractSelectorsFromAbi,
  extractSelectorsFromArtifact,
  compareSelectors,
} from "./utils/abi";

export {
  calculateErc7201Slot,
  readStorageSlot,
  decodeSlotValue,
  executeViewCall,
  verifySlot,
  verifyNamespace,
  verifyState,
} from "./utils/state";

export {
  calculateErc7201BaseSlot,
  loadStorageSchema,
  parsePath,
  computeSlot,
  verifyStoragePath,
} from "./utils/storage-path";

export { parseMarkdownConfig, loadMarkdownConfig } from "./utils/markdown-config";

// Type exports
export type {
  // Config types
  VerifierConfig,
  ChainConfig,
  ContractConfig,
  CliOptions,
  // Verification result types
  VerificationSummary,
  ContractVerificationResult,
  BytecodeComparisonResult,
  AbiComparisonResult,
  StateVerificationResult,
  VerificationStatus,
  // State verification config types
  StateVerificationConfig,
  ViewCallConfig,
  SlotConfig,
  NamespaceConfig,
  NamespaceVariable,
  StoragePathConfig,
  // State verification result types
  ViewCallResult,
  SlotResult,
  NamespaceResult,
  StoragePathResult,
  // Artifact types
  ArtifactFormat,
  NormalizedArtifact,
  HardhatArtifact,
  FoundryArtifact,
  AbiElement,
  AbiInput,
  // Storage schema types
  StorageSchema,
  StorageStructDef,
  StorageFieldDef,
  SolidityType,
  // Bytecode types
  BytecodeDifference,
  ImmutableDifference,
  ImmutableReference,
} from "./types";
