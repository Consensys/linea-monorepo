/**
 * Contract Integrity Verifier - ABI Utilities
 *
 * Utilities for parsing ABI files and comparing function selectors.
 * Supports both Hardhat and Foundry artifact formats.
 */

import { readFileSync } from "fs";
import { basename, dirname } from "path";
import { ethers } from "ethers";
import {
  ArtifactJson,
  AbiElement,
  AbiInput,
  AbiComparisonResult,
  NormalizedArtifact,
  ArtifactFormat,
  HardhatArtifact,
  FoundryArtifact,
  ImmutableReference,
} from "../types";

// ============================================================================
// Artifact Format Detection
// ============================================================================

/**
 * Detects whether an artifact is Hardhat or Foundry format.
 */
export function detectArtifactFormat(artifact: unknown): ArtifactFormat {
  const obj = artifact as Record<string, unknown>;

  // Foundry: bytecode is an object with 'object' property
  if (obj.bytecode && typeof obj.bytecode === "object" && (obj.bytecode as Record<string, unknown>).object) {
    return "foundry";
  }

  // Hardhat: bytecode is a string
  return "hardhat";
}

/**
 * Type guard for Foundry artifacts.
 */
function isFoundryArtifact(artifact: unknown): artifact is FoundryArtifact {
  return detectArtifactFormat(artifact) === "foundry";
}

// ============================================================================
// Artifact Loading and Normalization
// ============================================================================

/**
 * Loads and normalizes an artifact file (Hardhat or Foundry).
 */
export function loadArtifact(filePath: string): NormalizedArtifact {
  const content = readFileSync(filePath, "utf-8");
  const raw = JSON.parse(content);

  if (isFoundryArtifact(raw)) {
    return normalizeFoundryArtifact(raw, filePath);
  }

  return normalizeHardhatArtifact(raw as HardhatArtifact);
}

/**
 * Normalizes a Hardhat artifact to the common format.
 */
function normalizeHardhatArtifact(artifact: HardhatArtifact): NormalizedArtifact {
  return {
    format: "hardhat",
    contractName: artifact.contractName,
    abi: artifact.abi,
    bytecode: artifact.bytecode,
    deployedBytecode: artifact.deployedBytecode,
    // Hardhat doesn't provide immutableReferences or methodIdentifiers
    immutableReferences: undefined,
    methodIdentifiers: undefined,
  };
}

/**
 * Normalizes a Foundry artifact to the common format.
 */
function normalizeFoundryArtifact(artifact: FoundryArtifact, filePath: string): NormalizedArtifact {
  // Derive contract name from file path (e.g., "out/MyContract.sol/MyContract.json" -> "MyContract")
  const fileName = basename(filePath, ".json");
  const parentDir = basename(dirname(filePath));
  const contractName = parentDir.endsWith(".sol") ? fileName : fileName;

  // Extract immutable references
  let immutableReferences: ImmutableReference[] | undefined;
  if (artifact.deployedBytecode.immutableReferences) {
    immutableReferences = [];
    for (const refs of Object.values(artifact.deployedBytecode.immutableReferences)) {
      for (const ref of refs) {
        immutableReferences.push({ start: ref.start, length: ref.length });
      }
    }
  }

  // Convert methodIdentifiers to Map (selector -> signature)
  let methodIdentifiers: Map<string, string> | undefined;
  if (artifact.methodIdentifiers) {
    methodIdentifiers = new Map();
    for (const [signature, selector] of Object.entries(artifact.methodIdentifiers)) {
      methodIdentifiers.set(selector.toLowerCase(), signature);
    }
  }

  return {
    format: "foundry",
    contractName,
    abi: artifact.abi,
    bytecode: artifact.bytecode.object,
    deployedBytecode: artifact.deployedBytecode.object,
    immutableReferences,
    methodIdentifiers,
  };
}

/**
 * @deprecated Use loadArtifact() which returns NormalizedArtifact.
 * Kept for backward compatibility.
 */
export function loadArtifactLegacy(filePath: string): ArtifactJson {
  const content = readFileSync(filePath, "utf-8");
  return JSON.parse(content) as ArtifactJson;
}

/**
 * Generates a function signature from ABI element.
 * e.g., "transfer(address,uint256)"
 */
function getFunctionSignature(element: AbiElement): string {
  if (!element.name) {
    return "";
  }

  const inputs = element.inputs || [];
  const paramTypes = inputs.map((input) => getParamType(input)).join(",");
  return `${element.name}(${paramTypes})`;
}

/**
 * Recursively gets the parameter type string for tuple types.
 */
function getParamType(input: AbiInput): string {
  if (input.type === "tuple" || input.type === "tuple[]") {
    const components = input.components || [];
    const innerTypes = components.map((c) => getParamType(c)).join(",");
    const tupleType = `(${innerTypes})`;
    return input.type === "tuple[]" ? `${tupleType}[]` : tupleType;
  }
  return input.type;
}

/**
 * Computes the 4-byte function selector from a signature.
 */
export function computeSelector(signature: string): string {
  const hash = ethers.keccak256(ethers.toUtf8Bytes(signature));
  return hash.slice(2, 10).toLowerCase();
}

/**
 * Extracts all function selectors from an ABI.
 * Returns a map of selector -> function name for debugging.
 */
export function extractSelectorsFromAbi(abi: AbiElement[]): Map<string, string> {
  const selectorMap = new Map<string, string>();

  for (const element of abi) {
    if (element.type === "function" && element.name) {
      const signature = getFunctionSignature(element);
      const selector = computeSelector(signature);
      selectorMap.set(selector, element.name);
    }
  }

  return selectorMap;
}

/**
 * Extracts selectors from a normalized artifact.
 * Uses pre-computed methodIdentifiers for Foundry artifacts.
 */
export function extractSelectorsFromArtifact(artifact: NormalizedArtifact): Map<string, string> {
  // Foundry provides pre-computed selectors
  if (artifact.methodIdentifiers && artifact.methodIdentifiers.size > 0) {
    return artifact.methodIdentifiers;
  }

  // Fall back to computing from ABI
  return extractSelectorsFromAbi(artifact.abi);
}

/**
 * Extracts error selectors from an ABI.
 */
export function extractErrorSelectorsFromAbi(abi: AbiElement[]): Map<string, string> {
  const selectorMap = new Map<string, string>();

  for (const element of abi) {
    if (element.type === "error" && element.name) {
      const signature = getFunctionSignature(element);
      const selector = computeSelector(signature);
      selectorMap.set(selector, element.name);
    }
  }

  return selectorMap;
}

/**
 * Compares ABI selectors against bytecode-extracted selectors.
 */
export function compareSelectors(abiSelectors: Map<string, string>, bytecodeSelectors: string[]): AbiComparisonResult {
  const abiSelectorSet = new Set(abiSelectors.keys());
  const bytecodeSelectorSet = new Set(bytecodeSelectors);

  const missingSelectors: string[] = [];
  const extraSelectors: string[] = [];

  // Find selectors in ABI but not in bytecode
  for (const entry of Array.from(abiSelectors.entries())) {
    const [selector, name] = entry;
    if (!bytecodeSelectorSet.has(selector)) {
      missingSelectors.push(`${selector} (${name})`);
    }
  }

  // Find selectors in bytecode but not in ABI
  // Note: bytecode may have internal function selectors, so this is informational
  for (const selector of bytecodeSelectors) {
    if (!abiSelectorSet.has(selector)) {
      extraSelectors.push(selector);
    }
  }

  const localSelectors = Array.from(abiSelectors.keys()).sort();
  const remoteSelectors = bytecodeSelectors.sort();

  if (missingSelectors.length === 0) {
    return {
      status: "pass",
      message: `All ${abiSelectors.size} ABI function selectors found in bytecode`,
      localSelectors,
      remoteSelectors,
    };
  }

  return {
    status: "fail",
    message: `${missingSelectors.length} ABI function selectors not found in bytecode`,
    localSelectors,
    remoteSelectors,
    missingSelectors,
    extraSelectors: extraSelectors.length > 10 ? extraSelectors.slice(0, 10) : extraSelectors,
  };
}
