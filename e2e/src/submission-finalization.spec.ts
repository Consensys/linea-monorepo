import { describe, expect, it } from "@jest/globals";
import { JsonRpcProvider } from "ethers";
import {
  getMessageSentEventFromLogs,
  sendMessage,
  waitForEvents,
  wait,
  getBlockByNumberOrBlockTag,
  etherToWei,
} from "./common/utils";
import { config } from "./config/tests-config";
import { LineaRollup } from "./typechain";

describe("Submission and finalization test suite", () => {
  let l1Provider: JsonRpcProvider;

  beforeAll(() => {
    l1Provider = config.getL1Provider();
  });

  const sendMessages = async () => {
    const messageFee = etherToWei("0.0001");
    const messageValue = etherToWei("0.0051");
    const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

    const l1MessageSender = await config.getL1AccountManager().generateAccount();
    const lineaRollup = config.getLineaRollupContract();

    console.log("Sending messages on L1");

    // Send L1 messages
    const l1MessagesPromises = [];
    // eslint-disable-next-line prefer-const
    let [l1MessageSenderNonce, { maxPriorityFeePerGas, maxFeePerGas }] = await Promise.all([
      l1Provider.getTransactionCount(l1MessageSender.address),
      l1Provider.getFeeData(),
    ]);

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
            maxPriorityFeePerGas,
            maxFeePerGas,
          },
        ),
      );
      l1MessageSenderNonce++;
    }

    const l1Receipts = await Promise.all(l1MessagesPromises);

    console.log("Messages sent on L1.");

    // Extract message events
    const l1Messages = getMessageSentEventFromLogs(lineaRollup, l1Receipts);

    return { l1Messages, l1Receipts };
  };

  async function getFinalizedL2BlockNumber(lineaRollup: LineaRollup) {
    let blockNumber = null;

    while (!blockNumber) {
      try {
        blockNumber = await lineaRollup.currentL2BlockNumber({ blockTag: "finalized" });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } catch (error: any) {
        console.log("No finalized block yet, retrying in 5 seconds...");
        await new Promise((resolve) => setTimeout(resolve, 5_000));
      }
    }

    return blockNumber;
  }

  it.concurrent(
    "Check L2 anchoring",
    async () => {
      const lineaRollup = config.getLineaRollupContract();
      const l2MessageService = config.getL2MessageServiceContract();

      const { l1Messages } = await sendMessages();

      // Wait for the last L1->L2 message to be anchored on L2
      const lastNewL1MessageNumber = l1Messages.slice(-1)[0].messageNumber;

      console.log("Waiting for the anchoring using rolling hash...");
      const [rollingHashUpdatedEvent] = await waitForEvents(
        l2MessageService,
        l2MessageService.filters.RollingHashUpdated(),
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
    },
    150_000,
  );

  it.concurrent(
    "Check L1 data submission and finalization",
    async () => {
      const lineaRollup = config.getLineaRollupContract();

      const [currentL2BlockNumber, startingRootHash] = await Promise.all([
        lineaRollup.currentL2BlockNumber(),
        lineaRollup.stateRootHashes(await lineaRollup.currentL2BlockNumber()),
      ]);

      console.log("Waiting for data submission used to finalize with proof...");
      // Waiting for data submission starting from migration block number
      await waitForEvents(
        lineaRollup,
        lineaRollup.filters.DataSubmittedV2(undefined, currentL2BlockNumber + 1n),
        1_000,
      );

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
    },
    150_000,
  );

  it.concurrent(
    "Check L2 safe/finalized tag update on sequencer",
    async () => {
      const lineaRollup = config.getLineaRollupContract();
      const sequencerEndpoint = config.getSequencerEndpoint();
      if (!sequencerEndpoint) {
        console.log('Skipped the "Check L2 safe/finalized tag update on sequencer" test');
        return;
      }

      const lastFinalizedL2BlockNumberOnL1 = (await getFinalizedL2BlockNumber(lineaRollup)).toString();
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
    },
    150_000,
  );
});
