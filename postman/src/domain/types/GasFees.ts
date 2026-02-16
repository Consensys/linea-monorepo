export type GasFees = {
  maxFeePerGas: bigint;
  maxPriorityFeePerGas: bigint;
};

export type LineaGasFees = GasFees & {
  gasLimit: bigint;
};

export type FeeHistory = {
  oldestBlock: number;
  reward: string[][];
  baseFeePerGas: string[];
  gasUsedRatio: number[];
};

export type LineaEstimateGasResponse = {
  baseFeePerGas: string;
  priorityFeePerGas: string;
  gasLimit: string;
};

export function isLineaGasFees(fees: GasFees | LineaGasFees): fees is LineaGasFees {
  return "gasLimit" in fees;
}
