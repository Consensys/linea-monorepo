import { Wallet, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { TransactionRequest } from "@ethersproject/providers";
import { getAndIncreaseFeeData } from "./utils/helpers";
import { RollupGetZkEVMBlockNumberClient, getEvents, wait } from "./utils/utils";

const layer2TestSuite = (title: string) => {
  describe(title, () => {
    describe("Transaction data size", () => {
      it("Should revert if transaction data size is above the limit", async () => {
        const account = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
        await expect(
          dummyContract.connect(account).setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT)),
        ).rejects.toThrow("err: tx data is too large (in bytes)");
      });

      it("Should succeed if transaction data size is below the limit", async () => {
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);
        const tx = await dummyContract.connect(account).setPayload(ethers.utils.randomBytes(1000), {
          nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });

        const receipt = await tx.wait();
        expect(receipt.status).toEqual(1);
      });
    });

    describe("Block conflation", () => {
      it("Should succeed in conflating multiple blocks and proving on L1", async () => {
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);

        const l2BlockNumbers: number[] = [];
        for (let i = 0; i < 2; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          const tx: TransactionRequest = {
            type: 2,
            nonce,
            to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
            maxPriorityFeePerGas,
            maxFeePerGas,
            value: ethers.utils.parseEther("0.01"),
            gasLimit: "21000",
            chainId,
          };

          const signedTx = await account.signTransaction(tx);

          const receipt = await (await l2Provider.sendTransaction(signedTx)).wait();
          l2BlockNumbers.push(receipt.blockNumber);
        }

        for (let i = 0; i < 2; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          const tx = await dummyContract
            .connect(account)
            .setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
              nonce,
              maxPriorityFeePerGas,
              maxFeePerGas,
            });
          const receipt = await tx.wait();
          l2BlockNumbers.push(receipt.blockNumber);
        }

        // These is just to push the L1 verified block forward to the max number in
        // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
        for (let i = 0; i < 4; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          const tx = await dummyContract.connect(account).setPayload(ethers.utils.randomBytes(10), {
            nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });
          await tx.wait();
        }

        const maxL2BlockNumber = Math.max(...l2BlockNumbers);
        let currentL2BlockNumber = (await lineaRollup.currentL2BlockNumber()).toNumber();
        console.log(`maxL2BlockNumber: ${maxL2BlockNumber}`);
        console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
        console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber}`);

        while (maxL2BlockNumber > currentL2BlockNumber) {
          await wait(2000);
          currentL2BlockNumber = (await lineaRollup.currentL2BlockNumber()).toNumber();
        }

        const events = await getEvents(lineaRollup, lineaRollup.filters.BlocksVerificationDone());
        console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
        console.log(`currentL2BlockNumber: ${currentL2BlockNumber}`);

        expect(currentL2BlockNumber).toBeGreaterThanOrEqual(maxL2BlockNumber);
      }, 300000);

      it("Should succeed in conflating transactions with large calldata with low gas into multiple L1 blocks", async () => {
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
        const l2BlockNumbers: number[] = [];
        const txList = [];
        for (let i = 0; i < 4; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address, "pending"),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          const tx = await dummyContract
            .connect(account)
            .setPayload(ethers.utils.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
              nonce,
              maxPriorityFeePerGas,
              maxFeePerGas,
            });
          txList.push(tx);
        }

        await Promise.all(
          txList.map(async (tx) => {
            const receipt = await tx.wait();
            l2BlockNumbers.push(receipt.blockNumber);
          }),
        );

        // These is just to push the L1 verified block forward to the max number in
        // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
        for (let i = 0; i < 4; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          const tx = await dummyContract.connect(account).setPayload(ethers.utils.randomBytes(10), {
            nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });
          await tx.wait();
        }

        const maxL2BlockNumber = Math.max(...l2BlockNumbers);
        let currentL2BlockNumber = (await lineaRollup.currentL2BlockNumber()).toNumber();
        console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
        console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber}`);

        while (maxL2BlockNumber > currentL2BlockNumber) {
          await wait(2000);
          currentL2BlockNumber = (await lineaRollup.currentL2BlockNumber()).toNumber();
        }

        const events = await getEvents(lineaRollup, lineaRollup.filters.BlocksVerificationDone());
        console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
        console.log(`currentL2BlockNumber: ${currentL2BlockNumber}`);

        expect(currentL2BlockNumber).toBeGreaterThanOrEqual(maxL2BlockNumber);
      }, 600000);
    });

    describe("Different transaction types", () => {
      it("Should successfully send a legacy transaction", async () => {
        const account = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [, , gasPrice] = getAndIncreaseFeeData(feeData);

        const receipt = await (
          await account.sendTransaction({
            type: 0,
            nonce,
            to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
            gasPrice,
            value: ethers.utils.parseEther("0.01"),
            gasLimit: "0x466124",
            chainId,
          })
        ).wait();

        expect(receipt).not.toBeNull();
      });

      it("Should successfully send an EIP1559 transaction", async () => {
        const account = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);
        const receipt = await (
          await account.sendTransaction({
            type: 2,
            nonce,
            to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
            maxPriorityFeePerGas,
            maxFeePerGas,
            value: ethers.utils.parseEther("0.01"),
            gasLimit: "21000",
            chainId,
          })
        ).wait();

        expect(receipt).not.toBeNull();
      });

      it("Should successfully send an access list transaction with empty access list", async () => {
        const account = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);

        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [, , gasPrice] = getAndIncreaseFeeData(feeData);

        const receipt = await (
          await account.sendTransaction({
            type: 1,
            nonce,
            to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
            gasPrice,
            value: ethers.utils.parseEther("0.01"),
            gasLimit: "21000",
            chainId,
          })
        ).wait();

        expect(receipt).not.toBeNull();
      });

      it("Should successfully send an access list transaction with access list", async () => {
        const account = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);

        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [, , gasPrice] = getAndIncreaseFeeData(feeData);
        const accessList = {
          "0x8D97689C9818892B700e27F316cc3E41e17fBeb9": [
            "0x0000000000000000000000000000000000000000000000000000000000000000",
            "0x0000000000000000000000000000000000000000000000000000000000000001",
          ],
        };

        const receipt = await (
          await account.sendTransaction({
            type: 1,
            nonce,
            to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
            gasPrice,
            value: ethers.utils.parseEther("0.01"),
            gasLimit: "200000",
            chainId,
            accessList: ethers.utils.accessListify(accessList),
          })
        ).wait();

        expect(receipt).not.toBeNull();
      });
    });

    describe("Block finalization notifications", () => {
      // TODO: discuss new frontend
      it.skip("Shomei frontend always behind while conflating multiple blocks and proving on L1", async () => {
        if (SHOMEI_ENDPOINT == null || SHOMEI_FRONTEND_ENDPOINT == null) {
          // Skip this test for dev and uat environments
          return;
        }
        const account = new Wallet(L2_ACCOUNT_0_PRIVATE_KEY, l2Provider);
        const shomeiClient = new RollupGetZkEVMBlockNumberClient(SHOMEI_ENDPOINT);
        const shomeiFrontendClient = new RollupGetZkEVMBlockNumberClient(SHOMEI_FRONTEND_ENDPOINT);

        for (let i = 0; i < 5; i++) {
          const [nonce, feeData] = await Promise.all([
            l2Provider.getTransactionCount(account.address),
            l2Provider.getFeeData(),
          ]);

          const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

          await (
            await account.sendTransaction({
              type: 2,
              nonce,
              to: "0x8D97689C9818892B700e27F316cc3E41e17fBeb9",
              maxPriorityFeePerGas,
              maxFeePerGas,
              value: ethers.utils.parseEther("0.01"),
              gasLimit: "21000",
              chainId,
            })
          ).wait();

          const [shomeiBlock, shomeiFrontendBlock] = await Promise.all([
            shomeiClient.rollupGetZkEVMBlockNumber(),
            shomeiFrontendClient.rollupGetZkEVMBlockNumber(),
          ]);
          console.log(`shomeiBlock = ${shomeiBlock}, shomeiFrontendBlock = ${shomeiFrontendBlock}`);

          expect(shomeiBlock).toBeGreaterThan(shomeiFrontendBlock);
        }
      }, 300000);
    });
  });
};

export default layer2TestSuite;
