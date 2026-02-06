import { describe, expect, it } from "@jest/globals";
import { Hash, Hex, parseEther, parseGwei, toHex, TransactionReceipt } from "viem";

import {
  etherToWei,
  generateRandomUUIDv4,
  getRawTransactionHex,
  getTransactionHash,
  pollForBlockNumber,
} from "./common/utils";
import { createTestContext } from "./config/tests-config/setup";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";

describe("Send bundle test suite", () => {
  const context = createTestContext();
  const l2AccountManager = context.getL2AccountManager();
  const lineaCancelBundleClient = context.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
  const lineaSendBundleClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  it.concurrent(
    "Call sendBundle to RPC node and the bundled txs should get included",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = await lineaSendBundleClient.prepareTransactionRequest({
          type: "eip1559",
          account: senderAccount,
          to: recipientAccount.address,
          value: parseGwei("0.000001"),
          maxPriorityFeePerGas: parseGwei("1"),
          maxFeePerGas: parseGwei("10"),
          nonce: senderNonce++,
        });

        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 5n;
      const replacementUUID = generateRandomUUIDv4();

      const { bundleHash } = await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      logger.debug(`Bundle sent. bundleHash=${bundleHash}`);

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      for (const tx of txHashes) {
        const receipt = await l2PublicClient.waitForTransactionReceipt({ hash: tx, timeout: 20_000 });
        expect(receipt.status).toStrictEqual("success");
      }
    },
    120_000,
  );

  it.concurrent(
    "Call sendBundle to RPC node but the bundled txs should not get included as not all of them are valid",
    async () => {
      // Sender has 10 ETH and will try to send a bundle of three txs each sending 5 ETH, thus only the first
      // is valid, since we also need to account for the fee, so the second will fail and the entire bundle is reverted
      const senderAccount = await l2AccountManager.generateAccount(etherToWei("10"));
      const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = await lineaSendBundleClient.prepareTransactionRequest({
          type: "eip1559",
          account: senderAccount,
          to: recipientAccount.address,
          value: parseEther("5"),
          maxPriorityFeePerGas: parseGwei("1"),
          maxFeePerGas: parseGwei("10"),
          nonce: senderNonce++,
        });
        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 5n;
      const replacementUUID = generateRandomUUIDv4();

      const { bundleHash } = await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      logger.debug(`Bundle sent. bundleHash=${bundleHash}`);

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      // None of the bundled txs should be included as not all of them is valid
      for (const tx of txHashes) {
        let receipt: TransactionReceipt | undefined = undefined;
        try {
          receipt = await l2PublicClient.getTransactionReceipt({ hash: tx });
          // eslint-disable-next-line no-empty
        } catch {}
        expect(receipt?.status).toBeUndefined();
      }
    },
    120_000,
  );

  it.concurrent(
    "Call sendBundle to RPC node and then cancelBundle to sequencer and no bundled txs should get included",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const l2PublicClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = await lineaSendBundleClient.prepareTransactionRequest({
          type: "eip1559",
          account: senderAccount,
          to: recipientAccount.address,
          value: parseGwei("0.000001"),
          maxPriorityFeePerGas: parseGwei("1"),
          maxFeePerGas: parseGwei("10"),
          nonce: senderNonce++,
        });

        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 10n;
      const replacementUUID = generateRandomUUIDv4();

      const { bundleHash } = await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      logger.debug(`Bundle sent. bundleHash=${bundleHash}`);

      await pollForBlockNumber(l2PublicClient, targetBlockNumber - 5n);
      const cancelled = await lineaCancelBundleClient.lineaCancelBundle({ replacementUUID });
      expect(cancelled).toBeTruthy();

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      for (const tx of txHashes) {
        let receipt: TransactionReceipt | undefined = undefined;
        try {
          receipt = await l2PublicClient.getTransactionReceipt({ hash: tx });
          // eslint-disable-next-line no-empty
        } catch {}
        expect(receipt?.status).toBeUndefined();
      }
    },
    120_000,
  );
});
