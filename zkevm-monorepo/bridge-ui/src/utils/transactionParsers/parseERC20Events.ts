import { Chain, PublicClient, getAddress } from 'viem';
import log from 'loglevel';

import { NetworkTokens } from '@/config/config';

import { NetworkType } from '@/contexts/chain.context';
import { ERC20Event } from '@/models';
import fetchTokenInfo from '@/services/fetchTokenInfo';
import { TransactionHistory } from '@/models/history';
import { findTokenByAddress } from './helpers';

const parseERC20Events = async (
  events: ERC20Event[],
  client: PublicClient,
  fromChain: Chain,
  toChain: Chain,
  storedTokens: NetworkTokens,
  networkType: NetworkType,
) => {
  const history: TransactionHistory[] = [];

  if (
    networkType !== NetworkType.MAINNET &&
    networkType !== NetworkType.SEPOLIA &&
    networkType !== NetworkType.UNKNOWN
  ) {
    throw new Error('Invalid network type');
  }

  for (const event of events) {
    if (!event.args?.token) {
      log.warn('Token args not found');
      continue;
    }

    const tokenAddress = getAddress(event.args.token);
    let token = findTokenByAddress(tokenAddress, storedTokens, networkType);

    // Token list may change, skip old tokens
    if (!token) {
      token = await fetchTokenInfo(tokenAddress, networkType, fromChain);
      if (!token) {
        log.warn('Token not found');
        continue;
      }
    }

    // Get block timestamp
    const blockInfo = await client.getBlock({
      blockNumber: event.blockNumber,
    });

    const logHistory: TransactionHistory = {
      transactionHash: event.transactionHash,
      fromChain,
      toChain,
      tokenAddress,
      token,
      amount: event.args.amount,
      recipient: event.args.recipient,
      pending: true,
      event,
      timestamp: blockInfo.timestamp,
    };
    history.push(logHistory);
  }

  return history;
};

export default parseERC20Events;
