import { ILogger } from "@consensys/linea-shared-utils";
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
    private readonly logger: ILogger,
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

    const capped = this.config.enforceMaxGasFee && rawMaxFeePerGas > this.config.maxFeePerGasCap;
    const maxFeePerGas = capped ? this.config.maxFeePerGasCap : rawMaxFeePerGas;

    if (capped) {
      this.logger.debug("Linea gas fee capped to maxFeePerGasCap.", {
        rawMaxFeePerGas: rawMaxFeePerGas.toString(),
        cappedMaxFeePerGas: maxFeePerGas.toString(),
      });
    }

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
