import { decodeFunctionData, type Hex, type Abi } from "viem";

import type { ICalldataDecoder, DecodedCalldata } from "../../../domain/ports/ICalldataDecoder";

export class ViemCalldataDecoder implements ICalldataDecoder {
  private readonly abis: Abi[];

  constructor(functionInterfaces: string[]) {
    this.abis = functionInterfaces.map((iface) => [
      {
        type: "function" as const,
        name: iface.split("(")[0].trim(),
        inputs: this.parseInputs(iface),
        outputs: [],
        stateMutability: "nonpayable" as const,
      },
    ]);
  }

  public decode(calldata: string): DecodedCalldata | null {
    if (!calldata || calldata === "0x" || calldata.length < 10) {
      return null;
    }

    for (const abi of this.abis) {
      try {
        const decoded = decodeFunctionData({ abi, data: calldata as Hex });

        if (decoded) {
          const args: Record<string, unknown> = {};
          const func = abi[0];

          if ("inputs" in func && func.inputs) {
            (func.inputs as readonly { name?: string }[]).forEach((input: { name?: string }, index: number) => {
              args[input.name ?? `arg${index}`] = decoded.args?.[index];
            });
          }

          return {
            name: decoded.functionName,
            args,
          };
        }
      } catch {
        continue;
      }
    }

    return null;
  }

  private parseInputs(iface: string): { name: string; type: string }[] {
    const match = iface.match(/\(([^)]*)\)/);
    if (!match || !match[1]) return [];

    return match[1].split(",").map((param, index) => {
      const parts = param.trim().split(/\s+/);
      return {
        type: parts[0],
        name: parts.length > 1 ? parts[parts.length - 1] : `arg${index}`,
      };
    });
  }
}
