import { type PublicClient } from "viem";
import { estimateGas as lineaEstimateGas } from "viem/linea";

import {
  ILineaGasProvider,
  LineaGasFees,
  LineaGasProviderConfig,
} from "../../../../core/clients/blockchain/IGasProvider";
import { BaseError } from "../../../../core/errors";
import { TransactionRequest } from "../../../../core/types";

export class ViemLineaGasProvider implements ILineaGasProvider {
  constructor(
    private readonly client: PublicClient,
    private readonly config: LineaGasProviderConfig,
  ) {}

  public async getGasFees(transactionRequest: TransactionRequest): Promise<LineaGasFees> {
    if (!transactionRequest.from) {
      throw new BaseError("ViemLineaGasProvider: Transaction request must specify the 'from' address.");
    }

    const { gasLimit, baseFeePerGas, priorityFeePerGas } = await lineaEstimateGas(this.client, {
      account: transactionRequest.from,
      to: transactionRequest.to,
      data: transactionRequest.data,
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
