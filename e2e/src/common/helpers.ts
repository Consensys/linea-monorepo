import { ethers } from "ethers";

export function getAndIncreaseFeeData(feeData: ethers.FeeData): [bigint, bigint, bigint] {
  const maxPriorityFeePerGas = BigInt((parseFloat(feeData.maxPriorityFeePerGas!.toString()) * 1.1).toFixed(0));
  const maxFeePerGas = BigInt((parseFloat(feeData.maxFeePerGas!.toString()) * 1.1).toFixed(0));
  const gasPrice = BigInt((parseFloat(feeData.gasPrice!.toString()) * 1.1).toFixed(0));
  return [maxPriorityFeePerGas, maxFeePerGas, gasPrice];
}
