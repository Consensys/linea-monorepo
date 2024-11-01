import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { etherToWei, getTransactionHash, getWallet, TransactionExclusionClient, wait } from "./common/utils";
import { TransactionRequest } from "ethers";

const l2AccountManager = config.getL2AccountManager();

describe("Transaction exclusion test suite", () => {
  it.concurrent(
    "Should get the status of the rejected transaction reported from Besu RPC node",
    async () => {
      const transactionExclusionEndpoint = config.getTransactionExclusionEndpoint();
      expect(transactionExclusionEndpoint).toBeDefined();

      const transactionExclusionClient = new TransactionExclusionClient(transactionExclusionEndpoint!);
      const l2Account = await l2AccountManager.generateAccount();
      const l2AccountLocal = getWallet(l2Account.privateKey, config.getL2BesuNodeProvider()!);
      const testContract = config.getL2TestContract(l2AccountLocal)!;

      // This shall be rejected by the Besu node due to traces module limit overflow
      let rejectedTxHash;
      try {
        const txRequest: TransactionRequest = {
          to: await testContract.getAddress(),
          data: testContract.interface.encodeFunctionData("testAddmod", [13000, 31]),
          maxPriorityFeePerGas: etherToWei("0.000000001"), // 1 Gwei
          maxFeePerGas: etherToWei("0.00000001"), // 10 Gwei
        };
        rejectedTxHash = await getTransactionHash(txRequest, l2AccountLocal);
        await l2AccountLocal.sendTransaction(txRequest);
      } catch (err) {
        // This shall return error with traces limit overflow
        console.debug(`sendTransaction expected err: ${JSON.stringify(err)}`);
      }

      expect(rejectedTxHash).toBeDefined();
      console.log(`rejectedTxHash (RPC): ${rejectedTxHash}`);

      let getResponse;
      do {
        await wait(1_000);
        getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(rejectedTxHash!);
      } while (!getResponse?.result);

      expect(getResponse.result.txHash).toStrictEqual(rejectedTxHash);
      expect(getResponse.result.txRejectionStage).toStrictEqual("RPC");
      expect(getResponse.result.from.toLowerCase()).toStrictEqual(l2AccountLocal.address.toLowerCase());
    },
    120_000,
  );

  it.skip("Should get the status of the rejected transaction reported from Besu SEQUENCER node", async () => {
    expect(config.getTransactionExclusionEndpoint()).toBeDefined();

    const transactionExclusionClient = new TransactionExclusionClient(config.getTransactionExclusionEndpoint()!);
    const l2Account = await l2AccountManager.generateAccount();
    const l2AccountLocal = getWallet(l2Account.privateKey, config.getL2SequencerProvider()!);
    const testContract = config.getL2TestContract(l2AccountLocal);

    // This shall be rejected by sequencer due to traces module limit overflow
    const tx = await testContract!.connect(l2AccountLocal).testAddmod(13000, 31);
    const rejectedTxHash = tx.hash;
    console.log(`rejectedTxHash (SEQUENCER): ${rejectedTxHash}`);

    let getResponse;
    do {
      await wait(1_000);
      getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(rejectedTxHash);
    } while (!getResponse?.result);

    expect(getResponse.result.txHash).toStrictEqual(rejectedTxHash);
    expect(getResponse.result.txRejectionStage).toStrictEqual("SEQUENCER");
    expect(getResponse.result.from.toLowerCase()).toStrictEqual(l2AccountLocal.address.toLowerCase());
  }, 120_000);
});
