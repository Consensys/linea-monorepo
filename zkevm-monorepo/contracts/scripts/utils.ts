import { ethers } from "ethers";
import { HardhatEthersHelpers } from "hardhat/types";

/**
 * @param provider ethers JsonRpcProvider or HardhatEthersHelpers provider instance
 * @returns {Promise<{maxPriorityFeePerGas: *, maxFeePerGas: *}>}
 */
async function get1559Fees(
  provider: ethers.JsonRpcProvider | HardhatEthersHelpers["provider"],
): Promise<{ maxPriorityFeePerGas?: bigint; maxFeePerGas?: bigint }> {
  const { maxPriorityFeePerGas, maxFeePerGas } = await provider.getFeeData();

  return {
    ...(maxPriorityFeePerGas ? { maxPriorityFeePerGas } : {}),
    ...(maxFeePerGas ? { maxFeePerGas } : {}),
  };
}

export { get1559Fees };
