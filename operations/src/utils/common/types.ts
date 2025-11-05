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
