import type { GasFees, LineaGasFees } from "../types/blockchain";

export interface IGasProvider {
  getGasFees(): Promise<GasFees>;
}

export interface ILineaGasProvider {
  getGasFees(transactionCalldata: string): Promise<LineaGasFees>;
}

export type BaseGasProviderConfig = {
  maxFeePerGasCap: bigint;
  enforceMaxGasFee: boolean;
};

export type DefaultGasProviderConfig = BaseGasProviderConfig & {
  gasEstimationPercentile: number;
};

export type LineaGasProviderConfig = BaseGasProviderConfig;

export type GasProviderConfig = DefaultGasProviderConfig & {
  direction: string;
  enableLineaEstimateGas: boolean;
};
