import { Wallet, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { TransactionRequest } from "@ethersproject/providers";
import { getAndIncreaseFeeData } from "./utils/helpers";

describe("Layer 2 test suite", () => {
  describe("Transaction data size", () => {
    it("Should revert if transaction data size is above the limit", async () => {
      const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
      console.log("Running Should revert if transaction data size is above the limit");
      await expect(
        dummyContract.connect(account).setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT))
      ).rejects.toThrow("err: tx data is too large (in bytes)");
    });

    it("Should succeed if transaction data size is below the limit", async () => {
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];

      console.log("Running Should succeed if transaction data size is below the limit");
      const tx = await dummyContract.connect(account).setPayload(ethers.utils.randomBytes(1000), {
        nonce: nonce,
        maxPriorityFeePerGas: maxPriorityFeePerGas,
        maxFeePerGas: maxFeePerGas,
      });
      const receipt = await tx.wait();
      expect(receipt.status).toEqual(1);
    });

    // it.skip("Should succeed when block is full with data", async () => {
    //   const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);

    //   const tx = await dummyContract.connect(account).setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT - 20000));
    //   const receipt = await tx.wait();
    //   expect(receipt.status).toEqual(1);
    // });
  });

  describe("Block conflation", () => {
    it("Should succeed in conflating multiple blocks and proving on L1", async () => {
      console.log("Running Should succeed in conflating multiple blocks and proving on L1");
      const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);

      const l2BlockNumbers: number[] = [];
      for (let i = 0; i < 3; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address);
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx: TransactionRequest = {
          type: 2,
          nonce: nonce,
          to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
          maxPriorityFeePerGas: maxPriorityFeePerGas,
          maxFeePerGas: maxFeePerGas,
          value: ethers.utils.parseEther("0.01"),
          gasLimit: "21000",
          chainId
        };

        const signedTx = await account.signTransaction(tx);

        const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();
        console.log(receipt);
        l2BlockNumbers.push(receipt.blockNumber);
      }

      for (let i = 0; i < 3; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address);
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          }
        );
        const receipt = await tx.wait();
        l2BlockNumbers.push(receipt.blockNumber);
        console.log(receipt);
      }

      // These is just to push the L1 verified block forward to the max number in
      // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
      for (let i = 0; i < 6; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address);
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes(10), {
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          }
        );
        const receipt = await tx.wait();
        console.log(receipt);
      }

      const maxL2BlockNumber = Math.max(...l2BlockNumbers);
      let currentL2BlockNumber = (await zkEvmV2.currentL2BlockNumber()).toNumber();
      console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
      console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber}`);
      while (maxL2BlockNumber > currentL2BlockNumber) {
        await new Promise(res => setTimeout(res, 2000));
        currentL2BlockNumber = (await zkEvmV2.currentL2BlockNumber()).toNumber();
      }
      let events = await zkEvmV2.queryFilter(zkEvmV2.filters.BlocksVerificationDone());
      console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
      console.log(`currentL2BlockNumber: ${currentL2BlockNumber}`);
      expect(currentL2BlockNumber).toBeGreaterThanOrEqual(maxL2BlockNumber);
    }, 300000);

    it("Should succeed in conflating transactions with large calldata with low gas into multiple L1 blocks", async () => {
      console.log("Running Should succeed in conflating transactions with large calldata with low gas into multiple L1 blocks");const account = new Wallet(ACCOUNT_0_PRIVATE_KEY, l2Provider);
      const l2BlockNumbers: number[] = [];
      const txList = [];
      for (let i = 0; i < 6; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address, "pending");
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes((TRANSACTION_CALLDATA_LIMIT / 2) - 1000), {
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          }
        );
        txList.push(tx);
      }

      await Promise.all(txList.map(async tx => {
        const receipt = await tx.wait();
        l2BlockNumbers.push(receipt.blockNumber);
        console.log(receipt);
      }))

      // These is just to push the L1 verified block forward to the max number in
      // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
      for (let i = 0; i < 6; i++) {
        let nonce = await l2Provider.getTransactionCount(account.address);
        let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
        let maxPriorityFeePerGas = feeData[0];
        let maxFeePerGas = feeData[1];
        const tx = await dummyContract.connect(account).setPayload(
          ethers.utils.randomBytes(10), {
            nonce: nonce,
            maxPriorityFeePerGas: maxPriorityFeePerGas,
            maxFeePerGas: maxFeePerGas,
          }
        );
        const receipt = await tx.wait();
        console.log(receipt);
      }

      const maxL2BlockNumber = Math.max(...l2BlockNumbers);
      let currentL2BlockNumber = (await zkEvmV2.currentL2BlockNumber()).toNumber();
      console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
      console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber}`);
      while (maxL2BlockNumber > currentL2BlockNumber) {
        await new Promise(res => setTimeout(res, 2000));
        currentL2BlockNumber = (await zkEvmV2.currentL2BlockNumber()).toNumber();
      }
      let events = await zkEvmV2.queryFilter(zkEvmV2.filters.BlocksVerificationDone());
      console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
      console.log(`currentL2BlockNumber: ${currentL2BlockNumber}`);
      expect(currentL2BlockNumber).toBeGreaterThanOrEqual(maxL2BlockNumber);
    }, 300000);
  });

  describe("Different transaction types", () => {
    it("Should successfully send a legacy transaction", async () => {
      const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let gasPrice = feeData[2];
      const tx: TransactionRequest = {
        type: 0,
        nonce: nonce,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
        gasPrice: gasPrice,
        value: ethers.utils.parseEther("0.01"),
        gasLimit: "0x466124",
        chainId
      };

      const signedTx = await account.signTransaction(tx);
      const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();

      expect(receipt).not.toBeNull;
    });

    it("Should successfully send an EIP1559 transaction", async () => {
      const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let maxPriorityFeePerGas = feeData[0];
      let maxFeePerGas = feeData[1];
      const tx: TransactionRequest = {
        type: 2,
        nonce: nonce,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
        maxPriorityFeePerGas: maxPriorityFeePerGas,
        maxFeePerGas: maxFeePerGas,
        value: ethers.utils.parseEther("0.01"),
        gasLimit: "21000",
        chainId
      };

      const signedTx = await account.signTransaction(tx);
      const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();

      expect(receipt).not.toBeNull;
    });

    it("Should successfully send an access list transaction with empty access list", async () => {
      const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let gasPrice = feeData[2];
      const tx: TransactionRequest = {
        type: 1,
        nonce: nonce,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
        gasPrice: gasPrice,
        value: ethers.utils.parseEther("0.01"),
        gasLimit: "21000",
        chainId
      };

      const signedTx = await account.signTransaction(tx);
      const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();

      expect(receipt).not.toBeNull;
    });

    it("Should successfully send an access list transaction with access list", async () => {
      const account = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
      let nonce = await l2Provider.getTransactionCount(account.address);
      let feeData = getAndIncreaseFeeData(await l2Provider.getFeeData());
      let gasPrice = feeData[2];
      let accessList = {
        "0x8D97689C9818892B700e27F316cc3E41e17fBeb9":
        ["0x0000000000000000000000000000000000000000000000000000000000000000",
         "0x0000000000000000000000000000000000000000000000000000000000000001"]}
      const tx: TransactionRequest = {
        type: 1,
        nonce: nonce,
        to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
        gasPrice: gasPrice,
        value: ethers.utils.parseEther("0.01"),
        gasLimit: "200000",
        chainId,
        accessList: ethers.utils.accessListify(accessList)
      };

      const signedTx = await account.signTransaction(tx);
      const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();

      expect(receipt).not.toBeNull;
    });
  });
});
