import { type PublicClient, type Hex, type Address } from "viem";
import { estimateGas } from "viem/linea";

import type { ILineaGasProvider, LineaGasProviderConfig } from "../../../domain/ports/IGasProvider";
import type { LineaGasFees, LineaEstimateGasResponse } from "../../../domain/types/blockchain";

const BASE_FEE_MULTIPLIER = 1.35;
const PRIORITY_FEE_MULTIPLIER = 1.05;

export class ViemLineaGasProvider implements ILineaGasProvider {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly signerAddress: Address,
    private readonly contractAddress: Address,
    private readonly config: LineaGasProviderConfig,
  ) {}

  public async getGasFees(transactionCalldata: Hex): Promise<LineaGasFees> {
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

  private async getLineaGasFees(transactionCalldata: Hex): Promise<LineaGasFees> {
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

  private async fetchLineaResponse(transactionCalldata: Hex): Promise<LineaEstimateGasResponse> {
    return estimateGas(this.publicClient, {
      account: this.signerAddress,
      to: this.contractAddress,
      value: 0n,
      data: transactionCalldata,
    });
  }

  private getValueFromMultiplier(value: bigint, multiplier: number): bigint {
    return (value * BigInt(multiplier * 100)) / 100n;
  }
}
