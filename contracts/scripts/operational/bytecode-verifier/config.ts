/**
 * Bytecode Verifier - Configuration Loading
 *
 * Handles loading and validating the verifier configuration file.
 * Supports environment variable interpolation in config values.
 */

import { readFileSync, existsSync } from "fs";
import { resolve, dirname } from "path";
import { VerifierConfig, ChainConfig, ContractConfig } from "./types";

/**
 * Interpolates environment variables in a string.
 * Supports ${VAR_NAME} syntax.
 */
function interpolateEnvVars(value: string): string {
  return value.replace(/\$\{([^}]+)\}/g, (_, varName: string) => {
    const envValue = process.env[varName];
    if (envValue === undefined) {
      throw new Error(`Environment variable '${varName}' is not set`);
    }
    return envValue;
  });
}

/**
 * Validates a chain configuration.
 */
function validateChainConfig(name: string, config: ChainConfig): void {
  if (!config.chainId || typeof config.chainId !== "number") {
    throw new Error(`Chain '${name}' must have a valid numeric chainId`);
  }
  if (!config.rpcUrl || typeof config.rpcUrl !== "string") {
    throw new Error(`Chain '${name}' must have an rpcUrl`);
  }
}

/**
 * Validates a contract configuration.
 */
function validateContractConfig(config: ContractConfig, chains: Record<string, ChainConfig>): void {
  if (!config.name) {
    throw new Error("Contract must have a name");
  }
  if (!config.chain || !chains[config.chain]) {
    throw new Error(`Contract '${config.name}' references unknown chain '${config.chain}'`);
  }
  if (!config.address || !/^0x[a-fA-F0-9]{40}$/.test(config.address)) {
    throw new Error(`Contract '${config.name}' must have a valid address`);
  }
  if (!config.artifactFile) {
    throw new Error(`Contract '${config.name}' must have an artifactFile`);
  }
}

/**
 * Resolves file paths relative to the config file location.
 */
function resolveFilePaths(config: VerifierConfig, configDir: string): void {
  for (const contract of config.contracts) {
    contract.artifactFile = resolve(configDir, contract.artifactFile);
  }
}

/**
 * Loads and validates a configuration file.
 */
export function loadConfig(configPath: string): VerifierConfig {
  const absolutePath = resolve(configPath);

  if (!existsSync(absolutePath)) {
    throw new Error(`Configuration file not found: ${absolutePath}`);
  }

  let rawConfig: string;
  try {
    rawConfig = readFileSync(absolutePath, "utf-8");
  } catch (error) {
    throw new Error(`Failed to read configuration file: ${error}`);
  }

  // Interpolate environment variables before parsing
  const interpolatedConfig = interpolateEnvVars(rawConfig);

  let config: VerifierConfig;
  try {
    config = JSON.parse(interpolatedConfig) as VerifierConfig;
  } catch (error) {
    throw new Error(`Failed to parse configuration file as JSON: ${error}`);
  }

  // Validate chains
  if (!config.chains || typeof config.chains !== "object") {
    throw new Error("Configuration must have a 'chains' object");
  }
  for (const [name, chainConfig] of Object.entries(config.chains)) {
    validateChainConfig(name, chainConfig);
  }

  // Validate contracts
  if (!config.contracts || !Array.isArray(config.contracts)) {
    throw new Error("Configuration must have a 'contracts' array");
  }
  for (const contract of config.contracts) {
    validateContractConfig(contract, config.chains);
  }

  // Resolve relative file paths
  resolveFilePaths(config, dirname(absolutePath));

  return config;
}

/**
 * Checks if the artifact file exists for a contract configuration.
 */
export function checkArtifactExists(contract: ContractConfig): boolean {
  return existsSync(contract.artifactFile);
}
