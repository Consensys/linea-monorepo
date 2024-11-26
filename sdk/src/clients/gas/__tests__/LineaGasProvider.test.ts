import { describe, afterEach, jest, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { LineaGasProvider } from "../LineaGasProvider";
import { ChainQuerier } from "../../providers/provider";
import { generateTransactionRequest } from "../../../utils/testing/helpers";
import { toBeHex } from "ethers";

const MAX_FEE_PER_GAS = 100_000_000n;

describe("LineaGasProvider", () => {
  let chainQuerierMock: MockProxy<ChainQuerier>;
  let lineaGasProvider: LineaGasProvider;

  beforeEach(() => {
    chainQuerierMock = mock<ChainQuerier>();
    lineaGasProvider = new LineaGasProvider(chainQuerierMock, {
      maxFeePerGas: MAX_FEE_PER_GAS,
      enforceMaxGasFee: false,
    });
  });

  afterEach(() => {
    mockClear(chainQuerierMock);
  });

  describe("getGasFees", () => {
    it("should return maxFeePerGas, maxPriorityFeePerGas from config when enforceMaxGasFee option is enabled", async () => {
      lineaGasProvider = new LineaGasProvider(chainQuerierMock, {
        maxFeePerGas: MAX_FEE_PER_GAS,
        enforceMaxGasFee: true,
      });
      jest.spyOn(chainQuerierMock, "getCurrentBlockNumber").mockResolvedValueOnce(1);
      const sendRequestSpy = jest.spyOn(chainQuerierMock, "sendRequest").mockResolvedValueOnce({
        baseFeePerGas: "0x7",
        priorityFeePerGas: toBeHex(MAX_FEE_PER_GAS),
        gasLimit: toBeHex(50_000n),
      });

      const transactionRequest = generateTransactionRequest();

      const fees = await lineaGasProvider.getGasFees(transactionRequest);

      expect(fees).toStrictEqual({
        maxFeePerGas: MAX_FEE_PER_GAS,
        maxPriorityFeePerGas: MAX_FEE_PER_GAS,
        gasLimit: 50_000n,
      });

      expect(sendRequestSpy).toHaveBeenCalledTimes(1);
    });
    it("should return maxFeePerGas, maxPriorityFeePerGas and gasLimit", async () => {
      jest.spyOn(chainQuerierMock, "getCurrentBlockNumber").mockResolvedValueOnce(1);
      const sendRequestSpy = jest.spyOn(chainQuerierMock, "sendRequest").mockResolvedValueOnce({
        baseFeePerGas: "0x7",
        priorityFeePerGas: toBeHex(MAX_FEE_PER_GAS),
        gasLimit: toBeHex(50_000n),
      });

      const transactionRequest = generateTransactionRequest();

      const fees = await lineaGasProvider.getGasFees(transactionRequest);

      const expectedBaseFee = (BigInt("0x7") * BigInt(1.35 * 100)) / 100n;
      const expectedPriorityFeePerGas = (MAX_FEE_PER_GAS * BigInt(1.05 * 100)) / 100n;
      const expectedMaxFeePerGas = expectedBaseFee + expectedPriorityFeePerGas;

      expect(fees).toStrictEqual({
        maxFeePerGas: expectedMaxFeePerGas,
        maxPriorityFeePerGas: expectedPriorityFeePerGas,
        gasLimit: 50_000n,
      });

      expect(sendRequestSpy).toHaveBeenCalledTimes(1);
    });
  });
});
