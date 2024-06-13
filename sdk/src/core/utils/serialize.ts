/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Stringifier that handles bigint value.
 * @param {any} value Value to stringify.
 * @returns {string} the stringified output.
 */
export function serialize(value: any): string {
  return JSON.stringify(value, (_, value: any) => (typeof value === "bigint" ? value.toString() : value));
}
