/**
 * Contract Integrity Verifier - Browser-Safe Verifier
 *
 * This module contains only browser-compatible verification methods.
 * It does NOT include methods that depend on Node.js 'fs' module.
 *
 * For Node.js usage with file-based verification, use the main Verifier from ./verifier.
 */

import { EIP1967_IMPLEMENTATION_SLOT, BYTECODE_MATCH_THRESHOLD_PERCENT } from "./constants";
import type { Web3Adapter } from "./adapter";
import {
  ContractConfig,
  ChainConfig,
  ContractVerificationResult,
  StateVerificationConfig,
  StateVerificationResult,
  ViewCallResult,
  AbiElement,
  NormalizedArtifact,
  StorageSchema,
} from "./types";
import {
  compareBytecode,
  extractSelectorsFromBytecode,
  validateImmutablesAgainstArgs,
  verifyImmutableValues,
  definitiveCompareBytecode,
  groupImmutableDifferences,
  formatGroupedImmutables,
} from "./utils/bytecode";
import { extractSelectorsFromArtifact, compareSelectors } from "./utils/abi";
import { calculateErc7201BaseSlot, verifySlot, verifyNamespace, verifyStoragePath } from "./utils/storage";
import { formatValue, formatForDisplay, compareValues } from "./utils/comparison";

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
 * Browser-safe Verifier class.
 * Requires a Web3Adapter for blockchain interactions.
 *
 * This version only includes methods that don't require Node.js 'fs' module.
 * Use verifyContractWithContent() instead of verifyContract() for browser usage.
 */
export class BrowserVerifier {
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
   * Verifies a contract using pre-loaded content (browser-compatible).
   * Use this method in browser environments where artifacts are loaded via fetch/upload.
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

      // Bytecode verification
      if (!options.skipBytecode) {
        result.bytecodeResult = compareBytecode(
          artifact.deployedBytecode,
          remoteBytecode,
          artifact.immutableReferences,
        );

        // Validate immutable differences against constructor args if available
        if (
          result.bytecodeResult.onlyImmutablesDiffer &&
          result.bytecodeResult.immutableDifferences &&
          contract.constructorArgs
        ) {
          const immutableValidation = validateImmutablesAgainstArgs(
            result.bytecodeResult.immutableDifferences,
            contract.constructorArgs,
            options.verbose,
          );

          if (options.verbose && immutableValidation.details) {
            for (const detail of immutableValidation.details) {
              console.log(`    ${detail}`);
            }
          }

          if (immutableValidation.valid) {
            result.bytecodeResult.message += " - constructor args validated";
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${immutableValidation.message}`;
          }
        }

        // Named immutable values verification
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
            for (const r of result.immutableValuesResult.results) {
              const icon = r.status === "pass" ? "✓" : "✗";
              console.log(`    ${icon} ${r.message}`);
            }
          }

          if (result.immutableValuesResult.status === "pass") {
            // Upgrade status if immutable values were verified
            if (
              result.bytecodeResult.status === "fail" &&
              result.bytecodeResult.matchPercentage !== undefined &&
              result.bytecodeResult.matchPercentage >= BYTECODE_MATCH_THRESHOLD_PERCENT
            ) {
              result.bytecodeResult.status = "pass";
              result.bytecodeResult.message = `Bytecode matches (${result.bytecodeResult.immutableDifferences.length} immutable region(s) verified by name)`;
              result.bytecodeResult.onlyImmutablesDiffer = true;
            }

            if (result.bytecodeResult.onlyImmutablesDiffer) {
              result.bytecodeResult.matchPercentage = 100;
            }

            result.bytecodeResult.message += " - immutable values verified";
          } else {
            result.bytecodeResult.status = "warn";
            result.bytecodeResult.message += ` - ${result.immutableValuesResult.message}`;
          }
        }

        // Definitive bytecode comparison using known immutable positions
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
              const formatted = formatGroupedImmutables(result.groupedImmutables);
              for (const line of formatted) {
                console.log(`      ${line}`);
              }
            }
          }

          if (result.definitiveResult.exactMatch) {
            const fragmentedCount = result.groupedImmutables.filter((g) => g.isFragmented).length;
            const fragmentedNote = fragmentedCount > 0 ? ` (${fragmentedCount} fragmented)` : "";

            if (result.immutableValuesResult?.status === "fail") {
              result.bytecodeResult.status = "warn";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode structure matches (${result.definitiveResult.immutablesSubstituted} immutable(s)${fragmentedNote}) - ${result.immutableValuesResult.message}`;
            } else {
              result.bytecodeResult.status = "pass";
              result.bytecodeResult.matchPercentage = 100;
              result.bytecodeResult.message = `Bytecode matches exactly (${result.definitiveResult.immutablesSubstituted} immutable(s) verified${fragmentedNote})`;
            }
          } else if (result.definitiveResult.status === "fail") {
            result.bytecodeResult.status = "fail";
            result.bytecodeResult.message = result.definitiveResult.message;
          }
        }
      }

      // ABI verification
      if (!options.skipAbi) {
        const abiSelectors = extractSelectorsFromArtifact(this.adapter, artifact);
        const bytecodeSelectors = extractSelectorsFromBytecode(remoteBytecode);
        result.abiResult = compareSelectors(abiSelectors, bytecodeSelectors);
      }

      // State verification (browser-compatible version)
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
    } catch (err) {
      result.error = err instanceof Error ? err.message : String(err);
    }

    return result;
  }

  /**
   * Verifies state using pre-loaded schema (browser-compatible).
   */
  async verifyStateWithContent(
    address: string,
    abi: AbiElement[],
    config: StateVerificationConfig,
    schema?: StorageSchema,
  ): Promise<StateVerificationResult> {
    // Warn if storage paths are configured but schema is missing
    const storagePathsSkipped = config.storagePaths && config.storagePaths.length > 0 && !schema;

    const [viewCallResults, namespaceResults, slotResults, storagePathResults] = await Promise.all([
      // 1. Verify view calls in parallel
      config.viewCalls && config.viewCalls.length > 0
        ? Promise.all(config.viewCalls.map((call) => this.executeViewCall(address, abi, call)))
        : Promise.resolve([]),

      // 2. Verify ERC-7201 namespaces in parallel
      config.namespaces && config.namespaces.length > 0
        ? Promise.all(config.namespaces.map((ns) => verifyNamespace(this.adapter, address, ns)))
        : Promise.resolve([]),

      // 3. Verify explicit slots in parallel
      config.slots && config.slots.length > 0
        ? Promise.all(config.slots.map((slot) => verifySlot(this.adapter, address, slot)))
        : Promise.resolve([]),

      // 4. Verify storage paths (schema-based) in parallel
      config.storagePaths && config.storagePaths.length > 0 && schema
        ? Promise.all(config.storagePaths.map((pathConfig) => verifyStoragePath(this.adapter, address, pathConfig, schema)))
        : Promise.resolve([]),
    ]);

    // Aggregate results
    const allViewCallsPass = viewCallResults.every((r) => r.status === "pass");
    const allNamespacesPass = namespaceResults.every((r) => r.status === "pass");
    const allSlotsPass = slotResults.every((r) => r.status === "pass");
    const allStoragePathsPass = storagePathResults.every((r) => r.status === "pass");

    const totalChecks = viewCallResults.length + namespaceResults.length + slotResults.length + storagePathResults.length;
    const passedChecks =
      viewCallResults.filter((r) => r.status === "pass").length +
      namespaceResults.filter((r) => r.status === "pass").length +
      slotResults.filter((r) => r.status === "pass").length +
      storagePathResults.filter((r) => r.status === "pass").length;

    const allPass = allViewCallsPass && allNamespacesPass && allSlotsPass && allStoragePathsPass && !storagePathsSkipped;

    // Build message with optional warning about skipped storage paths
    let message: string;
    if (storagePathsSkipped) {
      const skippedCount = config.storagePaths!.length;
      message = allViewCallsPass && allNamespacesPass && allSlotsPass
        ? `${totalChecks} state checks passed, but ${skippedCount} storage path(s) SKIPPED (schema missing)`
        : `${passedChecks}/${totalChecks} state checks passed, ${skippedCount} storage path(s) SKIPPED (schema missing)`;
    } else {
      message = allPass
        ? `All ${totalChecks} state checks passed`
        : `${passedChecks}/${totalChecks} state checks passed`;
    }

    return {
      status: storagePathsSkipped ? "warn" : (allPass ? "pass" : "fail"),
      message,
      viewCallResults: viewCallResults.length > 0 ? viewCallResults : undefined,
      namespaceResults: namespaceResults.length > 0 ? namespaceResults : undefined,
      slotResults: slotResults.length > 0 ? slotResults : undefined,
      storagePathResults: storagePathResults.length > 0 ? storagePathResults : undefined,
    };
  }

  /**
   * Executes a view call and compares result with expected value.
   */
  private async executeViewCall(
    address: string,
    abi: AbiElement[],
    config: { function: string; params?: unknown[]; expected: unknown; comparison?: string },
  ): Promise<ViewCallResult> {
    try {
      // Find the function in ABI
      const abiFunc = abi.find((e) => e.type === "function" && e.name === config.function);
      if (!abiFunc) {
        return {
          function: config.function,
          params: config.params,
          expected: config.expected,
          actual: undefined,
          status: "fail",
          message: `Function '${config.function}' not found in ABI`,
        };
      }

      // Encode and execute call
      const callData = this.adapter.encodeFunctionData(abi, config.function, config.params ?? []);
      const result = await this.adapter.call(address, callData);
      const decoded = this.adapter.decodeFunctionResult(abi, config.function, result);

      // Format for comparison
      const actual = decoded.length === 1 ? formatValue(decoded[0]) : decoded.map(formatValue);
      const comparison = (config.comparison ?? "eq") as "eq" | "gt" | "gte" | "lt" | "lte" | "contains";
      const matches = compareValues(actual, config.expected, comparison);

      return {
        function: config.function,
        params: config.params,
        expected: config.expected,
        actual,
        status: matches ? "pass" : "fail",
        message: matches
          ? `${config.function}() = ${formatForDisplay(actual)}`
          : `Expected ${formatForDisplay(config.expected)}, got ${formatForDisplay(actual)}`,
      };
    } catch (err) {
      return {
        function: config.function,
        params: config.params,
        expected: config.expected,
        actual: undefined,
        status: "fail",
        message: `Call failed: ${err instanceof Error ? err.message : String(err)}`,
      };
    }
  }
}

// Export Verifier as alias for backwards compatibility
export { BrowserVerifier as Verifier };
