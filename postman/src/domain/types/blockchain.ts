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
  gasLimit: bigint;
  baseFeePerGas: bigint;
  priorityFeePerGas: bigint;
};

export function isLineaGasFees(fees: GasFees | LineaGasFees): fees is LineaGasFees {
  return "gasLimit" in fees;
}

type EventLogBase = {
  blockNumber: number;
  logIndex: number;
  contractAddress: string;
  transactionHash: string;
};

export type MessageSent = {
  messageSender: string;
  destination: string;
  fee: bigint;
  value: bigint;
  messageNonce: bigint;
  calldata: string;
  messageHash: string;
} & EventLogBase;

export type L2MessagingBlockAnchored = {
  l2Block: bigint;
} & EventLogBase;

export type MessageClaimed = {
  messageHash: string;
} & EventLogBase;

export type ServiceVersionMigrated = {
  version: bigint;
} & EventLogBase;
