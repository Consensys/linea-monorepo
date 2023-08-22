const HEXADECIMAL_REGEX = new RegExp("^0[xX][0-9a-fA-F]+$");

const ADDRESS_HEX_STR_SIZE = 42;
const PRIVKEY_HEX_STR_SIZE = 66;
const HASH_HEX_STR_SIZE = 66;

function sanitizeHexBytes(paramName: string, value: string, expectedSize: number) {
  // TODO: add hexadecimal regex match
  if (!value.startsWith("0x")) {
    value = "0x" + value;
  }

  if (!HEXADECIMAL_REGEX.test(value)) {
    throw new Error(`${paramName}: '${value}' is not a valid Hexadecimal notation!`);
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

function sanitizeHash(argName: string) {
  return (input: string) => {
    return sanitizeHexBytes(argName, input, HASH_HEX_STR_SIZE);
  };
}

function assertIsNumber(argName: string) {
  return (input: string) => {
    if (typeof input != "number" || isNaN(input)) {
      throw Error(`${argName} must be a valid number. Got: '${input}'`);
    }
    return input;
  };
}

export { sanitizeHexBytes, sanitizeAddress, sanitizePrivKey, sanitizeHash, assertIsNumber };
