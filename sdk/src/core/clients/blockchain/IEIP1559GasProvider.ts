export type Fees = {
  maxFeePerGas: bigint;
  maxPriorityFeePerGas?: bigint;
};

export type FeeHistory = {
  oldestBlock: number;
  reward: string[][];
  baseFeePerGas: string[];
  gasUsedRatio: number[];
};

export interface IEIP1559GasProvider {
  get1559Fees(percentile?: number): Promise<Fees>;
}
