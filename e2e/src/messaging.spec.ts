import { ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { config } from "../config";
import { encodeFunctionCall, sendTransactionsToGenerateTrafficWithInterval, waitForEvents } from "./common/utils";
import { getAndIncreaseFeeData } from "./common/helpers";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "./common/constants";

describe("Messaging test suite", () => {
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
        const [l1Account, l2Account] = await Promise.all([
          config.getL1AccountManager().generateAccount(),
          config.getL2AccountManager().generateAccount(),
        ]);

        const dummyContract = await config.getL2DummyContract(l2Account);
        const lineaRollup = config.getLineaRollupContract(l1Account);

        const valueAndFee = ethers.parseEther("1.1");
        const calldata = withCalldata
          ? encodeFunctionCall(dummyContract.interface, "setPayload", [ethers.randomBytes(100)])
          : "0x";
        const destinationAddress = withCalldata
          ? await dummyContract.getAddress()
          : "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

        const l1Provider = config.getL1Provider();
        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1Provider.getFeeData());
        const nonce = await l1Provider.getTransactionCount(l1Account.address, "pending");
        const tx = await lineaRollup.sendMessage(destinationAddress, valueAndFee, calldata, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas: maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas,
        });

        let receipt = await tx.wait();
        while (!receipt) {
          console.log("Waiting for transaction to be mined...");
          receipt = await tx.wait();
        }

        console.log("Moving the L2 chain forward to trigger anchoring...");
        const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(l2Account);

        const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
        const messageHash = messageSentEvent.topics[3];
        console.log(`L1 message sent: messageHash=${messageHash} transaction=${JSON.stringify(tx)}`);

        console.log("Waiting for MessageClaimed event on L2.");
        const l2MessageService = config.getL2MessageServiceContract();
        const [messageClaimedEvent] = await waitForEvents(
          l2MessageService,
          l2MessageService.filters.MessageClaimed(messageHash),
        );
        stopPolling();
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
        const [l1Account, [l2Account, l2AccountForLiveness]] = await Promise.all([
          config.getL1AccountManager().generateAccount(),
          config.getL2AccountManager().generateAccounts(2),
        ]);

        const l2Provider = config.getL2Provider();

        const dummyContract = await config.getL1DummyContract(l1Account);
        const l2MessageService = config.getL2MessageServiceContract(l2Account);
        const lineaRollup = config.getLineaRollupContract();

        const valueAndFee = ethers.parseEther("0.001");
        const calldata = withCalldata
          ? encodeFunctionCall(dummyContract.interface, "setPayload", [ethers.randomBytes(100)])
          : "0x";

        const destinationAddress = withCalldata ? await dummyContract.getAddress() : l1Account.address;
        const nonce = await l2Provider.getTransactionCount(l2Account.address, "pending");
        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());

        const tx = await l2MessageService.sendMessage(destinationAddress, valueAndFee, calldata, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });

        const receipt = await tx.wait();

        if (!receipt) {
          throw new Error("Transaction receipt is undefined");
        }

        const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
        const messageHash = messageSentEvent.topics[3];
        console.log(`L2 message sent: messageHash=${messageHash} transaction=${JSON.stringify(tx)}`);

        console.log("Moving the L2 chain forward to trigger conflation...");
        const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(l2AccountForLiveness);

        console.log("Waiting for MessageClaimed event on L1.");
        const [messageClaimedEvent] = await waitForEvents(lineaRollup, lineaRollup.filters.MessageClaimed(messageHash));
        stopPolling();

        console.log(`Message claimed on L1: ${JSON.stringify(messageClaimedEvent)}`);
        expect(messageClaimedEvent).toBeDefined();
      },
      300_000,
    );
  });
});
