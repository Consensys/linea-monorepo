import { decodeFunctionData, type Hex, type Abi, parseAbi, size } from "viem";

import type { ICalldataDecoder, DecodedCalldata } from "../../../domain/ports/ICalldataDecoder";

export class ViemCalldataDecoder implements ICalldataDecoder {
  private readonly abi: Abi;

  constructor(functionInterfaces: string[]) {
    this.abi = parseAbi(functionInterfaces);
  }

  public decode(calldata: Hex): DecodedCalldata | null {
    if (!calldata || calldata === "0x" || size(calldata) < 4) {
      return null;
    }
    try {
      const decoded = decodeFunctionData({ abi: this.abi, data: calldata });

      return {
        name: decoded.functionName,
        args: decoded.args,
      };
    } catch {
      return null;
    }
  }
}
