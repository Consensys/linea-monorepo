import { GasFees, LineaGasFees } from "../types";

export interface IGasProvider {
  getGasFees(transactionCalldata?: string): Promise<GasFees | LineaGasFees>;
  getMaxFeePerGas(): bigint;
}

export interface IEthereumGasProvider extends IGasProvider {
  getGasFees(): Promise<GasFees>;
}

export interface ILineaGasProvider extends IGasProvider {
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
