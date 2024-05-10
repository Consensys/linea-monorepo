import React, { createContext, useCallback, useContext, useEffect, useReducer, useState } from 'react';
import { useAccount } from 'wagmi';
import { fetchBlockNumber } from '@wagmi/core';
import log from 'loglevel';

import { generateKey, loadFromLocalStorage, saveToLocalStorage } from './storage';
import { ChainContext } from '@/contexts/chain.context';
import { useFetchBridgeTransactions } from '@/hooks';
import { TransactionHistory } from '@/models/history';
import transactionReducer, { TransactionState, getTransactionsByNetwork } from '@/reducers/transaction.reducer';

const DEFAULT_FIRST_BLOCK = BigInt(1000);

interface HistoryContextData {
  transactions: TransactionHistory[];
  fetchHistory(): void;
  clearHistory(): void;
  isLoading: boolean;
}

type Props = {
  children: JSX.Element;
};

export const HistoryContext = createContext<HistoryContextData>({} as HistoryContextData);

export const HistoryProvider = ({ children }: Props) => {
  // Prevent double fetching
  const [isFetching, setIsFetching] = useState<boolean>(false);

  // Wagmi
  const { address } = useAccount();

  // Context
  const context = useContext(ChainContext);
  const { l1Chain, l2Chain, networkType: currentNetworkType } = context;

  // Initialize transactions history
  const initialState: TransactionState = {
    MAINNET: [],
    SEPOLIA: [],
  };

  const [state, dispatch] = useReducer(transactionReducer, initialState);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Hooks
  const { fetchTransactions } = useFetchBridgeTransactions();

  const fetchHistory = useCallback(async () => {
    if (!l1Chain || !l2Chain || !address) {
      return;
    }
    // Prevent multiple call
    if (isFetching) return;

    try {
      setIsFetching(true);
      setIsLoading(true);

      const currentNetwork = currentNetworkType;

      // ToBlock: get last onchain block
      const l1ToBlockNumber = await fetchBlockNumber({
        chainId: l1Chain.id,
      });
      const l2ToBlockNumber = await fetchBlockNumber({
        chainId: l2Chain.id,
      });

      const transactions = loadFromLocalStorage(generateKey('transactions', address, currentNetworkType), []);

      const txs = await fetchTransactions({
        networkType: currentNetworkType,
        l1Chain,
        l2Chain,
        l1FromBlockNumber: DEFAULT_FIRST_BLOCK,
        l1ToBlockNumber,
        l2FromBlockNumber: DEFAULT_FIRST_BLOCK,
        l2ToBlockNumber,
        transactions,
      });

      currentNetwork === currentNetworkType &&
        dispatch({
          type: 'SET_TRANSACTIONS',
          networkType: currentNetwork,
          transactions: txs ?? [],
        });

      saveToLocalStorage(generateKey('transactions', address, currentNetwork), txs ?? []);
    } catch (error) {
      log.error(error);
    } finally {
      setIsFetching(false);
      setIsLoading(false);
    }
  }, [address, currentNetworkType, isFetching, l1Chain, l2Chain]);

  const clearHistory = useCallback(() => {
    // Clear local storage
    if (address) {
      localStorage.removeItem(generateKey('transactions', address, currentNetworkType));
    }

    // Clear local state
    dispatch({
      type: 'CLEAR_TRANSACTIONS',
      networkType: currentNetworkType,
    });

    // Trigger fetchHistory() to reload the transaction history faster
    setTimeout(() => fetchHistory(), 1000);
  }, [address, currentNetworkType, fetchHistory]);

  useEffect(() => {
    if (address && currentNetworkType) {
      // Load stored transactions
      const transactionsStored = loadFromLocalStorage(generateKey('transactions', address, currentNetworkType), []);
      // Reset history if bad format
      if (Object.prototype.toString.call(transactionsStored) !== '[object Array]') {
        log.error('Value not an array, clearing history');
        clearHistory();
        return;
      }

      dispatch({
        type: 'SET_TRANSACTIONS',
        networkType: currentNetworkType,
        transactions: transactionsStored,
      });
    }
  }, [address, currentNetworkType, clearHistory]);

  return (
    <HistoryContext.Provider
      value={{
        transactions: getTransactionsByNetwork(state, currentNetworkType),
        fetchHistory,
        clearHistory,
        isLoading,
      }}
    >
      {children}
    </HistoryContext.Provider>
  );
};
