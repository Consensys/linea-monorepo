import { describe, expect, it } from "@jest/globals";
import {
  getMessageSentEventFromLogs,
  sendMessage,
  waitForEvents,
  wait,
  getBlockByNumberOrBlockTag,
  etherToWei,
} from "./common/utils";
import { config } from "./config/tests-config/setup";
import { L2MessageServiceV1Abi, LineaRollupV6Abi } from "./generated";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";

const l1AccountManager = config.getL1AccountManager();

describe("Submission and finalization test suite", () => {
  const sendMessages = async () => {
    const messageFee = etherToWei("0.0001");
    const messageValue = etherToWei("0.0051");
    const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

    const l1MessageSender = await l1AccountManager.generateAccount();
    const l1WalletClient = config.l1WalletClient({ account: l1MessageSender });

    logger.debug("Sending messages on L1...");

    // Send L1 messages
    const l1MessagesPromises = [];

    for (let i = 0; i < 5; i++) {
      l1MessagesPromises.push(
        sendMessage(l1WalletClient, {
          account: l1MessageSender,
          chain: l1WalletClient.chain,
          args: {
            to: destinationAddress,
            fee: messageFee,
            calldata: "0x",
          },
          contractAddress: config.l1PublicClient().getLineaRollup().address,
          value: messageValue,
        }),
      );
    }

    const l1Receipts = await Promise.all(l1MessagesPromises);

    logger.debug("Messages sent on L1.");

    // Extract message events
    const l1Messages = getMessageSentEventFromLogs(l1Receipts);

    return { l1Messages, l1Receipts };
  };

  describe("Contracts v6", () => {
    it.concurrent(
      "Check L2 anchoring",
      async () => {
        const lineaRollupV6 = config.l1PublicClient().getLineaRollup();
        const l1PublicClient = config.l1PublicClient();
        const l2MessageService = config.l2PublicClient().getL2MessageServiceContract();

        const { l1Messages } = await sendMessages();

        // Wait for the last L1->L2 message to be anchored on L2
        const lastNewL1MessageNumber = l1Messages.slice(-1)[0].messageNumber;

        logger.debug(`Waiting for the anchoring using rolling hash... messageNumber=${lastNewL1MessageNumber}`);
        const [rollingHashUpdatedEvent] = await waitForEvents(l1PublicClient, {
          abi: L2MessageServiceV1Abi,
          eventName: "RollingHashUpdated",
          fromBlock: 0n,
          toBlock: "latest",
          pollingIntervalMs: 1_000,
          criteria: async (events) => events.filter((event) => event.args.messageNumber! >= lastNewL1MessageNumber),
        });

        const [lastNewMessageRollingHash, lastAnchoredL1MessageNumber] = await Promise.all([
          lineaRollupV6.read.rollingHashes([rollingHashUpdatedEvent.args.messageNumber!]),
          l2MessageService.read.lastAnchoredL1MessageNumber(),
        ]);
        expect(lastNewMessageRollingHash).toEqual(rollingHashUpdatedEvent.args.rollingHash);
        expect(lastAnchoredL1MessageNumber).toEqual(rollingHashUpdatedEvent.args.messageNumber);

        logger.debug(`New anchoring using rolling hash done. rollingHash=${lastNewMessageRollingHash}`);
      },
      150_000,
    );

    it.concurrent(
      "Check L1 data submission and finalization",
      async () => {
        const lineaRollupV6 = config.l1PublicClient().getLineaRollup();
        const l1PublicClient = config.l1PublicClient();
        const currentL2BlockNumber = await lineaRollupV6.read.currentL2BlockNumber();

        logger.debug("Waiting for DataSubmittedV3 used to finalize with proof...");
        await waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
          eventName: "DataSubmittedV3",
          fromBlock: 0n,
          toBlock: "latest",
          pollingIntervalMs: 1_000,
        });

        logger.debug("Waiting for DataFinalizedV3 event with proof...");
        const [dataFinalizedEvent] = await waitForEvents(l1PublicClient, {
          abi: LineaRollupV6Abi,
          eventName: "DataFinalizedV3",
          fromBlock: 0n,
          toBlock: "latest",
          args: {
            startBlockNumber: currentL2BlockNumber + 1n,
          },
          pollingIntervalMs: 1_000,
        });

        const [lastBlockFinalized, newStateRootHash] = await Promise.all([
          lineaRollupV6.read.currentL2BlockNumber(),
          lineaRollupV6.read.stateRootHashes([dataFinalizedEvent.args.endBlockNumber!]),
        ]);

        expect(lastBlockFinalized).toBeGreaterThanOrEqual(dataFinalizedEvent.args.endBlockNumber!);
        expect(newStateRootHash).toEqual(dataFinalizedEvent.args.finalStateRootHash);

        logger.debug(`Finalization with proof done. lastFinalizedBlockNumber=${lastBlockFinalized}`);
      },
      150_000,
    );

    it.concurrent(
      "Check L2 safe/finalized tag update on sequencer",
      async () => {
        const sequencerClient = config.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
        if (!config.isLocal()) {
          logger.warn('Skipped the "Check L2 safe/finalized tag update on sequencer" test');
          return;
        }

        const lastFinalizedL2BlockNumberOnL1 = 0;
        logger.debug(`lastFinalizedL2BlockNumberOnL1=${lastFinalizedL2BlockNumberOnL1}`);

        let safeL2BlockNumber = -1,
          finalizedL2BlockNumber = -1;
        while (
          safeL2BlockNumber < lastFinalizedL2BlockNumberOnL1 ||
          finalizedL2BlockNumber < lastFinalizedL2BlockNumberOnL1
        ) {
          const currentSafeL2BlockNumber = (await getBlockByNumberOrBlockTag(sequencerClient, { blockTag: "safe" }))
            ?.number;
          const currentFinalizedL2BlockNumber = (
            await getBlockByNumberOrBlockTag(sequencerClient, { blockTag: "finalized" })
          )?.number;

          safeL2BlockNumber = currentSafeL2BlockNumber
            ? parseInt(currentSafeL2BlockNumber.toString())
            : safeL2BlockNumber;
          finalizedL2BlockNumber = currentFinalizedL2BlockNumber
            ? parseInt(currentFinalizedL2BlockNumber.toString())
            : finalizedL2BlockNumber;
          await wait(1_000);
        }

        logger.debug(`safeL2BlockNumber=${safeL2BlockNumber} finalizedL2BlockNumber=${finalizedL2BlockNumber}`);

        expect(safeL2BlockNumber).toBeGreaterThanOrEqual(lastFinalizedL2BlockNumberOnL1);
        expect(finalizedL2BlockNumber).toBeGreaterThanOrEqual(lastFinalizedL2BlockNumberOnL1);

        logger.debug("L2 safe/finalized tag update on sequencer done.");
      },
      150_000,
    );
  });
});
