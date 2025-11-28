import { describe, expect, it } from "@jest/globals";
import {
  etherToWei,
  generateRandomUUIDv4,
  getRawTransactionHex,
  getTransactionHash,
  pollForBlockNumber,
} from "./common/utils";
import { config } from "./config/tests-config/setup";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";
import { Hash, Hex, parseEther, toHex } from "viem";

describe("Send bundle test suite", () => {
  const l2AccountManager = config.getL2AccountManager();
  const lineaCancelBundleClient = config.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
  const lineaSendBundleClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  it.concurrent(
    "Call sendBundle to RPC node and the bundled txs should get included",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = {
          account: senderAccount,
          to: recipientAccount.address,
          value: parseEther("1000"),
          maxPriorityFeePerGas: parseEther("0.000000001"), // 1 Gwei
          maxFeePerGas: parseEther("0.00000001"), // 10 Gwei
          nonce: senderNonce++,
        };
        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 5n;
      const replacementUUID = generateRandomUUIDv4();

      await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      for (const tx of txHashes) {
        const receipt = await l2PublicClient.getTransactionReceipt({ hash: tx });
        expect(receipt?.status).toStrictEqual(1);
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
      const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = {
          account: senderAccount,
          to: recipientAccount.address,
          value: parseEther("5"),
          maxPriorityFeePerGas: parseEther("0.000000001"), // 1 Gwei
          maxFeePerGas: parseEther("0.00000001"), // 10 Gwei
          nonce: senderNonce++,
        };
        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 5n;
      const replacementUUID = generateRandomUUIDv4();

      await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      // None of the bundled txs should be included as not all of them is valid
      for (const tx of txHashes) {
        const receipt = await l2PublicClient.getTransactionReceipt({ hash: tx });
        expect(receipt?.status).toBeUndefined();
      }
    },
    120_000,
  );

  it.concurrent(
    "Call sendBundle to RPC node and then cancelBundle to sequencer and no bundled txs should get included",
    async () => {
      const senderAccount = await l2AccountManager.generateAccount();
      const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });
      const recipientAccount = await l2AccountManager.generateAccount(0n);

      let senderNonce = await l2PublicClient.getTransactionCount({ address: senderAccount.address });
      const txHashes: Hash[] = [];
      const txs: Hex[] = [];

      for (let i = 0; i < 3; i++) {
        const txRequest = {
          account: senderAccount,
          to: recipientAccount.address,
          value: parseEther("1000"),
          maxPriorityFeePerGas: parseEther("0.000000001"), // 1 Gwei
          maxFeePerGas: parseEther("0.00000001"), // 10 Gwei
          nonce: senderNonce++,
        };

        txs.push(await getRawTransactionHex(l2PublicClient, txRequest));
        txHashes.push(await getTransactionHash(l2PublicClient, txRequest));
      }

      const targetBlockNumber = (await l2PublicClient.getBlockNumber()) + 10n;
      const replacementUUID = generateRandomUUIDv4();

      await lineaSendBundleClient.lineaSendBundle({
        txs,
        replacementUUID,
        blockNumber: toHex(targetBlockNumber),
      });

      await pollForBlockNumber(l2PublicClient, targetBlockNumber - 5n);
      const cancelled = await lineaCancelBundleClient.lineaCancelBundle({ replacementUUID });
      expect(cancelled).toBeTruthy();

      const hasReachedTargetBlockNumber = await pollForBlockNumber(l2PublicClient, targetBlockNumber);

      expect(hasReachedTargetBlockNumber).toBeTruthy();
      for (const tx of txHashes) {
        const receipt = await l2PublicClient.getTransactionReceipt({ hash: tx });
        expect(receipt?.status).toBeUndefined();
      }
    },
    120_000,
  );
});
