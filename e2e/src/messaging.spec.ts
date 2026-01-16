import { describe, expect, it } from "@jest/globals";
import {
  etherToWei,
  getMessageSentEventFromLogs,
  sendL1ToL2Message,
  sendL2ToL1Message,
  waitForEvents,
} from "./common/utils";
import { L2MessageServiceV1Abi, LineaRollupV6Abi } from "./generated";
import { config } from "./config/tests-config/setup";

const l1AccountManager = config.getL1AccountManager();
const l2AccountManager = config.getL2AccountManager();

describe("Messaging test suite", () => {
  it.concurrent(
    "Should send a transaction with fee and calldata to L1 message service, be successfully claimed it on L2",
    async () => {
      const l1Account = await l1AccountManager.generateAccount();

      const { txHash, receipt } = await sendL1ToL2Message({
        account: l1Account,
        fee: etherToWei("0.1"),
        value: etherToWei("0.1"),
        withCalldata: true,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        address: l2PublicClient.getL2MessageServiceContract().address,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        strict: true,
      });

      expect(messageClaimedEvent).toBeDefined();
      logger.debug(
        `Message claimed on L2. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );

  it.concurrent(
    "Should send a transaction with fee and without calldata to L1 message service, be successfully claimed it on L2",
    async () => {
      const l1Account = await l1AccountManager.generateAccount();

      const { txHash, receipt } = await sendL1ToL2Message({
        account: l1Account,
        fee: etherToWei("0.1"),
        value: etherToWei("0.2"),
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        address: l2PublicClient.getL2MessageServiceContract().address,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        strict: true,
      });
      expect(messageClaimedEvent).toBeDefined();
      logger.debug(
        `Message claimed on L2. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );

  // Test that Postman sponsoring works for L1->L2
  it.concurrent(
    "Should send a transaction without fee and without calldata to L1 message service, be successfully claimed it on L2",
    async () => {
      const l1Account = await l1AccountManager.generateAccount();

      const { txHash, receipt } = await sendL1ToL2Message({
        account: l1Account,
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        address: l2PublicClient.getL2MessageServiceContract().address,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
        strict: true,
      });

      expect(messageClaimedEvent).toBeDefined();
      logger.debug(
        `Message claimed on L2. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );

  it.concurrent(
    "Should send a transaction with fee and calldata to L2 message service, be successfully claimed it on L1",
    async () => {
      const l2Account = await l2AccountManager.generateAccount();
      const l1PublicClient = config.l1PublicClient();
      const lineaRollup = l1PublicClient.getLineaRollup();

      const { txHash, receipt } = await sendL2ToL1Message({
        account: l2Account,
        fee: etherToWei("0.001"),
        value: etherToWei("0.001"),
        withCalldata: true,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L2 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for L2MessagingBlockAnchored event... blockNumber=${messageSentEvent.blockNumber}`);
      const [l2MessagingBlockAnchoredEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "L2MessagingBlockAnchored",
        args: {
          l2Block: messageSentEvent.blockNumber,
        },
        pollingIntervalMs: 1_000,
        strict: true,
      });

      expect(l2MessagingBlockAnchoredEvent).toBeDefined();

      logger.debug(`Waiting for MessageClaimed event on L1... messageHash=${messageHash}`);
      const [messageClaimedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
        strict: true,
      });

      expect(messageClaimedEvent).toBeDefined();
      logger.debug(
        `Message claimed on L1. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );

  it.concurrent(
    "Should send a transaction with fee and without calldata to L2 message service, be successfully claimed it on L1",
    async () => {
      const l2Account = await l2AccountManager.generateAccount();
      const l1PublicClient = config.l1PublicClient();
      const lineaRollup = l1PublicClient.getLineaRollup();

      const { txHash, receipt } = await sendL2ToL1Message({
        account: l2Account,
        fee: etherToWei("0.001"),
        value: etherToWei("0.01"),
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L2 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for L2MessagingBlockAnchored event... blockNumber=${messageSentEvent.blockNumber}`);
      const [l2MessagingBlockAnchoredEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "L2MessagingBlockAnchored",
        args: {
          l2Block: messageSentEvent.blockNumber,
        },
        pollingIntervalMs: 1_000,
        strict: true,
      });

      expect(l2MessagingBlockAnchoredEvent).toBeDefined();

      logger.debug(`Waiting for MessageClaimed event on L1. messageHash=${messageHash}`);
      const [messageClaimedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        address: lineaRollup.address,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
        strict: true,
      });

      expect(messageClaimedEvent).toBeDefined();

      logger.debug(
        `Message claimed on L1. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );
});
