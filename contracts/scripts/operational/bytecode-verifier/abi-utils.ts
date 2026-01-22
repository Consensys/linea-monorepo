/**
 * Bytecode Verifier - ABI Utilities
 *
 * Utilities for parsing ABI files and comparing function selectors.
 */

import { readFileSync } from "fs";
import { ethers } from "ethers";
import { ArtifactJson, AbiElement, AbiInput, AbiComparisonResult } from "./types";

/**
 * Loads and parses a Hardhat artifact JSON file.
 */
export function loadArtifact(filePath: string): ArtifactJson {
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
