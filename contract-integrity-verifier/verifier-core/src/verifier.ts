/**
 * Contract Integrity Verifier - Core Verification Logic
 *
 * Main verification engine that fetches on-chain bytecode and compares
 * against local artifact files.
 */

import { EIP1967_IMPLEMENTATION_SLOT, BYTECODE_MATCH_THRESHOLD_PERCENT } from "./constants";
import {
  VerifierConfig,
  ContractConfig,
  ChainConfig,
  ContractVerificationResult,
  VerificationSummary,
  StateVerificationConfig,
  StateVerificationResult,
  ViewCallResult,
  AbiElement,
  NormalizedArtifact,
  StorageSchema,
} from "./types";
import { extractSelectorsFromArtifact, compareSelectors } from "./utils/abi";
import {
  compareBytecode,
  extractSelectorsFromBytecode,
  validateImmutablesAgainstArgs,
  verifyImmutableValues,
  definitiveCompareBytecode,
  groupImmutableDifferences,
  formatGroupedImmutables,
} from "./utils/bytecode";
// Browser-safe imports (no fs dependency)
import { formatValue, formatForDisplay, compareValues } from "./utils/comparison";
import { calculateErc7201BaseSlot, verifySlot, verifyNamespace, verifyStoragePath } from "./utils/storage";

import type { Web3Adapter } from "./adapter";

// Node.js-only imports loaded dynamically to avoid bundling 'fs' in browser builds
// These are only used in verifyContract() which is not called from browser code
type LoadArtifactFn = (filePath: string) => NormalizedArtifact;
type LoadStorageSchemaFn = (schemaPath: string, configDir: string) => StorageSchema;

let _loadArtifact: LoadArtifactFn | undefined;
let _loadStorageSchema: LoadStorageSchemaFn | undefined;

async function getLoadArtifact(): Promise<LoadArtifactFn> {
  if (!_loadArtifact) {
    // Dynamic import - will be excluded from browser bundle via tsup external config
    const mod = await import("./utils/abi-node.js");
    _loadArtifact = mod.loadArtifact;
  }
  return _loadArtifact;
}

async function getLoadStorageSchema(): Promise<LoadStorageSchemaFn> {
  if (!_loadStorageSchema) {
    // Dynamic import - will be excluded from browser bundle via tsup external config
    const mod = await import("./utils/storage-node.js");
    _loadStorageSchema = mod.loadStorageSchema;
  }
  return _loadStorageSchema;
}

/**
 * Options for verification operations.
 */
export interface VerifyOptions {
  verbose?: boolean;
  skipBytecode?: boolean;
  skipAbi?: boolean;
  skipState?: boolean;
  contractFilter?: string;
  chainFilter?: string;
}

/**
 * Content for browser-based verification.
 * Contains pre-loaded artifact and optional schema.
 */
export interface VerificationContent {
  /** Pre-loaded and parsed artifact */
  artifact: NormalizedArtifact;
  /** Pre-loaded and parsed storage schema (if needed for state verification) */
  schema?: StorageSchema;
}

/**
 * Main Verifier class.
 * Requires a Web3Adapter for blockchain interactions.
 */
export class Verifier {
  constructor(private readonly adapter: Web3Adapter) {}

  /**
   * Fetches bytecode from chain at a given address.
   */
  async fetchBytecode(address: string): Promise<string> {
    const bytecode = await this.adapter.getCode(address);
    if (bytecode === "0x" || bytecode === "") {
      throw new Error(`No bytecode found at address ${address}`);
    }
    return bytecode;
  }

  /**
   * Checks if a contract is an EIP-1967 proxy and returns the implementation address.
   */
  async getImplementationAddress(address: string): Promise<string | null> {
    try {
      const implementationSlot = await this.adapter.getStorageAt(address, EIP1967_IMPLEMENTATION_SLOT);
      const implementationAddress = this.adapter.checksumAddress("0x" + implementationSlot.slice(-40));

      if (implementationAddress === this.adapter.zeroAddress) {
        return null;
      }

      return implementationAddress;
    } catch {
      return null;
    }
  }

  /**
   * Calculates ERC-7201 base slot for a namespace.
   */
  calculateErc7201Slot(namespaceId: string): string {
    return calculateErc7201BaseSlot(this.adapter, namespaceId);
  }

  /**
   * Verifies a single contract.
   */
  async verifyContract(
    contract: ContractConfig,
    chain: ChainConfig,
    options: VerifyOptions = {},
    configDir: string = ".",
  ): Promise<ContractVerificationResult> {
    const result: ContractVerificationResult = {
      contract,
      chain,
    };

    try {
      // Load artifact (dynamic import to avoid fs in browser bundles)
      const loadArtifact = await getLoadArtifact();
      const artifact = loadArtifact(contract.artifactFile);

      if (options.verbose) {
        console.log(`  Artifact format: ${artifact.format}`);
        if (artifact.immutableReferences && artifact.immutableReferences.length > 0) {
          console.log(`  Known immutable positions: ${artifact.immutableReferences.length}`);
        }
      }

      // Fetch on-chain bytecode
      let remoteBytecode = await this.fetchBytecode(contract.address);
      let addressUsed = contract.address;

      // If marked as proxy, get implementation bytecode
      if (contract.isProxy) {
        const implAddress = await this.getImplementationAddress(contract.address);
        if (implAddress) {
          if (options.verbose) {
            console.log(`  Proxy detected, fetching implementation at ${implAddress}`);
          }
          remoteBytecode = await this.fetchBytecode(implAddress);
          addressUsed = implAddress;
        } else {
          console.warn(
            `  Warning: Contract marked as proxy but no EIP-1967 implementation found at ${contract.address}`,
          );
        }
      }

      // Bytecode comparison
      if (!options.skipBytecode) {
        result.bytecodeResult = compareBytecode(
          artifact.deployedBytecode,
          remoteBytecode,
          artifact.immutableReferences,
        );

        // Validate immutables against constructor args if provided
        if (
          result.bytecodeResult.onlyImmutablesDiffer &&
          result.bytecodeResult.immutableDifferences &&
          contract.constructorArgs
        ) {
          const validation = validateImmutablesAgainstArgs(
            result.bytecodeResult.immutableDifferences,
            contract.constructorArgs,
            options.verbose,
          );

          if (options.verbose && validation.details) {
            for (const detail of validation.details) {
              console.log(`    ${detail}`);
            }
          }

          if (validation.valid) {
            result.bytecodeResult.message += ` - constructor args validated`;
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${validation.message}`;
          }
        }

        // Verify named immutable values if provided
        // Run this even when onlyImmutablesDiffer is false, as the heuristic can fail
        // with many difference regions (e.g., split addresses with matching bytes)
        if (
          result.bytecodeResult.immutableDifferences &&
          result.bytecodeResult.immutableDifferences.length > 0 &&
          contract.immutableValues &&
          Object.keys(contract.immutableValues).length > 0
        ) {
          result.immutableValuesResult = verifyImmutableValues(
            contract.immutableValues,
            result.bytecodeResult.immutableDifferences,
          );

          if (options.verbose && result.immutableValuesResult.results) {
            for (const immResult of result.immutableValuesResult.results) {
              const icon = immResult.status === "pass" ? "✓" : "✗";
              console.log(`    ${icon} ${immResult.message}`);
            }
          }

          // Update bytecode status based on immutable values verification
          if (result.immutableValuesResult.status === "pass") {
            // If all named immutables verified, upgrade bytecode status to pass
            // This handles cases where heuristic fails but immutables are valid
            if (result.bytecodeResult.status === "fail" && result.bytecodeResult.matchPercentage !== undefined) {
              // Only upgrade if high match percentage suggests immutable-only differences
              if (result.bytecodeResult.matchPercentage >= BYTECODE_MATCH_THRESHOLD_PERCENT) {
                result.bytecodeResult.status = "pass";
                result.bytecodeResult.message = `Bytecode matches (${result.bytecodeResult.immutableDifferences.length} immutable region(s) verified by name)`;
                result.bytecodeResult.onlyImmutablesDiffer = true;
              }
            }
            // If immutables account for all differences, effective match is 100%
            if (result.bytecodeResult.onlyImmutablesDiffer) {
              result.bytecodeResult.matchPercentage = 100;
            }
            result.bytecodeResult.message += ` - immutable values verified`;
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${result.immutableValuesResult.message}`;
          }
        }

        // Definitive bytecode comparison (100% confidence, no ambiguity)
        // Requires Foundry artifact with immutableReferences
        if (
          artifact.immutableReferences &&
          artifact.immutableReferences.length > 0 &&
          result.bytecodeResult.immutableDifferences &&
          result.bytecodeResult.immutableDifferences.length > 0
        ) {
          result.definitiveResult = definitiveCompareBytecode(
            artifact.deployedBytecode,
            remoteBytecode,
            artifact.immutableReferences,
            result.bytecodeResult.immutableDifferences,
          );

          // Group immutable differences by their parent reference (handles fragmented immutables)
          result.groupedImmutables = groupImmutableDifferences(
            result.bytecodeResult.immutableDifferences,
            artifact.immutableReferences,
            remoteBytecode,
          );

          if (options.verbose) {
            const icon = result.definitiveResult.exactMatch ? "✓" : "✗";
            console.log(`    ${icon} Definitive: ${result.definitiveResult.message}`);

            // Show grouped immutables with fragment information
            if (result.groupedImmutables.length > 0) {
              const fragmentedCount = result.groupedImmutables.filter((g) => g.isFragmented).length;
              if (fragmentedCount > 0) {
                console.log(`    Note: ${fragmentedCount} immutable(s) are fragmented due to matching bytes`);
              }
              const formattedLines = formatGroupedImmutables(result.groupedImmutables);
              for (const line of formattedLines) {
                console.log(`      ${line}`);
              }
            }
          }

          // If definitive check passes, we have 100% confidence on bytecode structure
          if (result.definitiveResult.exactMatch) {
            const fragmentedCount = result.groupedImmutables.filter((g) => g.isFragmented).length;
            const fragmentNote = fragmentedCount > 0 ? ` (${fragmentedCount} fragmented)` : "";

            // Check if user-provided immutable values failed verification
            // If so, keep warn status - bytecode structure is correct but values don't match expectations
            if (result.immutableValuesResult?.status === "fail") {
              result.bytecodeResult.status = "warn";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode structure matches (${result.definitiveResult.immutablesSubstituted} immutable(s)${fragmentNote}) - ${result.immutableValuesResult.message}`;
            } else {
              result.bytecodeResult.status = "pass";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode matches exactly (${result.definitiveResult.immutablesSubstituted} immutable(s) verified${fragmentNote})`;
            }
          } else if (result.definitiveResult.status === "fail") {
            // Definitive check failed - this is a critical failure
            result.bytecodeResult.status = "fail";
            result.bytecodeResult.message = result.definitiveResult.message;
          }
        }
      }

      // ABI comparison
      if (!options.skipAbi) {
        const abiSelectors = extractSelectorsFromArtifact(this.adapter, artifact);
        const bytecodeSelectors = extractSelectorsFromBytecode(remoteBytecode);
        result.abiResult = compareSelectors(abiSelectors, bytecodeSelectors);
      }

      // State verification
      if (!options.skipState && contract.stateVerification) {
        result.stateResult = await this.verifyState(
          contract.address,
          artifact.abi,
          contract.stateVerification,
          configDir,
        );
      }

      if (options.verbose) {
        console.log(`  Address verified: ${addressUsed}`);
        console.log(`  Remote bytecode length: ${(remoteBytecode.length - 2) / 2} bytes`);
      }
    } catch (error) {
      result.error = error instanceof Error ? error.message : String(error);
    }

    return result;
  }

  /**
   * Verifies a single contract using pre-loaded content.
   * Browser-compatible - does not use filesystem.
   *
   * @param contract - Contract configuration
   * @param chain - Chain configuration
   * @param options - Verification options
   * @param content - Pre-loaded artifact and schema content
   */
  async verifyContractWithContent(
    contract: ContractConfig,
    chain: ChainConfig,
    options: VerifyOptions = {},
    content: VerificationContent,
  ): Promise<ContractVerificationResult> {
    const result: ContractVerificationResult = {
      contract,
      chain,
    };

    try {
      const artifact = content.artifact;

      if (options.verbose) {
        console.log(`  Artifact format: ${artifact.format}`);
        if (artifact.immutableReferences && artifact.immutableReferences.length > 0) {
          console.log(`  Known immutable positions: ${artifact.immutableReferences.length}`);
        }
      }

      // Fetch on-chain bytecode
      let remoteBytecode = await this.fetchBytecode(contract.address);
      let addressUsed = contract.address;

      // If marked as proxy, get implementation bytecode
      if (contract.isProxy) {
        const implAddress = await this.getImplementationAddress(contract.address);
        if (implAddress) {
          if (options.verbose) {
            console.log(`  Proxy detected, fetching implementation at ${implAddress}`);
          }
          remoteBytecode = await this.fetchBytecode(implAddress);
          addressUsed = implAddress;
        } else {
          console.warn(
            `  Warning: Contract marked as proxy but no EIP-1967 implementation found at ${contract.address}`,
          );
        }
      }

      // Bytecode comparison
      if (!options.skipBytecode) {
        result.bytecodeResult = compareBytecode(
          artifact.deployedBytecode,
          remoteBytecode,
          artifact.immutableReferences,
        );

        // Validate immutables against constructor args if provided
        if (
          result.bytecodeResult.onlyImmutablesDiffer &&
          result.bytecodeResult.immutableDifferences &&
          contract.constructorArgs
        ) {
          const validation = validateImmutablesAgainstArgs(
            result.bytecodeResult.immutableDifferences,
            contract.constructorArgs,
            options.verbose,
          );

          if (options.verbose && validation.details) {
            for (const detail of validation.details) {
              console.log(`    ${detail}`);
            }
          }

          if (validation.valid) {
            result.bytecodeResult.message += ` - constructor args validated`;
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${validation.message}`;
          }
        }

        // Verify named immutable values if provided
        if (
          result.bytecodeResult.immutableDifferences &&
          result.bytecodeResult.immutableDifferences.length > 0 &&
          contract.immutableValues &&
          Object.keys(contract.immutableValues).length > 0
        ) {
          result.immutableValuesResult = verifyImmutableValues(
            contract.immutableValues,
            result.bytecodeResult.immutableDifferences,
          );

          if (options.verbose && result.immutableValuesResult.results) {
            for (const immResult of result.immutableValuesResult.results) {
              const icon = immResult.status === "pass" ? "✓" : "✗";
              console.log(`    ${icon} ${immResult.message}`);
            }
          }

          // Update bytecode status based on immutable values verification
          if (result.immutableValuesResult.status === "pass") {
            if (result.bytecodeResult.status === "fail" && result.bytecodeResult.matchPercentage !== undefined) {
              // Only upgrade if high match percentage suggests immutable-only differences
              if (result.bytecodeResult.matchPercentage >= BYTECODE_MATCH_THRESHOLD_PERCENT) {
                result.bytecodeResult.status = "pass";
                result.bytecodeResult.message = `Bytecode matches (${result.bytecodeResult.immutableDifferences.length} immutable region(s) verified by name)`;
                result.bytecodeResult.onlyImmutablesDiffer = true;
              }
            }
            if (result.bytecodeResult.onlyImmutablesDiffer) {
              result.bytecodeResult.matchPercentage = 100;
            }
            result.bytecodeResult.message += ` - immutable values verified`;
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${result.immutableValuesResult.message}`;
          }
        }

        // Definitive bytecode comparison
        if (
          artifact.immutableReferences &&
          artifact.immutableReferences.length > 0 &&
          result.bytecodeResult.immutableDifferences &&
          result.bytecodeResult.immutableDifferences.length > 0
        ) {
          result.definitiveResult = definitiveCompareBytecode(
            artifact.deployedBytecode,
            remoteBytecode,
            artifact.immutableReferences,
            result.bytecodeResult.immutableDifferences,
          );

          result.groupedImmutables = groupImmutableDifferences(
            result.bytecodeResult.immutableDifferences,
            artifact.immutableReferences,
            remoteBytecode,
          );

          if (options.verbose) {
            const icon = result.definitiveResult.exactMatch ? "✓" : "✗";
            console.log(`    ${icon} Definitive: ${result.definitiveResult.message}`);

            if (result.groupedImmutables.length > 0) {
              const fragmentedCount = result.groupedImmutables.filter((g) => g.isFragmented).length;
              if (fragmentedCount > 0) {
                console.log(`    Note: ${fragmentedCount} immutable(s) are fragmented due to matching bytes`);
              }
              const formattedLines = formatGroupedImmutables(result.groupedImmutables);
              for (const line of formattedLines) {
                console.log(`      ${line}`);
              }
            }
          }

          // If definitive check passes, we have 100% confidence on bytecode structure
          if (result.definitiveResult.exactMatch) {
            const fragmentedCount = result.groupedImmutables.filter((g) => g.isFragmented).length;
            const fragmentNote = fragmentedCount > 0 ? ` (${fragmentedCount} fragmented)` : "";

            // Check if user-provided immutable values failed verification
            // If so, keep warn status - bytecode structure is correct but values don't match expectations
            if (result.immutableValuesResult?.status === "fail") {
              result.bytecodeResult.status = "warn";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode structure matches (${result.definitiveResult.immutablesSubstituted} immutable(s)${fragmentNote}) - ${result.immutableValuesResult.message}`;
            } else {
              result.bytecodeResult.status = "pass";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode matches exactly (${result.definitiveResult.immutablesSubstituted} immutable(s) verified${fragmentNote})`;
            }
          } else if (result.definitiveResult.status === "fail") {
            result.bytecodeResult.status = "fail";
            result.bytecodeResult.message = result.definitiveResult.message;
          }
        }
      }

      // ABI comparison
      if (!options.skipAbi) {
        const abiSelectors = extractSelectorsFromArtifact(this.adapter, artifact);
        const bytecodeSelectors = extractSelectorsFromBytecode(remoteBytecode);
        result.abiResult = compareSelectors(abiSelectors, bytecodeSelectors);
      }

      // State verification with pre-loaded schema
      if (!options.skipState && contract.stateVerification) {
        result.stateResult = await this.verifyStateWithContent(
          contract.address,
          artifact.abi,
          contract.stateVerification,
          content.schema,
        );
      }

      if (options.verbose) {
        console.log(`  Address verified: ${addressUsed}`);
        console.log(`  Remote bytecode length: ${(remoteBytecode.length - 2) / 2} bytes`);
      }
    } catch (error) {
      result.error = error instanceof Error ? error.message : String(error);
    }

    return result;
  }

  /**
   * Performs complete state verification for a contract.
   * All verification types run in parallel for efficiency.
   */
  async verifyState(
    address: string,
    abi: AbiElement[],
    config: StateVerificationConfig,
    configDir: string = ".",
  ): Promise<StateVerificationResult> {
    // Warn if storage paths are configured but schemaFile is missing
    const storagePathsSkipped = config.storagePaths && config.storagePaths.length > 0 && !config.schemaFile;

    // Run all verification types in parallel for efficiency
    const [viewCallResults, namespaceResults, slotResults, storagePathResults] = await Promise.all([
      // 1. Execute view calls in parallel
      config.viewCalls && config.viewCalls.length > 0
        ? Promise.all(config.viewCalls.map((viewCall) => this.executeViewCall(address, abi, viewCall)))
        : Promise.resolve([]),

      // 2. Verify namespaces (ERC-7201) in parallel
      config.namespaces && config.namespaces.length > 0
        ? Promise.all(config.namespaces.map((namespace) => verifyNamespace(this.adapter, address, namespace)))
        : Promise.resolve([]),

      // 3. Verify explicit slots in parallel
      config.slots && config.slots.length > 0
        ? Promise.all(config.slots.map((slot) => verifySlot(this.adapter, address, slot)))
        : Promise.resolve([]),

      // 4. Verify storage paths (schema-based) in parallel
      config.storagePaths && config.storagePaths.length > 0 && config.schemaFile && configDir
        ? (async () => {
            const loadStorageSchema = await getLoadStorageSchema();
            const schema = loadStorageSchema(config.schemaFile!, configDir);
            return Promise.all(
              config.storagePaths!.map((pathConfig) => verifyStoragePath(this.adapter, address, pathConfig, schema)),
            );
          })()
        : Promise.resolve([]),
    ]);

    // Aggregate results
    const allViewCallsPass = viewCallResults.every((r) => r.status === "pass");
    const allNamespacesPass = namespaceResults.every((r) => r.status === "pass");
    const allSlotsPass = slotResults.every((r) => r.status === "pass");
    const allStoragePathsPass = storagePathResults.every((r) => r.status === "pass");

    const totalChecks =
      viewCallResults.length + namespaceResults.length + slotResults.length + storagePathResults.length;
    const passedChecks =
      viewCallResults.filter((r) => r.status === "pass").length +
      namespaceResults.filter((r) => r.status === "pass").length +
      slotResults.filter((r) => r.status === "pass").length +
      storagePathResults.filter((r) => r.status === "pass").length;

    const allPass =
      allViewCallsPass && allNamespacesPass && allSlotsPass && allStoragePathsPass && !storagePathsSkipped;

    // Build message with optional warning about skipped storage paths
    let message: string;
    if (storagePathsSkipped) {
      const skippedCount = config.storagePaths!.length;
      message =
        allViewCallsPass && allNamespacesPass && allSlotsPass
          ? `${totalChecks} state checks passed, but ${skippedCount} storage path(s) SKIPPED (schemaFile missing)`
          : `${passedChecks}/${totalChecks} state checks passed, ${skippedCount} storage path(s) SKIPPED (schemaFile missing)`;
    } else {
      message = allPass
        ? `All ${totalChecks} state checks passed`
        : `${passedChecks}/${totalChecks} state checks passed`;
    }

    return {
      status: storagePathsSkipped ? "warn" : allPass ? "pass" : "fail",
      message,
      viewCallResults: viewCallResults.length > 0 ? viewCallResults : undefined,
      namespaceResults: namespaceResults.length > 0 ? namespaceResults : undefined,
      slotResults: slotResults.length > 0 ? slotResults : undefined,
      storagePathResults: storagePathResults.length > 0 ? storagePathResults : undefined,
    };
  }

  /**
   * Performs state verification using pre-loaded schema.
   * Browser-compatible - does not use filesystem.
   *
   * @param address - Contract address
   * @param abi - Contract ABI
   * @param config - State verification configuration
   * @param schema - Pre-loaded storage schema (optional, required if storagePaths are used)
   */
  async verifyStateWithContent(
    address: string,
    abi: AbiElement[],
    config: StateVerificationConfig,
    schema?: StorageSchema,
  ): Promise<StateVerificationResult> {
    // Warn if storage paths are configured but schema is missing
    const storagePathsSkipped = config.storagePaths && config.storagePaths.length > 0 && !schema;

    // Run all verification types in parallel for efficiency
    const [viewCallResults, namespaceResults, slotResults, storagePathResults] = await Promise.all([
      // 1. Execute view calls in parallel
      config.viewCalls && config.viewCalls.length > 0
        ? Promise.all(config.viewCalls.map((viewCall) => this.executeViewCall(address, abi, viewCall)))
        : Promise.resolve([]),

      // 2. Verify namespaces (ERC-7201) in parallel
      config.namespaces && config.namespaces.length > 0
        ? Promise.all(config.namespaces.map((namespace) => verifyNamespace(this.adapter, address, namespace)))
        : Promise.resolve([]),

      // 3. Verify explicit slots in parallel
      config.slots && config.slots.length > 0
        ? Promise.all(config.slots.map((slot) => verifySlot(this.adapter, address, slot)))
        : Promise.resolve([]),

      // 4. Verify storage paths (schema-based) in parallel - using pre-loaded schema
      config.storagePaths && config.storagePaths.length > 0 && schema
        ? Promise.all(
            config.storagePaths.map((pathConfig) => verifyStoragePath(this.adapter, address, pathConfig, schema)),
          )
        : Promise.resolve([]),
    ]);

    // Aggregate results
    const allViewCallsPass = viewCallResults.every((r) => r.status === "pass");
    const allNamespacesPass = namespaceResults.every((r) => r.status === "pass");
    const allSlotsPass = slotResults.every((r) => r.status === "pass");
    const allStoragePathsPass = storagePathResults.every((r) => r.status === "pass");

    const totalChecks =
      viewCallResults.length + namespaceResults.length + slotResults.length + storagePathResults.length;
    const passedChecks =
      viewCallResults.filter((r) => r.status === "pass").length +
      namespaceResults.filter((r) => r.status === "pass").length +
      slotResults.filter((r) => r.status === "pass").length +
      storagePathResults.filter((r) => r.status === "pass").length;

    const allPass =
      allViewCallsPass && allNamespacesPass && allSlotsPass && allStoragePathsPass && !storagePathsSkipped;

    // Build message with optional warning about skipped storage paths
    let message: string;
    if (storagePathsSkipped) {
      const skippedCount = config.storagePaths!.length;
      message =
        allViewCallsPass && allNamespacesPass && allSlotsPass
          ? `${totalChecks} state checks passed, but ${skippedCount} storage path(s) SKIPPED (schema missing)`
          : `${passedChecks}/${totalChecks} state checks passed, ${skippedCount} storage path(s) SKIPPED (schema missing)`;
    } else {
      message = allPass
        ? `All ${totalChecks} state checks passed`
        : `${passedChecks}/${totalChecks} state checks passed`;
    }

    return {
      status: storagePathsSkipped ? "warn" : allPass ? "pass" : "fail",
      message,
      viewCallResults: viewCallResults.length > 0 ? viewCallResults : undefined,
      namespaceResults: namespaceResults.length > 0 ? namespaceResults : undefined,
      slotResults: slotResults.length > 0 ? slotResults : undefined,
      storagePathResults: storagePathResults.length > 0 ? storagePathResults : undefined,
    };
  }

  /**
   * Executes a view function call and returns the result.
   */
  async executeViewCall(
    address: string,
    abi: AbiElement[],
    config: import("./types").ViewCallConfig,
  ): Promise<ViewCallResult> {
    try {
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

      const calldata = this.adapter.encodeFunctionData(abi, config.function, config.params ?? []);
      const result = await this.adapter.call(address, calldata);
      const decoded = this.adapter.decodeFunctionResult(abi, config.function, result);

      const actual = decoded.length === 1 ? formatValue(decoded[0]) : decoded.map(formatValue);

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
   * Runs verification for all contracts in a configuration.
   */
  async verify(
    config: VerifierConfig,
    options: VerifyOptions = {},
    configDir: string = ".",
  ): Promise<VerificationSummary> {
    const results: ContractVerificationResult[] = [];
    let passed = 0;
    let failed = 0;
    let warnings = 0;
    let skipped = 0;

    // Filter contracts if specified
    let contractsToVerify = config.contracts;
    if (options.contractFilter) {
      contractsToVerify = contractsToVerify.filter(
        (c) => c.name.toLowerCase() === options.contractFilter!.toLowerCase(),
      );
    }
    if (options.chainFilter) {
      contractsToVerify = contractsToVerify.filter((c) => c.chain.toLowerCase() === options.chainFilter!.toLowerCase());
    }

    if (contractsToVerify.length === 0) {
      console.log("No contracts to verify matching the specified filters.");
      return { total: 0, passed: 0, failed: 0, warnings: 0, skipped: 0, results: [] };
    }

    console.log(`\nVerifying ${contractsToVerify.length} contract(s)...\n`);

    for (const contract of contractsToVerify) {
      const chain = config.chains[contract.chain];
      if (!chain) {
        console.log(`Skipping ${contract.name}: unknown chain '${contract.chain}'`);
        skipped++;
        continue;
      }

      console.log(`Verifying ${contract.name} on ${contract.chain}...`);
      console.log(`  Address: ${contract.address}`);

      const result = await this.verifyContract(contract, chain, options, configDir);
      results.push(result);

      if (result.error) {
        console.log(`  ERROR: ${result.error}`);
        failed++;
      } else {
        // Print results
        if (result.bytecodeResult) {
          const br = result.bytecodeResult;
          const icon = br.status === "pass" ? "✓" : br.status === "fail" ? "✗" : br.status === "warn" ? "!" : "-";
          console.log(`  Bytecode: ${icon} ${br.message}`);
        }

        if (result.abiResult) {
          const ar = result.abiResult;
          const icon = ar.status === "pass" ? "✓" : ar.status === "fail" ? "✗" : ar.status === "warn" ? "!" : "-";
          console.log(`  ABI: ${icon} ${ar.message}`);
        }

        if (result.stateResult) {
          const sr = result.stateResult;
          const icon = sr.status === "pass" ? "✓" : sr.status === "fail" ? "✗" : sr.status === "warn" ? "!" : "-";
          console.log(`  State: ${icon} ${sr.message}`);
        }

        if (result.immutableValuesResult) {
          const ivr = result.immutableValuesResult;
          const icon = ivr.status === "pass" ? "✓" : ivr.status === "fail" ? "✗" : ivr.status === "warn" ? "!" : "-";
          console.log(`  Immutables: ${icon} ${ivr.message}`);
        }

        if (result.definitiveResult) {
          const dr = result.definitiveResult;
          const icon = dr.exactMatch ? "✓" : "✗";
          console.log(`  Definitive: ${icon} ${dr.message}`);
        }

        // Count results
        const bytecodeStatus = result.bytecodeResult?.status;
        const abiStatus = result.abiResult?.status;
        const stateStatus = result.stateResult?.status;
        const immutableValuesStatus = result.immutableValuesResult?.status;
        const definitiveStatus = result.definitiveResult?.status;

        if (
          bytecodeStatus === "fail" ||
          abiStatus === "fail" ||
          stateStatus === "fail" ||
          immutableValuesStatus === "fail" ||
          definitiveStatus === "fail"
        ) {
          failed++;
        } else if (
          bytecodeStatus === "warn" ||
          abiStatus === "warn" ||
          stateStatus === "warn" ||
          immutableValuesStatus === "warn"
        ) {
          warnings++;
        } else {
          const hasBytecodeResult = result.bytecodeResult !== undefined;
          const hasAbiResult = result.abiResult !== undefined;
          const hasStateResult = result.stateResult !== undefined;
          const hasAnyVerification = hasBytecodeResult || hasAbiResult || hasStateResult;

          if (!hasAnyVerification) {
            skipped++;
          } else if (
            (bytecodeStatus === "skip" || !hasBytecodeResult) &&
            (abiStatus === "skip" || !hasAbiResult) &&
            !hasStateResult
          ) {
            skipped++;
          } else {
            passed++;
          }
        }
      }

      console.log("");
    }

    return {
      total: contractsToVerify.length,
      passed,
      failed,
      warnings,
      skipped,
      results,
    };
  }
}

/**
 * Prints a summary of verification results.
 */
export function printSummary(summary: VerificationSummary): void {
  console.log("=".repeat(50));
  console.log("VERIFICATION SUMMARY");
  console.log("=".repeat(50));
  console.log(`Total contracts: ${summary.total}`);
  console.log(`  Passed:   ${summary.passed}`);
  console.log(`  Failed:   ${summary.failed}`);
  console.log(`  Warnings: ${summary.warnings}`);
  console.log(`  Skipped:  ${summary.skipped}`);
  console.log("=".repeat(50));

  if (summary.failed > 0) {
    console.log("\nFailed contracts:");
    for (const result of summary.results) {
      const hasBytecodeFailure = result.bytecodeResult?.status === "fail";
      const hasAbiFailure = result.abiResult?.status === "fail";
      const hasStateFailure = result.stateResult?.status === "fail";
      const hasImmutableValuesFailure = result.immutableValuesResult?.status === "fail";
      const hasDefinitiveFailure = result.definitiveResult?.status === "fail";
      const hasError = result.error;

      if (
        hasBytecodeFailure ||
        hasAbiFailure ||
        hasStateFailure ||
        hasImmutableValuesFailure ||
        hasDefinitiveFailure ||
        hasError
      ) {
        console.log(`  - ${result.contract.name} (${result.contract.chain})`);
        if (hasError) {
          console.log(`    Error: ${result.error}`);
        }
        if (hasBytecodeFailure) {
          console.log(`    Bytecode: ${result.bytecodeResult!.message}`);
        }
        if (hasAbiFailure) {
          console.log(`    ABI: ${result.abiResult!.message}`);
        }
        if (hasStateFailure) {
          console.log(`    State: ${result.stateResult!.message}`);
        }
        if (hasImmutableValuesFailure) {
          console.log(`    Immutables: ${result.immutableValuesResult!.message}`);
          for (const immResult of result.immutableValuesResult!.results) {
            if (immResult.status === "fail") {
              console.log(`      - ${immResult.name}: ${immResult.message}`);
            }
          }
        }
        if (hasDefinitiveFailure) {
          console.log(`    Definitive: ${result.definitiveResult!.message}`);
        }
      }
    }
  }
}
