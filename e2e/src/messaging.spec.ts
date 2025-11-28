import { describe, expect, it } from "@jest/globals";
import type { Logger } from "winston";
import {
  encodeFunctionCall,
  estimateLineaGas,
  etherToWei,
  getMessageSentEventFromLogs,
  waitForEvents,
} from "./common/utils";
import { DummyContractAbi, L2MessageServiceV1Abi, LineaRollupV6Abi } from "./generated";
import { PrivateKeyAccount, toHex } from "viem";
import { randomBytes } from "crypto";
import { config } from "./config/tests-config/setup";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";

async function sendL1ToL2Message(
  logger: Logger,
  {
    l1Account,
    l2Account,
    fee = 0n,
    withCalldata = false,
  }: {
    l1Account: PrivateKeyAccount;
    l2Account: PrivateKeyAccount;
    fee: bigint;
    withCalldata: boolean;
  },
) {
  const dummyContract = config.l2WalletClient({ account: l2Account }).getDummyContract();
  const lineaRollup = config.l1WalletClient({ account: l1Account }).getLineaRollup();

  const calldata = withCalldata
    ? encodeFunctionCall({
        abi: DummyContractAbi,
        functionName: "setPayload",
        args: [toHex(randomBytes(100).toString("hex"))],
      })
    : "0x";
  const destinationAddress = withCalldata ? dummyContract.address : "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

  const l1PublicClient = config.l1PublicClient();
  const { maxPriorityFeePerGas, maxFeePerGas } = await l1PublicClient.estimateFeesPerGas();

  logger.debug(`Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`);

  const txHash = await lineaRollup.write.sendMessage([destinationAddress, fee, calldata], {
    value: fee,
    maxPriorityFeePerGas,
    maxFeePerGas,
  });

  logger.debug(`sendMessage transaction sent. transactionHash=${txHash}`);

  logger.debug(`Waiting for transaction to be mined... transactionHash=${txHash}`);
  const receipt = await l1PublicClient.waitForTransactionReceipt({ hash: txHash });

  logger.debug(`Transaction mined. transactionHash=${txHash} status=${receipt.status}`);

  return { txHash, receipt };
}

async function sendL2ToL1Message(
  logger: Logger,
  {
    l1Account,
    l2Account,
    fee = 0n,
    withCalldata = false,
  }: {
    l1Account: PrivateKeyAccount;
    l2Account: PrivateKeyAccount;
    fee: bigint;
    withCalldata: boolean;
  },
) {
  const dummyContract = config.l1WalletClient({ account: l1Account }).getDummyContract();
  const l2MessageService = config.l2WalletClient({ account: l2Account }).getL2MessageServiceContract();
  const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  const calldata = withCalldata
    ? encodeFunctionCall({
        abi: DummyContractAbi,
        functionName: "setPayload",
        args: [toHex(randomBytes(100).toString("hex"))],
      })
    : "0x";

  const destinationAddress = withCalldata ? await dummyContract.address : l1Account.address;

  const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await estimateLineaGas(l2PublicClient, {
    account: l2Account,
    to: l2MessageService.address,
    data: encodeFunctionCall({
      abi: L2MessageServiceV1Abi,
      functionName: "sendMessage",
      args: [destinationAddress, fee, calldata],
    }),
    value: etherToWei("0.001"),
  });
  logger.debug(`Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`);

  const txHash = await l2MessageService.write.sendMessage([destinationAddress, fee, calldata], {
    value: fee,
    maxPriorityFeePerGas,
    maxFeePerGas,
    gasLimit,
  });

  logger.debug(`sendMessage transaction sent. transactionHash=${txHash}`);

  logger.debug(`Waiting for transaction to be mined... transactionHash=${txHash}`);
  const receipt = await config.l1PublicClient().waitForTransactionReceipt({ hash: txHash });

  logger.debug(`Transaction mined. transactionHash=${txHash} status=${receipt.status}`);

  return { txHash, receipt };
}

const l1AccountManager = config.getL1AccountManager();
const l2AccountManager = config.getL2AccountManager();

describe("Messaging test suite", () => {
  it.concurrent(
    "Should send a transaction with fee and calldata to L1 message service, be successfully claimed it on L2",
    async () => {
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const { txHash, receipt } = await sendL1ToL2Message(logger, {
        l1Account,
        l2Account,
        fee: etherToWei("1.1"),
        withCalldata: true,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
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
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const { txHash, receipt } = await sendL1ToL2Message(logger, {
        l1Account,
        l2Account,
        fee: etherToWei("1.1"),
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
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
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const { txHash, receipt } = await sendL1ToL2Message(logger, {
        l1Account,
        l2Account,
        fee: 0n,
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L1 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for MessageClaimed event on L2. messageHash=${messageHash}`);
      const l2PublicClient = config.l2PublicClient();
      const [messageClaimedEvent] = await waitForEvents(l2PublicClient, {
        abi: L2MessageServiceV1Abi,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
      });

      expect(messageClaimedEvent).toBeDefined();
      logger.debug(
        `Message claimed on L2. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );

  it.concurrent(
    "Should send a transaction with with fee and calldata to L2 message service, be successfully claimed it on L1",
    async () => {
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const l1PublicClient = config.l1PublicClient();
      const { txHash, receipt } = await sendL2ToL1Message(logger, {
        l1Account,
        l2Account,
        fee: etherToWei("0.001"),
        withCalldata: true,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L2 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for L2MessagingBlockAnchored event... blockNumber=${messageSentEvent.blockNumber}`);
      await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        eventName: "L2MessagingBlockAnchored",
        args: {
          l2Block: messageSentEvent.blockNumber,
        },
        pollingIntervalMs: 1_000,
      });

      logger.debug(`Waiting for MessageClaimed event on L1... messageHash=${messageHash}`);
      const [messageClaimedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
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
      const [l1Account, l2Account] = await Promise.all([
        l1AccountManager.generateAccount(),
        l2AccountManager.generateAccount(),
      ]);

      const l1PublicClient = config.l1PublicClient();
      const { txHash, receipt } = await sendL2ToL1Message(logger, {
        l1Account,
        l2Account,
        fee: etherToWei("0.001"),
        withCalldata: false,
      });

      const [messageSentEvent] = getMessageSentEventFromLogs([receipt]);
      const messageHash = messageSentEvent.messageHash;
      logger.debug(`L2 message sent. messageHash=${messageHash} transactionHash=${txHash}`);

      logger.debug(`Waiting for L2MessagingBlockAnchored event... blockNumber=${messageSentEvent.blockNumber}`);
      await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        eventName: "L2MessagingBlockAnchored",
        args: {
          l2Block: messageSentEvent.blockNumber,
        },
        pollingIntervalMs: 1_000,
      });

      logger.debug(`Waiting for MessageClaimed event on L1. messageHash=${messageHash}`);
      const [messageClaimedEvent] = await waitForEvents(l1PublicClient, {
        abi: LineaRollupV6Abi,
        eventName: "MessageClaimed",
        args: {
          _messageHash: messageHash,
        },
        pollingIntervalMs: 1_000,
      });

      expect(messageClaimedEvent).toBeDefined();

      logger.debug(
        `Message claimed on L1. messageHash=${messageClaimedEvent.args._messageHash} transactionHash=${messageClaimedEvent.transactionHash}`,
      );
    },
    150_000,
  );
});
