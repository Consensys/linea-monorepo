import { useMemo } from "react";
import { useReadContract } from "wagmi";
import { useChainStore } from "@/stores/chainStore";
import MessageService from "@/abis/MessageService.json";
import { ChainLayer } from "@/types";

const useMinimumFee = () => {
  const fromChain = useChainStore.useFromChain();

  const isL2Network = useMemo(() => fromChain.layer === ChainLayer.L2, [fromChain]);

  const { data, isLoading, error, queryKey, refetch } = useReadContract({
    address: fromChain.messageServiceAddress,
    abi: MessageService.abi,
    functionName: "minimumFeeInWei",
    chainId: fromChain?.id,
    query: {
      enabled: !!fromChain.messageServiceAddress && !!fromChain?.id && !!isL2Network,
    },
  });

  return {
    isLoading,
    minimumFee: (data as bigint | undefined) ?? 0n,
    error,
    queryKey,
    refetchMinimumFee: refetch,
  };
};

export default useMinimumFee;
