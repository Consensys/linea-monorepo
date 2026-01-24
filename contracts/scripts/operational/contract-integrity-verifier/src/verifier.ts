/**
 * Contract Integrity Verifier - Core Verification Logic
 *
 * Main verification engine that fetches on-chain bytecode and compares
 * against local artifact files.
 */

import { ethers } from "ethers";
import {
  VerifierConfig,
  ContractConfig,
  ChainConfig,
  ContractVerificationResult,
  VerificationSummary,
  CliOptions,
} from "./types";
import { checkArtifactExists } from "./config";
import { compareBytecode, extractSelectorsFromBytecode, validateImmutablesAgainstArgs } from "./utils/bytecode";
import { loadArtifact, extractSelectorsFromArtifact, compareSelectors } from "./utils/abi";
import { verifyState } from "./utils/state";

// EIP-1967 implementation slot
const IMPLEMENTATION_SLOT = "0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc";

/**
 * Formats a parameter for display.
 */
function formatParam(param: unknown): string {
  if (typeof param === "string" && param.length > 20) {
    return param.slice(0, 10) + "...";
  }
  return String(param);
}

/**
 * Fetches bytecode from a chain at a given address.
 */
async function fetchBytecode(provider: ethers.JsonRpcProvider, address: string): Promise<string> {
  const bytecode = await provider.getCode(address);
  if (bytecode === "0x" || bytecode === "") {
    throw new Error(`No bytecode found at address ${address}`);
  }
  return bytecode;
}

/**
 * Checks if a contract is an EIP-1967 proxy and returns the implementation address.
 */
async function getImplementationAddress(provider: ethers.JsonRpcProvider, address: string): Promise<string | null> {
  try {
    const implementationSlot = await provider.getStorage(address, IMPLEMENTATION_SLOT);
    // Slot contains the address padded to 32 bytes
    const implementationAddress = ethers.getAddress("0x" + implementationSlot.slice(-40));

    // Check if it's a valid address (not zero)
    if (implementationAddress === ethers.ZeroAddress) {
      return null;
    }

    return implementationAddress;
  } catch {
    return null;
  }
}

/**
 * Creates a provider for a chain configuration.
 */
function createProvider(chain: ChainConfig): ethers.JsonRpcProvider {
  return new ethers.JsonRpcProvider(chain.rpcUrl, {
    chainId: chain.chainId,
    name: `chain-${chain.chainId}`,
  });
}

/**
 * Verifies a single contract.
 */
async function verifyContract(
  contract: ContractConfig,
  chain: ChainConfig,
  options: CliOptions,
  configDir: string,
): Promise<ContractVerificationResult> {
  const result: ContractVerificationResult = {
    contract,
    chain,
  };

  const provider = createProvider(chain);

  try {
    // Check if artifact file exists
    if (!checkArtifactExists(contract)) {
      result.error = `Artifact file not found: ${contract.artifactFile}`;
      return result;
    }

    // Load artifact (supports both Hardhat and Foundry formats)
    const artifact = loadArtifact(contract.artifactFile);

    if (options.verbose) {
      console.log(`  Artifact format: ${artifact.format}`);
      if (artifact.immutableReferences && artifact.immutableReferences.length > 0) {
        console.log(`  Known immutable positions: ${artifact.immutableReferences.length}`);
      }
    }

    // Fetch on-chain bytecode
    let remoteBytecode = await fetchBytecode(provider, contract.address);
    let addressUsed = contract.address;

    // If marked as proxy, get implementation bytecode
    if (contract.isProxy) {
      const implAddress = await getImplementationAddress(provider, contract.address);
      if (implAddress) {
        if (options.verbose) {
          console.log(`  Proxy detected, fetching implementation at ${implAddress}`);
        }
        remoteBytecode = await fetchBytecode(provider, implAddress);
        addressUsed = implAddress;
      } else {
        // Warn that we couldn't get implementation address for a marked proxy
        console.warn(`  Warning: Contract marked as proxy but no EIP-1967 implementation found at ${contract.address}`);
        // Continue with proxy bytecode - this might be intentional (e.g., different proxy pattern)
      }
    }

    // Bytecode comparison (pass known immutable positions for Foundry artifacts)
    if (!options.skipBytecode) {
      result.bytecodeResult = compareBytecode(artifact.deployedBytecode, remoteBytecode, artifact.immutableReferences);

      // If there are immutable differences and constructor args provided, validate them
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

        // Update message with constructor arg validation result
        if (validation.valid) {
          result.bytecodeResult.message += ` - constructor args validated`;
        } else {
          result.bytecodeResult.status = "warn";
          result.bytecodeResult.message += ` - ${validation.message}`;
        }
      }
    }

    // ABI comparison (uses pre-computed selectors for Foundry artifacts)
    if (!options.skipAbi) {
      const abiSelectors = extractSelectorsFromArtifact(artifact);
      const bytecodeSelectors = extractSelectorsFromBytecode(remoteBytecode);
      result.abiResult = compareSelectors(abiSelectors, bytecodeSelectors);
    }

    // State verification (optional - only if configured)
    if (!options.skipState && contract.stateVerification) {
      // For proxies, state verification happens on the proxy address (where storage lives)
      const stateAddress = contract.address;
      result.stateResult = await verifyState(
        provider,
        stateAddress,
        artifact.abi,
        contract.stateVerification,
        configDir,
      );
    }

    // Log verbose info
    if (options.verbose) {
      console.log(`  Address verified: ${addressUsed}`);
      console.log(`  Remote bytecode length: ${(remoteBytecode.length - 2) / 2} bytes`);
      if (result.bytecodeResult?.immutableDifferences && result.bytecodeResult.immutableDifferences.length > 0) {
        console.log(`  Immutable differences detected: ${result.bytecodeResult.immutableDifferences.length}`);
        for (const imm of result.bytecodeResult.immutableDifferences) {
          console.log(`    Position ${imm.position}: ${imm.possibleType || "unknown"} = 0x${imm.remoteValue}`);
        }
      }
    }
  } catch (error) {
    result.error = error instanceof Error ? error.message : String(error);
  }

  return result;
}

/**
 * Runs verification for all contracts in the configuration.
 */
export async function runVerification(
  config: VerifierConfig,
  options: CliOptions,
  configDir: string,
): Promise<VerificationSummary> {
  const results: ContractVerificationResult[] = [];
  let passed = 0;
  let failed = 0;
  let warnings = 0;
  let skipped = 0;

  // Filter contracts if specified
  let contractsToVerify = config.contracts;
  if (options.contract) {
    contractsToVerify = contractsToVerify.filter((c) => c.name.toLowerCase() === options.contract!.toLowerCase());
  }
  if (options.chain) {
    contractsToVerify = contractsToVerify.filter((c) => c.chain.toLowerCase() === options.chain!.toLowerCase());
  }

  if (contractsToVerify.length === 0) {
    console.log("No contracts to verify matching the specified filters.");
    return { total: 0, passed: 0, failed: 0, warnings: 0, skipped: 0, results: [] };
  }

  console.log(`\nVerifying ${contractsToVerify.length} contract(s)...\n`);

  for (const contract of contractsToVerify) {
    const chain = config.chains[contract.chain];
    console.log(`Verifying ${contract.name} on ${contract.chain}...`);
    console.log(`  Address: ${contract.address}`);

    const result = await verifyContract(contract, chain, options, configDir);
    results.push(result);

    if (result.error) {
      console.log(`  ERROR: ${result.error}`);
      failed++;
    } else {
      // Bytecode result
      if (result.bytecodeResult) {
        const br = result.bytecodeResult;
        const icon = br.status === "pass" ? "✓" : br.status === "fail" ? "✗" : br.status === "warn" ? "!" : "-";
        console.log(`  Bytecode: ${icon} ${br.message}`);
        if (br.status === "fail" && br.differences && options.verbose) {
          console.log(`    First differences at positions: ${br.differences.map((d) => d.position).join(", ")}`);
        }
      }

      // ABI result
      if (result.abiResult) {
        const ar = result.abiResult;
        const icon = ar.status === "pass" ? "✓" : ar.status === "fail" ? "✗" : ar.status === "warn" ? "!" : "-";
        console.log(`  ABI: ${icon} ${ar.message}`);
        if (ar.status === "fail" && ar.missingSelectors && options.verbose) {
          console.log(`    Missing selectors: ${ar.missingSelectors.slice(0, 5).join(", ")}`);
        }
      }

      // State verification result
      if (result.stateResult) {
        const sr = result.stateResult;
        const icon = sr.status === "pass" ? "✓" : sr.status === "fail" ? "✗" : sr.status === "warn" ? "!" : "-";
        console.log(`  State: ${icon} ${sr.message}`);

        // Show view call results
        if (sr.viewCallResults && options.verbose) {
          for (const vcr of sr.viewCallResults) {
            const vcIcon = vcr.status === "pass" ? "✓" : "✗";
            const paramsStr = vcr.params ? `(${vcr.params.map((p) => formatParam(p)).join(", ")})` : "()";
            console.log(`    ${vcIcon} ${vcr.function}${paramsStr}: ${vcr.message}`);
          }
        }

        // Show namespace results
        if (sr.namespaceResults && options.verbose) {
          for (const nr of sr.namespaceResults) {
            console.log(`    Namespace: ${nr.namespaceId}`);
            for (const vr of nr.variables) {
              const vrIcon = vr.status === "pass" ? "✓" : "✗";
              console.log(`      ${vrIcon} ${vr.name}: ${vr.message}`);
            }
          }
        }

        // Show slot results
        if (sr.slotResults && options.verbose) {
          for (const slr of sr.slotResults) {
            const slIcon = slr.status === "pass" ? "✓" : "✗";
            console.log(`    ${slIcon} ${slr.name} (${slr.slot}): ${slr.message}`);
          }
        }

        // Show storage path results
        if (sr.storagePathResults && options.verbose) {
          for (const spr of sr.storagePathResults) {
            const sprIcon = spr.status === "pass" ? "✓" : "✗";
            console.log(`    ${sprIcon} ${spr.path}: ${spr.message}`);
          }
        }
      }

      // Count results
      const bytecodeStatus = result.bytecodeResult?.status;
      const abiStatus = result.abiResult?.status;
      const stateStatus = result.stateResult?.status;

      // Check for failures first
      if (bytecodeStatus === "fail" || abiStatus === "fail" || stateStatus === "fail") {
        failed++;
      } else if (bytecodeStatus === "warn" || abiStatus === "warn" || stateStatus === "warn") {
        warnings++;
      } else {
        // Determine if any verification was actually performed
        const hasBytecodeResult = result.bytecodeResult !== undefined;
        const hasAbiResult = result.abiResult !== undefined;
        const hasStateResult = result.stateResult !== undefined;
        const hasAnyVerification = hasBytecodeResult || hasAbiResult || hasStateResult;

        if (!hasAnyVerification) {
          // No verification was performed (all checks skipped via CLI or no state config)
          skipped++;
        } else if (
          (bytecodeStatus === "skip" || !hasBytecodeResult) &&
          (abiStatus === "skip" || !hasAbiResult) &&
          !hasStateResult
        ) {
          // All performed checks were skipped
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
      const hasError = result.error;

      const hasStateFailure = result.stateResult?.status === "fail";

      if (hasBytecodeFailure || hasAbiFailure || hasStateFailure || hasError) {
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
      }
    }
  }
}
