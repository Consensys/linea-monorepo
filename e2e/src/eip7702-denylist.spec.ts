import { describe, expect, it } from "@jest/globals";
import { readFileSync, writeFileSync } from "fs";
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

function writeDenyList(addresses: string[]): void {
  writeFileSync(DENY_LIST_PATH, addresses.length ? addresses.join("\n") + "\n" : "");
}

describe("EIP-7702 denylist test suite", () => {
  const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  it("should block EIP-7702 delegation transaction when authority is on the denylist", async () => {
    const [deployer, authority] = await l2AccountManager.generateAccounts(2);

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
    const targetContractAddress = getAddress(deployReceipt.contractAddress!);

    logger.debug(`TestEIP7702Delegation deployed. address=${targetContractAddress}`);

    const authorityWalletClient = context.l2WalletClient({ account: authority });

    const authorization = await authorityWalletClient.signAuthorization({
      contractAddress: targetContractAddress,
      executor: "self",
    });

    logger.debug(
      `EIP-7702 authorization signed. authorityAddress=${authority.address} target=${targetContractAddress}`,
    );

    const initializeData = encodeFunctionData({
      abi: TestEIP7702DelegationAbi,
      functionName: "initialize",
    });

    const authorityNonce = await l2PublicClient.getTransactionCount({ address: authority.address });

    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(l2PublicClient, {
      account: deployer,
      to: targetContractAddress,
      data: initializeData,
    });

    const originalDenyList = readFileSync(DENY_LIST_PATH, "utf-8");
    writeDenyList([authority.address.toLowerCase()]);
    await reloadDenyList(l2PublicClient);

    logger.debug(`Authority address added to deny list. address=${authority.address}`);

    try {
      await expect(
        authorityWalletClient.sendTransaction({
          authorizationList: [authorization],
          to: authority.address,
          data: initializeData,
          nonce: authorityNonce,
          gas: 100_000n,
          maxFeePerGas,
          maxPriorityFeePerGas,
        }),
      ).rejects.toThrow("blocked");

      logger.debug("EIP-7702 transaction correctly rejected for denied authority.");
    } finally {
      writeFileSync(DENY_LIST_PATH, originalDenyList);
      await reloadDenyList(l2PublicClient);
      logger.debug("Deny list restored to original state.");
    }
  }, 120_000);
});
