import { useCallback } from "react";
import { useAccount } from "wagmi";
import log from "loglevel";
import { useChainStore } from "@/stores/chainStore";
import { useHistoryStore } from "@/stores/historyStore";
import { generateKey } from "@/contexts/storage";
import useFetchBridgeTransactions from "./useFetchBridgeTransactions";
import { wagmiConfig } from "@/config";
import { getBlockNumber } from "@wagmi/core";

const useFetchHistory = () => {
  // Wagmi
  const { address } = useAccount();

  // Context
  const l1Chain = useChainStore((state) => state.l1Chain);
  const l2Chain = useChainStore((state) => state.l2Chain);
  const currentNetworkType = useChainStore((state) => state.networkType);
  const isLoading = useHistoryStore((state) => state.isLoading);
  const { setIsLoading, setTransactions, getTransactionsByKey, getFromBlockNumbers } = useHistoryStore((state) => ({
    setIsLoading: state.setIsLoading,
    setTransactions: state.setTransactions,
    getTransactionsByKey: state.getTransactionsByKey,
    getFromBlockNumbers: state.getFromBlockNumbers,
  }));

  // Hooks
  const { fetchTransactions } = useFetchBridgeTransactions();

  const fetchHistory = useCallback(async () => {
    if (!l1Chain || !l2Chain || !address) {
      return;
    }

    try {
      setIsLoading(true);

      const [l1ToBlockNumber, l2ToBlockNumber] = await Promise.all([
        getBlockNumber(wagmiConfig, {
          chainId: l1Chain.id,
        }),
        getBlockNumber(wagmiConfig, {
          chainId: l2Chain.id,
        }),
      ]);

      const key = generateKey("transactions", address, currentNetworkType);

      const transactions = getTransactionsByKey(key);
      const { l1FromBlock, l2FromBlock } = getFromBlockNumbers(key);

      const txs = await fetchTransactions({
        networkType: currentNetworkType,
        l1Chain,
        l2Chain,
        l1FromBlockNumber: l1FromBlock,
        l2FromBlockNumber: l2FromBlock,
        transactions,
      });

      setTransactions(
        generateKey("transactions", address, currentNetworkType),
        txs ?? [],
        l1ToBlockNumber,
        l2ToBlockNumber,
      );
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
