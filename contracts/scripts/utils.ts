import { ethers } from "ethers";

/**
 * @param provider ethers JsonRpcProvider or HardhatEthersHelpers provider instance
 * @returns {Promise<{maxPriorityFeePerGas: *, maxFeePerGas: *}>}
 */
async function get1559Fees(
  provider: ethers.Provider,
): Promise<{ maxPriorityFeePerGas?: bigint; maxFeePerGas?: bigint; gasPrice?: bigint }> {
  const { maxPriorityFeePerGas, maxFeePerGas, gasPrice } = await provider.getFeeData();
  return {
    ...(maxPriorityFeePerGas ? { maxPriorityFeePerGas } : {}),
    ...(maxFeePerGas ? { maxFeePerGas } : {}),
    ...(gasPrice ? { gasPrice } : {}),
  };
}

export { get1559Fees };
