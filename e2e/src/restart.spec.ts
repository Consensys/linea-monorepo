import { describe, expect, it } from "@jest/globals";

import {
  execDockerCommand,
  waitForEvents,
  getMessageSentEventFromLogs,
  sendMessage,
  etherToWei,
  serialize,
} from "./common/utils";
import { createTestContext } from "./config/setup";
import { L2MessageServiceV1Abi, LineaRollupV6Abi } from "./generated";

import type { Logger } from "winston";

/**
 * Barrier that allows N concurrent tests to synchronize at a restart point.
 * The first caller to `arrive()` creates the restart promise.
 * All callers await the same promise, which resolves once the restart completes.
 * No hardcoded participant count -- the first arrival triggers a short grace period
 * for others to join, then performs the restart.
 */
function createRestartBarrier(graceMs = 5_000) {
  let restartPromise: Promise<void> | null = null;
  let arrivals = 0;

  return {
    async arrive(logger: Logger): Promise<void> {
      arrivals++;
      logger.debug(`Test arrived at restart barrier (arrivals=${arrivals}).`);

      if (!restartPromise) {
        restartPromise = new Promise<void>((resolve, reject) => {
          setTimeout(async () => {
            logger.debug(`Grace period elapsed. ${arrivals} test(s) waiting. Restarting coordinator...`);
            try {
              await execDockerCommand("restart", "coordinator");
              logger.debug("Coordinator restarted.");
              resolve();
            } catch (error) {
              logger.error(`Failed to restart coordinator: ${error}`);
              reject(error);
            }
          }, graceMs);
        });
      }

      await restartPromise;
    },
  };
}

const restartBarrier = createRestartBarrier();
const context = createTestContext();
const l1AccountManager = context.getL1AccountManager();

describe("Coordinator restart test suite", () => {
  it.concurrent(
    "When the coordinator restarts it should resume blob submission and finalization",
    async () => {
      if (process.env.TEST_ENV !== "local") {
        logger.warn("Skipping test because it's not running on a local environment.");
        return;
      }
      const l1PublicClient = context.l1PublicClient();
      const lineaRollup = context.l1Contracts.lineaRollup(l1PublicClient);
      // await for a finalization to happen on L1
      const [dataSubmittedEventsBeforeRestart, dataFinalizedEventsBeforeRestart] = await Promise.all([
        waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
          address: lineaRollup.address,
          eventName: "DataSubmittedV3",
          fromBlock: 0n,
          toBlock: "latest",
          strict: true,
        }),
        waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
          address: lineaRollup.address,
          eventName: "DataFinalizedV3",
          fromBlock: 0n,
          toBlock: "latest",
          strict: true,
        }),
      ]);

      expect(dataSubmittedEventsBeforeRestart.length).toBeGreaterThan(0);
      expect(dataFinalizedEventsBeforeRestart.length).toBeGreaterThan(0);

      const lastDataSubmittedEventBeforeRestart = dataSubmittedEventsBeforeRestart.slice(-1)[0];
      const lastDataFinalizedEventsBeforeRestart = dataFinalizedEventsBeforeRestart.slice(-1)[0];

      logger.debug(
        `DataSubmittedV3 event before coordinator restart found. event=${serialize(lastDataSubmittedEventBeforeRestart)}`,
      );
      logger.debug(
        `DataFinalizedV3 event before coordinator restart found. event=${serialize(lastDataFinalizedEventsBeforeRestart)}`,
      );

      // Just some sanity checks
      // Check that the coordinator has submitted and finalized data before the restart
      expect(lastDataSubmittedEventBeforeRestart.blockNumber).toBeGreaterThan(0n);
      expect(lastDataFinalizedEventsBeforeRestart.args.endBlockNumber).toBeGreaterThan(0n);

      await restartBarrier.arrive(logger);

      const currentBlockNumberAfterRestart = await l1PublicClient.getBlockNumber();

      logger.debug("Waiting for DataSubmittedV3 event after coordinator restart...");
      const [dataSubmittedV3EventAfterRestart] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "DataSubmittedV3",
        fromBlock: currentBlockNumberAfterRestart,
        toBlock: "latest",
        pollingIntervalMs: 1_000,
        strict: true,
        criteria: async (events) =>
          events.filter((event) => event.blockNumber > lastDataSubmittedEventBeforeRestart.blockNumber),
      });

      expect(dataSubmittedV3EventAfterRestart).toBeDefined();

      logger.debug(
        `DataSubmittedV3 event after coordinator restart found. event=${serialize(dataSubmittedV3EventAfterRestart)}`,
      );

      logger.debug("Waiting for DataFinalizedV3 event after coordinator restart...");
      const [dataFinalizedEventAfterRestart] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "DataFinalizedV3",
        fromBlock: currentBlockNumberAfterRestart,
        toBlock: "latest",
        pollingIntervalMs: 1_000,
        strict: true,
        criteria: async (events) =>
          events.filter(
            (event) => event.args.endBlockNumber! > lastDataFinalizedEventsBeforeRestart.args.endBlockNumber,
          ),
      });

      expect(dataFinalizedEventAfterRestart).toBeDefined();

      logger.debug(
        `DataFinalizedV3 event after coordinator restart found. event=${serialize(dataFinalizedEventAfterRestart)}`,
      );

      expect(dataFinalizedEventAfterRestart.args.endBlockNumber).toBeGreaterThan(
        lastDataFinalizedEventsBeforeRestart.args.endBlockNumber,
      );
    },
    200_000,
  );

  it.concurrent(
    "When the coordinator restarts it should resume anchoring",
    async () => {
      if (process.env.TEST_ENV !== "local") {
        logger.warn("Skipping test because it's not running on a local environment.");
        return;
      }

      const l1PublicClient = context.l1PublicClient();
      const l1MessageSender = await l1AccountManager.generateAccount();

      const l1WalletClient = context.l1WalletClient({ account: l1MessageSender });

      // Send Messages L1 -> L2
      const messageFee = etherToWei("0.0001");
      const messageValue = etherToWei("0.0051");
      const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

      const l1MessagesPromises = [];
      let l1MessageSenderNonce = await l1PublicClient.getTransactionCount({ address: l1MessageSender.address });
      const { maxPriorityFeePerGas, maxFeePerGas } = await l1PublicClient.estimateFeesPerGas();
      logger.debug(`Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`);

      logger.debug("Sending messages L1 -> L2 before coordinator restart...");
      for (let i = 0; i < 5; i++) {
        l1MessagesPromises.push(
          sendMessage(l1WalletClient, {
            account: l1MessageSender,
            chain: l1PublicClient.chain,
            args: {
              to: destinationAddress,
              fee: messageFee,
              calldata: "0x",
            },
            contractAddress: context.l1Contracts.lineaRollup(context.l1PublicClient()).address,
            value: messageValue,
            nonce: l1MessageSenderNonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          }),
        );
        l1MessageSenderNonce++;
      }

      const l1Receipts = await Promise.all(l1MessagesPromises);
      const l1Messages = getMessageSentEventFromLogs(l1Receipts);

      // Wait for L2 Anchoring
      const lastNewL1MessageNumber = l1Messages.slice(-1)[0].messageNumber;

      logger.debug(`Waiting for L1->L2 anchoring before coordinator restart. messageNumber=${lastNewL1MessageNumber}`);

      const l2PublicClient = context.l2PublicClient();
      const l2MessageService = context.l2Contracts.l2MessageService(l2PublicClient);

      const [rollingHashUpdatedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        address: l2MessageService.address,
        eventName: "RollingHashUpdated",
        fromBlock: 0n,
        toBlock: "latest",
        pollingIntervalMs: 1_000,
        strict: true,
        criteria: async (events) => {
          return events.filter((event) => event.args.messageNumber >= lastNewL1MessageNumber);
        },
      });

      expect(rollingHashUpdatedEvent).toBeDefined();

      // Restart Coordinator
      await restartBarrier.arrive(logger);
      const l1Fees = await l1PublicClient.estimateFeesPerGas();

      logger.debug("Sending messages L1 -> L2 after coordinator restart...");

      // Send more messages L1 -> L2
      const l1MessagesPromisesAfterRestart = [];
      for (let i = 0; i < 5; i++) {
        l1MessagesPromisesAfterRestart.push(
          sendMessage(l1WalletClient, {
            account: l1MessageSender,
            chain: l1PublicClient.chain,
            args: {
              to: destinationAddress,
              fee: messageFee,
              calldata: "0x",
            },
            contractAddress: context.l1Contracts.lineaRollup(context.l1PublicClient()).address,
            value: messageValue,
            nonce: l1MessageSenderNonce,
            maxPriorityFeePerGas: l1Fees.maxPriorityFeePerGas,
            maxFeePerGas: l1Fees.maxFeePerGas,
          }),
        );
        l1MessageSenderNonce++;
      }

      const l1ReceiptsAfterRestart = await Promise.all(l1MessagesPromisesAfterRestart);
      const l1MessagesAfterRestart = getMessageSentEventFromLogs(l1ReceiptsAfterRestart);

      // Wait for messages to be anchored on L2
      const lastNewL1MessageNumberAfterRestart = l1MessagesAfterRestart.slice(-1)[0].messageNumber;

      logger.debug(
        `Waiting for L1->L2 anchoring after coordinator restart. messageNumber=${lastNewL1MessageNumberAfterRestart}`,
      );
      const [rollingHashUpdatedEventAfterRestart] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        address: l2MessageService.address,
        eventName: "RollingHashUpdated",
        fromBlock: 0n,
        toBlock: "latest",
        pollingIntervalMs: 1_000,
        strict: true,
        criteria: async (events) => {
          return events.filter((event) => event.args.messageNumber >= lastNewL1MessageNumberAfterRestart);
        },
      });

      expect(rollingHashUpdatedEventAfterRestart).toBeDefined();

      const lineaRollup = context.l1Contracts.lineaRollup(context.l1PublicClient());

      const [lastNewMessageRollingHashAfterRestart, lastAnchoredL1MessageNumberAfterRestart] = await Promise.all([
        lineaRollup.read.rollingHashes([rollingHashUpdatedEventAfterRestart.args.messageNumber]),
        l2MessageService.read.lastAnchoredL1MessageNumber(),
      ]);

      expect(lastNewMessageRollingHashAfterRestart).toEqual(rollingHashUpdatedEventAfterRestart.args.rollingHash);
      expect(lastAnchoredL1MessageNumberAfterRestart).toEqual(rollingHashUpdatedEventAfterRestart.args.messageNumber);
    },
    200_000,
  );
});
