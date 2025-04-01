import { useEffect } from "react";
import { useReadContract } from "wagmi";
import { ChainLayer } from "@/types";
import { useFormStore, useChainStore } from "@/stores";

const useMinimumFee = () => {
  const { fromChainId, messageServiceAddress, isL2Network } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    messageServiceAddress: state.fromChain.messageServiceAddress,
    isL2Network: state.fromChain.layer === ChainLayer.L2,
  }));
  const setMinimumFees = useFormStore((state) => state.setMinimumFees);

  const { data, isLoading, error, queryKey, refetch } = useReadContract({
    address: messageServiceAddress,
    abi: [
      {
        inputs: [],
        name: "minimumFeeInWei",
        outputs: [
          {
            internalType: "uint256",
            name: "",
            type: "uint256",
          },
        ],
        stateMutability: "view",
        type: "function",
      },
    ],
    functionName: "minimumFeeInWei",
    chainId: fromChainId,
    query: {
      enabled: !!isL2Network,
    },
  });

  useEffect(() => {
    if (!isL2Network) {
      setMinimumFees(0n);
    } else if (data) {
      setMinimumFees(data as bigint);
    }
  }, [data, isL2Network, setMinimumFees]);

  return {
    isLoading,
    minimumFee: (data as bigint | undefined) ?? 0n,
    error,
    queryKey,
    refetchMinimumFee: refetch,
  };
};

export default useMinimumFee;
