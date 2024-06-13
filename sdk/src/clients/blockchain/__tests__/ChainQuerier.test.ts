import { JsonRpcProvider, Wallet } from "ethers";
import { describe, afterEach, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { ChainQuerier } from "../ChainQuerier";
import { TEST_ADDRESS_1, TEST_L1_SIGNER_PRIVATE_KEY, TEST_TRANSACTION_HASH } from "../../../utils/testing/constants";
import { generateTransactionReceipt } from "../../../utils/testing/helpers";

describe("ChainQuerier", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let chainQuerier: ChainQuerier;
  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    chainQuerier = new ChainQuerier(providerMock, new Wallet(TEST_L1_SIGNER_PRIVATE_KEY, providerMock));
  });

  afterEach(() => {
    mockClear(providerMock);
  });

  describe("getCurrentNonce", () => {
    it("should throw an error when accountAddress param is undefined and no signer has been passed to the ChainQuerier class", async () => {
      const chainQuerier = new ChainQuerier(providerMock);
      await expect(chainQuerier.getCurrentNonce()).rejects.toThrow("Please provide a signer.");
    });

    it("should return the nonce of the accountAddress passed in parameter", async () => {
      const getTransactionCountSpy = jest.spyOn(providerMock, "getTransactionCount").mockResolvedValueOnce(10);
      expect(await chainQuerier.getCurrentNonce(TEST_ADDRESS_1)).toEqual(10);
      expect(getTransactionCountSpy).toHaveBeenCalledTimes(1);
      expect(getTransactionCountSpy).toHaveBeenCalledWith(TEST_ADDRESS_1);
    });

    it("should return the nonce of the signer address", async () => {
      const getTransactionCountSpy = jest.spyOn(providerMock, "getTransactionCount").mockResolvedValueOnce(10);
      expect(await chainQuerier.getCurrentNonce()).toEqual(10);
      expect(getTransactionCountSpy).toHaveBeenCalledTimes(1);
      expect(getTransactionCountSpy).toHaveBeenCalledWith("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf");
    });
  });

  describe("getCurrentBlockNumber", () => {
    it("should return the current block number", async () => {
      const mockBlockNumber = 10_000;
      jest.spyOn(providerMock, "getBlockNumber").mockResolvedValueOnce(mockBlockNumber);
      expect(await chainQuerier.getCurrentBlockNumber()).toEqual(mockBlockNumber);
    });
  });

  describe("getTransactionReceipt", () => {
    it("should return the transaction receipt", async () => {
      const mockTransactionReceipt = generateTransactionReceipt();
      jest.spyOn(providerMock, "getTransactionReceipt").mockResolvedValueOnce(mockTransactionReceipt);
      expect(await chainQuerier.getTransactionReceipt(TEST_TRANSACTION_HASH)).toEqual(mockTransactionReceipt);
    });
  });
});
