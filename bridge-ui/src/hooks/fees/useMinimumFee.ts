import { useEffect, useMemo } from "react";
import { useReadContract } from "wagmi";
import { useChainStore } from "@/stores/chainStore";
import MessageService from "@/abis/MessageService.json";
import { ChainLayer } from "@/types";
import { useFormStore } from "@/stores/formStoreProvider";

const useMinimumFee = () => {
  const fromChain = useChainStore.useFromChain();
  const setMinimumFees = useFormStore((state) => state.setMinimumFees);

  const isL2Network = useMemo(() => fromChain.layer === ChainLayer.L2, [fromChain]);

  const { data, isLoading, error, queryKey, refetch } = useReadContract({
    address: fromChain.messageServiceAddress,
    abi: MessageService.abi,
    functionName: "minimumFeeInWei",
    chainId: fromChain.id,
    query: {
      enabled: !!isL2Network,
    },
  });

  console.log(data);

  useEffect(() => {
    if (!isL2Network) {
      setMinimumFees(0n);
    } else {
      if (data) {
        setMinimumFees(data as bigint);
      }
    }
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
