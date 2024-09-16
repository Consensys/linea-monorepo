import { Wallet, ethers } from "ethers";
import { beforeAll, describe, expect, it } from "@jest/globals";
import {
  encodeFunctionCall,
  sendTransactionsToGenerateTrafficWithInterval,
  waitForEvents
} from "./utils/utils";
import { getAndIncreaseFeeData } from "./utils/helpers";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "./utils/constants";

const messagingTestSuite = (title: string) => {
  describe(title, () => {
    let l1Account: Wallet;
    let l2Account0: Wallet;

    beforeAll(() => {
      l1Account = new Wallet(L1_ACCOUNT_0_PRIVATE_KEY, l1Provider);
      l2Account0 = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
    });

    describe("Message Service L1 -> L2", () => {
      it.each([
        {
          subTitle: "with calldata",
          withCalldata: true,
        },
        {
          subTitle: "without calldata",
          withCalldata: false,
        },
      ])(
        "Should send a transaction $subTitle to L1 message service, be successfully claimed it on L2",
        async ({ withCalldata }) => {
          const valueAndFee = ethers.utils.parseEther("1.1");
          const calldata = withCalldata
            ? encodeFunctionCall(dummyContract.interface, "setPayload", [ethers.utils.randomBytes(100)])
            : "0x";
          const destinationAddress = withCalldata
            ? dummyContract.address
            : "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1Provider.getFeeData());
          const nonce = await l1Provider.getTransactionCount(l1Account.address, "pending");
          const tx = await lineaRollup.connect(l1Account).sendMessage(destinationAddress, valueAndFee, calldata, {
            value: valueAndFee,
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          });

          const receipt = await tx.wait();

          console.log("Moving the L2 chain forward to trigger anchoring...");
          const intervalId = await sendTransactionsToGenerateTrafficWithInterval(l2Account0);

          const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
          const messageHash = messageSentEvent.topics[3];

          console.log(`L1 message sent: messageHash=${messageHash} transaction=${JSON.stringify(tx)}`);

          //Extra transactions to trigger anchoring
          console.log("Waiting for MessageClaimed event on L2.");
          const [messageClaimedEvent] = await waitForEvents(
            l2MessageService,
            l2MessageService.filters.MessageClaimed(messageHash),
          );
          clearInterval(intervalId);
          console.log(`Message claimed on L2: ${JSON.stringify(messageClaimedEvent)}`);
          expect(messageClaimedEvent).toBeDefined();
        },
        300_000,
      );
    });

    describe("Message Service L2 -> L1", () => {
      it.each([
        {
          subTitle: "with calldata",
          withCalldata: true,
        },
        {
          subTitle: "without calldata",
          withCalldata: false,
        },
      ])(
        "Should send a transaction $subTitle to L2 message service, be successfully claimed it on L1",
        async ({ withCalldata }) => {
          const valueAndFee = ethers.utils.parseEther("0.001");
          const calldata = withCalldata
            ? encodeFunctionCall(l1DummyContract.interface, "setPayload", [ethers.utils.randomBytes(100)])
            : "0x";

          const destinationAddress = withCalldata ? l1DummyContract.address : l1Account.address;
          const nonce = await l2Provider.getTransactionCount(l2Account0.address, "pending");
          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());

          const tx = await l2MessageService.connect(l2Account0).sendMessage(destinationAddress, valueAndFee, calldata, {
            value: valueAndFee,
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });

          const receipt = await tx.wait();

          const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
          const messageHash = messageSentEvent.topics[3];
          console.log(`L2 message sent: messageHash=${messageHash} transaction=${JSON.stringify(tx)}`);

          console.log("Moving the L2 chain forward to trigger conflation...");
          const intervalId = await sendTransactionsToGenerateTrafficWithInterval(l2Account0);

          console.log("Waiting for MessageClaimed event on L1.");
          const [messageClaimedEvent] = await waitForEvents(
            lineaRollup,
            lineaRollup.filters.MessageClaimed(messageHash)
          );
          clearInterval(intervalId);

          console.log(`Message claimed on L1: ${JSON.stringify(messageClaimedEvent)}`);
          expect(messageClaimedEvent).toBeDefined();
        },
        300_000,
      );
    });
  });
};

export default messagingTestSuite;
