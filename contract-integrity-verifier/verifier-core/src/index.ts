/**
 * @consensys/linea-contract-integrity-verifier
 *
 * A tool to verify deployed smart contract integrity (bytecode, ABI, and state)
 * against local artifact files.
 *
 * @packageDocumentation
 */

// Adapter interface (for implementers)
export type { Web3Adapter, Web3AdapterOptions } from "./adapter";

// Main Verifier class
export { Verifier, printSummary } from "./verifier";
export type { VerifyOptions } from "./verifier";

// Config loading
export { loadConfig, checkArtifactExists } from "./config";

// Pure utility exports (no adapter needed)
export {
  stripCborMetadata,
  compareBytecode,
  extractSelectorsFromBytecode,
  validateImmutablesAgainstArgs,
} from "./utils/bytecode";

// ABI utilities (require adapter for selector computation)
export {
  loadArtifact,
  detectArtifactFormat,
  extractSelectorsFromAbi,
  extractSelectorsFromArtifact,
  compareSelectors,
} from "./utils/abi";

// Storage utilities (require adapter for hashing and RPC)
export {
  calculateErc7201BaseSlot,
  calculateErc7201Slot,
  readStorageSlot,
  decodeSlotValue,
  verifySlot,
  verifyNamespace,
  verifyStoragePath,
  loadStorageSchema,
  parsePath,
  computeSlot,
} from "./utils/storage";

// Markdown config parsing
export { parseMarkdownConfig } from "./utils/markdown-config";

// All types
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
  SlotType,
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
