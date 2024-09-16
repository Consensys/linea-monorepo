import {describe, expect, it} from "@jest/globals";
import {
  getEvents,
  execDockerCommand,
  waitForEvents,
  getMessageSentEventFromLogs,
  sendMessage,
  sendTransactionsWithInterval,
} from "./utils/utils";
import {getAndIncreaseFeeData} from "./utils/helpers";
import {Wallet, ethers} from "ethers";
// import { MessageEvent } from "./utils/types";

const coordinatorRestartTestSuite = (title: string) => {
  describe(title, () => {
    it("When the coordinator restarts it should resume blob submission and finalization", async () => {
      const l2Account0 = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());

      console.log("Moving the L2 chain forward to trigger conflation...");
      const intervalId = sendTransactionsWithInterval(
        l2Account0,
        {
          to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
          value: ethers.utils.parseEther("0.0001"),
          maxPriorityFeePerGas,
          maxFeePerGas,
        },
        1_000,
      )

      // await for a finalization to happen on L1
      await Promise.all([
        waitForEvents(lineaRollup, lineaRollup.filters.DataSubmittedV2(), 0, "latest"),
        waitForEvents(lineaRollup, lineaRollup.filters.DataFinalized(), 0, "latest"),
      ]);

      await execDockerCommand("stop", "coordinator");

      const currentBlockNumberBeforeRestart = await l1Provider.getBlockNumber();
      const [dataSubmittedEventsBeforeRestart, dataFinalizedEventsBeforeRestart] = await Promise.all([
        getEvents(lineaRollup, lineaRollup.filters.DataSubmittedV2(), 0, currentBlockNumberBeforeRestart),
        getEvents(lineaRollup, lineaRollup.filters.DataFinalized(), 0, currentBlockNumberBeforeRestart),
      ]);

      const lastDataSubmittedEventBeforeRestart = dataSubmittedEventsBeforeRestart.slice(-1)[0];
      const lastDataFinalizedEventsBeforeRestart = dataFinalizedEventsBeforeRestart.slice(-1)[0];
      // Just some sanity checks
      // Check that the coordinator has submitted and finalized data before the restart
      expect(lastDataSubmittedEventBeforeRestart.args.endBlock.toNumber()).toBeGreaterThan(0)
      expect(lastDataFinalizedEventsBeforeRestart.args.lastBlockFinalized.toNumber()).toBeGreaterThan(0)

      await execDockerCommand("start", "coordinator");
      const currentBlockNumberAfterRestart = await l1Provider.getBlockNumber();

      console.log("Waiting for DataSubmittedV2 event after coordinator restart...");
      const [dataSubmittedV2EventAfterRestart] = await waitForEvents(
        lineaRollup,
        lineaRollup.filters.DataSubmittedV2(null, lastDataSubmittedEventBeforeRestart.args.endBlock.add(1)),
        1_000,
        currentBlockNumberAfterRestart,
      );
      console.log(`New DataSubmittedV2 event found: event=${JSON.stringify(dataSubmittedV2EventAfterRestart)}`);

      console.log("Waiting for DataFinalized event after coordinator restart...");
      const [dataFinalizedEventAfterRestart] = await waitForEvents(
        lineaRollup,
        lineaRollup.filters.DataFinalized(),
        1_000,
        currentBlockNumberAfterRestart,
        "latest",
        async (events) => {
          return events.filter((event) =>
            event.args.lastBlockFinalized.gt(lastDataFinalizedEventsBeforeRestart.args.lastBlockFinalized),
          );
        },
      );
      console.log(`New DataFinalized event found: event=${JSON.stringify(dataFinalizedEventAfterRestart)}`);
      clearInterval(intervalId)

      expect(dataFinalizedEventAfterRestart.args.lastBlockFinalized.toNumber()).toBeGreaterThan(
        lastDataFinalizedEventsBeforeRestart.args.lastBlockFinalized.toNumber(),
      );
    }, 300_000);

    it("When the coordinator restarts it should resume anchoring", async () => {
      const l1MessageSender = new Wallet(L1_ACCOUNT_0_PRIVATE_KEY, l1Provider);
      const l2MessageSender = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

      // Send Messages L1 -> L2
      const messageFee = ethers.utils.parseEther("0.0001");
      const messageValue = ethers.utils.parseEther("0.0051");
      const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

      const l1MessagesPromises = [];
      let l1MessageSenderNonce = await l1Provider.getTransactionCount(l1MessageSender.address);
      const l1Fees = getAndIncreaseFeeData(await l1Provider.getFeeData());

      for (let i = 0; i < 5; i++) {
        l1MessagesPromises.push(
          sendMessage(
            lineaRollup.connect(l1MessageSender),
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
      }

      const l1Receipts = await Promise.all(l1MessagesPromises);
      const l1Messages = getMessageSentEventFromLogs(lineaRollup, l1Receipts);

      // Wait for L2 Anchoring
      const lastNewL1MessageNumber = l1Messages.slice(-1)[0].messageNumber;

      console.log(`Waiting L1->L2 anchoring messageNumber=${lastNewL1MessageNumber}`);
      await waitForEvents(l2MessageService, l2MessageService.filters.RollingHashUpdated(lastNewL1MessageNumber), 1_000);

      // Restart Coordinator
      await execDockerCommand("restart", "coordinator");

      // Send more messages L1 -> L2
      for (let i = 0; i < 5; i++) {
        l1MessagesPromises.push(
          sendMessage(
            lineaRollup.connect(l1MessageSender),
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
      }

      const l1ReceiptsAfterRestart = await Promise.all(l1MessagesPromises);
      const l1MessagesAfterRestart = getMessageSentEventFromLogs(lineaRollup, l1ReceiptsAfterRestart);

      console.log("Moving the L2 chain forward to trigger anchoring...");
      // Using 5 messages to give the coordinator time to restart
      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());
      const intervalId = sendTransactionsWithInterval(
        l2MessageSender,
        {
          to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
          value: ethers.utils.parseEther("0.0001"),
          maxPriorityFeePerGas,
          maxFeePerGas,
        },
        1_000,
      )

      // Wait for messages to be anchored on L2
      const lastNewL1MessageNumberAfterRestart = l1MessagesAfterRestart.slice(-1)[0].messageNumber;

      console.log(
        `Waiting L1->L2 anchoring after coordinator restart messageNumber=${lastNewL1MessageNumberAfterRestart}`
      );
      const [rollingHashUpdatedEventAfterRestart] = await waitForEvents(
        l2MessageService,
        l2MessageService.filters.RollingHashUpdated(lastNewL1MessageNumberAfterRestart),
        1_000,
      );

      const [lastNewMessageRollingHashAfterRestart, lastAnchoredL1MessageNumberAfterRestart] = await Promise.all([
        lineaRollup.rollingHashes(lastNewL1MessageNumberAfterRestart),
        l2MessageService.lastAnchoredL1MessageNumber(),
      ]);

      clearInterval(intervalId)

      expect(lastNewMessageRollingHashAfterRestart).toEqual(rollingHashUpdatedEventAfterRestart.args.rollingHash);
      expect(lastAnchoredL1MessageNumberAfterRestart).toEqual(lastNewL1MessageNumberAfterRestart);
    }, 300_000);
  });
};

export default coordinatorRestartTestSuite;
