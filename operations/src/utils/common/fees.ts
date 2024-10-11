import { ethers } from "ethers";
import { FeeHistory, Fees } from "./types.js";

export async function get1559Fees(
  provider: ethers.JsonRpcProvider,
  maxFeePerGasFromConfig: bigint,
  percentile: number,
): Promise<Fees> {
  const { reward, baseFeePerGas }: FeeHistory = await provider.send("eth_feeHistory", ["0x4", "latest", [percentile]]);

  const maxPriorityFeePerGas =
    reward.reduce((acc: bigint, currentValue: string[]) => acc + BigInt(currentValue[0]), 0n) / BigInt(reward.length);

  if (maxPriorityFeePerGas && maxPriorityFeePerGas > maxFeePerGasFromConfig) {
    throw new Error(
      `Estimated miner tip of ${maxPriorityFeePerGas} exceeds configured max fee per gas of ${maxFeePerGasFromConfig}.`,
    );
  }

  const maxFeePerGas = BigInt(baseFeePerGas[baseFeePerGas.length - 1]) * 2n + maxPriorityFeePerGas;

  if (maxFeePerGas > 0n && maxPriorityFeePerGas > 0n) {
    return {
      maxPriorityFeePerGas,
      maxFeePerGas: maxFeePerGas > maxFeePerGasFromConfig ? maxFeePerGasFromConfig : maxFeePerGas,
    };
  }

  return {
    maxFeePerGas: maxFeePerGasFromConfig,
  };
}
