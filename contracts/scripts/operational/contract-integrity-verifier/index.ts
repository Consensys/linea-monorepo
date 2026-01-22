#!/usr/bin/env ts-node
/**
 * Contract Integrity Verifier - CLI Entry Point
 *
 * Verifies deployed smart contract bytecode, ABI, and state against local artifact files.
 * Inspired by https://github.com/lidofinance/diffyscan
 *
 * Usage:
 *   npx ts-node scripts/operational/contract-integrity-verifier/index.ts --config <config.json>
 *
 * Options:
 *   --config, -c      Path to configuration file (required)
 *   --verbose, -v     Enable verbose output
 *   --contract        Filter to specific contract name
 *   --chain           Filter to specific chain
 *   --skip-bytecode   Skip bytecode comparison
 *   --skip-abi        Skip ABI comparison
 *   --skip-state      Skip state verification
 *   --help, -h        Show help
 *
 * Exit codes:
 *   0 - All verifications passed
 *   1 - One or more verifications failed
 *   2 - Configuration or runtime error
 */

import { dirname, resolve } from "path";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { loadConfig } from "./src/config";
import { runVerification, printSummary } from "./src/verifier";
import type { CliOptions } from "./src/types";

async function main(): Promise<void> {
  const argv = await yargs(hideBin(process.argv))
    .option("config", {
      alias: "c",
      type: "string",
      description: "Path to configuration file",
      demandOption: true,
    })
    .option("verbose", {
      alias: "v",
      type: "boolean",
      description: "Enable verbose output",
      default: false,
    })
    .option("contract", {
      type: "string",
      description: "Filter to specific contract name",
    })
    .option("chain", {
      type: "string",
      description: "Filter to specific chain",
    })
    .option("skip-bytecode", {
      type: "boolean",
      description: "Skip bytecode comparison",
      default: false,
    })
    .option("skip-abi", {
      type: "boolean",
      description: "Skip ABI comparison",
      default: false,
    })
    .option("skip-state", {
      type: "boolean",
      description: "Skip state verification",
      default: false,
    })
    .help()
    .alias("help", "h")
    .strict()
    .parse();

  const options: CliOptions = {
    config: argv.config,
    verbose: argv.verbose,
    contract: argv.contract,
    chain: argv.chain,
    skipBytecode: argv["skip-bytecode"],
    skipAbi: argv["skip-abi"],
    skipState: argv["skip-state"],
  };

  console.log("Contract Integrity Verifier");
  console.log("=".repeat(50));

  try {
    // Load configuration
    if (options.verbose) {
      console.log(`Loading configuration from: ${options.config}`);
    }
    const config = loadConfig(options.config);
    const configDir = dirname(resolve(options.config));

    if (options.verbose) {
      console.log(`Chains configured: ${Object.keys(config.chains).join(", ")}`);
      console.log(`Contracts configured: ${config.contracts.length}`);
    }

    // Run verification
    const summary = await runVerification(config, options, configDir);

    // Print summary
    printSummary(summary);

    // Exit with appropriate code
    if (summary.failed > 0) {
      process.exit(1);
    }
    process.exit(0);
  } catch (error) {
    console.error("\nFATAL ERROR:");
    console.error(error instanceof Error ? error.message : String(error));
    if (options.verbose && error instanceof Error && error.stack) {
      console.error("\nStack trace:");
      console.error(error.stack);
    }
    process.exit(2);
  }
}

main().catch((error) => {
  console.error("Unhandled error:", error);
  process.exit(2);
});
