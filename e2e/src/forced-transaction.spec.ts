import { etherToWei } from "@consensys/linea-shared-utils";
import { describe, expect, it } from "@jest/globals";
import { type Address, GetTransactionReceiptErrorType } from "viem";

import {
  buildSignedForcedTransaction,
  getDefaultLastFinalizedTimestamp,
  resolveLastFinalizedState,
} from "./common/test-helpers/forced-transaction";
import { getEvents, waitForEvents } from "./common/utils";
import { createTestContext } from "./config/setup";
import { LineaRollupV8Abi } from "./generated";

const context = createTestContext();
const l1AccountManager = context.getL1AccountManager();
const l2AccountManager = context.getL2AccountManager();

describe("Forced transaction test suite", () => {
  it.concurrent(
    "Should successfully submit a forced transaction containing a valid l2 transaction",
    async () => {
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const l1PublicClient = context.l1PublicClient();
      const l1WalletClient = context.l1WalletClient({ account: l1Account });

      const lineaRollup = context.l1Contracts.lineaRollup(l1PublicClient);
      const gateway = context.l1Contracts.forcedTransactionGateway(l1WalletClient);
      const gatewayRead = context.l1Contracts.forcedTransactionGateway(l1PublicClient);

      // Read gateway configuration
      const destinationChainId = await gatewayRead.read.DESTINATION_CHAIN_ID();

      logger.debug(`Gateway config — destinationChainId=${destinationChainId}`);

      // Resolve finalized state and fee
      const lastFinalizedState = await resolveLastFinalizedState(
        lineaRollup,
        l1PublicClient,
        getDefaultLastFinalizedTimestamp(),
      );

      logger.debug(
        `Resolved finalized state — timestamp=${lastFinalizedState.timestamp} messageNumber=${lastFinalizedState.messageNumber} messageRollingHash=${lastFinalizedState.messageRollingHash} forcedTransactionNumber=${lastFinalizedState.forcedTransactionNumber} forcedTransactionRollingHash=${lastFinalizedState.forcedTransactionRollingHash}`,
      );

      const { forcedTransaction, l2TxHash } = await buildSignedForcedTransaction(context, {
        l2Account,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9" as Address,
        nonce: 0n,
        value: etherToWei("0.1"),
        gasLimit: 21_000n,
        maxFeePerGas: 1_000_000_000n,
        maxPriorityFeePerGas: 100_000_000n,
      });

      logger.debug(
        `Built forced transaction — signer=${l2Account.address} to=${forcedTransaction.to} gasLimit=${forcedTransaction.gasLimit} l2TxHash=${l2TxHash}`,
      );

      const [, , , , feeAmount] = await lineaRollup.read.getRequiredForcedTransactionFields();
      const { maxPriorityFeePerGas, maxFeePerGas } = await l1PublicClient.estimateFeesPerGas();

      // Submit the forced transaction
      const txHash = await gateway.write.submitForcedTransaction([forcedTransaction, lastFinalizedState], {
        value: feeAmount,
        maxPriorityFeePerGas,
        maxFeePerGas,
      });

      logger.debug(`submitForcedTransaction sent. txHash=${txHash}`);

      const receipt = await l1PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: 30_000 });
      logger.debug(`Transaction confirmed. status=${receipt.status} blockNumber=${receipt.blockNumber}`);

      expect(receipt.status).toEqual("success");

      const [forcedTxEvent] = await getEvents(l1PublicClient, {
        abi: LineaRollupV8Abi,
        address: lineaRollup.address,
        eventName: "ForcedTransactionAdded",
        fromBlock: receipt.blockNumber,
        toBlock: receipt.blockNumber,
        strict: true,
      });

      expect(forcedTxEvent).toBeDefined();

      logger.debug(
        `ForcedTransactionAdded — forcedTransactionNumber=${forcedTxEvent.args.forcedTransactionNumber} from=${forcedTxEvent.args.from} blockNumberDeadline=${forcedTxEvent.args.blockNumberDeadline} forcedTransactionRollingHash=${forcedTxEvent.args.forcedTransactionRollingHash} rlpEncodedSignedTransaction=${forcedTxEvent.args.rlpEncodedSignedTransaction}`,
      );

      // Wait for the forced transaction to be executed on L2 by polling for its receipt
      const l2PublicClient = context.l2PublicClient();
      const submittedForcedTransactionNumber = forcedTxEvent.args.forcedTransactionNumber;

      logger.debug(`Waiting for forced transaction receipt on L2. l2TxHash=${l2TxHash}`);

      const l2Receipt = await l2PublicClient.waitForTransactionReceipt({ hash: l2TxHash, timeout: 120_000 });
      expect(l2Receipt.status).toEqual("success");

      logger.debug(
        `Forced transaction executed on L2. l2TxHash=${l2TxHash} blockNumber=${l2Receipt.blockNumber} status=${l2Receipt.status}`,
      );

      // Wait for L1 finalization that includes the forced transaction
      logger.debug(
        `Waiting for FinalizedStateUpdated with forcedTransactionNumber >= ${submittedForcedTransactionNumber}`,
      );

      const [finalizedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV8Abi,
        address: lineaRollup.address,
        eventName: "FinalizedStateUpdated",
        fromBlock: receipt.blockNumber,
        toBlock: "latest",
        pollingIntervalMs: 2_000,
        strict: true,
        criteria: async (events) =>
          events.filter((e) => e.args.forcedTransactionNumber >= submittedForcedTransactionNumber),
      });

      logger.debug(
        `Finalization includes forced transaction. blockNumber=${finalizedEvent.args.blockNumber} forcedTransactionNumber=${finalizedEvent.args.forcedTransactionNumber}`,
      );
    },
    300_000,
  );

  it.todo("Should send a forced transaction containing invalid L2 tx.");

  it.concurrent(
    "Should successfully submit a forced transaction containing an invalid L2 tx.",
    async () => {
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const l1PublicClient = context.l1PublicClient();
      const l1WalletClient = context.l1WalletClient({ account: l1Account });

      const lineaRollup = context.l1Contracts.lineaRollup(l1PublicClient);
      const gateway = context.l1Contracts.forcedTransactionGateway(l1WalletClient);
      const gatewayRead = context.l1Contracts.forcedTransactionGateway(l1PublicClient);

      // Read gateway configuration
      const destinationChainId = await gatewayRead.read.DESTINATION_CHAIN_ID();

      logger.debug(`Gateway config — destinationChainId=${destinationChainId}`);

      // Resolve finalized state and fee
      const lastFinalizedState = await resolveLastFinalizedState(
        lineaRollup,
        l1PublicClient,
        getDefaultLastFinalizedTimestamp(),
      );

      logger.debug(
        `Resolved finalized state — timestamp=${lastFinalizedState.timestamp} messageNumber=${lastFinalizedState.messageNumber} messageRollingHash=${lastFinalizedState.messageRollingHash} forcedTransactionNumber=${lastFinalizedState.forcedTransactionNumber} forcedTransactionRollingHash=${lastFinalizedState.forcedTransactionRollingHash}`,
      );

      const { forcedTransaction, l2TxHash } = await buildSignedForcedTransaction(context, {
        l2Account,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9" as Address,
        nonce: 1_000_000n, // HIGH NONCE TO ENSURE THE L2 TX REVERTS ON L2.
        value: etherToWei("0.1"),
        gasLimit: 21_000n,
        maxFeePerGas: 1_000_000_000n,
        maxPriorityFeePerGas: 100_000_000n,
      });

      logger.debug(
        `Built forced transaction — signer=${l2Account.address} to=${forcedTransaction.to} gasLimit=${forcedTransaction.gasLimit} l2TxHash=${l2TxHash}`,
      );

      const [, , , , feeAmount] = await lineaRollup.read.getRequiredForcedTransactionFields();
      const { maxPriorityFeePerGas, maxFeePerGas } = await l1PublicClient.estimateFeesPerGas();

      // Submit the forced transaction
      const txHash = await gateway.write.submitForcedTransaction([forcedTransaction, lastFinalizedState], {
        value: feeAmount,
        maxPriorityFeePerGas,
        maxFeePerGas,
      });

      logger.debug(`submitForcedTransaction sent. txHash=${txHash}`);

      const receipt = await l1PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: 30_000 });
      logger.debug(`Transaction confirmed. status=${receipt.status} blockNumber=${receipt.blockNumber}`);

      expect(receipt.status).toEqual("success");

      const [forcedTxEvent] = await getEvents(l1PublicClient, {
        abi: LineaRollupV8Abi,
        address: lineaRollup.address,
        eventName: "ForcedTransactionAdded",
        fromBlock: receipt.blockNumber,
        toBlock: receipt.blockNumber,
        strict: true,
      });

      expect(forcedTxEvent).toBeDefined();

      logger.debug(
        `ForcedTransactionAdded — forcedTransactionNumber=${forcedTxEvent.args.forcedTransactionNumber} from=${forcedTxEvent.args.from} blockNumberDeadline=${forcedTxEvent.args.blockNumberDeadline} forcedTransactionRollingHash=${forcedTxEvent.args.forcedTransactionRollingHash} rlpEncodedSignedTransaction=${forcedTxEvent.args.rlpEncodedSignedTransaction}`,
      );

      const submittedForcedTransactionNumber = forcedTxEvent.args.forcedTransactionNumber;

      // Wait for L1 finalization that includes the forced transaction
      logger.debug(
        `Waiting for FinalizedStateUpdated with forcedTransactionNumber >= ${submittedForcedTransactionNumber}`,
      );

      const [finalizedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV8Abi,
        address: lineaRollup.address,
        eventName: "FinalizedStateUpdated",
        fromBlock: receipt.blockNumber,
        toBlock: "latest",
        pollingIntervalMs: 2_000,
        strict: true,
        criteria: async (events) =>
          events.filter((e) => e.args.forcedTransactionNumber >= submittedForcedTransactionNumber),
      });

      logger.debug(
        `Finalization includes forced transaction. blockNumber=${finalizedEvent.args.blockNumber} forcedTransactionNumber=${finalizedEvent.args.forcedTransactionNumber}`,
      );

      const l2PublicClient = context.l2PublicClient();
      logger.debug(`Verifying that the forced transaction receipt is not available on L2 for l2TxHash=${l2TxHash}`);

      try {
        await l2PublicClient.getTransactionReceipt({ hash: l2TxHash });
        // This code path should not be reached; the getTransactionReceipt call is expected to throw.
        // If it does not throw, the test should fail.
        throw new Error(
          "Test failed: Expected getTransactionReceipt to throw TransactionReceiptNotFoundError, but it did not.",
        );
      } catch (error) {
        const e = error as GetTransactionReceiptErrorType;
        if (e.name !== "TransactionReceiptNotFoundError") {
          throw new Error("Test failed: Unexpected error type thrown. Expected TransactionReceiptNotFoundError.");
        }
      }

      logger.debug("Checked for forced transaction receipt on L2; confirmed that receipt was not found as expected.");
    },
    300_000,
  );
});
