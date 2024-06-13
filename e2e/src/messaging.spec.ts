import { Wallet, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { encodeFunctionCall, waitForEvents } from "./utils/utils";
import { getAndIncreaseFeeData } from "./utils/helpers";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "./utils/constants";

const messagingTestSuite = (title: string) => {
  describe(title, () => {
    describe("Message Service L1 -> L2", () => {
      it("Should send a transaction with calldata to L1 message service, be successfully claimed it on L2", async () => {
        const [lineaRollupBalance, l2MessageServiceBalance] = await Promise.all([
          l1Provider.getBalance(lineaRollup.address),
          l2Provider.getBalance(l2MessageService.address),
        ]);

        console.log(`T1 current L1 balance: ${ethers.utils.formatEther(lineaRollupBalance)} ETH`);
        console.log(`T1 current L2 balance: ${ethers.utils.formatEther(l2MessageServiceBalance)} ETH`);

        const account = new Wallet(L1_ACCOUNT_0_PRIVATE_KEY, l1Provider);
        const account2 = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const sendMessageCalldata = encodeFunctionCall(l1DummyContract.interface, "setPayload", [
          ethers.utils.randomBytes(100),
        ]);

        const valueAndFee = ethers.utils.parseEther("0.01");
        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1Provider.getFeeData());

        const nonce = await l1Provider.getTransactionCount(account.address, "pending");
        const tx = await lineaRollup
          .connect(account)
          .sendMessage(dummyContract.address, valueAndFee, sendMessageCalldata, {
            value: valueAndFee,
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          });

        //Retrieve message hashes
        console.log(`tx.hash: ${tx.hash}, dummyContract.address: ${dummyContract.address}`);

        const receipt = await tx.wait();

        console.log(`L1 receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);

        const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);

        const messageHash = messageSentEvent.topics[3];

        console.log(`tx1 messageHash: ${JSON.stringify(messageHash)}`);

        //Extra transactions to trigger anchoring
        for (let i = 0; i < 5; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account2.address, "pending"),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);
          const tx = await dummyContract
            .connect(account2)
            .setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
              nonce: nonce,
              maxPriorityFeePerGas,
              maxFeePerGas,
            });
          console.log(`T1: extra-tx nonce: ${nonce} hash: ${tx.hash}`);
        }

        const [messageClaimedEvent] = await waitForEvents(
          l2MessageService,
          l2MessageService.filters.MessageClaimed(messageHash),
          1_000,
        );

        console.log(`messageClaimed: ${JSON.stringify(messageClaimedEvent)}`);
        expect(messageClaimedEvent).toBeDefined();
      }, 320000);

      it("Should send a transaction without calldata to L1 message service, be successfully claimed it on L2", async () => {
        const [lineaRollupBalance, l2MessageServiceBalance] = await Promise.all([
          l1Provider.getBalance(lineaRollup.address),
          l2Provider.getBalance(l2MessageService.address),
        ]);

        console.log(`T2 current L1 balance: ${ethers.utils.formatEther(lineaRollupBalance)} ETH`);
        console.log(`T2 current L2 balance: ${ethers.utils.formatEther(l2MessageServiceBalance)} ETH`);

        const account = new Wallet(L1_ACCOUNT_0_PRIVATE_KEY, l1Provider);
        const account2 = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const value = ethers.utils.parseEther("0.001");
        const fee = ethers.utils.parseEther("0.0001");
        const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
        const calldata = "0x";

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1Provider.getFeeData());
        const tx = await lineaRollup.connect(account).sendMessage(to, fee, calldata, {
          value: value,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });

        console.log(`tx.hash: ${tx.hash}, dummyContract.address: ${dummyContract.address}`);
        const receipt = await tx.wait();
        console.log(`L1 receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);

        //Retrieve the message hash
        const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
        const messageHash = messageSentEvent.topics[3];

        console.log(`tx messageHash: ${JSON.stringify(messageHash)}`);

        //Extra transactions to trigger anchoring
        for (let i = 0; i < 5; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account2.address, "pending"),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);
          const tx = await dummyContract
            .connect(account2)
            .setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
              nonce: nonce,
              maxPriorityFeePerGas,
              maxFeePerGas,
            });
          console.log(`T2: extra-tx nonce: ${nonce} hash: ${tx.hash}`);
        }

        const [messageClaimed] = await waitForEvents(
          l2MessageService,
          l2MessageService.filters.MessageClaimed(messageHash),
          1_000,
        );
        expect(messageClaimed).not.toBeNull();
      }, 320000);
    });

    describe.skip("Message Service L2 -> L1", () => {
      it("Send transactions to L2 message service, should be delivered in L1", async () => {
        const [lineaRollupBalance, l2MessageServiceBalance] = await Promise.all([
          l1Provider.getBalance(lineaRollup.address),
          l2Provider.getBalance(l2MessageService.address),
        ]);

        console.log(`T3 current L1 balance: ${ethers.utils.formatEther(lineaRollupBalance)} ETH`);
        console.log(`T3 current L2 balance: ${ethers.utils.formatEther(l2MessageServiceBalance)} ETH`);

        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const sendMessageCalldata1 = encodeFunctionCall(dummyContract.interface, "setPayload", [
          ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000),
        ]);

        const sendMessageCalldata2 = encodeFunctionCall(dummyContract.interface, "setPayload", [
          ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000),
        ]);

        const valueAndFee = ethers.utils.parseEther("0.01");
        let nonce = await l2Provider.getTransactionCount(account.address, "pending");
        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l2Provider.getFeeData());

        const tx = await l2MessageService
          .connect(account)
          .sendMessage(l1DummyContract.address, valueAndFee, sendMessageCalldata1, {
            value: valueAndFee,
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });

        nonce = await l2Provider.getTransactionCount(account.address, "pending");

        const tx2 = await l2MessageService
          .connect(account)
          .sendMessage(l1DummyContract.address, valueAndFee, sendMessageCalldata2, {
            value: valueAndFee,
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });

        console.log(`tx.hash: ${tx.hash}`);
        console.log(`tx2.hash: ${tx2.hash}`);

        const [receipt, receipt2] = await Promise.all([tx.wait(), tx2.wait()]);

        console.log(`receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);
        console.log(`receipt2.blockNumber: ${JSON.stringify(receipt2.blockNumber)}`);

        //Extra transactions to trigger conflations
        // for (let i = 0; i < 5; i++) {
        //   const nonce = await l2Provider.getTransactionCount(account.address, "pending");
        //   const feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        //   const maxPriorityFeePerGas = feeData[0];
        //   const maxFeePerGas = feeData[1];
        //   await new Promise(res => setTimeout(res, 2000));
        // }

        //Retrieve the message hashes

        const [firstMessageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
        const [secondMessageSentEvent] = receipt2.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);

        const messageHash = firstMessageSentEvent.topics[3];
        const messageHash2 = secondMessageSentEvent.topics[3];

        console.log(`tx messageHash: ${JSON.stringify(messageHash)}`);
        console.log(`tx2 messageHash2: ${JSON.stringify(messageHash2)}`);

        expect(messageHash).not.toEqual(messageHash2);

        console.log(`currentL2BlockNumber: ${await lineaRollup.currentL2BlockNumber()}`);

        const [[firstMessageClaimedEvent], [secondMessageClaimedEvent]] = await Promise.all([
          waitForEvents(lineaRollup, lineaRollup.filters.MessageClaimed(messageHash), 1_000),
          waitForEvents(lineaRollup, lineaRollup.filters.MessageClaimed(messageHash2), 1_000),
        ]);

        console.log(`tx messageClaimed event: ${JSON.stringify(firstMessageClaimedEvent)}`);
        expect(firstMessageClaimedEvent).toBeDefined();

        console.log(`tx2 messageClaimed: ${JSON.stringify(secondMessageClaimedEvent)}`);
        expect(secondMessageClaimedEvent).toBeDefined();
      }, 720000);

      it("Should send a transaction without calldata to L2 message service, be successfully claimed it on L1", async () => {
        const [lineaRollupBalance, l2MessageServiceBalance] = await Promise.all([
          l1Provider.getBalance(lineaRollup.address),
          l2Provider.getBalance(l2MessageService.address),
        ]);

        console.log(`T4 current L1 balance: ${ethers.utils.formatEther(lineaRollupBalance)} ETH`);
        console.log(`T4 current L2 balance: ${ethers.utils.formatEther(l2MessageServiceBalance)} ETH`);

        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const value = ethers.utils.parseEther("0.001");
        const fee = ethers.utils.parseEther("0.00001");
        const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
        const calldata = "0x";

        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

        const tx = await l2MessageService.connect(account).sendMessage(to, fee, calldata, {
          value: value,
          maxPriorityFeePerGas,
          maxFeePerGas,
          gasLimit: 200000,
          nonce: nonce,
        });

        //Extra transactions to trigger conflations
        // for (let i = 0; i < 5; i++) {
        //   const nonce = await l2Provider.getTransactionCount(account.address, "pending");
        //   const feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        //   const maxPriorityFeePerGas = feeData[0];
        //   const maxFeePerGas = feeData[1];
        // }

        //Retrieve the message hash
        console.log(`tx.hash: ${tx.hash}`);
        const receipt = await tx.wait();

        const [messageSentEvent] = receipt.logs.filter((log) => log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE);
        console.log(`receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);

        expect(messageSentEvent).toBeDefined();

        const messageHash = messageSentEvent.topics[3];
        console.log(`tx messageHash: ${messageHash}`);

        const [messageClaimed] = await waitForEvents(
          lineaRollup,
          lineaRollup.filters.MessageClaimed(messageHash),
          1_000,
        );

        const [endLineaRollupBalance, endL2MessageServiceBalance] = await Promise.all([
          l1Provider.getBalance(lineaRollup.address),
          l2Provider.getBalance(l2MessageService.address),
        ]);

        console.log(`End current L1 balance: ${ethers.utils.formatEther(endLineaRollupBalance)} ETH`);
        console.log(`End current L2 balance: ${ethers.utils.formatEther(endL2MessageServiceBalance)} ETH`);

        expect(messageClaimed).toBeDefined();
      }, 720000);
    });
  });
};

export default messagingTestSuite;
