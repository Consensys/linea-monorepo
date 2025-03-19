import { useAccount } from "wagmi";
import { useQuery } from "@tanstack/react-query";
import { useChainStore } from "@/stores";
import { fetchTransactionsHistory } from "@/utils";
import useLineaSDK from "./useLineaSDK";
import useTokens from "./useTokens";

const useTransactionHistory = () => {
  const { address } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const { lineaSDK, lineaSDKContracts } = useLineaSDK();
  const tokens = useTokens();

  // TODO - For invalidateQueries with queryKey: "transactionHistory", we don't need to refetch the entire transaction history, just the single affected transaction
  const { data, isLoading, refetch } = useQuery({
    enabled: !!address && !!fromChain && !!toChain,
    queryKey: ["transactionHistory", address, fromChain.id, toChain.id],
    queryFn: () => fetchTransactionsHistory({ lineaSDK, fromChain, toChain, address: address!, tokens }),
    staleTime: 1000 * 60 * 2,
  });

  return {
    transactions: data || [],
    isLoading,
    refetch,
  };
};

export default useTransactionHistory;
