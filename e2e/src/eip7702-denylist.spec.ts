import { beforeAll, describe, expect, it } from "@jest/globals";
import { appendFileSync, readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import { encodeFunctionData, getAddress, isAddress } from "viem";

import { estimateLineaGas, expectSuccessfulTransaction, sendTransactionWithRetry } from "./common/utils";
import { L2RpcEndpoint } from "./config/clients/l2-client";
import { createTestContext } from "./config/setup";
import { TestEIP7702DelegationAbi, TestEIP7702DelegationAbiBytecode } from "./generated";

const DENY_LIST_PATH = resolve(__dirname, "../..", "docker/config/linea-besu-sequencer/deny-list.txt");

const POOL_VALIDATOR_PLUGIN = "net.consensys.linea.sequencer.txpoolvalidation.LineaTransactionPoolValidatorPlugin";

const context = createTestContext();
const l2AccountManager = context.getL2AccountManager();

async function reloadDenyList(client: any): Promise<void> {
  await client.request({
    method: "plugins_reloadPluginConfig",
    params: [POOL_VALIDATOR_PLUGIN],
  });
}

// Appends addresses to deny-list using atomic append (no read-modify-write).
function addToDenyList(addresses: string[]): void {
  const data = addresses.map((a) => a.toLowerCase()).join("\n") + "\n";
  appendFileSync(DENY_LIST_PATH, data);
}

// Removes only the specified addresses from deny-list, leaving other entries intact.
function removeFromDenyList(addresses: string[]): void {
  const current = readFileSync(DENY_LIST_PATH, "utf-8");
  const toRemove = new Set(addresses.map((a) => a.toLowerCase()));
  const remaining = current
    .split("\n")
    .filter(Boolean)
    .filter((a) => !toRemove.has(a.toLowerCase()));
  writeFileSync(DENY_LIST_PATH, remaining.length ? remaining.join("\n") + "\n" : "");
}

async function withDenyListAddresses(client: any, addresses: string[], run: () => Promise<void>): Promise<void> {
  addToDenyList(addresses);
  await reloadDenyList(client);

  try {
    await run();
  } finally {
    removeFromDenyList(addresses);
    await reloadDenyList(client);
  }
}

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

// deny-list.txt is a shared file modified by this suite. Tests within this suite
// MUST NOT use it.concurrent() because removeFromDenyList uses read-modify-write.
// Running concurrently with other test suites (e.g., eip7702.spec.ts) is safe
// because they don't touch deny-list.txt and use independently generated accounts.
describe("EIP-7702 denylist test suite", () => {
  const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
  const initializeData = encodeFunctionData({
    abi: TestEIP7702DelegationAbi,
    functionName: "initialize",
  });

  let targetContractAddress: `0x${string}`;

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
    // Arrange
    const [deployer] = await l2AccountManager.generateAccounts(1);
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

    // Assert
    expect(deployReceipt.status).toEqual("success");
    expect(deployReceipt.contractAddress).toBeDefined();
    expect(isAddress(deployReceipt.contractAddress as `0x${string}`)).toBe(true);
    targetContractAddress = getAddress(deployReceipt.contractAddress as `0x${string}`);

    logger.debug(`TestEIP7702Delegation deployed. address=${targetContractAddress}`);
  }, 120_000);

  it("should block EIP-7702 tx when authorization_list authority is denylisted", async () => {
    // Arrange
    // Path 1 (Puppet Bypass): A non-denylisted sponsor submits a Type 4 tx.
    // The denylisted authority only appears in the authorization_list.
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedAuthority);

    await withDenyListAddresses(l2PublicClient, [scenario.denyListAddress], async () => {
      // Act
      logger.debug(`Authority address added to deny list. address=${scenario.denyListAddress}`);
      // Sponsor submits tx on behalf of denylisted authority.
      // tx.from = sponsor (clean), authorization_list contains denylisted authority.
      const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

      // Assert
      await expectBlockedTransaction(sendTransactionPromise);
      logger.debug("EIP-7702 transaction correctly rejected for denied authority.");
    });

    logger.debug("Authority address removed from deny list.");
  }, 120_000);

  it("should block EIP-7702 tx when authorization_list delegates to denylisted contract", async () => {
    // Arrange
    // Path 2 (Parasite Bypass): A non-denylisted user delegates their EOA
    // to a denylisted contract. The denylisted contract is not tx.to or tx.from,
    // only referenced as the delegation target in authorization_list.
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedContract);
    await withDenyListAddresses(l2PublicClient, [scenario.denyListAddress], async () => {
      // Act
      logger.debug(`Contract address added to deny list. address=${scenario.denyListAddress}`);
      // Delegator sends tx delegating to the denylisted contract.
      // tx.from = delegator (clean), tx.to = delegator (clean),
      // authorization_list delegates to denylisted contract.
      const sendTransactionPromise = scenario.sendDelegatedInitializeTx();

      // Assert
      await expectBlockedTransaction(sendTransactionPromise);
      logger.debug("EIP-7702 transaction correctly rejected for denied contract delegation.");
    });

    logger.debug("Contract address removed from deny list.");
  }, 120_000);

  it("should allow EIP-7702 tx after denylisted authority is removed and plugin config is reloaded", async () => {
    // Arrange
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedAuthority);

    addToDenyList([scenario.denyListAddress]);
    await reloadDenyList(l2PublicClient);

    try {
      // Act
      await expectBlockedTransaction(scenario.sendDelegatedInitializeTx());

      removeFromDenyList([scenario.denyListAddress]);
      await reloadDenyList(l2PublicClient);

      // Assert
      await expectSuccessfulTransaction(l2PublicClient, scenario.sendDelegatedInitializeTx());
    } finally {
      removeFromDenyList([scenario.denyListAddress]);
      await reloadDenyList(l2PublicClient);
    }
  }, 120_000);

  it("should block EIP-7702 tx after non-denied authority is added to denylist and plugin config is reloaded", async () => {
    // Arrange
    const scenario = await createDelegationScenario(DelegationScenarioType.DenylistedAuthority);

    try {
      // Act
      await expectSuccessfulTransaction(l2PublicClient, scenario.sendDelegatedInitializeTx());

      addToDenyList([scenario.denyListAddress]);
      await reloadDenyList(l2PublicClient);

      // Assert
      await expectBlockedTransaction(scenario.sendDelegatedInitializeTx());
    } finally {
      removeFromDenyList([scenario.denyListAddress]);
      await reloadDenyList(l2PublicClient);
    }
  }, 120_000);
});
