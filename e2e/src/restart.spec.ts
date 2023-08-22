import { Wallet, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { encodeFunctionCall } from "./utils/utils";
import { ACCOUNT_0_PRIVATE_KEY, OPERATOR_0_PRIVATE_KEY, TRANSACTION_CALLDATA_LIMIT } from "./utils/constants.local";

describe.skip("Coordinator restart test suite", () => {
  describe("Message Service L1 -> L2", () => {
    it("When the coordinator restarts or crashes messages are not replayed", async () => {
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const l1Account = new Wallet(OPERATOR_0_PRIVATE_KEY, l1Provider);
      const sendMessageCalldata = encodeFunctionCall(dummyContract.interface, "setPayload", [
        ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000),
      ]);
      const tx = await zkEvmV2
        .connect(account)
        .sendMessage(dummyContract.address, 0, sendMessageCalldata, { value: ethers.utils.parseEther("1.0") });
      const receipt = await tx.wait(); 

      //TODO Claim message
      let messageClaimedEvent = await l2MessageService.filters.MessageClaimed()
      expect((await l2MessageService.queryFilter(messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length).toEqual(1);
      //Restart coordinator
      await testingEnv.restartCoordinator(useLocalSetup);

      //TODO Attempt to claim message again
      expect((await l2MessageService.queryFilter(messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length).toEqual(1);
    });
  });

  describe("Message Service L2 -> L1", () => {
    it("When the coordinator restarts or crashes messages are not replayed", async () => {
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const l1Account = new Wallet(OPERATOR_0_PRIVATE_KEY, l1Provider);
      const sendMessageCalldata = encodeFunctionCall(dummyContract.interface, "setPayload", [
        ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000),
      ]);
      const tx = await l2MessageService
        .connect(account)
        .sendMessage(dummyContract.address, 0, sendMessageCalldata, { value: ethers.utils.parseEther("1.0") });
        const receipt = await tx.wait(); 

        //TODO Claim message
        let messageClaimedEvent = await zkEvmV2.filters.MessageClaimed()
        expect((await zkEvmV2.queryFilter(messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length).toEqual(1);
        //Restart coordinator
        await testingEnv.restartCoordinator(useLocalSetup);
  
        //TODO Attempt to claim message again
        expect((await zkEvmV2.queryFilter(messageClaimedEvent, receipt.blockNumber, receipt.blockNumber)).length).toEqual(1);
    });
  });
});
