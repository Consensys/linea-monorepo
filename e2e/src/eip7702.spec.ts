import { describe, expect, it } from "@jest/globals";
import { encodeFunctionData, getAddress, isAddress } from "viem";

import {
  addToDenyList,
  expectBlockedTransaction,
  reloadDenyList,
  removeFromDenyList,
  withDenyListAddresses,
} from "./common/test-helpers";
import { estimateLineaGas, expectSuccessfulTransaction, sendTransactionWithRetry } from "./common/utils";
import { L2RpcEndpoint } from "./config/clients/l2-client";
import { createTestContext } from "./config/setup";
import { Eip7702TestEntrypointAbi, TestEIP7702DelegationAbi, TestEIP7702DelegationAbiBytecode } from "./generated";

const context = createTestContext();
const l2AccountManager = context.getL2AccountManager();

/** Predeployed EIP-7702 contract addresses from local stack (deployEIP7702Contracts). */
const EIP7702_NESTED = getAddress("0xE4392c8ecC46b304C83cDB5edaf742899b1bda93");
const EIP7702_DELEGATED = getAddress("0x997FC3aF1F193Cbdc013060076c67A13e218980e");
const EIP7702_ENTRYPOINT = getAddress("0xFCc2155b495B6Bf6701eb322D3a97b7817898306");

enum DelegationScenarioType {
  DenylistedAuthority,
  DenylistedContract,
}

type DelegationScenario = {
  denyListAddress: `0x${string}`;
  sendDelegatedInitializeTx: () => Promise<`0x${string}`>;
};

type CreateDelegationScenarioParams = {
  scenarioType: DelegationScenarioType;
  contractAddress: `0x${string}`;
};

describe("EIP-7702 test suite", () => {
  const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
  const sequencerClient = context.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
  const initializeData = encodeFunctionData({
    abi: TestEIP7702DelegationAbi,
    functionName: "initialize",
  });

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

  async function createDelegationScenario(params: CreateDelegationScenarioParams): Promise<DelegationScenario> {
    const { scenarioType, contractAddress } = params;
    const isDenylistedAuthorityCase = scenarioType === DelegationScenarioType.DenylistedAuthority;
    const [authority] = await l2AccountManager.generateAccounts(1);
    const [sponsor] = isDenylistedAuthorityCase ? await l2AccountManager.generateAccounts(1) : [authority];
    const authorityWalletClient = context.l2WalletClient({ account: authority });
    const sponsorWalletClient = context.l2WalletClient({ type: L2RpcEndpoint.Sequencer, account: sponsor });
    const authorization = await authorityWalletClient.signAuthorization({
      contractAddress,
      // Self-sponsored tx if target address is denylisted
      executor: isDenylistedAuthorityCase ? undefined : "self",
    });
    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: sponsor,
      to: contractAddress,
      data: initializeData,
    });

    return {
      denyListAddress: isDenylistedAuthorityCase ? authority.address : contractAddress,
      sendDelegatedInitializeTx: async () => {
        const nonce = await sequencerClient.getTransactionCount({ address: sponsor.address });

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

  it.concurrent("should execute EIP-7702 (Set Code) transactions", async () => {
    const [deployer] = await l2AccountManager.generateAccounts(1);
    const contractAddress = await deployDelegationContract(deployer);
    const scenario = await createDelegationScenario({
      scenarioType: DelegationScenarioType.DenylistedContract,
      contractAddress,
    });

    await expectSuccessfulTransaction(l2PublicClient, scenario.sendDelegatedInitializeTx());
  });

  it.concurrent("should execute EIP-7702 self-call with no calldata when delegating to a codeless EOA", async () => {
    const [accountA, accountB] = await l2AccountManager.generateAccounts(2);
    const accountAWalletClient = context.l2WalletClient({ account: accountA });

    const authorization = await accountAWalletClient.signAuthorization({
      contractAddress: accountB.address,
      executor: "self",
    });

    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: accountA,
      to: accountA.address,
      data: "0x",
    });

    const nonce = await l2PublicClient.getTransactionCount({ address: accountA.address });

    await expectSuccessfulTransaction(
      l2PublicClient,
      accountAWalletClient.sendTransaction({
        authorizationList: [authorization],
        to: accountA.address,
        data: "0x",
        nonce,
        gas: 100_000n,
        maxFeePerGas,
        maxPriorityFeePerGas,
      }),
    );
  });

  it.concurrent(
    "should block EIP-7702 tx when authorization_list authority is denylisted",
    async () => {
      const [deployer] = await l2AccountManager.generateAccounts(1);
      const contractAddress = await deployDelegationContract(deployer);
      const scenario = await createDelegationScenario({
        scenarioType: DelegationScenarioType.DenylistedAuthority,
        contractAddress,
      });

      await withDenyListAddresses(sequencerClient, [scenario.denyListAddress], async () => {
        logger.debug(`Authority address added to deny list. address=${scenario.denyListAddress}`);
        const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

        await expectBlockedTransaction(sendTransactionPromise);
        logger.debug("EIP-7702 transaction correctly rejected for denied authority.");
      });

      logger.debug("Authority address removed from deny list.");
    },
    120_000,
  );

  it.concurrent(
    "should block EIP-7702 tx when authorization_list delegates to denylisted contract",
    async () => {
      const [deployer] = await l2AccountManager.generateAccounts(1);
      const contractAddress = await deployDelegationContract(deployer);
      const scenario = await createDelegationScenario({
        scenarioType: DelegationScenarioType.DenylistedContract,
        contractAddress,
      });
      await withDenyListAddresses(sequencerClient, [scenario.denyListAddress], async () => {
        logger.debug(`Contract address added to deny list. address=${scenario.denyListAddress}`);
        const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

        await expectBlockedTransaction(sendTransactionPromise);
        logger.debug("EIP-7702 transaction correctly rejected for denied contract delegation.");
      });

      logger.debug("Contract address removed from deny list.");
    },
    120_000,
  );

  it.concurrent(
    "should block EIP-7702 tx after non-denied authority is added to denylist and plugin config is reloaded",
    async () => {
      const [deployer] = await l2AccountManager.generateAccounts(1);
      const contractAddress = await deployDelegationContract(deployer);
      const scenario = await createDelegationScenario({
        scenarioType: DelegationScenarioType.DenylistedAuthority,
        contractAddress,
      });

      try {
        await expectSuccessfulTransaction(l2PublicClient, scenario.sendDelegatedInitializeTx());

        addToDenyList([scenario.denyListAddress]);
        await reloadDenyList(sequencerClient);

        await expectBlockedTransaction(scenario.sendDelegatedInitializeTx());
      } finally {
        removeFromDenyList([scenario.denyListAddress]);
        await reloadDenyList(sequencerClient);
      }
    },
    120_000,
  );

  it.only("should block EIP-7702 entrypoint tx when authority is denylisted and allow after removal", async () => {
    const [accountA, accountB] = await l2AccountManager.generateAccounts(2);
    const accountAWalletClient = context.l2WalletClient({ account: accountA });
    const accountBWalletClient = context.l2WalletClient({ type: L2RpcEndpoint.Sequencer, account: accountB });

    const authorization = await accountAWalletClient.signAuthorization({
      contractAddress: EIP7702_DELEGATED,
      executor: undefined,
    });

    const setValueData = encodeFunctionData({
      abi: Eip7702TestEntrypointAbi,
      functionName: "setValue",
      args: [42n, accountA.address, EIP7702_NESTED],
    });

    // Use a no-op call for gas estimation; simulating the entrypoint would revert because
    // A has no delegated code during simulation (EIP-7702 auth is only applied at execution).
    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: accountB,
      to: accountB.address,
      data: "0x",
    });

    const nonce1 = await sequencerClient.getTransactionCount({ address: accountB.address });
    const txHash1 = await accountBWalletClient.sendTransaction({
      authorizationList: [authorization],
      to: EIP7702_ENTRYPOINT,
      data: setValueData,
      nonce: nonce1,
      gas: 200_000n,
      maxFeePerGas,
      maxPriorityFeePerGas,
    });

    const receipt1 = await l2PublicClient.waitForTransactionReceipt({ hash: txHash1 });
    expect(receipt1.status).toEqual("success");

    const logs = await l2PublicClient.getContractEvents({
      address: EIP7702_ENTRYPOINT,
      abi: Eip7702TestEntrypointAbi,
      eventName: "ValueSet",
      fromBlock: receipt1.blockNumber,
      toBlock: receipt1.blockNumber,
    });
    expect(logs.length).toBeGreaterThan(0);
    expect(logs.some((log) => log.args.sender === accountB.address && log.args.value === 42n)).toBe(true);

    addToDenyList([accountA.address]);
    await reloadDenyList(sequencerClient);

    const nonce2 = await sequencerClient.getTransactionCount({ address: accountB.address });
    const sendTxPromise = accountBWalletClient.sendTransaction({
      to: EIP7702_ENTRYPOINT,
      data: setValueData,
      nonce: nonce2,
      gas: 200_000n,
      maxFeePerGas,
      maxPriorityFeePerGas,
    });

    await expectBlockedTransaction(sendTxPromise);

    removeFromDenyList([accountA.address]);
    await reloadDenyList(sequencerClient);
  }, 120_000);
});
