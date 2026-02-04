/**
 * Contract Integrity Verifier - Hex Utilities
 *
 * Shared utilities for hex string manipulation.
 * Browser-compatible with no external dependencies.
 */

/**
 * Converts a hex string to a Uint8Array.
 * Handles both "0x" prefixed and unprefixed hex strings.
 *
 * @param hex - Hex string to convert (with or without 0x prefix)
 * @returns Uint8Array of bytes
 */
export function hexToBytes(hex: string): Uint8Array {
  const normalized = hex.startsWith("0x") ? hex.slice(2) : hex;
  const bytes = new Uint8Array(normalized.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(normalized.slice(i * 2, i * 2 + 2), 16);
  }
  return bytes;
}

/**
 * Checks if a string is a valid hex string (with or without 0x prefix).
 *
 * @param value - String to check
 * @returns true if the string is valid hex
 */
export function isHexString(value: string): boolean {
  const normalized = value.startsWith("0x") ? value.slice(2) : value;
  return /^[0-9a-fA-F]*$/.test(normalized);
}

/**
 * Checks if a string represents a decimal number.
 *
 * @param value - String to check
 * @returns true if the string is a valid decimal number
 */
export function isDecimalString(value: string): boolean {
  return /^-?\d+$/.test(value);
}
