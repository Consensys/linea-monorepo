import { decodeFunctionData, parseAbi } from "viem";

import { ICalldataDecoder } from "../../../core/services/ICalldataDecoder";

export class ViemCalldataDecoder implements ICalldataDecoder {
  public decode(calldataFunctionInterface: string, calldata: string): Record<string, unknown> {
    const abi = parseAbi([calldataFunctionInterface]);
    const { args } = decodeFunctionData({ abi, data: calldata as `0x${string}` });
    if (!args) return {};
    // Convert positional array to a record keyed by index for uniformity
    return Object.fromEntries((args as unknown[]).map((v, i) => [String(i), v]));
  }
}
