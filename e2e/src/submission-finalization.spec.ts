import { beforeAll, describe, expect, it } from "@jest/globals";
import { BigNumber, Wallet, ethers } from "ethers";
import { OPERATOR_ROLE, ROLLING_HASH_UPDATED_EVENT_SIGNATURE, VERIFIER_SETTER_ROLE } from "./utils/constants";
import { getAndIncreaseFeeData } from "./utils/helpers";
import { MessageEvent } from "./utils/types";
import { getMessageSentEventFromLogs, sendMessage, sendTransactionsWithInterval, waitForEvents } from "./utils/utils";

const submissionAndFinalizationTestSuite = (title: string) => {
  describe(title, () => {
    let securityCouncil: Wallet;
    let l1Messages: MessageEvent[];
    let l2Messages: MessageEvent[];

    beforeAll(async () => {
      // Deploy new contracts implementation and grant roles
      securityCouncil = new Wallet(SECURITY_COUNCIL_PRIVATE_KEY, l1Provider);
      const securityCouncilNonce = await securityCouncil.getTransactionCount();

      const rolesTransactions = await Promise.all([
        lineaRollup
          .connect(securityCouncil)
          .grantRole(OPERATOR_ROLE, OPERATOR_1_ADDRESS, { nonce: securityCouncilNonce }),
        lineaRollup
          .connect(securityCouncil)
          .grantRole(VERIFIER_SETTER_ROLE, securityCouncil.address, { nonce: securityCouncilNonce + 1 }),
      ]);

      await Promise.all(rolesTransactions.map((tx) => tx.wait()));
    });

    it("Send messages on L1 and L2", async () => {
      const messageFee = ethers.utils.parseEther("0.0001");
      const messageValue = ethers.utils.parseEther("0.0051");
      const destinationAddress = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

      const l1MessageSender = new Wallet(L1_ACCOUNT_0_PRIVATE_KEY, l1Provider);
      const l2MessageSender = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

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

        l2MessagesPromises.push(
          sendMessage(
            l2MessageService.connect(l2MessageSender),
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
      );

      const [lastNewMessageRollingHash, lastAnchoredL1MessageNumber] = await Promise.all([
        lineaRollup.rollingHashes(lastNewL1MessageNumber),
        l2MessageService.lastAnchoredL1MessageNumber(),
      ]);
      expect(lastNewMessageRollingHash).toEqual(rollingHashUpdatedEvent.args.rollingHash);
      expect(lastAnchoredL1MessageNumber).toEqual(lastNewL1MessageNumber);

      console.log("New anchoring using rolling hash done.");
    }, 300_000);

    it("Check L1 data submission and finalization", async () => {
      // Send transactions on L2 in the background to make the L2 chain moving forward
      const l2MessageSender = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());
      const sendTransactionsPromise = sendTransactionsWithInterval(
        l2MessageSender,
        {
          to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
          value: ethers.utils.parseEther("0.0001"),
          maxPriorityFeePerGas,
          maxFeePerGas,
        },
        5_000,
      );

      const [currentL2BlockNumber, startingRootHash] = await Promise.all([
        lineaRollup.currentL2BlockNumber(),
        lineaRollup.stateRootHashes(await lineaRollup.currentL2BlockNumber()),
      ]);

      console.log("Waiting for data submission used to finalize with proof...");
      // Waiting for data submission starting from migration block number
      await waitForEvents(lineaRollup, lineaRollup.filters.DataSubmittedV2(null, currentL2BlockNumber.add(1)), 1_000);

      console.log("Waiting for the first DataFinalized event with proof...");
      // Waiting for first DataFinalized event with proof
      const [dataFinalizedEvent] = await waitForEvents(
        lineaRollup,
        lineaRollup.filters.DataFinalized(null, startingRootHash),
        1_000,
      );

      const [lastBlockFinalized, newStateRootHash] = await Promise.all([
        lineaRollup.currentL2BlockNumber(),
        lineaRollup.stateRootHashes(dataFinalizedEvent.args.lastBlockFinalized),
      ]);

      expect(lastBlockFinalized).toEqual(BigNumber.from(dataFinalizedEvent.args.lastBlockFinalized));
      expect(newStateRootHash).toEqual(dataFinalizedEvent.args.finalRootHash);
      expect(dataFinalizedEvent.args.withProof).toBeTruthy();

      console.log("Finalization with proof done.");

      clearInterval(sendTransactionsPromise);
    }, 300_000);

    it("Check L1 claiming", async () => {
      // Send transactions on L2 in the background to make the L2 chain moving forward
      const l2MessageSender = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());
      const sendTransactionsPromise = sendTransactionsWithInterval(
        l2MessageSender,
        {
          to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
          value: ethers.utils.parseEther("0.0001"),
          maxPriorityFeePerGas,
          maxFeePerGas,
        },
        1_000,
      );

      const { messageHash, messageNumber, blockNumber } = l2Messages[0];

      console.log(`Waiting for L2MessagingBlockAnchored... with blockNumber=${blockNumber}`);
      await waitForEvents(lineaRollup, lineaRollup.filters.L2MessagingBlockAnchored(blockNumber), 1_000);

      console.log("L2MessagingBlockAnchored event found.");

      await waitForEvents(lineaRollup, lineaRollup.filters.MessageClaimed(messageHash), 1_000);

      expect(await lineaRollup.isMessageClaimed(messageNumber)).toBeTruthy();
      clearInterval(sendTransactionsPromise);
    }, 400_000);
  });
};

export default submissionAndFinalizationTestSuite;
