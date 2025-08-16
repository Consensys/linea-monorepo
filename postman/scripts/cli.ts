const HEXADECIMAL_REGEX = new RegExp("^0[xX][0-9a-fA-F]+$");

const ADDRESS_HEX_STR_SIZE = 42;
const PRIVKEY_HEX_STR_SIZE = 66;

function sanitizeHexBytes(paramName: string, value: string, expectedSize: number) {
  // Check if value already has 0x prefix
  const hasPrefix = value.startsWith("0x");

  // Validate hexadecimal format (with or without 0x prefix)
  const hexPattern = hasPrefix ? HEXADECIMAL_REGEX : new RegExp("^[0-9a-fA-F]+$");
  if (!hexPattern.test(value)) {
    throw new Error(`${paramName}: '${value}' is not a valid Hexadecimal notation!`);
  }

  // Add 0x prefix if not present
  if (!hasPrefix) {
    value = "0x" + value;
  }

  if (value.length !== expectedSize) {
    throw new Error(`${paramName} has size ${value.length} expected ${expectedSize}`);
  }
  return value;
}

function sanitizeAddress(argName: string) {
  return (input: string) => {
    return sanitizeHexBytes(argName, input, ADDRESS_HEX_STR_SIZE);
  };
}

function sanitizePrivKey(argName: string) {
  return (input: string) => {
    return sanitizeHexBytes(argName, input, PRIVKEY_HEX_STR_SIZE);
  };
}

export { sanitizeHexBytes, sanitizeAddress, sanitizePrivKey };
