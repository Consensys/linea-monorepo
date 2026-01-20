import { describe, expect, it } from "@jest/globals";
import { ethers, TransactionRequest } from "ethers";
import { config } from "./config/tests-config";
import {
  getRawTransactionHex,
  getTransactionHash,
  getWallet,
  LineaForcedTransactionClient,
  pollForBlockNumber,
} from "./common/utils";

describe("Forced transactions test suite", () => {
  const l2AccountManager = config.getL2AccountManager();
  const forcedTxClient = new LineaForcedTransactionClient(config.getSequencerEndpoint()!);

  it.concurrent(
    "Single forced transaction should get included",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const senderWallet = getWallet(senderAccount.privateKey, config.getL2BesuNodeProvider()!);
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      const senderNonce = await senderAccount.getNonce();
      const txRequest: TransactionRequest = {
        to: recipientAccount.address,
        value: ethers.parseUnits("1000", "wei"),
        maxPriorityFeePerGas: ethers.parseEther("0.000000001"), // 1 Gwei
        maxFeePerGas: ethers.parseEther("0.00000001"), // 10 Gwei
        nonce: senderNonce,
      };

      const rawTx = await getRawTransactionHex(txRequest, senderWallet);
      const expectedTxHash = await getTransactionHash(txRequest, senderWallet);

      const resultHashes = await forcedTxClient.lineaSendForcedRawTransaction([{ transaction: rawTx }]);

      expect(resultHashes).toHaveLength(1);
      expect(resultHashes[0]).toEqual(expectedTxHash);

      // Wait for inclusion
      const startBlockNumber = await config.getL2Provider().getBlockNumber();
      const targetBlockNumber = startBlockNumber + 5;
      const hasReachedTargetBlockNumber = await pollForBlockNumber(config.getL2Provider(), targetBlockNumber);
      expect(hasReachedTargetBlockNumber).toBeTruthy();

      // Verify transaction was included
      const receipt = await config.getL2Provider().getTransactionReceipt(expectedTxHash);
      expect(receipt?.status).toStrictEqual(1);

      // Verify inclusion status
      const status = await forcedTxClient.lineaGetForcedTransactionInclusionStatus(expectedTxHash);
      expect(status).not.toBeNull();
      expect(status!.inclusionResult).toEqual("INCLUDED");
      expect(status!.transactionHash).toEqual(expectedTxHash.toLowerCase());
    },
    120_000,
  );

  it.concurrent(
    "Multiple forced transactions should get included in order",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const senderWallet = getWallet(senderAccount.privateKey, config.getL2BesuNodeProvider()!);
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await senderAccount.getNonce();
      const txHashes: string[] = [];
      const txs: Array<{ transaction: string }> = [];

      for (let i = 0; i < 3; i++) {
        const txRequest: TransactionRequest = {
          to: recipientAccount.address,
          value: ethers.parseUnits("1000", "wei"),
          maxPriorityFeePerGas: ethers.parseEther("0.000000001"), // 1 Gwei
          maxFeePerGas: ethers.parseEther("0.00000001"), // 10 Gwei
          nonce: senderNonce++,
        };
        txs.push({ transaction: await getRawTransactionHex(txRequest, senderWallet) });
        txHashes.push(await getTransactionHash(txRequest, senderWallet));
      }

      const resultHashes = await forcedTxClient.lineaSendForcedRawTransaction(txs);

      expect(resultHashes).toHaveLength(3);
      expect(resultHashes).toEqual(txHashes);

      // Wait for inclusion
      const startBlockNumber = await config.getL2Provider().getBlockNumber();
      const targetBlockNumber = startBlockNumber + 10;
      const hasReachedTargetBlockNumber = await pollForBlockNumber(config.getL2Provider(), targetBlockNumber);
      expect(hasReachedTargetBlockNumber).toBeTruthy();

      // Verify all transactions were included
      for (const txHash of txHashes) {
        const receipt = await config.getL2Provider().getTransactionReceipt(txHash);
        expect(receipt?.status).toStrictEqual(1);

        const status = await forcedTxClient.lineaGetForcedTransactionInclusionStatus(txHash);
        expect(status?.inclusionResult).toEqual("INCLUDED");
      }
    },
    120_000,
  );

  it.concurrent(
    "Forced transaction with insufficient balance should be rejected with BAD_BALANCE",
    async () => {
      // Create account with very small balance
      const senderAccount = await l2AccountManager.generateAccount(ethers.parseUnits("1", "wei"));
      const senderWallet = getWallet(senderAccount.privateKey, config.getL2BesuNodeProvider()!);
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      const senderNonce = await senderAccount.getNonce();
      const txRequest: TransactionRequest = {
        to: recipientAccount.address,
        value: ethers.parseEther("1000"), // Much more than available
        maxPriorityFeePerGas: ethers.parseEther("0.000000001"),
        maxFeePerGas: ethers.parseEther("0.00000001"),
        nonce: senderNonce,
      };

      const rawTx = await getRawTransactionHex(txRequest, senderWallet);
      const expectedTxHash = await getTransactionHash(txRequest, senderWallet);

      await forcedTxClient.lineaSendForcedRawTransaction([{ transaction: rawTx }]);

      // Wait for processing
      const startBlockNumber = await config.getL2Provider().getBlockNumber();
      const targetBlockNumber = startBlockNumber + 5;
      await pollForBlockNumber(config.getL2Provider(), targetBlockNumber);

      // Verify rejection status
      const status = await forcedTxClient.lineaGetForcedTransactionInclusionStatus(expectedTxHash);
      expect(status).not.toBeNull();
      expect(status!.inclusionResult).toEqual("BAD_BALANCE");
    },
    120_000,
  );

  it.concurrent(
    "Forced transaction with wrong nonce should be rejected with BAD_NONCE",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const senderWallet = getWallet(senderAccount.privateKey, config.getL2BesuNodeProvider()!);
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      const senderNonce = await senderAccount.getNonce();
      // Use future nonce
      const txRequest: TransactionRequest = {
        to: recipientAccount.address,
        value: ethers.parseUnits("1000", "wei"),
        maxPriorityFeePerGas: ethers.parseEther("0.000000001"),
        maxFeePerGas: ethers.parseEther("0.00000001"),
        nonce: senderNonce + 100, // Nonce too high
      };

      const rawTx = await getRawTransactionHex(txRequest, senderWallet);
      const expectedTxHash = await getTransactionHash(txRequest, senderWallet);

      await forcedTxClient.lineaSendForcedRawTransaction([{ transaction: rawTx }]);

      // Wait for processing
      const startBlockNumber = await config.getL2Provider().getBlockNumber();
      const targetBlockNumber = startBlockNumber + 5;
      await pollForBlockNumber(config.getL2Provider(), targetBlockNumber);

      // Verify rejection status
      const status = await forcedTxClient.lineaGetForcedTransactionInclusionStatus(expectedTxHash);
      expect(status).not.toBeNull();
      expect(status!.inclusionResult).toEqual("BAD_NONCE");
    },
    120_000,
  );

  it.concurrent(
    "Query inclusion status for unknown transaction should return null",
    async () => {
      const unknownTxHash = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
      const status = await forcedTxClient.lineaGetForcedTransactionInclusionStatus(unknownTxHash);
      expect(status).toBeNull();
    },
    30_000,
  );
});
