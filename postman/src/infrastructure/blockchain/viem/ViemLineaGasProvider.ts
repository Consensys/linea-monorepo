import { type Hex, type PublicClient } from "viem";
import { estimateGas as lineaEstimateGas } from "viem/linea";

import { ILineaGasProvider, LineaGasFees, LineaGasProviderConfig } from "../../../core/clients/blockchain/IGasProvider";
import { TransactionRequest } from "../../../core/types";

export class ViemLineaGasProvider implements ILineaGasProvider {
  constructor(
    private readonly client: PublicClient,
    private readonly config: LineaGasProviderConfig,
  ) {}

  public async getGasFees(transactionRequest: TransactionRequest): Promise<LineaGasFees> {
    const { gasLimit, baseFeePerGas, priorityFeePerGas } = await lineaEstimateGas(this.client, {
      account: transactionRequest.from as Hex,
      to: transactionRequest.to as Hex,
      data: transactionRequest.data as Hex | undefined,
      value: transactionRequest.value,
    });

    const rawMaxFeePerGas = baseFeePerGas + priorityFeePerGas;

    const maxFeePerGas =
      this.config.enforceMaxGasFee && rawMaxFeePerGas > this.config.maxFeePerGasCap
        ? this.config.maxFeePerGasCap
        : rawMaxFeePerGas;

    return {
      gasLimit,
      maxFeePerGas,
      maxPriorityFeePerGas: priorityFeePerGas,
    };
  }

  public getMaxFeePerGas(): bigint {
    return this.config.maxFeePerGasCap;
  }
}
