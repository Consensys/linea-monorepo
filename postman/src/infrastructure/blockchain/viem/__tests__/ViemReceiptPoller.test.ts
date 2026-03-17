import { describe, it, beforeEach, expect } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { ITransactionProvider } from "../../../../core/clients/blockchain/IProvider";
import { BaseError } from "../../../../core/errors";
import { Hash, TransactionReceipt } from "../../../../core/types";
import { ViemReceiptPoller } from "../ViemReceiptPoller";

const TEST_TX_HASH: Hash = "0x2020202020202020202020202020202020202020202020202020202020202020";

const generateReceipt = (): TransactionReceipt => ({
  hash: TEST_TX_HASH,
  blockNumber: 100,
  status: "success",
  gasUsed: 50_000n,
  gasPrice: 100_000_000_000n,
  logs: [],
});

describe("ViemReceiptPoller", () => {
  const provider = mock<ITransactionProvider>();
  let poller: ViemReceiptPoller;

  beforeEach(() => {
    jest.resetAllMocks();
    poller = new ViemReceiptPoller(provider);
  });

  it("returns receipt immediately when available", async () => {
    const receipt = generateReceipt();
    jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(receipt);

    const result = await poller.poll(TEST_TX_HASH, 5000, 100);

    expect(result).toBe(receipt);
    expect(provider.getTransactionReceipt).toHaveBeenCalledTimes(1);
  });

  it("polls until receipt is found", async () => {
    const receipt = generateReceipt();
    jest
      .spyOn(provider, "getTransactionReceipt")
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce(null)
      .mockResolvedValueOnce(receipt);

    const result = await poller.poll(TEST_TX_HASH, 5000, 0);

    expect(result).toBe(receipt);
    expect(provider.getTransactionReceipt).toHaveBeenCalledTimes(3);
  });

  it("throws BaseError when timeout is reached", async () => {
    jest.spyOn(provider, "getTransactionReceipt").mockResolvedValue(null);

    await expect(poller.poll(TEST_TX_HASH, 0, 0)).rejects.toThrow(BaseError);
  });
});
