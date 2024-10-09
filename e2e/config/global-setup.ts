import { ethers } from "ethers";
import { getAndIncreaseFeeData } from "../src/common/helpers";
import { config } from ".";

export default async (): Promise<void> => {
  const l1JsonRpcProvider = config.getL1Provider();

  const account = await config.getL1AccountManager().generateAccount(ethers.parseEther("510"));

  const lineaRollup = config.getLineaRollupContract(account);
  // Send ETH to the LineaRollup contract
  const value = ethers.parseEther("500");
  const fee = ethers.parseEther("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";
  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1JsonRpcProvider.getFeeData());
  const tx = await lineaRollup.sendMessage(to, fee, calldata, { value, maxPriorityFeePerGas, maxFeePerGas });
  await tx.wait();
};
