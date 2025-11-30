import { describe, expect, it } from "@jest/globals";
import { etherToWei, getTransactionHash, serialize, wait } from "./common/utils";
import { config } from "./config/tests-config/setup";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";
import { BaseError, encodeFunctionData } from "viem";
import { TestContractAbi } from "./generated";

const l2AccountManager = config.getL2AccountManager();

describe("Transaction exclusion test suite", () => {
  it.concurrent(
    "Should get the status of the rejected transaction reported from Besu RPC node",
    async () => {
      const transactionExclusionClient = config.l2PublicClient({ type: L2RpcEndpoint.TransactionExclusion });

      const l2Account = await l2AccountManager.generateAccount();

      const l2WalletClient = config.l2WalletClient({ type: L2RpcEndpoint.BesuNode, account: l2Account });
      const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

      const txRequest = {
        account: l2Account,
        to: config.l2PublicClient().getTestContract().address,
        data: encodeFunctionData({
          abi: TestContractAbi,
          functionName: "testAddmod",
          args: [13000n, 31n],
        }),
        maxPriorityFeePerGas: etherToWei("0.000000001"), // 1 Gwei
        maxFeePerGas: etherToWei("0.00000001"), // 10 Gwei
      };
      const rejectedTxHash = await getTransactionHash(l2PublicClient, txRequest);

      try {
        // This shall be rejected by the Besu node due to traces module limit overflow
        await l2WalletClient.sendTransaction(txRequest);
      } catch (err) {
        if (err instanceof BaseError) {
          // This shall return error with traces limit overflow
          logger.debug(`sendTransaction expected rejection: ${serialize(err)}`);
          // assert it was indeed rejected by the traces module limit
          expect(err.message).toContain("is above the limit");
        }
        throw new Error("Transaction was expected to be rejected, but it was not.");
      }

      expect(rejectedTxHash).toBeDefined();
      logger.debug(`Transaction rejected as expected (RPC). transactionHash=${rejectedTxHash}`);

      let getResponse;
      do {
        await wait(1_000);
        getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1({ txHash: rejectedTxHash! });
      } while (!getResponse);

      logger.debug(`Transaction exclusion status received. response=${serialize(getResponse)}`);

      expect(getResponse.txHash).toStrictEqual(rejectedTxHash);
      expect(getResponse.txRejectionStage).toStrictEqual("RPC");
      expect(getResponse.from.toLowerCase()).toStrictEqual(l2Account.address.toLowerCase());
    },
    120_000,
  );

  it.skip("Should get the status of the rejected transaction reported from Besu SEQUENCER node", async () => {
    const transactionExclusionClient = config.l2PublicClient({ type: L2RpcEndpoint.TransactionExclusion });
    const l2Account = await l2AccountManager.generateAccount();
    const testContract = config.l2WalletClient({ type: L2RpcEndpoint.Sequencer, account: l2Account }).getTestContract();

    // This shall be rejected by sequencer due to traces module limit overflow
    const rejectedTxHash = await testContract.write.testAddmod([13000n, 31n]);
    logger.debug(`Transaction rejected as expected (SEQUENCER). transactionHash=${rejectedTxHash}`);

    let getResponse;
    do {
      await wait(1_000);
      getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1({ txHash: rejectedTxHash! });
    } while (!getResponse);

    logger.debug(`Transaction exclusion status received. response=${serialize(getResponse)}`);

    expect(getResponse.txHash).toStrictEqual(rejectedTxHash);
    expect(getResponse.txRejectionStage).toStrictEqual("SEQUENCER");
    expect(getResponse.from.toLowerCase()).toStrictEqual(l2Account.address.toLowerCase());
  }, 120_000);
});
