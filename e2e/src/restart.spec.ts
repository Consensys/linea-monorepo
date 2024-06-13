import { Wallet, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { encodeFunctionCall, getEvents } from "./utils/utils";
import { TRANSACTION_CALLDATA_LIMIT } from "./utils/constants.local";

const coordinatorRestartTestSuite = (title: string) => {
  describe.skip(title, () => {
    describe("Message Service L1 -> L2", () => {
      it("When the coordinator restarts or crashes messages are not replayed", async () => {
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
        const sendMessageCalldata = encodeFunctionCall(dummyContract.interface, "setPayload", [
          ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000),
        ]);

        const tx = await lineaRollup
          .connect(account)
          .sendMessage(dummyContract.address, 0, sendMessageCalldata, { value: ethers.utils.parseEther("1.0") });
        const receipt = await tx.wait();

        //TODO Claim message
        const messageClaimedEvent = l2MessageService.filters.MessageClaimed();
        expect(
          (await getEvents(l2MessageService, messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length,
        ).toEqual(1);
        //Restart coordinator
        await testingEnv.restartCoordinator(useLocalSetup);

        //TODO Attempt to claim message again
        expect(
          (await getEvents(l2MessageService, messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length,
        ).toEqual(1);
      });
    });

    describe("Message Service L2 -> L1", () => {
      it("When the coordinator restarts or crashes messages are not replayed", async () => {
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
        const sendMessageCalldata = encodeFunctionCall(dummyContract.interface, "setPayload", [
          ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000),
        ]);

        const tx = await l2MessageService
          .connect(account)
          .sendMessage(dummyContract.address, 0, sendMessageCalldata, { value: ethers.utils.parseEther("1.0") });
        const receipt = await tx.wait();

        //TODO Claim message
        const messageClaimedEvent = lineaRollup.filters.MessageClaimed();
        expect(
          (await getEvents(lineaRollup, messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length,
        ).toEqual(1);
        //Restart coordinator
        await testingEnv.restartCoordinator(useLocalSetup);

        //TODO Attempt to claim message again
        expect(
          (await getEvents(lineaRollup, messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length,
        ).toEqual(1);
      });
    });
  });
};

export default coordinatorRestartTestSuite;
