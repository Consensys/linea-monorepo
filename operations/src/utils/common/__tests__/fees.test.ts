import { describe, afterEach, jest, it, expect, beforeEach } from "@jest/globals";
import { JsonRpcProvider } from "ethers";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { get1559Fees } from "../fees.js";

describe("Utils", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  const MAX_FEE_PER_GAS_FROM_CONFIG = 100_000_000n;
  const GAS_ESTIMATION_PERCENTILE = 15;

  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
  });

  afterEach(() => {
    mockClear(providerMock);
  });

  describe("get1559Fees", () => {
    it("should throw an error if maxPriorityFee is greater than maxFeePerGas", async () => {
      jest.spyOn(providerMock, "getBlockNumber").mockResolvedValueOnce(1);
      const sendSpy = jest.spyOn(providerMock, "send").mockResolvedValueOnce({
        baseFeePerGas: ["0x3da8e7618", "0x3e1ba3b1b", "0x3dfd72b90", "0x3d64eee76", "0x3d4da2da0", "0x3ccbcac6b"],
        gasUsedRatio: [0.5290747666666666, 0.49240453333333334, 0.4615576, 0.49407083333333335, 0.4669053],
        oldestBlock: "0xfab8ac",
        reward: [
          ["0x59682f00", "0x59682f00"],
          ["0x59682f00", "0x59682f00"],
          ["0x3b9aca00", "0x59682f00"],
          ["0x510b0870", "0x59682f00"],
          ["0x3b9aca00", "0x59682f00"],
        ],
      });
      await expect(
        get1559Fees(providerMock, MAX_FEE_PER_GAS_FROM_CONFIG, GAS_ESTIMATION_PERCENTILE),
      ).rejects.toThrowError(
        `Estimated miner tip of ${1_271_935_510} exceeds configured max fee per gas of ${MAX_FEE_PER_GAS_FROM_CONFIG.toString()}.`,
      );
      expect(sendSpy).toHaveBeenCalledTimes(1);
    });

    it("should return maxFeePerGas and maxPriorityFeePerGas", async () => {
      jest.spyOn(providerMock, "getBlockNumber").mockResolvedValueOnce(1);
      const sendSpy = jest.spyOn(providerMock, "send").mockResolvedValueOnce({
        baseFeePerGas: ["0x3da8e7618", "0x3e1ba3b1b", "0x3dfd72b90", "0x3d64eee76", "0x3d4da2da0", "0x3ccbcac6b"],
        gasUsedRatio: [0.5290747666666666, 0.49240453333333334, 0.4615576, 0.49407083333333335, 0.4669053],
        oldestBlock: "0xfab8ac",
        reward: [
          ["0xe4e1c0", "0xe4e1c0"],
          ["0xe4e1c0", "0xe4e1c0"],
          ["0xe4e1c0", "0xe4e1c0"],
          ["0xcf7867", "0xe4e1c0"],
          ["0x5f5e100", "0xe4e1c0"],
        ],
      });

      const fees = await get1559Fees(providerMock, MAX_FEE_PER_GAS_FROM_CONFIG, GAS_ESTIMATION_PERCENTILE);

      expect(fees).toStrictEqual({
        maxFeePerGas: MAX_FEE_PER_GAS_FROM_CONFIG,
        maxPriorityFeePerGas: 31_719_355n,
      });

      expect(sendSpy).toHaveBeenCalledTimes(1);
    });

    it("should return maxFeePerGas from config when maxFeePerGas and maxPriorityFeePerGas === 0", async () => {
      jest.spyOn(providerMock, "getBlockNumber").mockResolvedValueOnce(1);

      const sendSpy = jest.spyOn(providerMock, "send").mockResolvedValueOnce({
        baseFeePerGas: ["0x0", "0x0", "0x0", "0x0", "0x0", "0x0"],
        gasUsedRatio: [0, 0, 0, 0, 0],
        oldestBlock: "0xfab8ac",
        reward: [
          ["0x0", "0x0"],
          ["0x0", "0x0"],
          ["0x0", "0x0"],
          ["0x0", "0x0"],
          ["0x0", "0x0"],
        ],
      });

      const fees = await get1559Fees(providerMock, MAX_FEE_PER_GAS_FROM_CONFIG, GAS_ESTIMATION_PERCENTILE);

      expect(fees).toStrictEqual({
        maxFeePerGas: MAX_FEE_PER_GAS_FROM_CONFIG,
      });

      expect(sendSpy).toHaveBeenCalledTimes(1);
    });
  });
});
