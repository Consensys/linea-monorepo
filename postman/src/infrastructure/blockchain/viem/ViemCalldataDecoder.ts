import { type Abi, decodeFunctionData, type Hex, parseAbi } from "viem";

import { ICalldataDecoder } from "../../../core/services/ICalldataDecoder";

export class ViemCalldataDecoder implements ICalldataDecoder {
  public decode(calldataFunctionInterface: string, calldata: Hex): Record<string, unknown> {
    const abi = parseAbi([calldataFunctionInterface]) as unknown as Abi;
    const { functionName, args } = decodeFunctionData({ abi, data: calldata });
    if (!args) return {};

    const abiFunction = abi.find((item) => item.type === "function" && item.name === functionName);
    if (!abiFunction || abiFunction.type !== "function") return {};

    return Object.fromEntries(abiFunction.inputs.map((input, i) => [input.name || String(i), args[i]]));
  }
}
