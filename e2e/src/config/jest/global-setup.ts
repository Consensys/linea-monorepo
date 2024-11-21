/* eslint-disable no-var */
import { ethers } from "ethers";
import { config } from "../tests-config";
import { deployContract } from "../../common/deployments";
import { DummyContract__factory, TestContract__factory } from "../../typechain";
import { etherToWei, sendTransactionsToGenerateTrafficWithInterval } from "../../common/utils";

declare global {
  var stopL2TrafficGeneration: () => void;
}

export default async (): Promise<void> => {
  const l1AccountManager = config.getL1AccountManager();
  const l2AccountManager = config.getL2AccountManager();

  const account = config.getL1AccountManager().whaleAccount(0);
  const l2Account = config.getL2AccountManager().whaleAccount(0);
  const lineaRollup = config.getLineaRollupContract(account);

  const l1TokenBridge = config.getL1TokenBridgeContract();
  const l2TokenBridge = config.getL2TokenBridgeContract();
  const l1SecurityCouncil = l1AccountManager.whaleAccount(3);
  const l2SecurityCouncil = l2AccountManager.whaleAccount(3);

  const [l1AccountNonce, l2AccountNonce] = await Promise.all([account.getNonce(), l2Account.getNonce()]);

  const fee = etherToWei("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";

  const [dummyContract, l2DummyContract, l2TestContract] = await Promise.all([
    deployContract(new DummyContract__factory(), account, [{ nonce: l1AccountNonce }]),
    deployContract(new DummyContract__factory(), l2Account, [{ nonce: l2AccountNonce }]),
    deployContract(new TestContract__factory(), l2Account, [{ nonce: l2AccountNonce + 1 }]),
    // Send ETH to the LineaRollup contract
    (
      await lineaRollup.sendMessage(to, fee, calldata, {
        value: etherToWei("500"),
        gasPrice: ethers.parseUnits("300", "gwei"),
        nonce: l1AccountNonce + 1,
      })
    ).wait(),
    (
      await l1TokenBridge.connect(l1SecurityCouncil).setRemoteTokenBridge(await l2TokenBridge.getAddress(), {
        gasPrice: ethers.parseUnits("300", "gwei"),
      })
    ).wait(),
    (
      await l2TokenBridge.connect(l2SecurityCouncil).setRemoteTokenBridge(await l1TokenBridge.getAddress(), {
        gasPrice: ethers.parseUnits("300", "gwei"),
      })
    ).wait(),
  ]);

  console.log(`L1 Dummy contract deployed at address: ${await dummyContract.getAddress()}`);
  console.log(`L2 Dummy contract deployed at address: ${await l2DummyContract.getAddress()}`);
  console.log(`L2 Test contract deployed at address: ${await l2TestContract.getAddress()}`);

  console.log("Generating L2 traffic...");
  const pollingAccount = await config.getL2AccountManager().generateAccount(etherToWei("200"));
  const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(pollingAccount, 2_000);

  global.stopL2TrafficGeneration = stopPolling;
};
