/* eslint-disable no-var */
import { config } from "../tests-config";
import { deployContract } from "../../common/deployments";
import { DummyContract__factory } from "../../typechain";
import { etherToWei, sendTransactionsToGenerateTrafficWithInterval } from "../../common/utils";

declare global {
  var stopL2TrafficGeneration: () => void;
}

export default async (): Promise<void> => {
  const l1AccountManager = config.getL1AccountManager();
  const l2AccountManager = config.getL2AccountManager();

  const l1TokenBridge = config.getL1TokenBridgeContract();
  const l2TokenBridge = config.getL2TokenBridgeContract();
  const l1SecurityCouncil = l1AccountManager.whaleAccount(3);
  const l2SecurityCouncil = l2AccountManager.whaleAccount(3);
  console.log("l2SecurityCouncil.address", l2SecurityCouncil.address);

  const account = l1AccountManager.whaleAccount(0);
  const l2Account = l2AccountManager.whaleAccount(0);

  const [dummyContract, l2DummyContract] = await Promise.all([
    deployContract(new DummyContract__factory(), account),
    deployContract(new DummyContract__factory(), l2Account),
  ]);

  console.log(`L1 Dummy contract deployed at address: ${await dummyContract.getAddress()}`);
  console.log(`L2 Dummy contract deployed at address: ${await l2DummyContract.getAddress()}`);

  // Send ETH to the LineaRollup contract
  const lineaRollup = config.getLineaRollupContract(account);
  const l1JsonRpcProvider = config.getL1Provider();
  const l2JsonRpcProvider = config.getL2Provider();

  const value = etherToWei("500");
  const fee = etherToWei("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";
  const { maxPriorityFeePerGas, maxFeePerGas } = await l1JsonRpcProvider.getFeeData();
  const tx = await lineaRollup.sendMessage(to, fee, calldata, { value, maxPriorityFeePerGas, maxFeePerGas });
  await tx.wait();

  console.log("Generating L2 traffic...");
  const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(l2Account, 2_000);

  // Setting the Remote TokenBridge
  console.log("Setting the TokenBridge L1 Remote");
  await (await l1TokenBridge.connect(l1SecurityCouncil).setRemoteTokenBridge(await l2TokenBridge.getAddress())).wait();
  let remoteSender = await l1TokenBridge.remoteSender();

  console.log("L1 TokenBridge remote sender :", remoteSender);
  const l1TokenBridgeAddress = await l1TokenBridge.getAddress();

  console.log("Setting the TokenBridge L2 remote");

  const { maxPriorityFeePerGas: l2MaxPriorityFeePerGas, maxFeePerGas: l2MaxFeePerGas } =
    await l2JsonRpcProvider.getFeeData();

  const setRemoteTx = await l2TokenBridge.connect(l2SecurityCouncil).setRemoteTokenBridge(l1TokenBridgeAddress, {
    maxPriorityFeePerGas: l2MaxPriorityFeePerGas,
    maxFeePerGas: l2MaxFeePerGas,
  });

  await setRemoteTx.wait();

  remoteSender = await l2TokenBridge.remoteSender();
  console.log("L2 TokenBridge remote sender :", remoteSender);

  global.stopL2TrafficGeneration = stopPolling;
};
