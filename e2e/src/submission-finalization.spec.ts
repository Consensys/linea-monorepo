import { describe, expect, it } from "@jest/globals";
import { ethers, JsonRpcProvider } from "ethers";
import { ROLLING_HASH_UPDATED_EVENT_SIGNATURE } from "./common/constants";
import { getAndIncreaseFeeData } from "./common/helpers";
import { MessageEvent } from "./common/types";
import {
  getMessageSentEventFromLogs,
  sendMessage,
  waitForEvents,
  wait,
  getBlockByNumberOrBlockTag,
  sendTransactionsToGenerateTrafficWithInterval,
} from "./common/utils";
import { config } from "../config";
import { L2MessageService, LineaRollup } from "./typechain";

describe("Submission and finalization test suite", () => {
  let l1Messages: MessageEvent[];
  let l2Messages: MessageEvent[];
  let l1Provider: JsonRpcProvider;
  let l2Provider: JsonRpcProvider;
  let lineaRollup: LineaRollup;
  let l2MessageService: L2MessageService;

  beforeAll(() => {
    l1Provider = config.getL1Provider();
    l2Provider = config.getL2Provider();
    lineaRollup = config.getLineaRollupContract();
    l2MessageService = config.getL2MessageServiceContract();
  });

  it("Send messages on L1 and L2", async () => {
    const messageFee = ethers.parseEther("0.0001");
    const messageValue = ethers.parseEther("0.0051");
    const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

    const [l1MessageSender, l2MessageSender] = await Promise.all([
      config.getL1AccountManager().generateAccount(),
      config.getL2AccountManager().generateAccount(),
    ]);

    console.log("Sending messages on L1 and L2...");

    // Send L1 messages
    const l1MessagesPromises = [];
    let l1MessageSenderNonce = await l1Provider.getTransactionCount(l1MessageSender.address);
    const l1Fees = getAndIncreaseFeeData(await l1Provider.getFeeData());

    const l2MessagesPromises = [];
    let l2MessageSenderNonce = await l2Provider.getTransactionCount(l2MessageSender.address);
    const l2Fees = getAndIncreaseFeeData(await l2Provider.getFeeData());

    for (let i = 0; i < 5; i++) {
      l1MessagesPromises.push(
        sendMessage(
          l1MessageSender,
          lineaRollup,
          {
            to: destinationAddress,
            fee: messageFee,
            calldata: "0x",
          },
          {
            value: messageValue,
            nonce: l1MessageSenderNonce,
            maxPriorityFeePerGas: l1Fees[0],
            maxFeePerGas: l1Fees[1],
          },
        ),
      );
      l1MessageSenderNonce++;

      l2MessagesPromises.push(
        sendMessage(
          l2MessageSender,
          l2MessageService,
          {
            to: destinationAddress,
            fee: messageFee,
            calldata: "0x",
          },
          {
            value: messageValue,
            nonce: l2MessageSenderNonce,
            maxPriorityFeePerGas: l2Fees[0],
            maxFeePerGas: l2Fees[1],
          },
        ),
      );
      l2MessageSenderNonce++;
    }

    const l1Receipts = await Promise.all(l1MessagesPromises);
    const l2Receipts = await Promise.all(l2MessagesPromises);

    console.log("Messages sent on L1 and L2.");

    // Check that L1 messages emit RollingHashUpdated events
    expect(l1Receipts.length).toBeGreaterThan(0);

    const newL1MessagesRollingHashUpdatedLogs = l1Receipts
      .flatMap((receipt) => receipt.logs)
      .filter((log) => log.topics[0] === ROLLING_HASH_UPDATED_EVENT_SIGNATURE);

    expect(newL1MessagesRollingHashUpdatedLogs).toHaveLength(l1Receipts.length);

    // Check that there are L2 messages
    expect(l2Receipts.length).toBeGreaterThan(0);

    l1Messages = getMessageSentEventFromLogs(lineaRollup, l1Receipts);
    l2Messages = getMessageSentEventFromLogs(l2MessageService, l2Receipts);
  }, 300_000);

  it("Check L2 anchoring", async () => {
    // Wait for the last L1->L2 message to be anchored on L2
    const lastNewL1MessageNumber = l1Messages.slice(-1)[0].messageNumber;

    console.log("Waiting for the anchoring using rolling hash...");
    const [rollingHashUpdatedEvent] = await waitForEvents(
      l2MessageService,
      l2MessageService.filters.RollingHashUpdated(lastNewL1MessageNumber),
      1_000,
      0,
      "latest",
      async (events) => events.filter((event) => event.args.messageNumber >= lastNewL1MessageNumber),
    );

    const [lastNewMessageRollingHash, lastAnchoredL1MessageNumber] = await Promise.all([
      lineaRollup.rollingHashes(rollingHashUpdatedEvent.args.messageNumber),
      l2MessageService.lastAnchoredL1MessageNumber(),
    ]);
    expect(lastNewMessageRollingHash).toEqual(rollingHashUpdatedEvent.args.rollingHash);
    expect(lastAnchoredL1MessageNumber).toEqual(rollingHashUpdatedEvent.args.messageNumber);

    console.log("New anchoring using rolling hash done.");
  }, 300_000);

  it("Check L1 data submission and finalization", async () => {
    const l2MessageSender = await config.getL2AccountManager().generateAccount();

    // Send transactions on L2 in the background to make the L2 chain moving forward
    const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(l2MessageSender, 5_000);

    const [currentL2BlockNumber, startingRootHash] = await Promise.all([
      lineaRollup.currentL2BlockNumber(),
      lineaRollup.stateRootHashes(await lineaRollup.currentL2BlockNumber()),
    ]);

    console.log("Waiting for data submission used to finalize with proof...");
    // Waiting for data submission starting from migration block number
    await waitForEvents(lineaRollup, lineaRollup.filters.DataSubmittedV2(undefined, currentL2BlockNumber + 1n), 1_000);

    console.log("Waiting for the first DataFinalized event with proof...");
    // Waiting for first DataFinalized event with proof
    const [dataFinalizedEvent] = await waitForEvents(
      lineaRollup,
      lineaRollup.filters.DataFinalized(undefined, startingRootHash),
      1_000,
    );

    const [lastBlockFinalized, newStateRootHash] = await Promise.all([
      lineaRollup.currentL2BlockNumber(),
      lineaRollup.stateRootHashes(dataFinalizedEvent.args.lastBlockFinalized),
    ]);

    expect(lastBlockFinalized).toBeGreaterThanOrEqual(dataFinalizedEvent.args.lastBlockFinalized);
    expect(newStateRootHash).toEqual(dataFinalizedEvent.args.finalRootHash);
    expect(dataFinalizedEvent.args.withProof).toBeTruthy();

    console.log("Finalization with proof done.");

    stopPolling();
  }, 300_000);

  it("Check L2 safe/finalized tag update on sequencer", async () => {
    const sequencerEndpoint = config.getSequencerEndpoint();
    if (!sequencerEndpoint) {
      console.log('Skipped the "Check L2 safe/finalized tag update on sequencer" test');
      return;
    }

    const lastFinalizedL2BlockNumberOnL1 = (
      await lineaRollup.currentL2BlockNumber({ blockTag: "finalized" })
    ).toString();
    console.log(`lastFinalizedL2BlockNumberOnL1=${lastFinalizedL2BlockNumberOnL1}`);

    let safeL2BlockNumber = -1,
      finalizedL2BlockNumber = -1;
    while (
      safeL2BlockNumber < parseInt(lastFinalizedL2BlockNumberOnL1) ||
      finalizedL2BlockNumber < parseInt(lastFinalizedL2BlockNumberOnL1)
    ) {
      safeL2BlockNumber = (await getBlockByNumberOrBlockTag(sequencerEndpoint, "safe"))?.number || safeL2BlockNumber;
      finalizedL2BlockNumber =
        (await getBlockByNumberOrBlockTag(sequencerEndpoint, "finalized"))?.number || finalizedL2BlockNumber;
      await wait(1_000);
    }

    console.log(`safeL2BlockNumber=${safeL2BlockNumber} finalizedL2BlockNumber=${finalizedL2BlockNumber}`);

    expect(safeL2BlockNumber).toBeGreaterThanOrEqual(parseInt(lastFinalizedL2BlockNumberOnL1));
    expect(finalizedL2BlockNumber).toBeGreaterThanOrEqual(parseInt(lastFinalizedL2BlockNumberOnL1));

    console.log("L2 safe/finalized tag update on sequencer done.");
  }, 300_000);

  it("Check L1 claiming", async () => {
    const l2MessageSender = await config.getL2AccountManager().generateAccount();

    // Send transactions on L2 in the background to make the L2 chain moving forward
    const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(l2MessageSender, 2_000);

    const { messageHash, messageNumber, blockNumber } = l2Messages[0];

    console.log(`Waiting for L2MessagingBlockAnchored... with blockNumber=${blockNumber}`);
    await waitForEvents(lineaRollup, lineaRollup.filters.L2MessagingBlockAnchored(blockNumber), 1_000);

    console.log("L2MessagingBlockAnchored event found.");

    await waitForEvents(lineaRollup, lineaRollup.filters.MessageClaimed(messageHash), 1_000);

    expect(await lineaRollup.isMessageClaimed(messageNumber)).toBeTruthy();

    console.log("L1 claiming done.");

    stopPolling();
  }, 400_000);
});
