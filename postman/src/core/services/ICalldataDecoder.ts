export interface ICalldataDecoder {
  /** Decode function calldata according to the given ABI signature string. */
  decode(calldataFunctionInterface: string, calldata: string): Record<string, unknown>;
}
