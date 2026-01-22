/**
 * Bytecode Verifier - TypeScript Types
 *
 * Types for the bytecode verification tool that compares deployed contracts
 * against local artifact files.
 */

export interface ChainConfig {
  chainId: number;
  rpcUrl: string;
  explorerUrl?: string;
}

export interface ContractConfig {
  name: string;
  chain: string;
  address: string;
  artifactFile: string;
  isProxy?: boolean;
  /**
   * Constructor arguments for contracts with immutables.
   * Can be provided as:
   * - Array of values (will be ABI-encoded using constructor ABI)
   * - Hex string of already-encoded args
   */
  constructorArgs?: unknown[] | string;
}

export interface ImmutableReference {
  start: number; // byte position in deployed bytecode
  length: number; // always 32 for immutables
}

export interface ImmutableDifference {
  position: number;
  length: number;
  localValue: string;
  remoteValue: string;
  possibleType: string | undefined;
}

export interface VerifierConfig {
  chains: Record<string, ChainConfig>;
  contracts: ContractConfig[];
}


export interface ArtifactJson {
  _format?: string;
  contractName: string;
  sourceName?: string;
  abi: AbiElement[];
  bytecode: string;
  deployedBytecode: string;
  linkReferences?: Record<string, unknown>;
  deployedLinkReferences?: Record<string, unknown>;
}

export interface AbiElement {
  type: string;
  name?: string;
  inputs?: AbiInput[];
  outputs?: AbiInput[];
  stateMutability?: string;
  anonymous?: boolean;
  indexed?: boolean;
}

export interface AbiInput {
  internalType: string;
  name: string;
  type: string;
  indexed?: boolean;
  components?: AbiInput[];
}

export type VerificationStatus = "pass" | "fail" | "warn" | "skip";

export interface BytecodeComparisonResult {
  status: VerificationStatus;
  message: string;
  localBytecodeLength: number | undefined;
  remoteBytecodeLength: number | undefined;
  matchPercentage: number | undefined;
  differences: BytecodeDifference[] | undefined;
  immutableDifferences: ImmutableDifference[] | undefined;
  onlyImmutablesDiffer: boolean | undefined;
}

export interface BytecodeDifference {
  position: number;
  localByte: string;
  remoteByte: string;
}

export interface AbiComparisonResult {
  status: VerificationStatus;
  message: string;
  localSelectors?: string[];
  remoteSelectors?: string[];
  missingSelectors?: string[];
  extraSelectors?: string[];
}

export interface ContractVerificationResult {
  contract: ContractConfig;
  chain: ChainConfig;
  bytecodeResult?: BytecodeComparisonResult;
  abiResult?: AbiComparisonResult;
  error?: string;
}

export interface VerificationSummary {
  total: number;
  passed: number;
  failed: number;
  warnings: number;
  skipped: number;
  results: ContractVerificationResult[];
}

export interface CliOptions {
  config: string;
  verbose: boolean;
  contract: string | undefined;
  chain: string | undefined;
  skipBytecode: boolean;
  skipAbi: boolean;
}
