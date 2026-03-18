import { ILogger } from "@consensys/linea-shared-utils";
import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { LineaGasProviderConfig } from "../../../../../core/clients/blockchain/IGasProvider";
import { BaseError } from "../../../../../core/errors/BaseError";
import { TransactionRequest } from "../../../../../core/types";
import { ViemLineaGasProvider } from "../ViemLineaGasProvider";

import type { PublicClient } from "viem";

const TEST_TX_REQUEST: TransactionRequest = {
  from: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  to: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  data: "0x",
  value: 0n,
};

describe("ViemLineaGasProvider", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let logger: ReturnType<typeof mock<ILogger>>;

  const baseConfig: LineaGasProviderConfig = {
    maxFeePerGasCap: 100_000_000_000n,
    enforceMaxGasFee: false,
  };

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    logger = mock<ILogger>();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getGasFees", () => {
    it("maps linea_estimateGas hex response to LineaGasFees", async () => {
      jest.spyOn(publicClient, "request").mockResolvedValue({
        gasLimit: "0x5208", // 21000
        baseFeePerGas: "0x3B9ACA00", // 1_000_000_000
        priorityFeePerGas: "0x3B9ACA00", // 1_000_000_000
      } as never);

      const provider = new ViemLineaGasProvider(publicClient, baseConfig, logger);
      const fees = await provider.getGasFees(TEST_TX_REQUEST);

      expect(fees.gasLimit).toBe(21000n);
      expect(fees.maxPriorityFeePerGas).toBe(1_000_000_000n);
      expect(fees.maxFeePerGas).toBe(2_000_000_000n); // base + priority
    });

    it("calls linea_estimateGas via viem/linea with correct params", async () => {
      jest.spyOn(publicClient, "request").mockResolvedValue({
        gasLimit: "0x5208",
        baseFeePerGas: "0x1",
        priorityFeePerGas: "0x1",
      } as never);

      const provider = new ViemLineaGasProvider(publicClient, baseConfig, logger);
      await provider.getGasFees(TEST_TX_REQUEST);

      expect(publicClient.request).toHaveBeenCalledWith(
        expect.objectContaining({
          method: "linea_estimateGas",
          params: [
            expect.objectContaining({
              from: TEST_TX_REQUEST.from,
              to: TEST_TX_REQUEST.to,
            }),
          ],
        }),
      );
    });

    it("caps maxFeePerGas when enforceMaxGasFee is true and computed fee exceeds cap", async () => {
      jest.spyOn(publicClient, "request").mockResolvedValue({
        gasLimit: "0x5208",
        baseFeePerGas: "0x174876E800", // 100_000_000_000
        priorityFeePerGas: "0x174876E800", // 100_000_000_000
      } as never);

      const capConfig: LineaGasProviderConfig = { ...baseConfig, enforceMaxGasFee: true };
      const provider = new ViemLineaGasProvider(publicClient, capConfig, logger);
      const fees = await provider.getGasFees(TEST_TX_REQUEST);

      expect(fees.maxFeePerGas).toBe(100_000_000_000n);
    });

    it("throws when from address is missing", async () => {
      const provider = new ViemLineaGasProvider(publicClient, baseConfig, logger);
      const noFromRequest: TransactionRequest = { to: TEST_TX_REQUEST.to, data: "0x", value: 0n };
      await expect(provider.getGasFees(noFromRequest)).rejects.toThrow(BaseError);
      await expect(provider.getGasFees(noFromRequest)).rejects.toThrow(
        "ViemLineaGasProvider: Transaction request must specify the 'from' address.",
      );
    });

    it("does not cap when enforceMaxGasFee is false", async () => {
      jest.spyOn(publicClient, "request").mockResolvedValue({
        gasLimit: "0x5208",
        baseFeePerGas: "0x174876E800", // 100_000_000_000
        priorityFeePerGas: "0x174876E800", // 100_000_000_000
      } as never);

      const provider = new ViemLineaGasProvider(publicClient, baseConfig, logger);
      const fees = await provider.getGasFees(TEST_TX_REQUEST);

      expect(fees.maxFeePerGas).toBe(200_000_000_000n);
    });
  });

  describe("getMaxFeePerGas", () => {
    it("returns the cap from config", () => {
      const provider = new ViemLineaGasProvider(publicClient, baseConfig, logger);
      expect(provider.getMaxFeePerGas()).toBe(100_000_000_000n);
    });
  });
});
