export function bigintReplacer(key: string, value: unknown) {
  return typeof value === "bigint" ? value.toString() : value;
}

export function serialize(value: unknown): string {
  return JSON.stringify(value, (_, value: unknown) => (typeof value === "bigint" ? value.toString() : value));
}

/**
 * Type guard function to check if a given value is a string.
 *
 * @param {unknown} value - The value to check.
 * @returns {boolean} `true` if the value is a string, `false` otherwise.
 */
export function isString(value: unknown): value is string {
  return typeof value === "string";
}
