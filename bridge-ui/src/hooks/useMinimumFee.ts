import { useState, useEffect, useCallback } from "react";
import { readContract } from "@wagmi/core";
import { useChainStore } from "@/stores/chainStore";
import { NetworkLayer, wagmiConfig } from "@/config";
import MessageService from "@/abis/MessageService.json";
import { getChainNetworkLayer } from "@/utils/chainsUtil";

const useMinimumFee = () => {
  const [minimumFee, setMinimumFee] = useState(BigInt(0));
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [error, setError] = useState<any | null>(null);

  const { messageServiceAddress, fromChain } = useChainStore((state) => ({
    messageServiceAddress: state.messageServiceAddress,
    fromChain: state.fromChain,
  }));

  const fetchMinimumFee = useCallback(async () => {
    if (!messageServiceAddress) {
      return;
    }

    try {
      let fees = BigInt(0);
      if (fromChain && getChainNetworkLayer(fromChain) === NetworkLayer.L2) {
        fees = (await readContract(wagmiConfig, {
          address: messageServiceAddress,
          abi: MessageService.abi,
          functionName: "minimumFeeInWei",
          chainId: fromChain.id,
        })) as bigint;
      }
      setMinimumFee(fees);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      setError(err);
    }
  }, [messageServiceAddress, fromChain]);

  useEffect(() => {
    fetchMinimumFee();
  }, [fetchMinimumFee]);

  return { minimumFee, error };
};

export default useMinimumFee;
