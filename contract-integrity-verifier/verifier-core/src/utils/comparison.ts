/**
 * Contract Integrity Verifier - Comparison Utilities
 *
 * Shared utilities for value comparison and formatting.
 * Used by both verifier.ts and storage.ts.
 */

// ============================================================================
// Value Formatting
// ============================================================================

/**
 * Formats a value for internal processing (e.g., bigint to string).
 */
export function formatValue(value: unknown): unknown {
  if (typeof value === "bigint") {
    return value.toString();
  }
  if (Array.isArray(value)) {
    return value.map(formatValue);
  }
  return value;
}

/**
 * Formats a value for display, truncating long strings.
 */
export function formatForDisplay(value: unknown, maxLength: number = 20): string {
  if (typeof value === "string" && value.length > maxLength) {
    return value.slice(0, 10) + "..." + value.slice(-8);
  }
  return String(value);
}

// ============================================================================
// Value Normalization
// ============================================================================

/**
 * Checks if a string represents a numeric value (decimal or hex).
 */
export function isNumericString(value: string): boolean {
  return /^-?\d+$/.test(value) || /^0x[0-9a-fA-F]+$/.test(value);
}

/**
 * Normalizes a value for comparison.
 * - Addresses are lowercased
 * - Numbers/bigints are converted to strings
 * - Booleans are converted to strings
 */
export function normalizeForComparison(value: unknown): string {
  if (typeof value === "string") {
    // Lowercase addresses for case-insensitive comparison
    if (value.startsWith("0x") && value.length === 42) {
      return value.toLowerCase();
    }
    return value;
  }
  if (typeof value === "bigint" || typeof value === "number") {
    return String(value);
  }
  if (typeof value === "boolean") {
    return String(value);
  }
  return String(value);
}

// ============================================================================
// Value Comparison
// ============================================================================

/**
 * Supported comparison operators.
 */
export type ComparisonOperator = "eq" | "gt" | "gte" | "lt" | "lte" | "contains";

/**
 * Compares two values using the specified comparison operator.
 *
 * @param actual - The actual value from the chain
 * @param expected - The expected value from config
 * @param comparison - The comparison operator (default: "eq")
 * @returns true if the comparison passes
 */
export function compareValues(
  actual: unknown,
  expected: unknown,
  comparison: ComparisonOperator | string = "eq",
): boolean {
  const normalizedActual = normalizeForComparison(actual);
  const normalizedExpected = normalizeForComparison(expected);

  switch (comparison) {
    case "eq":
      return normalizedActual === normalizedExpected;

    case "gt":
    case "gte":
    case "lt":
    case "lte": {
      // For numeric comparisons, both values must be numeric
      if (!isNumericString(normalizedActual) || !isNumericString(normalizedExpected)) {
        // Fall back to string equality for non-numeric values
        return normalizedActual === normalizedExpected;
      }
      const actualBigInt = BigInt(normalizedActual);
      const expectedBigInt = BigInt(normalizedExpected);

      if (comparison === "gt") return actualBigInt > expectedBigInt;
      if (comparison === "gte") return actualBigInt >= expectedBigInt;
      if (comparison === "lt") return actualBigInt < expectedBigInt;
      return actualBigInt <= expectedBigInt; // lte
    }

    case "contains":
      return String(normalizedActual).includes(String(normalizedExpected));

    default:
      // Unknown comparison type, fall back to equality
      return normalizedActual === normalizedExpected;
  }
}
