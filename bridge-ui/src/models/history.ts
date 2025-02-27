import { WaitForTransactionReceiptReturnType } from "@wagmi/core";
import { Address, Log } from "viem";
import { NetworkType, TokenInfo } from "@/config";
import { MessageWithStatus } from "@/hooks";
import { Chain } from "@/types";

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
  message?: MessageWithStatus;
  isWaiting?: boolean;
}

export interface BlockRange {
  networkType: NetworkType;
  l1Chain: Chain;
  l2Chain: Chain;
  l1FromBlockNumber: bigint;
  l2FromBlockNumber: bigint;
  transactions: TransactionHistory[];
}
