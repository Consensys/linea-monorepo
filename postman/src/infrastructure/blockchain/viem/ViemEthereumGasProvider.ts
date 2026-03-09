import { type PublicClient } from "viem";

import { DefaultGasProviderConfig, GasFees, IEthereumGasProvider } from "../../../core/clients/blockchain/IGasProvider";

export class ViemEthereumGasProvider implements IEthereumGasProvider {
  constructor(
    private readonly client: PublicClient,
    private readonly config: DefaultGasProviderConfig,
  ) {}

  public async getGasFees(): Promise<GasFees> {
    const fees = await this.client.estimateFeesPerGas();
    const rawMaxFeePerGas = fees.maxFeePerGas;
    const maxFeePerGas =
      this.config.enforceMaxGasFee && rawMaxFeePerGas > this.config.maxFeePerGasCap
        ? this.config.maxFeePerGasCap
        : rawMaxFeePerGas;

    return {
      maxFeePerGas,
      maxPriorityFeePerGas: fees.maxPriorityFeePerGas,
    };
  }

  public getMaxFeePerGas(): bigint {
    return this.config.maxFeePerGasCap;
  }
}
