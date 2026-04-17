import { ethers } from "ethers";
import fs from "fs";
import path from "path";

const REGISTRY_DIR = path.join(__dirname, "..", "..", "deployments", "addresses");

const REGISTRY_NETWORKS = new Set(["mainnet", "sepolia", "hoodi", "linea_mainnet", "linea_sepolia"]);

interface RegistryEntry {
  address: string;
  notes?: string;
}

interface NetworkRegistry {
  network: string;
  chainId: number;
  contracts: Record<string, RegistryEntry>;
}

function loadRegistry(networkName: string): NetworkRegistry | undefined {
  const filePath = path.join(REGISTRY_DIR, `${networkName}.json`);
  if (!fs.existsSync(filePath)) {
    return undefined;
  }
  const raw = fs.readFileSync(filePath, "utf-8");
  return JSON.parse(raw) as NetworkRegistry;
}

/**
 * Reads a contract address from the per-network registry file.
 * Returns `undefined` if the network has no registry, the key is missing, or the
 * address is the zero address (used as a placeholder for entries not yet populated).
 */
export function getAddressFromRegistry(networkName: string, contractKey: string): string | undefined {
  if (!REGISTRY_NETWORKS.has(networkName)) {
    return undefined;
  }
  const registry = loadRegistry(networkName);
  if (!registry) {
    return undefined;
  }
  const entry = registry.contracts[contractKey];
  if (!entry) {
    return undefined;
  }
  if (!ethers.isAddress(entry.address)) {
    throw new Error(
      `Registry entry for "${contractKey}" on "${networkName}" has an invalid address: "${entry.address}". ` +
        `Fix contracts/deployments/addresses/${networkName}.json.`,
    );
  }
  const checksummed = ethers.getAddress(entry.address);
  // Treat the zero address as "not populated" — registry entries are initialised with
  // address(0) as a placeholder until the real address is known.
  if (checksummed === ethers.ZeroAddress) {
    return undefined;
  }
  return checksummed;
}

/**
 * Resolves an address for deployment, combining the registry with an optional env var.
 *
 * Resolution order and conflict rules:
 *
 * | Registry entry        | Env var set              | Outcome                                              |
 * |-----------------------|--------------------------|------------------------------------------------------|
 * | Present (non-zero)    | Not set                  | Returns registry address — env var is optional       |
 * | Present (non-zero)    | Matches registry         | Returns registry address                             |
 * | Present (non-zero)    | Conflicts with registry  | Hard fail — both values printed, deploy aborted      |
 * | Absent / zero         | Set (valid address)      | Returns env var value (validated and checksummed)    |
 * | Absent / zero         | Not set                  | Hard fail — no source available                      |
 *
 * Networks without a registry file (custom, zkevm_dev, l2) bypass registry lookup
 * entirely and fall back to the env var only.
 *
 * @param networkName  Hardhat network name (e.g. "mainnet", "sepolia", "custom")
 * @param contractKey  Key in the registry JSON (e.g. "LineaRollup", "L1_SECURITY_COUNCIL")
 * @param envVarName   Optional env var to cross-check / fall back to
 */
export function requireAddressOrRegistry(networkName: string, contractKey: string, envVarName?: string): string {
  const registryAddress = getAddressFromRegistry(networkName, contractKey);
  const rawEnvValue = envVarName ? process.env[envVarName] : undefined;

  const envAddress = (() => {
    if (!rawEnvValue) return undefined;
    if (!ethers.isAddress(rawEnvValue)) {
      throw new Error(
        `Environment variable "${envVarName}" is not a valid EVM address. Got: "${rawEnvValue}". ` +
          `Expected a 0x-prefixed 40-hex-character address.`,
      );
    }
    return ethers.getAddress(rawEnvValue);
  })();

  if (registryAddress !== undefined) {
    if (envAddress !== undefined && envAddress !== registryAddress) {
      throw new Error(
        `Address conflict for "${contractKey}" on network "${networkName}":\n` +
          `  Registry (contracts/deployments/addresses/${networkName}.json): ${registryAddress}\n` +
          `  Environment variable ${envVarName}: ${envAddress}\n` +
          `Either remove the env var override or update the registry to match.`,
      );
    }
    console.log(
      `Using registry address for ${contractKey} on ${networkName}: ${registryAddress}${envAddress ? " (matches env var)" : ""}`,
    );
    return registryAddress;
  }

  if (envAddress !== undefined) {
    console.log(`Using environment variable ${envVarName}=${envAddress} for ${contractKey} (no registry entry)`);
    return envAddress;
  }

  const registryNote = REGISTRY_NETWORKS.has(networkName)
    ? ` Add it to contracts/deployments/addresses/${networkName}.json, or set env var ${envVarName ?? "<env var>"}.`
    : ` Set env var ${envVarName ?? "<env var>"}.`;

  throw new Error(`No address found for "${contractKey}" on network "${networkName}".${registryNote}`);
}

/**
 * @deprecated Use getAddressFromRegistry instead.
 *
 * Reads a deployed contract address from the legacy hardhat-deploy artefact format at
 * `deployments/<networkName>/<contractName>.json`. This path is no longer written by
 * deploy scripts; prefer the address registry under `deployments/addresses/`.
 */
export const getDeployedContractOnNetwork = async (
  networkName: string,
  contractName: string,
): Promise<string | undefined> => {
  const filePath = path.join(__dirname, "..", "..", "deployments", `${networkName}`, `${contractName}.json`);
  if (!fs.existsSync(filePath)) {
    return undefined;
  }
  const data = fs.readFileSync(filePath, "utf-8");
  const address: string = JSON.parse(data).address;
  return ethers.isAddress(address) ? ethers.getAddress(address) : undefined;
};
