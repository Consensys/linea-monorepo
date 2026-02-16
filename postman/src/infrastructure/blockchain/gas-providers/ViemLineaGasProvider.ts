import { type PublicClient, type Hex, type Address } from "viem";

import type { ILineaGasProvider, LineaGasProviderConfig } from "../../../domain/ports/IGasProvider";
import type { LineaGasFees, LineaEstimateGasResponse } from "../../../domain/types";

const BASE_FEE_MULTIPLIER = 1.35;
const PRIORITY_FEE_MULTIPLIER = 1.05;

export class ViemLineaGasProvider implements ILineaGasProvider {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly signerAddress: Address,
    private readonly contractAddress: Address,
    private readonly config: LineaGasProviderConfig,
  ) {}

  public async getGasFees(transactionCalldata: string): Promise<LineaGasFees> {
    const gasFees = await this.getLineaGasFees(transactionCalldata);

    if (this.config.enforceMaxGasFee) {
      return {
        ...gasFees,
        maxPriorityFeePerGas: this.config.maxFeePerGasCap,
        maxFeePerGas: this.config.maxFeePerGasCap,
      };
    }

    return gasFees;
  }

  public getMaxFeePerGas(): bigint {
    return this.config.maxFeePerGasCap;
  }

  private async getLineaGasFees(transactionCalldata: string): Promise<LineaGasFees> {
    const lineaGasFees = await this.fetchLineaResponse(transactionCalldata);

    const baseFee = this.getValueFromMultiplier(BigInt(lineaGasFees.baseFeePerGas), BASE_FEE_MULTIPLIER);
    const maxPriorityFeePerGas = this.getValueFromMultiplier(
      BigInt(lineaGasFees.priorityFeePerGas),
      PRIORITY_FEE_MULTIPLIER,
    );
    const maxFeePerGas = baseFee + maxPriorityFeePerGas;
    const gasLimit = BigInt(lineaGasFees.gasLimit);

    return { maxPriorityFeePerGas, maxFeePerGas, gasLimit };
  }

  private async fetchLineaResponse(transactionCalldata: string): Promise<LineaEstimateGasResponse> {
    return this.publicClient.request({
      method: "linea_estimateGas" as "eth_estimateGas",
      params: [
        {
          from: this.signerAddress,
          to: this.contractAddress,
          value: "0x0" as Hex,
          data: transactionCalldata as Hex,
        },
      ],
    }) as unknown as Promise<LineaEstimateGasResponse>;
  }

  private getValueFromMultiplier(value: bigint, multiplier: number): bigint {
    return (value * BigInt(multiplier * 100)) / 100n;
  }
}
