import { type PublicClient, type Hex, numberToHex } from "viem";

import { BaseError } from "../../../domain/errors/BaseError";

import type { IEthereumGasProvider, DefaultGasProviderConfig } from "../../../domain/ports/IGasProvider";
import type { GasFees } from "../../../domain/types";

export class ViemEthereumGasProvider implements IEthereumGasProvider {
  private gasFeesCache: GasFees;
  private cacheIsValidForBlockNumber: bigint;

  constructor(
    private readonly publicClient: PublicClient,
    private readonly config: DefaultGasProviderConfig,
  ) {
    this.cacheIsValidForBlockNumber = 0n;
    this.gasFeesCache = {
      maxFeePerGas: config.maxFeePerGasCap,
      maxPriorityFeePerGas: config.maxFeePerGasCap,
    };
  }

  public async getGasFees(): Promise<GasFees> {
    if (this.config.enforceMaxGasFee) {
      return {
        maxPriorityFeePerGas: this.config.maxFeePerGasCap,
        maxFeePerGas: this.config.maxFeePerGasCap,
      };
    }

    const currentBlockNumber = await this.publicClient.getBlockNumber();
    if (this.cacheIsValidForBlockNumber >= currentBlockNumber) {
      return this.gasFeesCache;
    }

    const feeHistory = await this.fetchFeeHistory();
    const maxPriorityFeePerGas = this.calculateMaxPriorityFee(feeHistory.reward);

    if (maxPriorityFeePerGas > this.config.maxFeePerGasCap) {
      throw new BaseError(
        `Estimated miner tip of ${maxPriorityFeePerGas} exceeds configured max fee per gas of ${this.config.maxFeePerGasCap}!`,
      );
    }

    this.updateCache(currentBlockNumber, feeHistory.baseFeePerGas, maxPriorityFeePerGas);
    return this.gasFeesCache;
  }

  public getMaxFeePerGas(): bigint {
    return this.config.maxFeePerGasCap;
  }

  private async fetchFeeHistory(): Promise<{ reward: Hex[][]; baseFeePerGas: Hex[] }> {
    return this.publicClient.request({
      method: "eth_feeHistory",
      params: [numberToHex(4), "latest", [this.config.gasEstimationPercentile]],
    }) as Promise<{ reward: Hex[][]; baseFeePerGas: Hex[] }>;
  }

  private calculateMaxPriorityFee(reward: Hex[][]): bigint {
    return (
      reward.reduce((acc: bigint, currentValue: Hex[]) => acc + BigInt(currentValue[0]), 0n) / BigInt(reward.length)
    );
  }

  private updateCache(currentBlockNumber: bigint, baseFeePerGas: Hex[], maxPriorityFeePerGas: bigint): void {
    this.cacheIsValidForBlockNumber = currentBlockNumber;
    const maxFeePerGas = BigInt(baseFeePerGas[baseFeePerGas.length - 1]) * 2n + maxPriorityFeePerGas;

    if (maxFeePerGas > 0n && maxPriorityFeePerGas > 0n) {
      this.gasFeesCache = {
        maxPriorityFeePerGas,
        maxFeePerGas: maxFeePerGas > this.config.maxFeePerGasCap ? this.config.maxFeePerGasCap : maxFeePerGas,
      };
    } else {
      this.gasFeesCache = {
        maxPriorityFeePerGas: this.config.maxFeePerGasCap,
        maxFeePerGas: this.config.maxFeePerGasCap,
      };
    }
  }
}
