import { BigNumber, Wallet, ethers } from "ethers";
import { describe, expect, it, jest } from "@jest/globals";
import { encodeFunctionCall } from "./utils/utils";
import { MessageClaimedEvent } from "./typechain/L2MessageService";
import { getAndIncreaseFeeData } from "./utils/helpers";

describe("Messaging test suite", () => {
  describe("Message Service L1 -> L2", () => {
    it.skip("Should send a transaction with calldata to L1 message service, be successfully claimed it on L2", async () => {
      console.log(`T1 current L2 balance: ${await l2Provider.getBalance(l2MessageService.address)}`);
      console.log(`T1 current L1 balance: ${await l1Provider.getBalance(zkEvmV2.address)}`);
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l1Provider);
      const account2 = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const sendMessageCalldata1 = encodeFunctionCall(l1DummyContract.interface, "setPayload", [
        ethers.utils.randomBytes(1000),
      ]);
      const sendMessageCalldata2 = encodeFunctionCall(l1DummyContract.interface, "setPayload", [
        ethers.utils.randomBytes(1000),
      ]);

      const valueAndFee = ethers.utils.parseEther("0.01")
      let feeData = getAndIncreaseFeeData(await l1Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];
      let nonce = await l1Provider.getTransactionCount(account.address, "pending");
      const tx = await zkEvmV2
        .connect(account)
        .sendMessage(dummyContract.address, valueAndFee, sendMessageCalldata1, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas: maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas
        });

      nonce = await l1Provider.getTransactionCount(account.address, "pending");
      const tx2 = await zkEvmV2
        .connect(account)
        .sendMessage(dummyContract.address, valueAndFee, sendMessageCalldata2, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas: maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas
        });
      const messageSentFilter = zkEvmV2.filters.MessageSent();

      //Retrieve the message hash
      console.log(`tx.hash: ${tx.hash}`);
      console.log(`tx2.hash: ${tx2.hash}`);
      const receipt = await tx.wait();
      const receipt2 = await tx2.wait();
      console.log(`L1 receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);
      console.log(`L1 receipt2.blockNumber: ${JSON.stringify(receipt2.blockNumber)}`);
      let messageHash: string | null | undefined = null;
      while (!messageHash) {
        let events = (await zkEvmV2.queryFilter(messageSentFilter, receipt.blockNumber, receipt.blockNumber));
        messageHash = events.filter((event) => {
          return (
            event.args._to == dummyContract.address &&
            event.args._value.eq(BigNumber.from(0)) &&
            event.args._calldata == sendMessageCalldata1)
        }).at(0)?.args._messageHash;
      }
      console.log(`tx1 messageHash: ${JSON.stringify(messageHash)}`);

      //Retrieve the message hash
      let messageHash2: string | null | undefined = null;
      while (!messageHash2) {
        let events = (await zkEvmV2.queryFilter(messageSentFilter, receipt2.blockNumber, receipt2.blockNumber));
        messageHash2 = events.filter((event) => {
          return (
            event.args._to == dummyContract.address &&
            event.args._value.eq(BigNumber.from(0)) &&
            event.args._calldata == sendMessageCalldata2);
        }).at(0)?.args._messageHash;
      }
      console.log(`tx2 messageHash2: ${JSON.stringify(messageHash2)}`);

      expect(messageHash).not.toEqual(messageHash2);

      //Extra transactions to trigger anchoring
      for (let i = 0; i < 10; i++) {
        let nonce = await l2Provider.getTransactionCount(account2.address, "pending");
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account2).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas
          });
        const receipt = await tx.wait();
        console.log(`extra-tx receipt.blockNumber: ${receipt.blockNumber}`);
        //await new Promise(res => setTimeout(res, 2000));
      }

      let messageClaimedEvent = l2MessageService.filters.MessageClaimed();
      let messageClaimed = null;
      let fromBlock = (await l2Provider.getBlock("latest")).number - 100;
      while (!messageClaimed) {
        let events = (await l2MessageService.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return (event.args._messageHash == messageHash);
        }).at(0);
      }
      console.log(`messageClaimed1: ${JSON.stringify(messageClaimed)}`);
      expect(messageClaimed).not.toBeNull;

      messageClaimed = null;
      while (!messageClaimed) {
        let events = (await l2MessageService.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return (event.args._messageHash == messageHash2);
        });
      }
      console.log(`messageClaimed2: ${JSON.stringify(messageClaimed)}`);
      expect(messageClaimed).not.toBeNull;
    }, 320000);

    it("Should send a transaction without calldata to L1 message service, be successfully claimed it on L2", async () => {
      console.log(`T2 current L2 balance: ${await l2Provider.getBalance(l2MessageService.address)}`);
      console.log(`T2 current L1 balance: ${await l1Provider.getBalance(zkEvmV2.address)}`);
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l1Provider);
      const account2 = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);

      let value = ethers.utils.parseEther("0.001");
      let fee = ethers.utils.parseEther("0.0001");
      let to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
      let calldata = "0x";
      let feeData = getAndIncreaseFeeData(await l1Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];
      const tx = await zkEvmV2
        .connect(account)
        .sendMessage(to, fee, calldata, {
          value: value,
          maxPriorityFeePerGas: maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas
         });

      console.log(`tx.hash: ${tx.hash}`);
      const receipt = await tx.wait();
      console.log(`L1 receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);
      const messageSentFilter = zkEvmV2.filters.MessageSent();

      //Retrieve the message hash
      let messageHash: string | null | undefined = null;
      while (!messageHash) {
        let events = (await zkEvmV2.queryFilter(messageSentFilter, receipt.blockNumber, receipt.blockNumber));
        messageHash = events.filter((event) => {
          return (
            event.args._to == to &&
            event.args._value.eq(value.sub(fee)) &&
            event.args._calldata == calldata);
        }).at(0)?.args._messageHash;
      }
      console.log(`tx messageHash: ${JSON.stringify(messageHash)}`);

      //Extra transactions to trigger anchoring
      for (let i = 0; i < 10; i++) {
        let nonce = await l2Provider.getTransactionCount(account2.address, "pending");
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account2).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas
          });
        const receipt = await tx.wait();
        console.log(`extra-tx receipt.blockNumber: ${receipt.blockNumber}`);
        //await new Promise(res => setTimeout(res, 2000));
      }

      let messageClaimedEvent = l2MessageService.filters.MessageClaimed();
      let messageClaimed = null;
      let fromBlock = (await l2Provider.getBlock("latest")).number - 100;
      while (!messageClaimed) {
        let events = (await l2MessageService.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return event.args._messageHash == messageHash;
        }).at(0)?.args._messageHash;
      }
      expect(messageClaimed).not.toBeNull;
    }, 320000);
  });


  describe.skip("Message Service L2 -> L1", () => {
    it("Send transactions to L2 message service, should be delivered in L1", async () => {
      console.log(`T3 current L2 balance: ${await l2Provider.getBalance(l2MessageService.address)}`);
      console.log(`T3 current L1 balance: ${await l1Provider.getBalance(zkEvmV2.address)}`);
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const sendMessageCalldata1 = encodeFunctionCall(dummyContract.interface, "setPayload", [
        ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000),
      ]);
      const sendMessageCalldata2 = encodeFunctionCall(dummyContract.interface, "setPayload", [
        ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000),
      ]);

      //const valueAndFee = BigNumber.from(0)
      const valueAndFee = ethers.utils.parseEther("0.01")
      let nonce = await l2Provider.getTransactionCount(account.address, "pending");
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];
      const tx = await l2MessageService
        .connect(account)
        .sendMessage(l1DummyContract.address, valueAndFee, sendMessageCalldata1, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas,
          maxFeePerGas
        });

      nonce = await l2Provider.getTransactionCount(account.address, "pending");
      const tx2 = await l2MessageService
        .connect(account)
        .sendMessage(l1DummyContract.address, valueAndFee, sendMessageCalldata2, {
          value: valueAndFee,
          nonce: nonce,
          maxPriorityFeePerGas,
          maxFeePerGas
        });

      //Extra transactions to trigger conflations
      for (let i = 0; i < 5; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address, "pending");
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas
          });
        //await new Promise(res => setTimeout(res, 2000));
      }
      const messageSentFilter = l2MessageService.filters.MessageSent();

      //Retrieve the message hashes
      console.log(`tx.hash: ${tx.hash}`);
      console.log(`tx2.hash: ${tx2.hash}`);
      const receipt = await tx.wait();
      const receipt2 = await tx2.wait();
      console.log(`receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);
      console.log(`receipt2.blockNumber: ${JSON.stringify(receipt2.blockNumber)}`);
      let messageHash: string | null | undefined = null;
      while (!messageHash) {
        let events = (await l2MessageService.queryFilter(messageSentFilter, receipt.blockNumber, receipt.blockNumber));
        messageHash = events.filter((event) => {
          return (
            event.args._to == l1DummyContract.address &&
            event.args._value.eq(BigNumber.from(0)) &&
            event.args._calldata == sendMessageCalldata1)
        }).at(0)?.args._messageHash;
      }

      let messageHash2: string | null | undefined = null;
      while (!messageHash2) {
        let events = (await l2MessageService.queryFilter(messageSentFilter, receipt2.blockNumber, receipt2.blockNumber));
        messageHash2 = events.filter((event) => {
          return (
            event.args._to == l1DummyContract.address &&
            event.args._value.eq(BigNumber.from(0)) &&
            event.args._calldata == sendMessageCalldata2)
        }).at(0)?.args._messageHash;
      }

      console.log(`tx messageHash: ${JSON.stringify(messageHash)}`);
      console.log(`tx2 messageHash2: ${JSON.stringify(messageHash2)}`);

      expect(messageHash).not.toEqual(messageHash2);

      console.log(`currentL2BlockNumber: ${await zkEvmV2.currentL2BlockNumber()}`);

      let messageClaimedEvent = zkEvmV2.filters.MessageClaimed();
      let fromBlock = (await l1Provider.getBlock("latest")).number - 100;
      let messageClaimed: MessageClaimedEvent[] = [];
      while (messageClaimed.length == 0) {
        let events = (await zkEvmV2.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return event.args._messageHash == messageHash;
        });
      }
      console.log(`tx messageClaimed: ${JSON.stringify(messageClaimed)}`);
      expect(messageClaimed).not.toBeNull;

      messageClaimed = [];
      while (messageClaimed.length == 0) {
        let events = (await zkEvmV2.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return event.args._messageHash == messageHash2;
        });
      }
      console.log(`tx2 messageClaimed: ${JSON.stringify(messageClaimed)}`);
      expect(messageClaimed.length).toBeGreaterThan(0);
    }, 720000);

    it("Should send a transaction without calldata to L2 message service, be successfully claimed it on L1", async () => {
      console.log(`T4 current L2 balance: ${await l2Provider.getBalance(l2MessageService.address)}`);
      console.log(`T4 current L1 balance: ${await l1Provider.getBalance(zkEvmV2.address)}`);

      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);


      let value = ethers.utils.parseEther("0.001");
      let fee = ethers.utils.parseEther("0.00001");
      let to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
      let calldata = "0x";

      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];
      const tx = await l2MessageService
        .connect(account)
        .sendMessage(to, fee, calldata, {
          value: value,
          maxPriorityFeePerGas,
          maxFeePerGas,
          gasLimit: 200000,
          nonce: nonce
        });

      //Extra transactions to trigger conflations
      for (let i = 0; i < 5; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address, "pending");
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas,
            maxFeePerGas
          });
      }

      const messageSentFilter = l2MessageService.filters.MessageSent();

      //Retrieve the message hash
      console.log(`tx.hash: ${tx.hash}`);
      const receipt = await tx.wait();
      console.log(`receipt.blockNumber: ${JSON.stringify(receipt.blockNumber)}`);
      let messageHash: string | null | undefined = null;
      while (!messageHash) {
        let events = (await l2MessageService.queryFilter(messageSentFilter, receipt.blockNumber, receipt.blockNumber));
        messageHash = events.filter((event) => {
          return (
            event.args._to == to &&
            event.args._value.eq(value.sub(fee)) &&
            event.args._calldata == calldata)
        }).at(0)?.args._messageHash;
      }
      console.log(`tx messageHash: ${messageHash}`);
      let messageClaimedEvent = zkEvmV2.filters.MessageClaimed();
      let messageClaimed: MessageClaimedEvent[] = [];
      let fromBlock = (await l1Provider.getBlock("latest")).number - 100;
      while (messageClaimed.length == 0) {
        let events = (await zkEvmV2.queryFilter(messageClaimedEvent));
        messageClaimed = events.filter((event) => {
          return event.args._messageHash == messageHash;
        });
      }
      console.log(`End current L2 balance: ${await l2Provider.getBalance(l2MessageService.address)}`);
      console.log(`End current L1 balance: ${await l1Provider.getBalance(zkEvmV2.address)}`);
      expect(messageClaimed.length).toBeGreaterThan(0);
    }, 720000);
  });
});
