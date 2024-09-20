import { WaitForTransactionReceiptReturnType } from "@wagmi/core";
import { Address, Chain, Log } from "viem";
import { NetworkType, TokenInfo } from "@/config";
import { MessageWithStatus } from "@/hooks";

export interface TransactionHistory {
  transactionHash: Address;
  fromChain: Chain;
  toChain: Chain;
  tokenAddress: Address | null;
  token: TokenInfo;
  amount: bigint;
  recipient: Address;
  pending: boolean;
  success?: boolean;
  error?: boolean;
  event: Log;
  data?: WaitForTransactionReceiptReturnType;
  timestamp: bigint;
  messages?: MessageWithStatus[];
  isWaiting?: boolean;
}

export interface BlockRange {
  networkType: NetworkType;
  l1Chain: Chain;
  l2Chain: Chain;
  l1FromBlockNumber: bigint;
  l1ToBlockNumber: bigint;
  l2FromBlockNumber: bigint;
  l2ToBlockNumber: bigint;
  transactions: TransactionHistory[];
}
