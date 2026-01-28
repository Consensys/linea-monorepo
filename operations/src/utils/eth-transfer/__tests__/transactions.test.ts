import { describe, jest, it, expect, afterAll } from "@jest/globals";
import axios from "axios";
import { JsonRpcProvider, Signature, TransactionLike, TransactionResponse, ethers } from "ethers";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { estimateTransactionGas, executeTransaction, getWeb3SignerSignature } from "../transactions.js";

jest.mock("axios");

const transaction: TransactionLike = {
  to: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
  value: 1n,
};

describe("Transactions", () => {
  let providerMock: MockProxy<JsonRpcProvider>;

  beforeAll(() => {
    providerMock = mock<JsonRpcProvider>();
  });

  afterAll(() => {
    mockClear(providerMock);
  });

  describe("getWeb3SignerSignature", () => {
    const web3SignerUrl = "http://localhost:9000";
    const web3SignerPublicKey = ethers.hexlify(ethers.randomBytes(64));

    it("should throw an error when the axios request failed", async () => {
      jest.spyOn(axios, "post").mockRejectedValueOnce(new Error("http error"));

      await expect(getWeb3SignerSignature(web3SignerUrl, web3SignerPublicKey, transaction)).rejects.toThrowError(
        `Web3SignerError: ${JSON.stringify("http error")}`,
      );
    });

    it("should return the signature", async () => {
      jest.spyOn(axios, "post").mockResolvedValueOnce({ data: "0xaaaaaa" });
      expect(await getWeb3SignerSignature(web3SignerUrl, web3SignerPublicKey, transaction)).toStrictEqual("0xaaaaaa");
    });
  });

  describe("estimateTransactionGas", () => {
    it("should throw an error when the transaction gas estimation failed", async () => {
      jest.spyOn(providerMock, "estimateGas").mockRejectedValueOnce(new Error("estimated gas error"));

      await expect(estimateTransactionGas(providerMock, transaction)).rejects.toThrow(
        `GasEstimationError: ${JSON.stringify("estimated gas error")}`,
      );
    });

    it("should return estimated transaction gas limit", async () => {
      jest.spyOn(providerMock, "estimateGas").mockResolvedValueOnce(100_000n);
      expect(await estimateTransactionGas(providerMock, transaction)).toStrictEqual(100_000n);
    });
  });

  describe("executeTransaction", () => {
    it("should throw an error when the transaction sent failed", async () => {
      jest.spyOn(providerMock, "broadcastTransaction").mockRejectedValueOnce(new Error("broadcast transaction error"));

      await expect(
        executeTransaction(providerMock, {
          from: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
          to: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
          value: ethers.parseEther("1").toString(),
          gasPrice: "1606627962",
          maxPriorityFeePerGas: "1000000000",
          maxFeePerGas: "2213255924",
          gasLimit: "21001",
          nonce: 3,
          data: "0x",
          chainId: 31337,
          signature: Signature.from({
            r: "0xac6fbe2571f913aa5e88596b50e6a9ab01833e94187b9cf4b0cc86e7fccb6ca8",
            s: "0x405936c6e570b8e877ac391124d04ac7bff11a49c6b50b78bc057442d9f98262",
            v: 1,
          }),
        }),
      ).rejects.toThrowError(`TransactionError: ${JSON.stringify("broadcast transaction error")}`);
    });

    it("should successfully execute the transaction", async () => {
      const expectedTransactionReceipt = {
        blockHash: "0xcd224ee1fc35433bb96dbc81f3c2a0dc67d97f84ef41d9826b02b039bc3da055",
        blockNumber: 4,
        confirmations: 1,
        contractAddress: null,
        cumulativeGasUsed: "21000",
        effectiveGasPrice: "1606627962",
        from: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
        gasUsed: "21000",
        logs: [],
        logsBloom:
          "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        status: 1,
        to: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
        transactionHash: "0xcc0a7710f2bdaf8a1b34d1d62d0e5f65dab5c89e61d5cfab76a3e6f1fdc745dc",
        transactionIndex: 0,
        type: 2,
      };

      jest.spyOn(providerMock, "broadcastTransaction").mockResolvedValueOnce({
        hash: "0x81a954827d5ed7b1693f1bc844fd99895e9c4ac9f47ff12f280cd9b5b7a200e5",
        type: 2,
        accessList: [],
        blockHash: "0x33c47baf1c92aa0472fc2bc071acf7d5c1336f3f298d1074593ba87e9565518a",
        blockNumber: 4,
        index: 0,
        from: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
        gasPrice: 1606627962n,
        maxPriorityFeePerGas: 1000000000n,
        maxFeePerGas: 2213255924n,
        gasLimit: 21001n,
        to: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
        value: 1n,
        nonce: 3,
        data: "0x",
        signature: Signature.from({
          r: "0x65623e79d4875e745c0f390b05bae72786ef8920087c93b6be22a199932970b4",
          s: "0x5969afc485f0d32debe16ac290b812d41e2538dac32f345bb5174f49f35789bf",
          v: 1,
        }),
        chainId: 31337n,
        wait: jest.fn().mockImplementationOnce(() => expectedTransactionReceipt),
      } as unknown as TransactionResponse);

      expect(
        await executeTransaction(providerMock, {
          hash: "0x81a954827d5ed7b1693f1bc844fd99895e9c4ac9f47ff12f280cd9b5b7a200e5",
          type: 2,
          accessList: [],
          from: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
          gasPrice: 1606627962n,
          maxPriorityFeePerGas: 1000000000n,
          maxFeePerGas: 2213255924n,
          gasLimit: 21001n,
          to: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
          value: 1n,
          nonce: 3,
          data: "0x",
          signature: {
            r: "0x65623e79d4875e745c0f390b05bae72786ef8920087c93b6be22a199932970b4",
            s: "0x5969afc485f0d32debe16ac290b812d41e2538dac32f345bb5174f49f35789bf",
            v: 1,
          },
          chainId: 31337,
        }),
      ).toStrictEqual(expectedTransactionReceipt);
    });
  });
});
