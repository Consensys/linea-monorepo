import { useCallback } from "react";
import { useAccount } from "wagmi";
import log from "loglevel";
import { useChainStore } from "@/stores/chainStore";
import { useHistoryStore } from "@/stores/historyStore";
import { generateKey } from "@/contexts/storage";
import useFetchBridgeTransactions from "./useFetchBridgeTransactions";

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

  const { isLoading, setIsLoading, setTransactions, getTransactionsByKey, getMinEventBlockNumber } = useHistoryStore(
    (state) => ({
      isLoading: state.isLoading,
      setIsLoading: state.setIsLoading,
      setTransactions: state.setTransactions,
      getTransactionsByKey: state.getTransactionsByKey,
      getMinEventBlockNumber: state.getMinEventBlockNumber,
    }),
  );

  // Hooks
  const { fetchTransactions } = useFetchBridgeTransactions();

  const fetchHistory = useCallback(async () => {
    if (!l1Chain || !l2Chain || !address) {
      return;
    }

    try {
      setIsLoading(true);

      const key = generateKey("transactions", address, currentNetworkType);

      const transactions = getTransactionsByKey(key);
      const minL1EventBlockNumber = getMinEventBlockNumber(key, l1Chain.id);
      const minL2EventBlockNumber = getMinEventBlockNumber(key, l2Chain.id);

      const txs = await fetchTransactions({
        networkType: currentNetworkType,
        l1Chain,
        l2Chain,
        l1FromBlockNumber: minL1EventBlockNumber,
        l2FromBlockNumber: minL2EventBlockNumber,
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

  return {
    fetchHistory,
    isLoading,
    transactions: address ? getTransactionsByKey(generateKey("transactions", address, currentNetworkType)) : [],
  };
};

export default useFetchHistory;
