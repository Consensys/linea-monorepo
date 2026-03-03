import { beforeAll, describe, expect, it } from "@jest/globals";
import { encodeFunctionData, getAddress, isAddress } from "viem";

import {
  addToDenyList,
  reloadDenyList,
  removeFromDenyList,
  withDenyListAddresses,
} from "./common/test-helpers/deny-list";
import { estimateLineaGas, expectSuccessfulTransaction, sendTransactionWithRetry } from "./common/utils";
import { L2RpcEndpoint } from "./config/clients/l2-client";
import { createTestContext } from "./config/setup";
import { TestEIP7702DelegationAbi, TestEIP7702DelegationAbiBytecode } from "./generated";

const context = createTestContext();
const l2AccountManager = context.getL2AccountManager();

async function expectBlockedTransaction(sendTransactionPromise: Promise<`0x${string}`>): Promise<void> {
  await expect(sendTransactionPromise).rejects.toThrow("blocked");
}

enum DelegationScenarioType {
  DenylistedAuthority,
  DenylistedContract,
}

type DelegationScenario = {
  denyListAddress: `0x${string}`;
  sendDelegatedInitializeTx: () => Promise<`0x${string}`>;
};

// deny-list.txt is a shared file modified by this suite. Denylist tests below
// MUST NOT use it.concurrent() because removeFromDenyList uses read-modify-write.
describe("EIP-7702 test suite", () => {
  const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
  const sequencerClient = context.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
  const initializeData = encodeFunctionData({
    abi: TestEIP7702DelegationAbi,
    functionName: "initialize",
  });

  let targetContractAddress: `0x${string}`;

  async function deployDelegationContract(deployer: { address: `0x${string}` }): Promise<`0x${string}`> {
    const deployerWalletClient = context.l2WalletClient({ account: deployer });
    const deployNonce = await l2PublicClient.getTransactionCount({ address: deployer.address });

    const deployEstimate = await estimateLineaGas(l2PublicClient, {
      account: deployer,
      data: TestEIP7702DelegationAbiBytecode,
    });

    const { receipt: deployReceipt } = await sendTransactionWithRetry(l2PublicClient, (fees) =>
      deployerWalletClient.sendTransaction({
        data: TestEIP7702DelegationAbiBytecode,
        nonce: deployNonce,
        ...deployEstimate,
        ...fees,
      }),
    );

    expect(deployReceipt.status).toEqual("success");
    expect(deployReceipt.contractAddress).toBeDefined();
    expect(isAddress(deployReceipt.contractAddress as `0x${string}`)).toBe(true);

    const deployedContractAddress = getAddress(deployReceipt.contractAddress as `0x${string}`);
    logger.debug(`TestEIP7702Delegation deployed. address=${deployedContractAddress}`);
    return deployedContractAddress;
  }

  async function createDelegationScenario(scenarioType: DelegationScenarioType): Promise<DelegationScenario> {
    const isDenylistedAuthorityCase = scenarioType === DelegationScenarioType.DenylistedAuthority;
    const [authority] = await l2AccountManager.generateAccounts(1);
    const [sponsor] = isDenylistedAuthorityCase ? await l2AccountManager.generateAccounts(1) : [authority];
    const authorityWalletClient = context.l2WalletClient({ account: authority });
    const sponsorWalletClient = context.l2WalletClient({ account: sponsor });
    const authorization = await authorityWalletClient.signAuthorization({
      contractAddress: targetContractAddress,
      executor: isDenylistedAuthorityCase ? undefined : "self",
    });
    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: sponsor,
      to: targetContractAddress,
      data: initializeData,
    });

    return {
      denyListAddress: isDenylistedAuthorityCase ? authority.address : targetContractAddress,
      sendDelegatedInitializeTx: async () => {
        const nonce = await l2PublicClient.getTransactionCount({ address: sponsor.address });

        return sponsorWalletClient.sendTransaction({
          authorizationList: [authorization],
          to: authority.address,
          data: initializeData,
          nonce,
          gas: 100_000n,
          maxFeePerGas,
          maxPriorityFeePerGas,
        });
      },
    };
  }

  beforeAll(async () => {
    const [deployer] = await l2AccountManager.generateAccounts(1);
    targetContractAddress = await deployDelegationContract(deployer);
  }, 120_000);

  it.concurrent("should execute EIP-7702 (Set Code) transactions", async () => {
    const [deployer, eoa] = await l2AccountManager.generateAccounts(2);
    const testTargetContractAddress = await deployDelegationContract(deployer);

    const eoaWalletClient = context.l2WalletClient({ account: eoa });
    const authorization = await eoaWalletClient.signAuthorization({
      contractAddress: testTargetContractAddress,
      executor: "self",
    });

    logger.debug(`EIP-7702 authorization signed. eoaAddress=${eoa.address} target=${testTargetContractAddress}`);

    const eoaNonce = await l2PublicClient.getTransactionCount({ address: eoa.address });

    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: deployer,
      to: testTargetContractAddress,
      data: initializeData,
    });

    const txHash = await eoaWalletClient.sendTransaction({
      authorizationList: [authorization],
      to: eoa.address,
      data: initializeData,
      nonce: eoaNonce,
      gas: 100_000n,
      maxFeePerGas,
      maxPriorityFeePerGas,
    });

    logger.debug(`EIP-7702 transaction sent. transactionHash=${txHash}`);

    const receipt = await l2PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: 60_000 });
    logger.debug(`EIP-7702 transaction receipt received. transactionHash=${txHash} status=${receipt.status}`);

    expect(receipt.status).toEqual("success");
  });

  it("should block EIP-7702 tx when authorization_list authority is denylisted", async () => {
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedAuthority);

    await withDenyListAddresses(sequencerClient, [scenario.denyListAddress], async () => {
      logger.debug(`Authority address added to deny list. address=${scenario.denyListAddress}`);
      const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

      await expectBlockedTransaction(sendTransactionPromise);
      logger.debug("EIP-7702 transaction correctly rejected for denied authority.");
    });

    logger.debug("Authority address removed from deny list.");
  }, 120_000);

  it("should block EIP-7702 tx when authorization_list delegates to denylisted contract", async () => {
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedContract);
    await withDenyListAddresses(sequencerClient, [scenario.denyListAddress], async () => {
      logger.debug(`Contract address added to deny list. address=${scenario.denyListAddress}`);
      const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

      await expectBlockedTransaction(sendTransactionPromise);
      logger.debug("EIP-7702 transaction correctly rejected for denied contract delegation.");
    });

    logger.debug("Contract address removed from deny list.");
  }, 120_000);

  it("should block EIP-7702 tx after non-denied authority is added to denylist and plugin config is reloaded", async () => {
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedAuthority);

    try {
      await expectSuccessfulTransaction(l2PublicClient, scenario.sendDelegatedInitializeTx());

      addToDenyList([scenario.denyListAddress]);
      await reloadDenyList(sequencerClient);

      await expectBlockedTransaction(scenario.sendDelegatedInitializeTx());
    } finally {
      removeFromDenyList([scenario.denyListAddress]);
      await reloadDenyList(sequencerClient);
    }
  }, 120_000);
});
