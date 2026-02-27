import { beforeAll, describe, expect, it } from "@jest/globals";
import { appendFileSync, readFileSync, writeFileSync } from "fs";
import { resolve } from "path";
import { encodeFunctionData, getAddress } from "viem";

import { estimateLineaGas, sendTransactionWithRetry } from "./common/utils";
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

// deny-list.txt is a shared file modified by this suite. Tests within this suite
// MUST NOT use it.concurrent() because removeFromDenyList uses read-modify-write.
// Running concurrently with other test suites (e.g., eip7702.spec.ts) is safe
// because they don't touch deny-list.txt and use independently generated accounts.
describe("EIP-7702 denylist test suite", () => {
  const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  let targetContractAddress: `0x${string}`;

  beforeAll(async () => {
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

    expect(deployReceipt.status).toEqual("success");
    expect(deployReceipt.contractAddress).toBeTruthy();
    targetContractAddress = getAddress(deployReceipt.contractAddress!);

    logger.debug(`TestEIP7702Delegation deployed. address=${targetContractAddress}`);
  }, 120_000);

  it("should block EIP-7702 tx when authorization_list authority is denylisted", async () => {
    // Path 1 (Puppet Bypass): A non-denylisted sponsor submits a Type 4 tx.
    // The denylisted authority only appears in the authorization_list.
    const [sponsor, authority] = await l2AccountManager.generateAccounts(2);

    const authorityWalletClient = context.l2WalletClient({ account: authority });

    // Authority signs authorization - no executor:"self" so sponsor can submit it
    const authorization = await authorityWalletClient.signAuthorization({
      contractAddress: targetContractAddress,
    });

    logger.debug(
      `EIP-7702 authorization signed. authorityAddress=${authority.address} target=${targetContractAddress}`,
    );

    const initializeData = encodeFunctionData({
      abi: TestEIP7702DelegationAbi,
      functionName: "initialize",
    });

    const sponsorWalletClient = context.l2WalletClient({ account: sponsor });
    const sponsorNonce = await l2PublicClient.getTransactionCount({ address: sponsor.address });

    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: sponsor,
      to: targetContractAddress,
      data: initializeData,
    });

    addToDenyList([authority.address]);
    await reloadDenyList(l2PublicClient);

    logger.debug(`Authority address added to deny list. address=${authority.address}`);

    try {
      // Sponsor submits tx on behalf of denylisted authority.
      // tx.from = sponsor (clean), authorization_list contains denylisted authority.
      await expect(
        sponsorWalletClient.sendTransaction({
          authorizationList: [authorization],
          to: authority.address,
          data: initializeData,
          nonce: sponsorNonce,
          gas: 100_000n,
          maxFeePerGas,
          maxPriorityFeePerGas,
        }),
      ).rejects.toThrow("blocked");

      logger.debug("EIP-7702 transaction correctly rejected for denied authority.");
    } finally {
      removeFromDenyList([authority.address]);
      await reloadDenyList(l2PublicClient);
      logger.debug("Authority address removed from deny list.");
    }
  }, 120_000);

  it("should block EIP-7702 tx when authorization_list delegates to denylisted contract", async () => {
    // Path 2 (Parasite Bypass): A non-denylisted user delegates their EOA
    // to a denylisted contract. The denylisted contract is not tx.to or tx.from,
    // only referenced as the delegation target in authorization_list.
    const [delegator] = await l2AccountManager.generateAccounts(1);

    const delegatorWalletClient = context.l2WalletClient({ account: delegator });

    const authorization = await delegatorWalletClient.signAuthorization({
      contractAddress: targetContractAddress,
      executor: "self",
    });

    logger.debug(
      `EIP-7702 authorization signed. delegatorAddress=${delegator.address} target=${targetContractAddress}`,
    );

    const initializeData = encodeFunctionData({
      abi: TestEIP7702DelegationAbi,
      functionName: "initialize",
    });

    const delegatorNonce = await l2PublicClient.getTransactionCount({ address: delegator.address });

    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: delegator,
      to: targetContractAddress,
      data: initializeData,
    });

    addToDenyList([targetContractAddress]);
    await reloadDenyList(l2PublicClient);

    logger.debug(`Contract address added to deny list. address=${targetContractAddress}`);

    try {
      // Delegator sends tx delegating to the denylisted contract.
      // tx.from = delegator (clean), tx.to = delegator (clean),
      // authorization_list delegates to denylisted contract.
      await expect(
        delegatorWalletClient.sendTransaction({
          authorizationList: [authorization],
          to: delegator.address,
          data: initializeData,
          nonce: delegatorNonce,
          gas: 100_000n,
          maxFeePerGas,
          maxPriorityFeePerGas,
        }),
      ).rejects.toThrow("blocked");

      logger.debug("EIP-7702 transaction correctly rejected for denied contract delegation.");
    } finally {
      removeFromDenyList([targetContractAddress]);
      await reloadDenyList(l2PublicClient);
      logger.debug("Contract address removed from deny list.");
    }
  }, 120_000);
});
