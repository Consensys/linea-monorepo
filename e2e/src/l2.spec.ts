import { JsonRpcProvider, TransactionRequest, ethers } from "ethers";
import { beforeAll, describe, expect, it } from "@jest/globals";
import { config } from "../config";
import { getAndIncreaseFeeData } from "./common/helpers";
import { RollupGetZkEVMBlockNumberClient, getEvents, wait } from "./common/utils";
import { TRANSACTION_CALLDATA_LIMIT } from "./common/constants";

describe("Layer 2 test suite", () => {
  let l2Provider: JsonRpcProvider;

  beforeAll(() => {
    l2Provider = config.getL2Provider();
  });

  describe("Transaction data size", () => {
    it("Should revert if transaction data size is above the limit", async () => {
      const account = await config.getL2AccountManager().generateAccount();
      const dummyContract = await config.getL2DummyContract(account);

      await expect(
        dummyContract.connect(account).setPayload(ethers.randomBytes(TRANSACTION_CALLDATA_LIMIT)),
      ).rejects.toThrow("missing revert data");
    });

    it("Should succeed if transaction data size is below the limit", async () => {
      const account = await config.getL2AccountManager().generateAccount();
      const dummyContract = await config.getL2DummyContract(account);

      const [nonce, feeData] = await Promise.all([
        l2Provider.getTransactionCount(account.address),
        l2Provider.getFeeData(),
      ]);

      const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);
      const tx = await dummyContract.connect(account).setPayload(ethers.randomBytes(1000), {
        nonce,
        maxPriorityFeePerGas,
        maxFeePerGas,
      });

      const receipt = await tx.wait();
      expect(receipt?.status).toEqual(1);
    });
  });

  describe("Block conflation", () => {
    it("Should succeed in conflating multiple blocks and proving on L1", async () => {
      const account = await config.getL2AccountManager().generateAccount();
      const dummyContract = await config.getL2DummyContract(account);

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
          value: ethers.parseEther("0.01"),
          gasLimit: "21000",
          chainId: config.getL2ChainId(),
        };

        const receipt = await (await account.sendTransaction(tx)).wait();
        if (receipt?.blockNumber) {
          l2BlockNumbers.push(receipt.blockNumber);
        }
      }

      for (let i = 0; i < 2; i++) {
        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

        const tx = await dummyContract
          .connect(account)
          .setPayload(ethers.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
            nonce,
            maxPriorityFeePerGas,
            maxFeePerGas,
          });
        const receipt = await tx.wait();
        l2BlockNumbers.push(receipt?.blockNumber || 0);
      }

      // These is just to push the L1 verified block forward to the max number in
      // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
      for (let i = 0; i < 4; i++) {
        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

        const tx = await dummyContract.connect(account).setPayload(ethers.randomBytes(10), {
          nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });
        await tx.wait();
      }

      const lineaRollup = config.getLineaRollupContract();

      const maxL2BlockNumber = Math.max(...l2BlockNumbers);
      let currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();
      console.log(`maxL2BlockNumber: ${maxL2BlockNumber}`);
      console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
      console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber.toString()}`);

      while (BigInt(maxL2BlockNumber) > currentL2BlockNumber) {
        await wait(2000);
        currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();
      }

      const events = await getEvents(lineaRollup, lineaRollup.filters.BlocksVerificationDone());
      console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
      console.log(`currentL2BlockNumber: ${currentL2BlockNumber.toString()}`);

      expect(currentL2BlockNumber).toBeGreaterThanOrEqual(BigInt(maxL2BlockNumber));
    }, 300000);

    it("Should succeed in conflating transactions with large calldata with low gas into multiple L1 blocks", async () => {
      const account = await config.getL2AccountManager().generateAccount();
      const dummyContract = await config.getL2DummyContract(account);

      let nonce = await l2Provider.getTransactionCount(account.address, "pending");

      const l2BlockNumbers: number[] = [];
      const txList = [];
      for (let i = 0; i < 4; i++) {
        const feeData = await l2Provider.getFeeData();

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

        const tx = await dummyContract.setPayload(ethers.randomBytes(TRANSACTION_CALLDATA_LIMIT / 2 - 1000), {
          nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });
        txList.push(tx);
        nonce += 1;
      }

      await Promise.all(
        txList.map(async (tx) => {
          const receipt = await tx.wait();
          l2BlockNumbers.push(receipt?.blockNumber || 0);
        }),
      );

      // These is just to push the L1 verified block forward to the max number in
      // l2BlockNumbers as it's always 2 blocks behind the current L2 block number
      for (let i = 0; i < 4; i++) {
        const [nonce, feeData] = await Promise.all([
          l2Provider.getTransactionCount(account.address, "pending"),
          l2Provider.getFeeData(),
        ]);

        const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(feeData);

        const tx = await dummyContract.setPayload(ethers.randomBytes(10), {
          nonce,
          maxPriorityFeePerGas,
          maxFeePerGas,
        });
        await tx.wait();
      }

      const lineaRollup = config.getLineaRollupContract();

      const maxL2BlockNumber = Math.max(...l2BlockNumbers);
      let currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();
      console.log(`l2BlockNumbers: ${l2BlockNumbers}`);
      console.log(`initial currentL2BlockNumber: ${currentL2BlockNumber.toString()}`);

      while (BigInt(maxL2BlockNumber) > currentL2BlockNumber) {
        await wait(2000);
        currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();
      }

      const events = await getEvents(lineaRollup, lineaRollup.filters.BlocksVerificationDone());
      console.log(`Last blockVerification: ${JSON.stringify(events.at(-1))}`);
      console.log(`currentL2BlockNumber: ${currentL2BlockNumber}`);

      expect(currentL2BlockNumber).toBeGreaterThanOrEqual(BigInt(maxL2BlockNumber));
    }, 600000);
  });

  describe("Different transaction types", () => {
    it("Should successfully send a legacy transaction", async () => {
      const account = await config.getL2AccountManager().generateAccount();

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
          value: ethers.parseEther("0.01"),
          gasLimit: "0x466124",
          chainId: config.getL2ChainId(),
        })
      ).wait();

      expect(receipt).not.toBeNull();
    });

    it("Should successfully send an EIP1559 transaction", async () => {
      const account = await config.getL2AccountManager().generateAccount();

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
          value: ethers.parseEther("0.01"),
          gasLimit: "21000",
          chainId: config.getL2ChainId(),
        })
      ).wait();

      expect(receipt).not.toBeNull();
    });

    it("Should successfully send an access list transaction with empty access list", async () => {
      const account = await config.getL2AccountManager().generateAccount();

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
          value: ethers.parseEther("0.01"),
          gasLimit: "21000",
          chainId: config.getL2ChainId(),
        })
      ).wait();

      expect(receipt).not.toBeNull();
    });

    it("Should successfully send an access list transaction with access list", async () => {
      const account = await config.getL2AccountManager().generateAccount();

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
          value: ethers.parseEther("0.01"),
          gasLimit: "200000",
          chainId: config.getL2ChainId(),
          accessList: ethers.accessListify(accessList),
        })
      ).wait();

      expect(receipt).not.toBeNull();
    });
  });

  describe("Block finalization notifications", () => {
    // TODO: discuss new frontend
    it.skip("Shomei frontend always behind while conflating multiple blocks and proving on L1", async () => {
      const account = await config.getL2AccountManager().generateAccount();

      const shomeiEndpoint = config.getShomeiEndpoint();
      const shomeiFrontendEndpoint = config.getShomeiFrontendEndpoint();

      if (!shomeiEndpoint || !shomeiFrontendEndpoint) {
        // Skip this test for dev and uat environments
        return;
      }
      const shomeiClient = new RollupGetZkEVMBlockNumberClient(shomeiEndpoint);
      const shomeiFrontendClient = new RollupGetZkEVMBlockNumberClient(shomeiFrontendEndpoint);

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
            value: ethers.parseEther("0.01"),
            gasLimit: "21000",
            chainId: config.getL2ChainId(),
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
