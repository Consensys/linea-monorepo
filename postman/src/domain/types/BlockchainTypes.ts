export type TransactionReceipt = {
  transactionHash: string;
  blockNumber: number;
  status: "success" | "reverted";
  gasPrice: bigint;
  gasUsed: bigint;
};

export type TransactionResponse = {
  hash: string;
  gasLimit: bigint;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
  nonce: number;
};

export type Block = {
  number: number;
  timestamp: number;
};

export type BlockExtraData = {
  version: number;
  fixedCost: number;
  variableCost: number;
  ethGasPrice: number;
};
