import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { DefaultGasProviderConfig } from "../../../../core/clients/blockchain/IGasProvider";
import { ViemEthereumGasProvider } from "../ViemEthereumGasProvider";

import type { PublicClient } from "viem";

describe("ViemEthereumGasProvider", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;

  const baseConfig: DefaultGasProviderConfig = {
    maxFeePerGasCap: 100_000_000_000n,
    enforceMaxGasFee: false,
    gasEstimationPercentile: 50,
  };

  beforeEach(() => {
    publicClient = mock<PublicClient>();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("getGasFees", () => {
    it("returns fees from estimateFeesPerGas", async () => {
      publicClient.estimateFeesPerGas.mockResolvedValue({
        maxFeePerGas: 20_000_000_000n,
        maxPriorityFeePerGas: 2_000_000_000n,
      } as never);

      const provider = new ViemEthereumGasProvider(publicClient, baseConfig);
      const fees = await provider.getGasFees();

      expect(fees.maxFeePerGas).toBe(20_000_000_000n);
      expect(fees.maxPriorityFeePerGas).toBe(2_000_000_000n);
    });

    it("caps maxFeePerGas when enforceMaxGasFee is true and fee exceeds cap", async () => {
      publicClient.estimateFeesPerGas.mockResolvedValue({
        maxFeePerGas: 200_000_000_000n,
        maxPriorityFeePerGas: 5_000_000_000n,
      } as never);

      const capConfig: DefaultGasProviderConfig = { ...baseConfig, enforceMaxGasFee: true };
      const provider = new ViemEthereumGasProvider(publicClient, capConfig);
      const fees = await provider.getGasFees();

      expect(fees.maxFeePerGas).toBe(100_000_000_000n);
      expect(fees.maxPriorityFeePerGas).toBe(5_000_000_000n);
    });

    it("does not cap when enforceMaxGasFee is true but fee is under cap", async () => {
      publicClient.estimateFeesPerGas.mockResolvedValue({
        maxFeePerGas: 50_000_000_000n,
        maxPriorityFeePerGas: 1_000_000_000n,
      } as never);

      const capConfig: DefaultGasProviderConfig = { ...baseConfig, enforceMaxGasFee: true };
      const provider = new ViemEthereumGasProvider(publicClient, capConfig);
      const fees = await provider.getGasFees();

      expect(fees.maxFeePerGas).toBe(50_000_000_000n);
    });

    it("does not cap when enforceMaxGasFee is false even if fee exceeds cap", async () => {
      publicClient.estimateFeesPerGas.mockResolvedValue({
        maxFeePerGas: 500_000_000_000n,
        maxPriorityFeePerGas: 1_000_000_000n,
      } as never);

      const provider = new ViemEthereumGasProvider(publicClient, baseConfig);
      const fees = await provider.getGasFees();

      expect(fees.maxFeePerGas).toBe(500_000_000_000n);
    });
  });

  describe("getMaxFeePerGas", () => {
    it("returns the cap from config", () => {
      const provider = new ViemEthereumGasProvider(publicClient, baseConfig);
      expect(provider.getMaxFeePerGas()).toBe(100_000_000_000n);
    });
  });
});
