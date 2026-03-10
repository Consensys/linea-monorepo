import type { Hash } from "./hex";

export type TransactionSubmission = {
  hash: Hash;
  nonce: number;
  gasLimit: bigint;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
};
