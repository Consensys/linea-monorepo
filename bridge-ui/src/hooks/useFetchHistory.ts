import { useCallback } from "react";
import { useAccount } from "wagmi";
import log from "loglevel";
import { useChainStore } from "@/stores/chainStore";
import { useHistoryStore } from "@/stores/historyStore";
import { getBlockNumber } from "@wagmi/core";
import { generateKey } from "@/contexts/storage";
import { wagmiConfig } from "@/config";
import useFetchBridgeTransactions from "./useFetchBridgeTransactions";

const DEFAULT_FIRST_BLOCK = BigInt(1000);

const useFetchHistory = () => {
  // Wagmi
  const { address } = useAccount();

  // Context
  const {
    l1Chain,
    l2Chain,
    networkType: currentNetworkType,
  } = useChainStore((state) => ({
    l1Chain: state.l1Chain,
    l2Chain: state.l2Chain,
    networkType: state.networkType,
  }));

  const { isLoading, setIsLoading, setTransactions, clearStorage, getTransactionsByKey } = useHistoryStore((state) => ({
    isLoading: state.isLoading,
    setIsLoading: state.setIsLoading,
    setTransactions: state.setTransactions,
    clearStorage: state.clearStorage,
    getTransactionsByKey: state.getTransactionsByKey,
  }));

  // Hooks
  const { fetchTransactions } = useFetchBridgeTransactions();

  const fetchHistory = useCallback(async () => {
    if (!l1Chain || !l2Chain || !address) {
      return;
    }

    try {
      setIsLoading(true);

      // ToBlock: get last onchain block
      const l1ToBlockNumber = await getBlockNumber(wagmiConfig, {
        chainId: l1Chain.id,
      });
      const l2ToBlockNumber = await getBlockNumber(wagmiConfig, {
        chainId: l2Chain.id,
      });

      const transactions = getTransactionsByKey(generateKey("transactions", address, currentNetworkType));

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

      setTransactions(generateKey("transactions", address, currentNetworkType), txs ?? []);
    } catch (error) {
      log.error(error);
    } finally {
      setIsLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [address, currentNetworkType, l1Chain, l2Chain]);

  const clearHistory = useCallback(() => {
    // Clear local storage
    if (address) {
      clearStorage(generateKey("transactions", address, currentNetworkType));
    }

    // Trigger fetchHistory() to reload the transaction history faster
    setTimeout(() => fetchHistory(), 1000);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [address, currentNetworkType, fetchHistory]);

  return {
    fetchHistory,
    clearHistory,
    isLoading,
    transactions: address ? getTransactionsByKey(generateKey("transactions", address, currentNetworkType)) : [],
  };
};

export default useFetchHistory;
