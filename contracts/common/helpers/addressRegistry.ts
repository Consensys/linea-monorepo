import { ethers } from "ethers";
import fs from "fs";
import path from "path";

export const REGISTRY_DIR = path.join(__dirname, "..", "..", "deployments", "addresses");

export const REGISTRY_NETWORKS = new Set(["mainnet", "sepolia", "hoodi", "linea_mainnet", "linea_sepolia"]);

export const EXPECTED_CHAIN_IDS: Record<string, number> = {
  mainnet: 1,
  sepolia: 11155111,
  hoodi: 560048,
  linea_mainnet: 59144,
  linea_sepolia: 59141,
};

export interface AddressObject {
  address: string;
  notes?: string;
}

export interface SingleRegistryEntry {
  address: string;
  notes?: string;
}

export interface MultiRegistryEntry {
  addresses: AddressObject[];
  notes?: string;
}

export type RegistryContractEntry = SingleRegistryEntry | MultiRegistryEntry;

export interface NetworkRegistry {
  $schema?: string;
  network: string;
  chainId: number;
  contracts: Record<string, RegistryContractEntry>;
}

export interface RegistryValidationIssue {
  file: string;
  path: string;
  message: string;
}

interface DuplicateJsonKey {
  path: string;
  key: string;
}

export function isMultiRegistryEntry(entry: RegistryContractEntry): entry is MultiRegistryEntry {
  return "addresses" in entry && Array.isArray(entry.addresses);
}

export function parseChecksummedAddress(raw: string, context: string): string {
  if (!ethers.isAddress(raw)) {
    throw new Error(`${context}: invalid EVM address "${raw}"`);
  }
  return ethers.getAddress(raw);
}

export function assertChecksummedAddress(address: string, context: string): void {
  const checksummed = parseChecksummedAddress(address, context);
  if (address !== checksummed) {
    throw new Error(`${context}: address must be EIP-55 checksummed. Expected "${checksummed}", got "${address}".`);
  }
}

export function resolveRegistryEntryAddresses(entry: RegistryContractEntry): string[] {
  if (isMultiRegistryEntry(entry)) {
    return entry.addresses.map((item, index) => parseChecksummedAddress(item.address, `addresses[${index}].address`));
  }
  return [parseChecksummedAddress(entry.address, "address")];
}

/**
 * Returns populated addresses for a registry entry, or `undefined` when the entry is
 * treated as unpopulated (single zero address, or an `addresses` array where every item is zero).
 */
export function getPopulatedAddresses(entry: RegistryContractEntry): string[] | undefined {
  const addresses = resolveRegistryEntryAddresses(entry);
  const nonZero = addresses.filter((value) => value !== ethers.ZeroAddress);
  if (nonZero.length === 0) {
    return undefined;
  }
  if (nonZero.length !== addresses.length) {
    throw new Error(
      "Registry entry mixes zero and non-zero addresses. Populate every item or use zero placeholders for the entire entry.",
    );
  }
  return addresses;
}

export function loadRegistry(networkName: string): NetworkRegistry | undefined {
  const filePath = path.join(REGISTRY_DIR, `${networkName}.json`);
  if (!fs.existsSync(filePath)) {
    return undefined;
  }
  const raw = fs.readFileSync(filePath, "utf-8");
  return JSON.parse(raw) as NetworkRegistry;
}

export function lookupRegistryEntry(
  registry: NetworkRegistry,
  contractKey: string,
  envVarName?: string,
): RegistryContractEntry | undefined {
  if (registry.contracts[contractKey] !== undefined) {
    return registry.contracts[contractKey];
  }
  if (envVarName !== undefined && envVarName !== contractKey && registry.contracts[envVarName] !== undefined) {
    return registry.contracts[envVarName];
  }
  return undefined;
}

export function parseEnvAddressList(envVarName: string): string[] | undefined {
  const raw = process.env[envVarName];
  if (!raw?.trim()) {
    return undefined;
  }
  return raw.split(",").map((part, index) => parseChecksummedAddress(part.trim(), `${envVarName}[${index}]`));
}

export function parseEnvAddress(envVarName: string): string | undefined {
  const addresses = parseEnvAddressList(envVarName);
  if (addresses === undefined) {
    return undefined;
  }
  if (addresses.length !== 1) {
    throw new Error(
      `Environment variable "${envVarName}" must contain exactly one address, got ${addresses.length}. ` +
        "Use the multi-address registry helper for comma-delimited address lists.",
    );
  }
  return addresses[0];
}

export function sameAddressSet(left: string[], right: string[]): boolean {
  if (left.length !== right.length) {
    return false;
  }
  const sortedLeft = [...left].sort();
  const sortedRight = [...right].sort();
  return sortedLeft.every((value, index) => value === sortedRight[index]);
}

function findDuplicateJsonKeys(source: string): DuplicateJsonKey[] {
  const duplicates: DuplicateJsonKey[] = [];
  let index = 0;

  const skipWhitespace = () => {
    while (/\s/.test(source[index] ?? "")) {
      index += 1;
    }
  };

  const parseString = (): string => {
    if (source[index] !== '"') {
      throw new Error(`Expected string at offset ${index}.`);
    }
    let value = "";
    index += 1;
    while (index < source.length) {
      const char = source[index];
      if (char === '"') {
        index += 1;
        return value;
      }
      if (char === "\\") {
        const escapeStart = index;
        index += 1;
        const escaped = source[index];
        if (escaped === "u") {
          const hex = source.slice(index + 1, index + 5);
          if (!/^[0-9a-fA-F]{4}$/.test(hex)) {
            throw new Error(`Invalid unicode escape at offset ${escapeStart}.`);
          }
          value += String.fromCharCode(parseInt(hex, 16));
          index += 5;
          continue;
        }
        const escapes: Record<string, string> = {
          '"': '"',
          "\\": "\\",
          "/": "/",
          b: "\b",
          f: "\f",
          n: "\n",
          r: "\r",
          t: "\t",
        };
        if (escaped === undefined || escapes[escaped] === undefined) {
          throw new Error(`Invalid escape at offset ${escapeStart}.`);
        }
        value += escapes[escaped];
        index += 1;
        continue;
      }
      value += char;
      index += 1;
    }
    throw new Error("Unterminated string.");
  };

  const parseValue = (pathParts: string[]): void => {
    skipWhitespace();
    const char = source[index];
    if (char === "{") {
      parseObject(pathParts);
      return;
    }
    if (char === "[") {
      parseArray(pathParts);
      return;
    }
    if (char === '"') {
      parseString();
      return;
    }
    while (index < source.length && !/[,\]}]/.test(source[index])) {
      index += 1;
    }
  };

  const parseArray = (pathParts: string[]): void => {
    index += 1;
    skipWhitespace();
    let itemIndex = 0;
    while (source[index] !== "]") {
      parseValue([...pathParts, `[${itemIndex}]`]);
      itemIndex += 1;
      skipWhitespace();
      if (source[index] === ",") {
        index += 1;
        skipWhitespace();
        continue;
      }
      if (source[index] !== "]") {
        throw new Error(`Expected "," or "]" at offset ${index}.`);
      }
    }
    index += 1;
  };

  const parseObject = (pathParts: string[]): void => {
    index += 1;
    skipWhitespace();
    const seen = new Set<string>();
    while (source[index] !== "}") {
      const key = parseString();
      const pathLabel = pathParts.length > 0 ? pathParts.join(".") : "$";
      if (seen.has(key)) {
        duplicates.push({ path: pathLabel, key });
      }
      seen.add(key);
      skipWhitespace();
      if (source[index] !== ":") {
        throw new Error(`Expected ":" at offset ${index}.`);
      }
      index += 1;
      parseValue([...pathParts, key]);
      skipWhitespace();
      if (source[index] === ",") {
        index += 1;
        skipWhitespace();
        continue;
      }
      if (source[index] !== "}") {
        throw new Error(`Expected "," or "}" at offset ${index}.`);
      }
    }
    index += 1;
  };

  skipWhitespace();
  parseValue([]);
  return duplicates;
}

function validateRegistryEntryShape(
  file: string,
  contractKey: string,
  entry: unknown,
  issues: RegistryValidationIssue[],
): entry is RegistryContractEntry {
  const entryPath = `contracts.${contractKey}`;
  const issueCountBefore = issues.length;

  if (entry === null || typeof entry !== "object" || Array.isArray(entry)) {
    issues.push({ file, path: entryPath, message: "Entry must be an object." });
    return false;
  }

  const record = entry as Record<string, unknown>;
  const hasAddress = typeof record.address === "string";
  const hasAddresses = Array.isArray(record.addresses);

  if (hasAddress && hasAddresses) {
    issues.push({ file, path: entryPath, message: "Entry must not define both `address` and `addresses`." });
    return false;
  }

  if (!hasAddress && !hasAddresses) {
    issues.push({ file, path: entryPath, message: "Entry must define either `address` or `addresses`." });
    return false;
  }

  if (hasAddresses) {
    const addressItems = record.addresses as unknown[];
    if (addressItems.length === 0) {
      issues.push({ file, path: `${entryPath}.addresses`, message: "`addresses` must contain at least one item." });
      return false;
    }

    for (const [index, item] of addressItems.entries()) {
      const itemPath = `${entryPath}.addresses[${index}]`;
      if (item === null || typeof item !== "object" || Array.isArray(item)) {
        issues.push({ file, path: itemPath, message: "Address item must be an object." });
        continue;
      }
      const addressItem = item as Record<string, unknown>;
      if (typeof addressItem.address !== "string") {
        issues.push({ file, path: `${itemPath}.address`, message: "`address` must be a string." });
        continue;
      }
      try {
        assertChecksummedAddress(addressItem.address, `${file} ${itemPath}.address`);
      } catch (error) {
        issues.push({
          file,
          path: `${itemPath}.address`,
          message: error instanceof Error ? error.message : String(error),
        });
      }
      if (addressItem.notes !== undefined && typeof addressItem.notes !== "string") {
        issues.push({ file, path: `${itemPath}.notes`, message: "`notes` must be a string when present." });
      }
    }

    const seen = new Set<string>();
    for (const item of addressItems as AddressObject[]) {
      if (!ethers.isAddress(item.address)) {
        continue;
      }
      const normalized = ethers.getAddress(item.address).toLowerCase();
      if (seen.has(normalized)) {
        issues.push({
          file,
          path: `${entryPath}.addresses`,
          message: `Duplicate address "${item.address}" in \`addresses\` array.`,
        });
        break;
      }
      seen.add(normalized);
    }
  } else if (typeof record.address === "string") {
    try {
      assertChecksummedAddress(record.address, `${file} ${entryPath}.address`);
    } catch (error) {
      issues.push({
        file,
        path: `${entryPath}.address`,
        message: error instanceof Error ? error.message : String(error),
      });
    }
  } else {
    issues.push({ file, path: `${entryPath}.address`, message: "`address` must be a string." });
    return false;
  }

  if (record.notes !== undefined && typeof record.notes !== "string") {
    issues.push({ file, path: `${entryPath}.notes`, message: "`notes` must be a string when present." });
  }

  return issues.length === issueCountBefore;
}

function validateMixedZeroPopulation(
  file: string,
  contractKey: string,
  entry: RegistryContractEntry,
  issues: RegistryValidationIssue[],
): void {
  try {
    getPopulatedAddresses(entry);
  } catch (error) {
    issues.push({
      file,
      path: `contracts.${contractKey}`,
      message: error instanceof Error ? error.message : String(error),
    });
  }
}

export function validateRegistryFile(fileName: string): RegistryValidationIssue[] {
  const filePath = path.join(REGISTRY_DIR, fileName);
  const issues: RegistryValidationIssue[] = [];

  if (!fs.existsSync(filePath)) {
    issues.push({ file: fileName, path: fileName, message: "Registry file is missing." });
    return issues;
  }

  let registry: NetworkRegistry;
  const raw = fs.readFileSync(filePath, "utf-8");
  try {
    for (const duplicate of findDuplicateJsonKeys(raw)) {
      issues.push({
        file: fileName,
        path: duplicate.path,
        message: `Duplicate JSON key "${duplicate.key}". JSON parsers keep only the last value.`,
      });
    }
    registry = JSON.parse(raw) as NetworkRegistry;
  } catch (error) {
    issues.push({
      file: fileName,
      path: fileName,
      message: `Invalid JSON: ${error instanceof Error ? error.message : String(error)}`,
    });
    return issues;
  }

  const networkName = fileName.replace(/\.json$/, "");
  if (registry.network !== networkName) {
    issues.push({
      file: fileName,
      path: "network",
      message: `Expected network "${networkName}", got "${registry.network}".`,
    });
  }

  const expectedChainId = EXPECTED_CHAIN_IDS[networkName];
  if (expectedChainId === undefined) {
    issues.push({ file: fileName, path: "network", message: `Unknown registry network "${networkName}".` });
  } else if (registry.chainId !== expectedChainId) {
    issues.push({
      file: fileName,
      path: "chainId",
      message: `Expected chainId ${expectedChainId}, got ${registry.chainId}.`,
    });
  }

  if (registry.contracts === null || typeof registry.contracts !== "object" || Array.isArray(registry.contracts)) {
    issues.push({ file: fileName, path: "contracts", message: "`contracts` must be an object." });
    return issues;
  }

  for (const [contractKey, entry] of Object.entries(registry.contracts)) {
    if (validateRegistryEntryShape(fileName, contractKey, entry, issues)) {
      validateMixedZeroPopulation(fileName, contractKey, entry, issues);
    }
  }

  return issues;
}

export function validateAllAddressRegistries(): RegistryValidationIssue[] {
  return [...REGISTRY_NETWORKS].flatMap((networkName) => validateRegistryFile(`${networkName}.json`));
}
