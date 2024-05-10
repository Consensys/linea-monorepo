import { Address } from 'wagmi';
import { WaitForTransactionResult } from '@wagmi/core';
import { Chain, Log } from 'viem';

import { TokenInfo } from '@/config/config';

import { MessageWithStatus } from '@/hooks';
import { NetworkType } from '@/contexts/chain.context';

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
  data?: WaitForTransactionResult;
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
