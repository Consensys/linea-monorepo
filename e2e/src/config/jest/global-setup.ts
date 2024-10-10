import { getAndIncreaseFeeData } from "../../common/helpers";
import { config } from "../tests-config";
import { deployContract } from "../../common/deployments";
import { DummyContract__factory } from "../../typechain";
import { etherToWei } from "../../common/utils";

export default async (): Promise<void> => {
  const account = config.getL1AccountManager().whaleAccount(0);
  const l2Account = config.getL2AccountManager().whaleAccount(0);

  const [dummyContract, l2DummyContract] = await Promise.all([
    deployContract(new DummyContract__factory(), account),
    deployContract(new DummyContract__factory(), l2Account),
  ]);

  console.log(`L1 Dummy contract deployed at address: ${await dummyContract.getAddress()}`);
  console.log(`L2 Dummy contract deployed at address: ${await l2DummyContract.getAddress()}`);

  // Send ETH to the LineaRollup contract
  const lineaRollup = config.getLineaRollupContract(account);
  const l1JsonRpcProvider = config.getL1Provider();

  const value = etherToWei("500");
  const fee = etherToWei("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";
  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1JsonRpcProvider.getFeeData());
  const tx = await lineaRollup.sendMessage(to, fee, calldata, { value, maxPriorityFeePerGas, maxFeePerGas });
  await tx.wait();
};
