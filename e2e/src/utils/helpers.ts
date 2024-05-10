/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { BigNumber, ethers } from "ethers";

export function getAndIncreaseFeeData(feeData: ethers.providers.FeeData): [BigNumber, BigNumber, BigNumber] {
  const maxPriorityFeePerGas = BigNumber.from((parseFloat(feeData.maxPriorityFeePerGas!.toString()) * 1.1).toFixed(0));
  const maxFeePerGas = BigNumber.from((parseFloat(feeData.maxFeePerGas!.toString()) * 1.1).toFixed(0));
  const gasPrice = BigNumber.from((parseFloat(feeData.gasPrice!.toString()) * 1.1).toFixed(0));
  return [maxPriorityFeePerGas, maxFeePerGas, gasPrice];
}
