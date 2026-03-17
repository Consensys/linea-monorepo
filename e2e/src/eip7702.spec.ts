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
import {
  Eip7702TestEntrypointAbi,
  Eip7702TestNestedAbi,
  TestEIP7702DelegationAbi,
  TestEIP7702DelegationAbiBytecode,
} from "./generated";

const context = createTestContext();
const l2AccountManager = context.getL2AccountManager();

/** Predeployed EIP-7702 contract addresses from local stack (deployEIP7702Contracts, nonces 14–16). */
const eip7702Addresses = context.getEip7702Addresses();
const EIP7702_NESTED = eip7702Addresses ? getAddress(eip7702Addresses.nested) : null;
const EIP7702_DELEGATED = eip7702Addresses ? getAddress(eip7702Addresses.delegated) : null;
const EIP7702_ENTRYPOINT = eip7702Addresses ? getAddress(eip7702Addresses.entrypoint) : null;

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

  it.concurrent(
    "should block EIP-7702 entrypoint tx when authority is denylisted and allow after removal",
    async () => {
      if (!EIP7702_NESTED || !EIP7702_DELEGATED || !EIP7702_ENTRYPOINT) {
        logger.debug("Skipping: EIP-7702 predeployed addresses not configured (local only).");
        return;
      }
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

      const innerLogs = await l2PublicClient.getContractEvents({
        address: EIP7702_NESTED,
        abi: Eip7702TestEntrypointAbi,
        eventName: "ValueSet",
        fromBlock: receipt1.blockNumber,
        toBlock: receipt1.blockNumber,
      });
      expect(innerLogs.length).toBeGreaterThan(0);
      expect(innerLogs.some((log) => log.args.sender === accountA.address && log.args.value === 42n)).toBe(true);

      addToDenyList([accountA.address]);
      await reloadDenyList(sequencerClient);

      const nonce2 = await sequencerClient.getTransactionCount({ address: accountB.address });

      const setValueDataAlt = encodeFunctionData({
        abi: Eip7702TestEntrypointAbi,
        functionName: "setValue",
        args: [43n, accountA.address, EIP7702_NESTED],
      });

      const txHash = await accountBWalletClient.sendTransaction({
        to: EIP7702_ENTRYPOINT,
        data: setValueDataAlt,
        nonce: nonce2,
        gas: 200_000n,
        maxFeePerGas,
        maxPriorityFeePerGas,
      });

      // make sure the transaction is mined
      await l2PublicClient.waitForTransactionReceipt({ hash: txHash });

      // make sure the value is still 42
      const value = await l2PublicClient.readContract({
        address: EIP7702_NESTED,
        abi: Eip7702TestNestedAbi,
        functionName: "getValue",
        args: [accountA.address],
      });
      expect(value).toBe(42n);

      removeFromDenyList([accountA.address]);
      await reloadDenyList(sequencerClient);
    },
    120_000,
  );
});
