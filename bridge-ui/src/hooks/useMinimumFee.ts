import { useMemo } from "react";
import { useReadContract } from "wagmi";
import { useChainStore } from "@/stores/chainStore";
import { NetworkLayer } from "@/config";
import MessageService from "@/abis/MessageService.json";
import { getChainNetworkLayer } from "@/utils/chainsUtil";

const useMinimumFee = () => {
  const { messageServiceAddress, fromChain } = useChainStore((state) => ({
    messageServiceAddress: state.messageServiceAddress,
    fromChain: state.fromChain,
  }));

  const isL2Network = useMemo(() => fromChain && getChainNetworkLayer(fromChain) === NetworkLayer.L2, [fromChain]);

  const { data, isLoading, error, queryKey, refetch } = useReadContract({
    address: messageServiceAddress ?? "0x",
    abi: MessageService.abi,
    functionName: "minimumFeeInWei",
    chainId: fromChain?.id,
    query: {
      enabled: !!messageServiceAddress && !!fromChain?.id && !!isL2Network,
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
