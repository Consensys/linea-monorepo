import { describe, afterEach, jest, it, expect, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { testL1NetworkConfig } from "../../../../utils/testing/constants";
import { LineaGasProvider } from "../LineaGasProvider";
import { ChainQuerier } from "../../ChainQuerier";
import { generateTransactionRequest } from "../../../../utils/testing/helpers";
import { toBeHex } from "ethers";

describe("LineaGasProvider", () => {
  let chainQuerierMock: MockProxy<ChainQuerier>;
  let lineaGasProvider: LineaGasProvider;

  beforeEach(() => {
    chainQuerierMock = mock<ChainQuerier>();
    lineaGasProvider = new LineaGasProvider(chainQuerierMock, {
      maxFeePerGas: testL1NetworkConfig.claiming.maxFeePerGas,
      enforceMaxGasFee: false,
    });
  });

  afterEach(() => {
    mockClear(chainQuerierMock);
  });

  describe("getGasFees", () => {
    it("should return maxFeePerGas, maxPriorityFeePerGas from config when enforceMaxGasFee option is enabled", async () => {
      lineaGasProvider = new LineaGasProvider(chainQuerierMock, {
        maxFeePerGas: testL1NetworkConfig.claiming.maxFeePerGas,
        enforceMaxGasFee: true,
      });
      jest.spyOn(chainQuerierMock, "getCurrentBlockNumber").mockResolvedValueOnce(1);
      const sendRequestSpy = jest.spyOn(chainQuerierMock, "sendRequest").mockResolvedValueOnce({
        baseFeePerGas: "0x7",
        priorityFeePerGas: toBeHex(testL1NetworkConfig.claiming.maxFeePerGas),
        gasLimit: toBeHex(50_000n),
      });

      const transactionRequest = generateTransactionRequest();

      const fees = await lineaGasProvider.getGasFees(transactionRequest);

      expect(fees).toStrictEqual({
        maxFeePerGas: testL1NetworkConfig.claiming.maxFeePerGas,
        maxPriorityFeePerGas: testL1NetworkConfig.claiming.maxFeePerGas,
        gasLimit: 50_000n,
      });

      expect(sendRequestSpy).toHaveBeenCalledTimes(1);
    });
    it("should return maxFeePerGas, maxPriorityFeePerGas and gasLimit", async () => {
      jest.spyOn(chainQuerierMock, "getCurrentBlockNumber").mockResolvedValueOnce(1);
      const sendRequestSpy = jest.spyOn(chainQuerierMock, "sendRequest").mockResolvedValueOnce({
        baseFeePerGas: "0x7",
        priorityFeePerGas: toBeHex(testL1NetworkConfig.claiming.maxFeePerGas),
        gasLimit: toBeHex(50_000n),
      });

      const transactionRequest = generateTransactionRequest();

      const fees = await lineaGasProvider.getGasFees(transactionRequest);

      const expectedBaseFee = (BigInt("0x7") * BigInt(1.35 * 100)) / 100n;
      const expectedPriorityFeePerGas = (testL1NetworkConfig.claiming.maxFeePerGas * BigInt(1.05 * 100)) / 100n;
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
