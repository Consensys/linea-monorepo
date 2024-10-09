import { beforeAll, describe, expect, it } from "@jest/globals";
import { TransactionExclusionClient, wait } from "./utils/utils";
import { ethers, Wallet } from "ethers";
import { getAndIncreaseFeeData } from "./utils/helpers";

const transactionExclusionTestSuite = (title: string) => {
  describe(title, () => {
    let transactionExclusionClient: TransactionExclusionClient;
    const rejectedBlockNumber = 12345;
    const overflows = [
      {
        module: "ADD",
        count: 402,
        limit: 70,
      },
      {
        module: "MUL",
        count: 587,
        limit: 400,
      },
    ];
    const transactionRLP =
      "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c496d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8";
    const expectedTxHash = "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7";
    const expectedNonce = 100;
    const expectedFromAddress = "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2";

    beforeAll(async () => {
      if (TRANSACTION_EXCLUSION_ENDPOINT != null) {
        transactionExclusionClient = new TransactionExclusionClient(TRANSACTION_EXCLUSION_ENDPOINT);
      }
    });

    it("Should get the status of the rejected transaction reported from Besu P2P node", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const account = new Wallet(
        L2_DEPLOYER_ACCOUNT_PRIVATE_KEY,
        new ethers.providers.JsonRpcProvider("http://localhost:9045")
      );

      const [nonce, feeData] = await Promise.all([
        l2Provider.getTransactionCount(account.address),
        l2Provider.getFeeData(),
      ]);

      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

      // This shall be rejected by the Besu node due to traces module limit overflow (as reduced traces limits)
      let rejectedTxHash = "";
      try {
        await testContract.connect(account).testAddmod(1200, 31, {
          nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });
      } catch (err) {
        // This shall return SERVER_ERROR with traces limit overflow 
        rejectedTxHash = (err as any).transactionHash
        console.log(`rejectedTxHash: ${JSON.stringify(rejectedTxHash)}`)
      }

      let getResponse;
      do {
        await wait(5_000);
        getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(rejectedTxHash);
      } while (!getResponse?.result);

      expect(getResponse.result.txHash).toStrictEqual(rejectedTxHash);
      expect(getResponse.result.txRejectionStage).toStrictEqual("P2P");
      expect(getResponse.result.from.toLowerCase()).toStrictEqual(account.address.toLowerCase());
    }, 120_000);

    it("Should get the status of the rejected transaction reported from Besu SEQUENCER node", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const account = new Wallet(
        L2_DEPLOYER_ACCOUNT_PRIVATE_KEY,
        new ethers.providers.JsonRpcProvider("http://localhost:8545")
      );

      const [nonce, feeData] = await Promise.all([
        l2Provider.getTransactionCount(account.address),
        l2Provider.getFeeData(),
      ]);

      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

      // This shall be rejected by sequencer due to traces module limit overflow (as reduced traces limits)
      const tx = await testContract.connect(account).testAddmod(1200, 31, {
        nonce,
        maxPriorityFeePerGas,
        maxFeePerGas,
      });

      const rejectedTxHash = tx.hash;
      console.log(`rejectedTxHash: ${rejectedTxHash}`);

      let getResponse;
      do {
        await wait(5_000);
        getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(rejectedTxHash);
      } while (!getResponse?.result);

      expect(getResponse.result.txHash).toStrictEqual(rejectedTxHash);
      expect(getResponse.result.txRejectionStage).toStrictEqual("SEQUENCER");
      expect(getResponse.result.from.toLowerCase()).toStrictEqual(account.address.toLowerCase());
    }, 120_000);

    it("Should save the first rejected transaction from P2P without rejected block number", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const rejectedTimestamp = new Date().toISOString();
      const saveResponse = await transactionExclusionClient.saveRejectedTransactionV1(
        "P2P",
        rejectedTimestamp,
        null,
        transactionRLP,
        "Transaction line count for module ADD=402 is above the limit 1000 (from e2e test)",
        overflows,
      );

      console.log(`saveResponse: ${JSON.stringify(saveResponse)}`);
      expect(saveResponse.result.status).toStrictEqual("SAVED");
      expect(saveResponse.result.txHash).toStrictEqual(expectedTxHash);

      const getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(expectedTxHash);

      console.log(`getResponse: ${JSON.stringify(getResponse)}`);
      expect(getResponse.result.txHash).toStrictEqual(expectedTxHash);
      expect(getResponse.result.txRejectionStage).toStrictEqual("P2P");
      expect(getResponse.result.from).toStrictEqual(expectedFromAddress);
      expect(getResponse.result.nonce).toStrictEqual(`0x${expectedNonce.toString(16)}`);
      expect(getResponse.result.blockNumber).toBeUndefined();
      expect(getResponse.result.timestamp).toStrictEqual(rejectedTimestamp);
    }, 10_000);

    it("Should save the rejected transaction from SEQUENCER with same txHash but different reason message", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const rejectedTimestamp = new Date().toISOString();
      const saveResponse = await transactionExclusionClient.saveRejectedTransactionV1(
        "SEQUENCER",
        rejectedTimestamp,
        rejectedBlockNumber,
        transactionRLP,
        "Transaction line count for module MUL=587 is above the limit 400 (from e2e test)",
        overflows,
      );

      console.log(`saveResponse: ${JSON.stringify(saveResponse)}`);
      expect(saveResponse.result.status).toStrictEqual("SAVED");
      expect(saveResponse.result.txHash).toStrictEqual(expectedTxHash);

      const getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(expectedTxHash);

      console.log(`getResponse: ${JSON.stringify(getResponse)}`);
      expect(getResponse.result.txHash).toStrictEqual(expectedTxHash);
      expect(getResponse.result.txRejectionStage).toStrictEqual("SEQUENCER");
      expect(getResponse.result.from).toStrictEqual(expectedFromAddress);
      expect(getResponse.result.nonce).toStrictEqual(`0x${expectedNonce.toString(16)}`);
      expect(getResponse.result.blockNumber).toStrictEqual(`0x${rejectedBlockNumber.toString(16)}`);
      expect(getResponse.result.timestamp).toStrictEqual(rejectedTimestamp);
    }, 10_000);

    it("Should return DUPLICATE_ALREADY_SAVED_BEFORE when saving the rejected transaction from SEQUENCER with same txHash and reason message", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const saveResponse = await transactionExclusionClient.saveRejectedTransactionV1(
        "SEQUENCER",
        new Date().toISOString(),
        rejectedBlockNumber,
        transactionRLP,
        "Transaction line count for module MUL=587 is above the limit 400 (from e2e test)",
        overflows,
      );

      console.log(`saveResponse: ${JSON.stringify(saveResponse)}`);
      expect(saveResponse.result.status).toStrictEqual("DUPLICATE_ALREADY_SAVED_BEFORE");
      expect(saveResponse.result.txHash).toStrictEqual(expectedTxHash);
    }, 10_000);

    it("Should return result as null when getting the rejected transaction with unknown transaction hash", async () => {
      if (transactionExclusionClient == null) {
        // Skip this test for dev and uat environments
        return;
      }

      const unknownTxHash = "0x7b37edcaacaceff0dc70a9ace28bd8e2284021c2df63d8e6b4f2f7673f032977";
      const getResponse = await transactionExclusionClient.getTransactionExclusionStatusV1(unknownTxHash);

      console.log(`getResponse: ${JSON.stringify(getResponse)}`);
      expect(getResponse.result).toStrictEqual(null);
    }, 10_000);
  });
};

export default transactionExclusionTestSuite;
