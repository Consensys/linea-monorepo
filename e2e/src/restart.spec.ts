import { describe, expect, it } from "@jest/globals";
import type { Logger } from "winston";
import {
  execDockerCommand,
  waitForEvents,
  getMessageSentEventFromLogs,
  sendMessage,
  etherToWei,
  wait,
  serialize,
} from "./common/utils";
import { config } from "./config/tests-config/setup";
import { L2MessageServiceV1Abi, LineaRollupV6Abi } from "./generated";

let testsWaitingForRestart = 0;
const TOTAL_TESTS_WAITING = 2;
let coordinatorHasRestarted = false;

async function waitForCoordinatorRestart(logger: Logger) {
  testsWaitingForRestart += 1;

  while (testsWaitingForRestart < TOTAL_TESTS_WAITING) {
    logger.debug(`Waiting for other test to reach restart point... (${testsWaitingForRestart}/${TOTAL_TESTS_WAITING})`);
    await wait(1000);
  }

  if (!coordinatorHasRestarted) {
    coordinatorHasRestarted = true;
    logger.debug("Both tests have reached the restart point. Restarting coordinator...");
    try {
      await execDockerCommand("restart", "coordinator");
      logger.debug("Coordinator restarted.");
    } catch (error) {
      logger.error(`Failed to restart coordinator: ${error}`);
      throw error;
    }
  } else {
    logger.debug("Coordinator has already been restarted by another test.");
  }
}
const l1AccountManager = config.getL1AccountManager();

describe("Coordinator restart test suite", () => {
  it.concurrent(
    "When the coordinator restarts it should resume blob submission and finalization",
    async () => {
      if (process.env.TEST_ENV !== "local") {
        logger.warn("Skipping test because it's not running on a local environment.");
        return;
      }
      const l1PublicClient = config.l1PublicClient();
      // await for a finalization to happen on L1
      const [dataSubmittedEventsBeforeRestart, dataFinalizedEventsBeforeRestart] = await Promise.all([
        waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
          eventName: "DataSubmittedV3",
          fromBlock: 0n,
          toBlock: "latest",
          strict: true,
        }),
        waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
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

      await waitForCoordinatorRestart(logger);

      const currentBlockNumberAfterRestart = await l1PublicClient.getBlockNumber();

      logger.debug("Waiting for DataSubmittedV3 event after coordinator restart...");
      const [dataSubmittedV3EventAfterRestart] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
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
    150_000,
  );

  it.concurrent(
    "When the coordinator restarts it should resume anchoring",
    async () => {
      if (process.env.TEST_ENV !== "local") {
        logger.warn("Skipping test because it's not running on a local environment.");
        return;
      }

      const l1PublicClient = config.l1PublicClient();
      const l1MessageSender = await l1AccountManager.generateAccount();

      const l1WalletClient = config.l1WalletClient({ account: l1MessageSender });

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
            contractAddress: config.l1PublicClient().getLineaRollup().address,
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

      const l2PublicClient = config.l2PublicClient();
      const [rollingHashUpdatedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
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
      await waitForCoordinatorRestart(logger);
      const l1Fees = await l1PublicClient.estimateFeesPerGas();

      logger.debug("Sending messages L1 -> L2 after coordinator restart...");
      // Send more messages L1 -> L2
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
            contractAddress: config.l1PublicClient().getLineaRollup().address,
            value: messageValue,
            nonce: l1MessageSenderNonce,
            maxPriorityFeePerGas: l1Fees.maxPriorityFeePerGas,
            maxFeePerGas: l1Fees.maxFeePerGas,
          }),
        );
        l1MessageSenderNonce++;
      }

      const l1ReceiptsAfterRestart = await Promise.all(l1MessagesPromises);
      const l1MessagesAfterRestart = getMessageSentEventFromLogs(l1ReceiptsAfterRestart);

      // Wait for messages to be anchored on L2
      const lastNewL1MessageNumberAfterRestart = l1MessagesAfterRestart.slice(-1)[0].messageNumber;

      logger.debug(
        `Waiting for L1->L2 anchoring after coordinator restart. messageNumber=${lastNewL1MessageNumberAfterRestart}`,
      );
      const [rollingHashUpdatedEventAfterRestart] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
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

      const lineaRollup = config.l1PublicClient().getLineaRollup();
      const l2MessageService = config.l2PublicClient().getL2MessageServiceContract();

      const [lastNewMessageRollingHashAfterRestart, lastAnchoredL1MessageNumberAfterRestart] = await Promise.all([
        lineaRollup.read.rollingHashes([rollingHashUpdatedEventAfterRestart.args.messageNumber]),
        l2MessageService.read.lastAnchoredL1MessageNumber(),
      ]);

      expect(lastNewMessageRollingHashAfterRestart).toEqual(rollingHashUpdatedEventAfterRestart.args.rollingHash);
      expect(lastAnchoredL1MessageNumberAfterRestart).toEqual(rollingHashUpdatedEventAfterRestart.args.messageNumber);
    },
    150_000,
  );
});
